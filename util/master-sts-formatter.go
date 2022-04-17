// This filr gives names and attributes of master statefulset used in openGauss-operator
package util

import (
	"fmt"
	"strconv"

	v1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
)

type MasterFormatter struct {
	OpenGauss *v1.OpenGauss
}

func Master(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &MasterFormatter{OpenGauss: og}
}

func (formatter *MasterFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-masters"
}

func (formatter *MasterFormatter) PodName(id int) string {
	return formatter.StatefulSetName() + "-" + strconv.Itoa(id)
}

func (formatter *MasterFormatter) DataResourceName(id int) string {
	return formatter.OpenGauss.Name + "_masters_" + strconv.Itoa(id)
}

func (formatter *MasterFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-master-service"
}

func (formatter *MasterFormatter) ReplConnInfo() string {
	replica := ReplicaSmall(formatter.OpenGauss)
	replicaStatefulsetName := replica.StatefulSetName()
	replInfo := ""
	for i := 0; i < 1; i++ {
		replInfo += fmt.Sprintf("replconninfo%d='localhost=%s-0 remotehost=%s-%d", i+1, formatter.StatefulSetName(), replicaStatefulsetName, i)
		replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	}
	return replInfo
}

func (formatter *MasterFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-master-config"
}

func (formatter *MasterFormatter) NodeSelector() map[string]string {
	ans := make(map[string]string)
	ans["app"] = "opengauss"
	return ans
}
