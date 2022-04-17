package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/prometheus/common/model"
	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
)

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
	   return false
	}
	return true
}

func temp() {
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Not enough parameter")
		return
	}
	filepath := "processor/cpu_load_" + os.Args[1] + ".txt"
	var file *os.File
	file, _ = os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	// if checkFileIsExist(filepath) {
	// 	file, _ = os.OpenFile(filepath, os.O_APPEND, 0666)
	// } else {
	// 	file, _ = os.Create(filepath)
	// }
	defer file.Close()

	_, queryClient, err := prometheusUtil.GetPrometheusClient("http://10.77.50.203:31111")
	if err != nil {
		fmt.Println("error opening client")
		return
	}

	iterate := 6
	sum := 0.0
	for i := 0; i < iterate; i++ {
		result, err := prometheusUtil.QueryWorkerCpuUsagePercentage("d", queryClient)
		if err != nil {
			fmt.Printf("Cannot query prometheus: ", err.Error())
		}
		vec, _ := result.(model.Vector)
		fmt.Println("CPU load ", vec[0].Value)
		sum += float64(vec[0].Value)
		time.Sleep(20*time.Second)
	}
	_, err = io.WriteString(file, fmt.Sprintf("%f", sum/float64(iterate)) + " ")
	if err != nil {
		fmt.Println("write to file error: ", err)
	}
}