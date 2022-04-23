/*
	This files implements helpful utils to manage components of openGauss.
*/
package v1

import (
	"strconv"

	"k8s.io/klog"
)

const (
	readyStatus   string = "READY"
	unreadyStatus string = "NOT-READY"
)

// IsReady check if opengauss is ready
func (og *OpenGauss) IsReady() bool {
	return og.Status.OpenGaussStatus == readyStatus
}

// IsMasterDeployed check if opengauss's master is deployed
func (og *OpenGauss) IsMasterDeployed() bool {
	return stringToInt32(og.Status.ReadyMaster) == *og.Spec.OpenGauss.Master.Replicas
}

// IsReplicaDeployed check if opengauss's replicas is deployed
func (og *OpenGauss) IsReplicaDeployed() bool {
	ans := stringToInt32(og.Status.ReadyReplicasSmall) == int32(*og.Spec.OpenGauss.WorkerSmall.Replicas)
	ans = ans && stringToInt32(og.Status.ReadyReplicasMid) == int32(*og.Spec.OpenGauss.WorkerMid.Replicas)
	ans = ans && stringToInt32(og.Status.ReadyReplicasLarge) == int32(*og.Spec.OpenGauss.WorkerLarge.Replicas)
	return ans
}

func stringToInt32(str string) int32 {
	ans, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		klog.Error(err)
	}
	return int32(ans)
}