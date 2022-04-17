package main

import (
	"context"
	"log"
	"time"

	pb "github.com/waterme7on/openGauss-operator/rpc/protobuf"
	"google.golang.org/grpc"
	"k8s.io/klog"
)

type VmType int
const (
	VmSmall VmType = 0
	VmMid	VmType = 1
	VmLarge	VmType = 2
)

const cpuPerSec float64 = 0.000049
const memoryPerSec float64 = 0.00000613

type Vmpool struct {
	cpuCore int
	cost float64
	conn *grpc.ClientConn
	client pb.OpenGaussControllerClient
	scaleRequest *pb.ScaleRequest
	ch chan int
}

func vmCost(vmtype VmType, vmCnt int, sec int) float64 {
	switch vmtype {
	case VmSmall:
		return float64(sec) * float64(vmCnt) * (1*cpuPerSec + 2*memoryPerSec)
	case VmMid:
		return float64(sec) * float64(vmCnt) * (2*cpuPerSec + 4*memoryPerSec)
	case VmLarge:
		return float64(sec) * float64(vmCnt) * (4*cpuPerSec + 8*memoryPerSec)
	default:
		return 0
	}
}

func NewVmpool(serverAddr string, ogkey string) (vmpool *Vmpool){
	vmpool = &Vmpool{cpuCore: 0, cost: 0, ch: make(chan int)}

	// set up conncetion and get information of cluster
	var err error
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	vmpool.conn, err = grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	vmpool.client = pb.NewOpenGaussControllerClient(vmpool.conn)
	request := &pb.GetRequest{
		OpenGaussObjectKey: ogkey,
	}
	response, err := vmpool.client.Get(context.TODO(), request)
	if err != nil {
		klog.Error("Error when get meta about opengauss")
		klog.Error(err.Error())
		return nil
	}
	vmpool.scaleRequest = &pb.ScaleRequest{
		OpenGaussObjectKey: "test/d",
		MasterReplication:  response.MasterReplication,
		WorkerSmallReplication: response.WorkerSmallReplication,
		WorkerMidReplication: response.WorkerMidReplication,
		WorkerLargeReplication: response.WorkerLargeReplication,
	}
	
	// goroutine that calculate cost 
	go func(vmpool *Vmpool) {
		for _ = range vmpool.ch {
			time.Sleep(2 * time.Second)
			vmpool.cost += vmCost(VmSmall, int(vmpool.scaleRequest.WorkerSmallReplication), 2)
			vmpool.cost += vmCost(VmSmall, int(vmpool.scaleRequest.WorkerMidReplication), 2)
			vmpool.cost += vmCost(VmSmall, int(vmpool.scaleRequest.WorkerLargeReplication), 2)
		}
	}(vmpool)

	return
}

func (vmpool *Vmpool) Close() {
	vmpool.conn.Close()
	close(vmpool.ch)
}

func (vm *Vmpool) Price(vtype VmType) float64 {
	return vm.cost
}

func (vmpool *Vmpool) ScaleUp(vtype VmType) {
	switch vtype {
	case VmSmall:
		vmpool.scaleRequest.WorkerSmallReplication++
	case VmMid:
		vmpool.scaleRequest.WorkerMidReplication++
	case VmLarge:
		vmpool.scaleRequest.WorkerLargeReplication++
	}
	response, err := vmpool.client.Scale(context.TODO(), vmpool.scaleRequest)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(response)
}

func (vmpool *Vmpool) ScaleDown(vtype VmType) {
	switch vtype {
	case VmSmall:
		vmpool.scaleRequest.WorkerSmallReplication--
	case VmMid:
		vmpool.scaleRequest.WorkerMidReplication--
	case VmLarge:
		vmpool.scaleRequest.WorkerLargeReplication--
	}
	response, err := vmpool.client.Scale(context.TODO(), vmpool.scaleRequest)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(response)
}