// This file implements some function related to og.Status.xxIPs
package util

import v1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"


func SliceSame(s1 []string, s2[]string) bool{
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func IPsChange(ogOld *v1.OpenGauss, ogNew *v1.OpenGauss) bool {
	ans := MasterIPsChange(ogOld, ogNew) || ReplicaSmallIPsChange(ogOld, ogNew)
	ans = ans || ReplicaMidIPsChange(ogOld, ogNew) || ReplicaLargeIPsChange(ogOld, ogNew)
	return  ans
}

func MasterIPsChange(ogOld *v1.OpenGauss, ogNew *v1.OpenGauss) bool {
	return !SliceSame(ogOld.Status.MasterIPs, ogNew.Status.MasterIPs)
}

func ReplicaSmallIPsChange(ogOld *v1.OpenGauss, ogNew *v1.OpenGauss) bool {
	return !SliceSame(ogOld.Status.ReplicasSmallIPs, ogNew.Status.ReplicasSmallIPs)
}

func ReplicaMidIPsChange(ogOld *v1.OpenGauss, ogNew *v1.OpenGauss) bool {
	return !SliceSame(ogOld.Status.ReplicasMidIPs, ogNew.Status.ReplicasMidIPs)
}

func ReplicaLargeIPsChange(ogOld *v1.OpenGauss, ogNew *v1.OpenGauss) bool {
	return !SliceSame(ogOld.Status.ReplicasLargeIPs, ogNew.Status.ReplicasLargeIPs)
}