// package processor
package main

import (
	"fmt"
	// "strconv"
	"time"

	// "github.com/waterme7on/openGauss-operator/util/prometheusUtil"
)

type Load struct {
	thread	int
	tps		int
}

type ActionType int
const (
	ScaleUp 	= 0
	ScaleDown	= 1
	Set			= 2
)

type Decision struct {
	atype ActionType
	vtype VmType
	vcnt  []int
}

type LoadPredictor struct {
	*ScalerBase
}

func NewLoadPredictor(
	paddress string, 
	vmpool *Vmpool, 
	ogname string,
	start time.Time,
	scaleInterval time.Duration,
	probeInterval time.Duration,
	ch <-chan int,
	upper int,
	lower int) *LoadPredictor {
	
	return &LoadPredictor{
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

// Monitor measurements and invoke analyzer at some time
func (predictor *LoadPredictor)Monitor() {
	// lastScaleTime := time.Now()
	for {
		select {
		case <- predictor.ch:
			return
		default:
			// Acquire cpu utilizatin
			// _, queryClient, err := prometheusUtil.GetPrometheusClient(predictor.prometheusAddress)
			// if err != nil {
			// 	fmt.Println("Cannot connect to prometheus client " + predictor.prometheusAddress)
			// }
			// result, err := prometheusUtil.QueryWorkerCpuUsagePercentage(predictor.ogname, queryClient)
			// if err != nil {
			// 	fmt.Printf("Cannot query prometheus: %s, %s\n", address, err.Error())
			// }
			// cpuUtil, _ := strconv.ParseFloat(extractValue(&result), 64)
			// fmt.Println("CPU util ", cpuUtil)

			// If cpuUtil not in expected range
			// if (cpuUtil > float64(predictor.upper) || cpuUtil < float64(predictor.lower)) && time.Since(lastScaleTime) >= predictor.scaleInterval {
			// 	fmt.Println("Indicator is not int expected range cpuUtil ", cpuUtil)
			// 	load := predictor.Analyze(time.Now())
			// 	decisions := predictor.Plan(load)
			// 	if predictor.Execute(decisions) {
			// 		lastScaleTime = time.Now()
			// 	}
			// }
			
			// scale at propriate moment
			var sleepTime []int
			sleepTime = append(sleepTime, 0,7,17,26,35)
			for i := 1; i < len(sleepTime); i++ {
				time.Sleep(time.Duration(sleepTime[i] - sleepTime[i-1]) * time.Minute)
				load := predictor.Analyze(sleepTime[i])
				decisions := predictor.Plan(load)
				predictor.Execute(decisions)
			}
		}
	}
}

func (predictor *LoadPredictor) Analyze(obj interface{}) Load{
	// var tick time.Time
	// switch obj := obj.(type) {
	// case time.Time:
	// 	tick = obj
	// default:
	// 	fmt.Println("Analyze Unknown")
	// }

	// Use previous day's load to predict
	// loads := make(map[int]Load)
	// for i := 0; i < 5; i++ {
	// 	loads[i] = Load{thread: 1, tps: 5}
	// }
	// for i := 5; i < 10; i++ {
	// 	loads[i] = Load{thread: 3, tps: 20}
	// }
	// for i := 10; i < 15; i++ {
	// 	loads[i] = Load{thread: 1, tps: 5}
	// }
	// for i := 15; i < 20; i++ {
	// 	loads[i] = Load{thread: 3, tps: 20}
	// }
	// for i := 20; i < 25; i++ {
	// 	loads[i] = Load{thread: 4, tps: 30}
	// }
	// for i := 25; i < 30; i++ {
	// 	loads[i] = Load{thread: 3, tps: 20}
	// }
	// for i := 30; i < 35; i++ {
	// 	loads[i] = Load{thread: 1, tps: 5}
	// }

	// minute := int(tick.Sub(predictor.start).Minutes())

	var minute int
	switch obj := obj.(type) {
	case int:
		minute = obj
	default:
		fmt.Println("Analyze Unknown")
	}
	loads := make(map[int]Load)
	loads[7] = Load{thread: 2, tps: 25}
	loads[17] = Load{thread: 1, tps: 8}
	loads[26] = Load{thread: 4, tps: 35}
	// loads[28] = Load{thread: 4, tps: 35}
	loads[35] = Load{thread: 1, tps: 6}
	fmt.Println("The predicted load is ", loads[minute])
	return loads[minute]
}

func (predictor *LoadPredictor) Plan(obj interface{}) []Decision {
	var load Load
	switch obj := obj.(type) {
	case Load:
		load = obj
	default:
		fmt.Println("Plan Unknown")
		return nil
	}

	var decisions []Decision
	if load.tps > predictor.vmpool.Capacity().tps {
		fmt.Println("predicted tps ", load.tps)
		fmt.Println("capacity tps ", predictor.vmpool.Capacity())
		gap := load.tps - predictor.vmpool.Capacity().tps
		largeCnt := predictor.vmpool.vmLargeCnt()
		midCnt := predictor.vmpool.vmMidCnt()
		smallCnt := predictor.vmpool.vmSmallCnt()
		if gap >= capacityLarge {
			largeCnt += 2
			gap -= 2*capacityLarge
		}
		for gap >= capacityMid {
			largeCnt++
			gap -= capacityLarge
		}
		for gap >= capacitySmall {
			midCnt++
			gap -= capacityMid
		}
		for gap > 0 {
			smallCnt++
			gap -= capacitySmall
		}
		var cnts []int
		cnts = append(cnts, largeCnt, midCnt, smallCnt)
		fmt.Println("cnt ", cnts)
		decisions = append(decisions, Decision{atype: Set, vtype: VmSmall, vcnt: cnts})
	} else if load.tps < predictor.vmpool.Capacity().tps {
		fmt.Println("predicted tps ", load.tps)
		fmt.Println("capacity tps ", predictor.vmpool.Capacity())
		gap := predictor.vmpool.Capacity().tps - load.tps
		doaction := false
		largeCnt := predictor.vmpool.vmLargeCnt()
		midCnt := predictor.vmpool.vmMidCnt()
		smallCnt := predictor.vmpool.vmSmallCnt()
		for gap - capacityLarge >= 0 && largeCnt > 0 && largeCnt + midCnt + smallCnt > 1 {
			largeCnt--
			gap -= capacityLarge
			doaction = true
		}
		for gap - capacityMid >= 0 && midCnt > 0 && largeCnt + midCnt + smallCnt > 1 {
			midCnt--
			gap -= capacityMid
			doaction = true
		}
		for gap - capacitySmall >= 0 && smallCnt > 0 && largeCnt + midCnt + smallCnt > 1 {
			smallCnt--
			gap -= capacitySmall			
			doaction = true
		}

		// have to add replicas with small config and delete replicas with high config
		var cnts []int
		if !doaction {
			if midCnt > 0 {
				gap = tpsPerCpu * cpuMid - gap
				for gap > 0 {
					smallCnt++
					gap -= tpsPerCpu * cpuSmall			
				}
				cnts = append(cnts, largeCnt, midCnt, smallCnt)
				decisions = append(decisions, Decision{atype: Set, vtype: VmSmall, vcnt: cnts})
				cnts[1] -= 1
				decisions = append(decisions, Decision{atype: Set, vtype: VmSmall, vcnt: cnts})
			} else if largeCnt > 0 {
				gap = tpsPerCpu * cpuLarge - gap
				for gap > 0 && tpsPerCpu * cpuMid <= gap && midCnt < smallCnt {
					midCnt++
					gap -= tpsPerCpu * cpuMid	
				}
				for gap > 0 {
					smallCnt++
					gap -= tpsPerCpu * cpuSmall		
				}	
				cnts = append(cnts, largeCnt, midCnt, smallCnt)			
				cnts[0] -= 1
				decisions = append(decisions, Decision{atype: Set, vtype: VmSmall, vcnt: cnts})
			}
		} else {
			cnts = append(cnts, largeCnt, midCnt, smallCnt)
			decisions = append(decisions, Decision{atype: Set, vtype: VmSmall, vcnt: cnts})
		}
		fmt.Println("cnt ", cnts)
		// cnt := (predictor.vmpool.Capacity().tps - load.tps) / tpsPerCpu
		// decisions = append(decisions, Decision{atype: ScaleDown, vtype: VmSmall, vcnt: cnt})
	}

	return decisions
}

// func (predictor *LoadPredictor) Execute(obj interface{}) bool {
// 	var decisons []Decision
// 	switch obj := obj.(type) {
// 	case []Decision:
// 		decisons = obj
// 	default:
// 		fmt.Println("Execute Unknown")
// 		return false
// 	}

// 	doaction := false
// 	for _, decision := range decisons {
// 		if decision.atype == ScaleUp {
// 			doaction = doaction || predictor.vmpool.ScaleUp(decision.vtype, int32(decision.vcnt[0]))
// 		} else if decision.atype == ScaleDown {
// 			doaction = doaction || predictor.vmpool.ScaleDown(decision.vtype, int32(decision.vcnt[0]))
// 		} else {
// 			doaction = doaction || predictor.vmpool.Set(uint32(decision.vcnt[0]), uint32(decision.vcnt[1]), uint32(decision.vcnt[2]))
// 		}

// 		// interval between two scale action must be larger than 1min
// 		if len(decisons) > 1 {
// 			time.Sleep(1 * time.Minute)
// 		}
// 	}
// 	return true
// }