/*
	this file defines formatter for names and attributes used in openGauss-operator
*/
package util

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	v1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
	"k8s.io/klog/v2"
)

func OpenGaussClusterFormatter(og *v1.OpenGauss) *openGaussClusterFormatter {
	return &openGaussClusterFormatter{
		OpenGauss: og,
	}
}

type openGaussClusterFormatter struct {
	OpenGauss *v1.OpenGauss
}

func (formatter *openGaussClusterFormatter) PersistentVolumeCLaimName() string {
	return formatter.OpenGauss.Name + "-pvc"
}

func (formatter *openGaussClusterFormatter) ShardingsphereConfigMapName() string {
	return formatter.OpenGauss.Name + "-shardingsphere-cm"
}

func (formatter *openGaussClusterFormatter) ShardingsphereStatefulsetName() string {
	return formatter.OpenGauss.Name + "-sts"
}

func (formatter *openGaussClusterFormatter) ShardingsphereServiceName() string {
	return formatter.OpenGauss.Name + "-shardingsphere-svc"
}

func (formatter *openGaussClusterFormatter) ShardingSphereServerConfig() string {
	ret, err := ioutil.ReadFile("shardingsphere-proxy/server.yaml")
	if err != nil {
		klog.Error("Unable to open file shardingsphere-proxy/server.yaml")
	}
	return string(ret)		
}

func (formatter *openGaussClusterFormatter) ShardingsphereReadwriteConfig() (ret string) {
	if formatter.OpenGauss.Status != nil {
		if len(formatter.OpenGauss.Status.MasterIPs) > 1 {
			err := ""
			for _, v := range formatter.OpenGauss.Status.MasterIPs {
				err += v
			}
			klog.Error("MasterIPs: " + err)
			panic("Number of opengauss master is bigger than 1")
		}
		content, err := ioutil.ReadFile("shardingsphere-proxy/config-readwrite-splitting.yaml")
		if err != nil {
			klog.Error("Unable to open file shardingsphere-proxy/config-readwrite-splitting.yaml")
			return ""
		}	
		ret = strings.Replace(string(content), "[master-ip]", formatter.OpenGauss.Status.MasterIPs[0], 1)
		// create config for multi replicas
		var left, right int
		if left = strings.Index(ret, "replica_ds_0"); left == -1 {
			klog.Error("No replicas configuration in config-readwrite-splitting.yaml")
			return ""
		}
		if right = strings.Index(ret, "rules"); right == -1 {
			klog.Error("No rule configuration in config-readwrite-splitting.yaml")
			return ""
		}

		p := ""
		for i := 1; i < len(formatter.OpenGauss.Status.ReplicasIPs); i++ {
			p = p + "  " + strings.Replace(ret[left:right], "replica_ds_0", "replica_ds_" + strconv.Itoa(i), 1) 
			ret += ",replica_ds_" + strconv.Itoa(i)
		}		
		ret = ret[:right] + p + ret[right:]

		for _, ip := range(formatter.OpenGauss.Status.ReplicasIPs) {
			ret = strings.Replace(ret, "[replica-ip]", ip, 1)
		}
		// ret = strings.Replace(ret, "[replica-ip]", formatter.OpenGauss.Status.ReplicasServiceIP, 1)
		return
	}
	return ""
}

type StatefulsetFormatterInterface interface {
	StatefulSetName() string
	ServiceName() string
	ReplConnInfo() string
	ConfigMapName() string
	NodeSelector() map[string]string
}

func Master(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &MasterFormatter{OpenGauss: og}
}
func Replica(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &ReplicaFormatter{OpenGauss: og}
}

type MasterFormatter struct {
	OpenGauss *v1.OpenGauss
}

func (formatter *MasterFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-masters"
}

func (formatter *MasterFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-master-service"
}

func (formatter *MasterFormatter) ReplConnInfo() string {
	replica := Replica(formatter.OpenGauss)
	replicaStatefulsetName := replica.StatefulSetName()
	// workerSize := int(math.Max(float64(*formatter.OpenGauss.Spec.OpenGauss.Worker.Replicas), 1))
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

type ReplicaFormatter struct {
	OpenGauss *v1.OpenGauss
}

func (formatter *ReplicaFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-replicas"
}

func (formatter *ReplicaFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-replicas-service"
}

func (formatter *ReplicaFormatter) ReplConnInfo() string {
	master := Master(formatter.OpenGauss)
	masterStatefulsetName := master.StatefulSetName()
	replInfo := ""
	replInfo += fmt.Sprintf("replconninfo1='localhost=127.0.0.1 remotehost=%s-0", masterStatefulsetName)
	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	// workerSize := int(math.Max(float64(*formatter.OpenGauss.Spec.OpenGauss.Worker.Replicas), 1))
	// replInfo := ""
	// for i := 0; i < workerSize; i++ {
	// 	replInfo += fmt.Sprintf("replconninfo%d='localhost=%s-%d remotehost=%s-0", i+1, formatter.StatefulSetName(), i, masterStatefulsetName)
	// 	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	// }
	return replInfo
}

func (formatter *ReplicaFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-replicas-config"
}

func (formatter *ReplicaFormatter) NodeSelector() map[string]string {
	ans := make(map[string]string)
	ans["app"] = "opengauss"
	return ans	
}