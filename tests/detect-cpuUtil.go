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
)

const (
	// Prometheus server address & controller server address
	address        = "http://10.77.50.203:31111"
	serverAddr	   = "localhost:17173"

	adjustSec	   = 30
	adjustInterval = time.Second * adjustSec   // 获取CPU利用率间隔

	// Key of opengauss cluster
	ogkey		   = "test/d"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Not enough parameter")
		return
	}

	cycle, _ := strconv.Atoi(os.Args[1])
	cycle = cycle / adjustSec
	maxCpu := 0.0
	maxCpu2 := 0.0
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
		fmt.Println("Worker Cpu Usage", cpuUtil)
		if cpuUtil > maxCpu {
			maxCpu = cpuUtil
			maxCpu2 = maxCpu
		} else if cpuUtil > maxCpu2 {
			maxCpu2 = cpuUtil
		}
	}
	fmt.Println("max cpu util: ", (maxCpu + maxCpu2) / 2.0)
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
