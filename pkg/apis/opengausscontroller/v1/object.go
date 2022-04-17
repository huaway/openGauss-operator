/*
	This files implements helpful utils to manage components of openGauss.
*/
package v1

import "strconv"

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
	return og.Status.ReadyMaster == strconv.Itoa(int(*og.Spec.OpenGauss.Master.Replicas))
}

// IsReplicaDeployed check if opengauss's replicas is deployed
func (og *OpenGauss) IsReplicaDeployed() bool {
	ans := og.Status.ReadyReplicasSmall == strconv.Itoa(int(*og.Spec.OpenGauss.WorkerSmall.Replicas))
	ans = ans && og.Status.ReadyReplicasMid == strconv.Itoa(int(*og.Spec.OpenGauss.WorkerMid.Replicas))
	ans = ans && og.Status.ReadyReplicasLarge == strconv.Itoa(int(*og.Spec.OpenGauss.WorkerLarge.Replicas))
	return ans
}
