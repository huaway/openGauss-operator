// 通过修改runtime, scaleInterval, isScale配置实验参数

package main

import (
	// "fmt"
	// "log"
	// "strconv"
	"time"

	"github.com/prometheus/common/model"
	// "github.com/waterme7on/openGauss-operator/util/prometheusUtil"
)

const (
	// Prometheus server address & controller server address
	address        = "http://10.77.50.203:31111"
	serverAddr	   = "localhost:17173"

	// Key of opengauss cluster
	ogkey		   = "test/d"

	// Attributes about scale
	runtime        = time.Minute * 300 // 测试时间
	scaleInterval  = time.Second * 150 // 弹性伸缩间隔
	adjustInterval = time.Second * 5   // 获取CPU利用率间隔
	isScale        = true              // 是否开启弹性伸缩

	// related threshold-based method
	upper		   = 80
	lower   	   = 40
	maxVM		   = 4
)

func main() {
	vmpool := NewVmpool(serverAddr, ogkey)
	// lastScaleTime := time.Now()
	// for {
	// 	time.Sleep(adjustInterval)
		
	// 	// 获得第一个集群的平均CPU利用率，以判断是否伸缩备机
	// 	_, queryClient, err := prometheusUtil.GetPrometheusClient(address)
	// 	if err != nil {
	// 		fmt.Println("Cannot connect to prometheus client " + address)
	// 	}
	// 	result, err := prometheusUtil.QueryWorkerCpuUsagePercentage("d", queryClient)
	// 	if err != nil {
	// 		log.Fatalf("Cannot query prometheus: %s, %s", address, err.Error())
	// 	}
	// 	cpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
	// 	fmt.Println("Worker Cpu Usage", cpuUtil)

	// 	// Scale every once in a while
	// 	if time.Since(lastScaleTime) >= scaleInterval {
	// 		// threshold-based method
	// 		if cpuUtil > upper {
	// 			vmpool.ScaleUp(VmLarge)
	// 		} else if cpuUtil < lower {
	// 			vmpool.ScaleDown(VmLarge)
	// 		}
	// 	}
	// }
	// vmpool.ScaleUp(VmMid)
	// time.Sleep(20*time.Second)
	vmpool.ScaleDown(VmLarge)
}

// Return the value when only one kv pair
func extractValue(v *model.Value) string {
	m := extractResult(v)
	var values []string
	for _, v := range m {
		values = append(values, v)
	}
	return values[0]
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
