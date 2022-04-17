// This file gives names and attributes of replica statefulset used in openGauss-operator
package util

import (
	"fmt"
	"strconv"
	// "strconv"

	v1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
)

// Replicas with minimum configuration
type ReplicaSmallFormatter struct {
	OpenGauss *v1.OpenGauss
}

func ReplicaSmall(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &ReplicaSmallFormatter{OpenGauss: og}
}

func (formatter *ReplicaSmallFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-replicas-small"
}

func (formatter *ReplicaSmallFormatter) PodName(id int) string {
	return formatter.StatefulSetName() + "-" + strconv.Itoa(id)
}

func (formatter *ReplicaSmallFormatter) DataResourceName(id int) string {
	return formatter.OpenGauss.Name + "_replicas_small_"  + strconv.Itoa(id)
}

func (formatter *ReplicaSmallFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-replicas-small-service"
}

func (formatter *ReplicaSmallFormatter) ReplConnInfo() string {
	master := Master(formatter.OpenGauss)
	masterStatefulsetName := master.StatefulSetName()
	replInfo := ""
	replInfo += fmt.Sprintf("replconninfo1='localhost=127.0.0.1 remotehost=%s-0", masterStatefulsetName)
	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	return replInfo
}

func (formatter *ReplicaSmallFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-replicas-small-config"
}

func (formatter *ReplicaSmallFormatter) NodeSelector() map[string]string {
	ans := make(map[string]string)
	ans["app"] = "opengauss"
	return ans	
}

// Replica with medium configuration
type ReplicaMidFormatter struct {
	OpenGauss *v1.OpenGauss
}

func ReplicaMid(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &ReplicaMidFormatter{OpenGauss: og}
}

func (formatter *ReplicaMidFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-replicas-mid"
}

func (formatter *ReplicaMidFormatter) PodName(id int) string {
	return formatter.StatefulSetName() + "-" + strconv.Itoa(id)
}

func (formatter *ReplicaMidFormatter) DataResourceName(id int) string {
	return formatter.OpenGauss.Name + "_replicas_mid_"  + strconv.Itoa(id)
}

func (formatter *ReplicaMidFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-replicas-mid-service"
}

func (formatter *ReplicaMidFormatter) ReplConnInfo() string {
	master := Master(formatter.OpenGauss)
	masterStatefulsetName := master.StatefulSetName()
	replInfo := ""
	replInfo += fmt.Sprintf("replconninfo1='localhost=127.0.0.1 remotehost=%s-0", masterStatefulsetName)
	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	return replInfo
}

func (formatter *ReplicaMidFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-replicas-mid-config"
}

func (formatter *ReplicaMidFormatter) NodeSelector() map[string]string {
	ans := make(map[string]string)
	ans["app"] = "opengauss"
	return ans	
}

// Replica with maximum configuration
type ReplicaLargeFormatter struct {
	OpenGauss *v1.OpenGauss
}

func ReplicaLarge(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &ReplicaLargeFormatter{OpenGauss: og}
}

func (formatter *ReplicaLargeFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-replicas-large"
}

func (formatter *ReplicaLargeFormatter) PodName(id int) string {
	return formatter.StatefulSetName() + "-" + strconv.Itoa(id)
}

func (formatter *ReplicaLargeFormatter) DataResourceName(id int) string {
	return formatter.OpenGauss.Name + "_replicas_large_"  + strconv.Itoa(id)
}

func (formatter *ReplicaLargeFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-replicas-large-service"
}

func (formatter *ReplicaLargeFormatter) ReplConnInfo() string {
	master := Master(formatter.OpenGauss)
	masterStatefulsetName := master.StatefulSetName()
	replInfo := ""
	replInfo += fmt.Sprintf("replconninfo1='localhost=127.0.0.1 remotehost=%s-0", masterStatefulsetName)
	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	return replInfo
}

func (formatter *ReplicaLargeFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-replicas-large-config"
}

func (formatter *ReplicaLargeFormatter) NodeSelector() map[string]string {
	ans := make(map[string]string)
	ans["app"] = "opengauss"
	return ans	
}
