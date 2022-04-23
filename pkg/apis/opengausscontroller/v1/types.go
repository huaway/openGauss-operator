package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenGauss is a top-level type
type OpenGauss struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Status *OpenGaussStatus `json:"status,omitempty"`
	Spec   *OpenGaussSpec   `json:"spec,omitempty"`
}

type OpenGaussSpec struct {
	Image            string                         `json:"image"`
	ImagePullPolicy  string                         `json:"imagePullPolicy"`
	OpenGauss        *OpenGaussClusterConfiguration `json:"opengauss"`
	Resources        *corev1.ResourceRequirements   `json:"resources,omitempty"`
	StorageClassName string                         `json:"storageClassName,omitempty"`
}

// Define OpenGauss's needs for master and replicas
type OpenGaussClusterConfiguration struct {
	Master 			*OpenGaussStatefulSet   		`json:"master"` // Master Configuration
	WorkerSmall 	*OpenGaussStatefulSet   		`json:"workerSmall"` // Small replicas Configuration
	WorkerMid		*OpenGaussStatefulSet			`json:"workerMid"` // Mid replicas Configration
	WorkerLarge		*OpenGaussStatefulSet			`json:"workerLarge"` // Large replicas Configration
	Shardingsphere  *ShardingsphereStatefulSet      `json:"shardingsphere"`  // Shardingsphere Configuration
	Origin 			*OriginOpenGaussCluster 		`json:"origin"` // Multi-Master shared info
}

type OriginOpenGaussCluster struct {
	PVC              string `json:"pvc"`
	MycatClusterName string `json:"mycatCluster"`
	Master           string `json:"master"`
}

type OpenGaussStatefulSet struct {
	Replicas  *int32                       `json:"replicas"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ShardingsphereStatefulSet struct {
	Replicas  *int32                       `json:"replicas"`
	Image     string                       `json:"image"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// OpenGauss Cluster's status
type OpenGaussStatus struct {
	OpenGaussStatus           string   `json:"opengaussStatus"`         // OpenGauss if ready or not
	ReadyMaster               string   `json:"readyMaster,omitempty"`   // Ready Master number
	ReadyReplicasSmall    	  string   `json:"readyReplicasSmall"` // Number of Ready Replicas with minimum configration
	ReadyReplicasMid		  string   `json:"readyReplicasMid"` // Number of Ready Replicas with medium configration
	ReadyReplicasLarge		  string   `json:"readyReplicasLarge"` // Number of Ready Replicas with maximum configration
	MasterIPs                 []string `json:"masterIPs,omitempty"`                 // master ips
	ReplicasSmallIPs		  []string `json:"replicasSmallIPs,omitempty"`
	ReplicasMidIPs		      []string `json:"replicasMidIPs,omitempty"`
	ReplicasLargeIPs		  []string `json:"replicasLargeIPs,omitempty"`
	ReadyShardingsphere       string   `json:"readyShardingsphere"`
	PersistentVolumeClaimName string   `json:"persistentVolumeClaimName,omitempty"` // name of pvc
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type OpenGaussList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenGauss `json:"items"`
}
