package main

import (
	"fmt"
	"testing"
	"time"
)

func TestAnalyze(t *testing.T) {
	ch := make(chan int)
	vmpool := NewVmpool(serverAddr, ogkey)
	go vmpool.Charge(ch)
	// go vmpool.Config(ch)

	scaler := NewLoadPredictor(
		address,
		vmpool,
		"d",
		time.Now(),
		scaleInterval,
		adjustInterval,
		ch,
		upper,
		lower,
	)
	// go scaler.Monitor()

	for i := 0; i < 35; i++ {
		time.Sleep(time.Minute)
		load := scaler.Analyze(time.Now())
		fmt.Println(load)
	}
}