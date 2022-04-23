package prometheusUtil

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	api "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// GetPrometheusClient returns prometheus apiConfig and apiClient
func GetPrometheusClient(address string) (*api.Config, *prometheus.API, error) {
	// 连接到Prometheus Client
	config := api.Config{
		Address: address,
	}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, nil, errors.New("connect to prometheus error")
	}
	// 执行query
	queryClient := prometheus.NewAPI(client)
	return &config, &queryClient, nil
}

// Return the value when only one kv pair
func ExtractValue(v *model.Value) string {
	m := ExtractResult(v)
	var values []string
	for _, v := range m {
		values = append(values, v)
	}
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// 返回查询结果
// 输出如：map[{pod="prometheus-6d75d99cb9-lx8w2"}:4.93641914680743 {pod="prometheus-adapter-5b8db7955f-6zs2j"}:0 {pod="prometheus-adapter-5b8db7955f-ktp2k"}:3.571457910076159 {pod="prometheus-k8s-0"}:311.1957729587634 {pod="prometheus-operator-75d9b475d9-955fv"}:0.6592752119650527]
// key: {pod="prometheus-6d75d99cb9-lx8w2"}
// value: 4.93641914680743
// 均为string
func ExtractResult(v *model.Value) (m map[string]string) {
	switch (*v).(type) {
	case model.Vector:
		vec, _ := (*v).(model.Vector)
		m = VectorToMap(&vec)
	default:
		break
	}
	return
}

func VectorToMap(v *model.Vector) (m map[string]string) {
	m = make(map[string]string)
	for i := range *v {
		m[(*v)[i].Metric.String()] = (*v)[i].Value.String()
	}
	return
}

func QueryPodCpuUsage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodCpuUsage, podPrefix), time.Now())
	return value, err
}

func QueryPodCpuUsagePercentage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodCpuUsagePercentage, podPrefix), time.Now())
	return value, err
}

func QueryPodMemoryUsage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodMemoryUsage, podPrefix), time.Now())
	return value, err
}

func QueryPodMemoryUsagePercentage(podPrefix string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(PodMemoryUsagePercentage, podPrefix), time.Now())
	return value, err
}

func QueryClusterNumber(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ClusterNum, cluster), time.Now())
	return value, err
}

func QueryMasterCpuUsagePercentage(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(MasterCpuUsagePercentage, cluster, cluster), time.Now())
	return value, err
}
func QueryWorkerCpuUsagePercentage(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(WorkerCpuUsagePercentage, cluster, cluster), time.Now())
	return value, err
}
func QueryClusterCpuUsagePercentage(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ClusterCpuUsagePercentage, cluster, cluster), time.Now())
	return value, err
}
func QueryClusterMemoryUsagePercentage(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ClusterMemoryUsagePercentage, cluster, cluster), time.Now())
	return value, err
}
func QueryReplicaMidCount(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ReplicaMidCount, cluster), time.Now())
	return value, err
}
func QueryScaledTwoReplicasMidCpuUsage(cluster string, cnt int, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ReplicaMidTwoCpuUsage, cluster, strconv.Itoa(cnt), 
														strconv.Itoa(cnt+1), cluster, strconv.Itoa(cnt), strconv.Itoa(cnt+1)), time.Now())
	return value, err
}
func QueryReplicaSmallCpuUsage(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ReplicaSmallCpuUsage, cluster, cluster), time.Now())
	return value, err
}
func QueryReplicaMidCpuUsage(cluster string, client *prometheus.API) (model.Value, error) {
	value, _, err := (*client).Query(context.TODO(), fmt.Sprintf(ReplicaMidCpuUsage, cluster, cluster), time.Now())
	return value, err
}