package main

import (
	"context"
	"fmt"
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

const (
	cpuPerSec float64 		= 0.000049
	memoryPerSec float64 	= 0.00000613
	maxVM					= 5
	minVM					= 1
)

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
	return
}

func (vmpool *Vmpool) Close() {
	vmpool.conn.Close()
	close(vmpool.ch)
}

func (vmpool *Vmpool) Charge(ch <-chan int) {
	for {
		select {
		case <- ch:
			return
		default:
			time.Sleep(2 * time.Second)
			vmpool.cost += vmCost(VmSmall, int(vmpool.scaleRequest.WorkerSmallReplication), 2)
			vmpool.cost += vmCost(VmSmall, int(vmpool.scaleRequest.WorkerMidReplication), 2)
			vmpool.cost += vmCost(VmSmall, int(vmpool.scaleRequest.WorkerLargeReplication), 2)
		}
	}
}

// Print number of master and replicas to stdout every half minute.
func (vmpool *Vmpool) Config(ch <-chan int) {
	for {
		select {
		case <- ch:
			return
		default:
			fmt.Println("---------------------")
			fmt.Println("replicas-small: ", vmpool.scaleRequest.WorkerSmallReplication)
			fmt.Println("replicas-mid: ", vmpool.scaleRequest.WorkerMidReplication)
			fmt.Println("replicas-large: ", vmpool.scaleRequest.WorkerLargeReplication)
			fmt.Println("---------------------")
			time.Sleep(30 * time.Second)
		}
	}
}

func (vm *Vmpool) Price(vtype VmType) float64 {
	return vm.cost
}

func (vmpool *Vmpool) ScaleUp(vtype VmType, cnt int32) {
	if vmpool.scaleRequest.WorkerSmallReplication + vmpool.scaleRequest.WorkerMidReplication + vmpool.scaleRequest.WorkerLargeReplication + cnt > maxVM {
		return
	}
	switch vtype {
	case VmSmall:
		vmpool.scaleRequest.WorkerSmallReplication += cnt
	case VmMid:
		vmpool.scaleRequest.WorkerMidReplication += cnt
	case VmLarge:
		vmpool.scaleRequest.WorkerLargeReplication += cnt
	}
	fmt.Println("vmpool scale req: ", vmpool.scaleRequest)
	
	response, err := vmpool.client.Scale(context.TODO(), vmpool.scaleRequest)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(response)
}

func (vmpool *Vmpool) ScaleDown(vtype VmType, cnt int32) {
	if vmpool.scaleRequest.WorkerSmallReplication + vmpool.scaleRequest.WorkerMidReplication + vmpool.scaleRequest.WorkerLargeReplication - cnt < 1 {
		return
	}
	switch vtype {
	case VmSmall:
		if vmpool.scaleRequest.WorkerSmallReplication >= cnt {
			vmpool.scaleRequest.WorkerSmallReplication -= cnt
		} else {
			return
		}
	case VmMid:
		if vmpool.scaleRequest.WorkerMidReplication >= cnt {
			vmpool.scaleRequest.WorkerMidReplication -= cnt
		} else {
			return
		}
	case VmLarge:
		if vmpool.scaleRequest.WorkerLargeReplication >= cnt {
			vmpool.scaleRequest.WorkerLargeReplication -= cnt
		} else {
			return
		}
	}
	response, err := vmpool.client.Scale(context.TODO(), vmpool.scaleRequest)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(response)
}