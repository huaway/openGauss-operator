// 通过修改runtime, scaleInterval, isScale配置实验参数

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
)

const (
	// Prometheus server address & controller server address
	address        = "http://10.77.50.203:31111"
	serverAddr	   = "localhost:17173"

	// Key of opengauss cluster
	ogkey		   = "test/d"

	// Attributes about scale
	runtime        = time.Minute * 300 // 测试时间
	scaleInterval  = time.Second * 60 // 弹性伸缩间隔
	adjustSec	   = 5
	adjustInterval = time.Second * adjustSec   // 获取CPU利用率间隔
	neuralInterval = time.Second * 60 // Seconds of tentavie action
	isScale        = true              // 是否开启弹性伸缩

	// related threshold-based method
	upper		   = 80
	lower   	   = 40

	// measures about cpu state
	cpuIdle		   = 25
	cpuOneFull	   = 40
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Not enough parameter")
		return
	}

	ch := make(chan int)
	vmpool := NewVmpool(serverAddr, ogkey)
	go vmpool.Charge(ch)
	go vmpool.Config(ch)

	cycle, _ := strconv.Atoi(os.Args[1])
	cycle = cycle / adjustSec
	fmt.Println("cycle ", cycle)
	lastScaleTime := time.Now()
	for i := 0; i < cycle; i++ {
		time.Sleep(adjustInterval)
		
		// 获得第一个集群的平均CPU利用率，以判断是否伸缩备机
		_, queryClient, err := prometheusUtil.GetPrometheusClient(address)
		if err != nil {
			fmt.Println("Cannot connect to prometheus client " + address)
		}
		result, err := prometheusUtil.QueryWorkerCpuUsagePercentage("d", queryClient)
		if err != nil {
			log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
		}
		cpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
		// fmt.Println("Worker Cpu Usage", cpuUtil)

		// Scale every once in a while
		vmpool.threshold(&lastScaleTime, cpuUtil)
		// vmpool.tentative(queryClient, &lastScaleTime, cpuUtil)
	}

	ch <- 3
	close(ch)
	fmt.Println(vmpool.cost)

	// vmpool := NewVmpool(serverAddr, ogkey)
	// vmpool.ScaleUp(VmLarge, 1)
	// time.Sleep(30 * time.Second)
	// vmpool.ScaleDown(VmMid, 1)
	vmpool.Close()
}

func (vmpool *Vmpool) threshold(lastScaleTime *time.Time, cpuUtil float64) {
	if time.Since(*lastScaleTime) >= scaleInterval {
		// threshold-based method
		if cpuUtil > upper {
			*lastScaleTime = time.Now()
			vmpool.ScaleUp(VmLarge, 1)
		} else if cpuUtil < lower {
			*lastScaleTime = time.Now()
			vmpool.ScaleDown(VmLarge, 1)
		}
	}
}

func (vmpool *Vmpool) tentative(client *prometheus.API, lastScaleTime *time.Time, cpuUtil float64) {
	// still during cool period
	if time.Since(*lastScaleTime) < scaleInterval {
		return
	}
	
	if cpuUtil > upper {
		*lastScaleTime = time.Now()
		// acquire the count of replicas-mid
		result, err := prometheusUtil.QueryReplicaMidCount("d", client)
		if err != nil {
			log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
		}
		midCnt, err := strconv.ParseInt(extractValue(&result), 10, 0)
		if err != nil {
			fmt.Println("When convert int error: ", err)
		}
		// add two replicas-mid and query effect
		vmpool.ScaleUp(VmMid, 2)
		time.Sleep(neuralInterval)
		// examine if cpuUtil of two scaled replicas lower than lower
		// if so, scale down one replica
		result, err = prometheusUtil.QueryScaledTwoReplicasMidCpuUsage("d", int(midCnt), client)
		if err != nil {
			log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
		}
		scaleCpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
		if scaleCpuUtil < lower {
			vmpool.ScaleDown(VmMid, 1)
		}
	} else if cpuUtil < lower {
		*lastScaleTime = time.Now()
		// acquire the avg cpuUtil of replica-samll
		result, err := prometheusUtil.QueryReplicaSmallCpuUsage("d", client)
		if err != nil {
			log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
		}
		smallCpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
		// acquire the avg cpuUtil of replica-mid
		result, err = prometheusUtil.QueryReplicaMidCpuUsage("d", client)
		if err != nil {
			log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
		}
		midCpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
		// examine if avg replica-mid's cpuUtil < 25% and avg replica-small's cpuUtil < 40%
		// if so, scale down one relica-mid
		// else scale down one replica-small
		if smallCpuUtil != 0 && midCpuUtil != 0 {
			if midCpuUtil <= cpuIdle && smallCpuUtil < cpuOneFull {
				vmpool.ScaleDown(VmMid, 1)
			}
		} else if smallCpuUtil != 0 {
			vmpool.ScaleDown(VmSmall, 1)
		} else if midCpuUtil != 0{
			vmpool.ScaleDown(VmMid, 1)
		}
	}
}

// Return the value when only one kv pair
func extractValue(v *model.Value) string {
	m := extractResult(v)
	var values []string
	for _, v := range m {
		values = append(values, v)
	}
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// 返回查询结果
// 输出如：map[{pod="prometheus-6d75d99cb9-lx8w2"}:4.93641914680743 {pod="prometheus-adapter-5b8db7955f-6zs2j"}:0 {pod="prometheus-adapter-5b8db7955f-ktp2k"}:3.571457910076159 {pod="prometheus-k8s-0"}:311.1957729587634 {pod="prometheus-operator-75d9b475d9-955fv"}:0.6592752119650527]
// key: {pod="prometheus-6d75d99cb9-lx8w2"}
// value: 4.93641914680743
// 均为string
func extractResult(v *model.Value) (m map[string]string) {
	switch (*v).(type) {
	case model.Vector:
		vec, _ := (*v).(model.Vector)
		m = vectorToMap(&vec)
	default:
		break
	}
	return
}

func vectorToMap(v *model.Vector) (m map[string]string) {
	m = make(map[string]string)
	for i := range *v {
		m[(*v)[i].Metric.String()] = (*v)[i].Value.String()
	}
	return
}
