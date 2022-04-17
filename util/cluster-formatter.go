// This file defines formatter for names and attributes of cluster used in openGauss-operator
package util

import (
	"io/ioutil"
	"strings"

	v1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
	"k8s.io/klog/v2"
)

type openGaussClusterFormatter struct {
	OpenGauss *v1.OpenGauss
}

type StatefulsetFormatterInterface interface {
	StatefulSetName() string
	PodName(int) string
	DataResourceName(int) string
	ServiceName() string
	ReplConnInfo() string
	ConfigMapName() string
	NodeSelector() map[string]string
}

func OpenGaussClusterFormatter(og *v1.OpenGauss) *openGaussClusterFormatter {
	return &openGaussClusterFormatter{
		OpenGauss: og,
	}
}

func (formatter *openGaussClusterFormatter) MasterStsName() string {
	return formatter.OpenGauss.Name + "-masters"
}

func (formatter *openGaussClusterFormatter) ReplicasSmallStsName() string {
	return formatter.OpenGauss.Name + "-replicas-small"
}

func (formatter *openGaussClusterFormatter) ReplicasMidStsName() string {
	return formatter.OpenGauss.Name + "-replicas-mid"
}

func (formatter *openGaussClusterFormatter) ReplicasLargeStsName() string {
	return formatter.OpenGauss.Name + "-replicas-large"
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

// When cluster is established, create readwrite config to boost shardingspgere
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
		// replace write data source name and ip
		ret = strings.Replace(string(content), "primary_ds", Master(formatter.OpenGauss).DataResourceName(0), -1)
		ret = strings.Replace(ret, "[master-ip]", formatter.OpenGauss.Status.MasterIPs[0], 1)

		// create config for multi replicas
		// Find [left,right) corresponding to replica_ds_0:
		var left, right int
		if left = strings.Index(ret, "replica_ds_0"); left == -1 {
			klog.Error("No replicas configuration in config-readwrite-splitting.yaml")
			return ""
		}
		if right = strings.Index(ret, "rules"); right == -1 {
			klog.Error("No rule configuration in config-readwrite-splitting.yaml")
			return ""
		}

		readSources := ""
		readNames := ""
		for i := 0; i < len(formatter.OpenGauss.Status.ReplicasSmallIPs); i++ {
			readSources = readSources + "  " + strings.Replace(ret[left:right], "replica_ds_0", ReplicaSmall(formatter.OpenGauss).DataResourceName(i), 1)
			readSources = strings.Replace(readSources, "[replica-ip]", formatter.OpenGauss.Status.ReplicasSmallIPs[i], 1)
			readNames += ReplicaSmall(formatter.OpenGauss).DataResourceName(i) +  ","
		}
		for i := 0; i < len(formatter.OpenGauss.Status.ReplicasMidIPs); i++ {
			readSources = readSources + "  " + strings.Replace(ret[left:right], "replica_ds_0", ReplicaMid(formatter.OpenGauss).DataResourceName(i), 1)
			readSources = strings.Replace(readSources, "[replica-ip]", formatter.OpenGauss.Status.ReplicasMidIPs[i], 1)
			readNames += ReplicaMid(formatter.OpenGauss).DataResourceName(i) + ","
		}
		for i := 0; i < len(formatter.OpenGauss.Status.ReplicasLargeIPs); i++ {
			readSources = readSources + "  " + strings.Replace(ret[left:right], "replica_ds_0", ReplicaLarge(formatter.OpenGauss).DataResourceName(i), 1)
			readSources = strings.Replace(readSources, "[replica-ip]", formatter.OpenGauss.Status.ReplicasLargeIPs[i], 1)
			readNames += ReplicaLarge(formatter.OpenGauss).DataResourceName(i) + ","
		}
		readNames = readNames[:len(readNames)-1]
		ret = ret[:left-2] + readSources + ret[right:]

		// replace props "read-data-source-names:" to correct value
		if left = strings.Index(ret, "replica_ds_0"); left == -1 {
			klog.Error("No \"read-data-source-names: replica_ds_0\" in config-readwrite-splitting.yaml")
			return ""			
		}
		ret = ret[:left] + readNames

		return
	}
	return ""
}
