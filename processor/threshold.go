package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/waterme7on/openGauss-operator/util/prometheusUtil"
)

type ThresholdScaler struct {
	*ScalerBase	
}

func NewThresholdScaler(
	paddress string, 
	vmpool *Vmpool, 
	ogname string,
	start time.Time,
	scaleInterval time.Duration,
	probeInterval time.Duration,
	ch <-chan int,
	upper int,
	lower int) *ThresholdScaler {
	
	return &ThresholdScaler{
		ScalerBase: newScalerBase(
			paddress,
			vmpool,
			ogname,
			start,
			scaleInterval,
			probeInterval,
			ch,
			upper,
			lower,
		),
	}
}

func (scaler *ThresholdScaler) Monitor() {
	lastScaleTime := time.Now()
	violate := 0
	for {
		select {
		case <- scaler.ch:
			return
		default:
			// Acquire cpu utilizatin
			_, queryClient, err := prometheusUtil.GetPrometheusClient(scaler.prometheusAddress)
			if err != nil {
				fmt.Println("Cannot connect to prometheus client " + scaler.prometheusAddress)
			}
			result, err := prometheusUtil.QueryWorkerCpuUsagePercentage(scaler.ogname, queryClient)
			if err != nil {
				fmt.Printf("Cannot query prometheus: %s, %s\n", address, err.Error())
			}
			cpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
			// fmt.Println("CPU util ", cpuUtil)

			// If cpuUtil not in expected range
			if (cpuUtil > float64(scaler.upper) || cpuUtil < float64(scaler.lower)) && time.Since(lastScaleTime) >= scaler.scaleInterval {
				fmt.Println("Indicator is not int expected range, cpuUtil is ", cpuUtil)
				violate++
				if violate >= 1 {
					violate = 0
					load := scaler.Analyze(cpuUtil)
					decisions := scaler.Plan(load)
					if scaler.Execute(decisions) {
						lastScaleTime = time.Now()
					}
				}
			}

			time.Sleep(scaler.probeInterval)
		}
	}
}

func (scaler *ThresholdScaler) Analyze(obj interface{}) Load {
	fmt.Println("This is from threshold analyze")
	var cpuUtil float64
	switch obj := obj.(type) {
	case float64:
		cpuUtil = obj
	default:
		fmt.Println("Threshold scaler analyze unknown")
		return Load{}
	}

	if cpuUtil > float64(scaler.upper) {
		return Load{thread: math.MaxInt, tps: math.MaxInt}
	} else if cpuUtil < float64(scaler.lower) {
		return Load{thread: math.MinInt, tps: math.MinInt}
	}

	return Load{}
}

func (scaler *ThresholdScaler) Plan(obj interface{}) []Decision{
	fmt.Println("This is from threshold plan")
	var load Load
	switch obj := obj.(type) {
	case Load:
		load = obj
	default:
		fmt.Println("Plan Unknown")
		return nil
	}

	var decisions []Decision
	var cnts []int
	cnts = append(cnts, 1)
	if load.tps > scaler.vmpool.Capacity().tps {
		fmt.Println("send scaleup to server")
		decisions = append(decisions, Decision{atype: ScaleUp, vtype: VmSmall, vcnt: cnts})
	} else if load.tps < scaler.vmpool.Capacity().tps {
		fmt.Println("send scaledown to server")
		decisions = append(decisions, Decision{atype: ScaleDown, vtype: VmSmall, vcnt: cnts})
	}
	
	return decisions
}
