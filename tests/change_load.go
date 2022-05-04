package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Load struct {
	thread	int
	tps		int
}

func main() {
	rand.Seed(time.Now().Unix())

	loads := make(map[int]Load)
	// stable low workload
	i := 0
	for ; i < 5; i++ {
		loads[i] = Load{thread: 1, tps: 5}
	}
	// increasing slowly
	loads[i] = Load{thread: 1, tps: 7}
	i++
	loads[i] = Load{thread: 1, tps: 9}
	i++
	loads[i] = Load{thread: 2, tps: 11}
	i++
	loads[i] = Load{thread: 2, tps: 13}
	i++
	loads[i] = Load{thread: 2, tps: 15}
	i++
	// stable mid workload
	for j := 0; j < 5; j++ {
		loads[i] = Load{thread: 2, tps: 20}
		i++
	}
	// decreasing slowly
	loads[i] = Load{thread: 2, tps: 15}
	i++
	loads[i] = Load{thread: 2, tps: 13}
	i++
	loads[i] = Load{thread: 2, tps: 11}
	i++
	loads[i] = Load{thread: 1, tps: 9}
	i++
	loads[i] = Load{thread: 1, tps: 7}
	i++
	// stable low workload
	for j := 0; j < 5; j++ {
		loads[i] = Load{thread: 1, tps: 5}
		i++
	}
	// increasing rapidly
	loads[i] = Load{thread: 1, tps: 10}
	i++
	loads[i] = Load{thread: 2, tps: 20}
	i++
	loads[i] = Load{thread: 3, tps: 20}
	i++
	// stable high workload
	loads[i] = Load{thread: 4, tps: 32}
	i++
	loads[i] = Load{thread: 4, tps: 32}
	i++
	loads[i] = Load{thread: 4, tps: 32}
	i++
	// descreasing rapidly
	loads[i] = Load{thread: 3, tps: 20}
	i++
	loads[i] = Load{thread: 2, tps: 20}
	i++
	loads[i] = Load{thread: 1, tps: 10}
	i++
	// stable low workload
	for j := 0; j < 5; j++ {
		loads[i] = Load{thread: 1, tps: 5}
		i++
	}


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

	for j := 0; j < i; j++ {
		var wave float64
		if rand.Intn(100) < 90 {
			wave = rand.Float64() * 0.15 - 0.05
		} else {
			wave = rand.Float64()
		}
	
		fmt.Printf("\"%d ", loads[j].thread)
		fmt.Print(int((1 + wave) * float64(loads[j].tps)))
		fmt.Print("\" ")
	}
	// 39minutes
	fmt.Print("\n minutes: ", i, "\n")
}