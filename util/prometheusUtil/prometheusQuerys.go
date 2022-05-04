package prometheusUtil

const (
	PodCpuUsage                  = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod)"
	PodCpuUsagePercentage        = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s.*\"}[1m])) by (pod)/sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests) by (pod) *100"
	PodMemoryUsage               = "sum(container_memory_rss{pod=~\"%s.*\"}) by(pod)" // /1024/1024/1024 = GiB
	PodMemoryUsagePercentage     = "sum(container_memory_rss{pod=~\"%s-.*\"}) by(pod)/sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests) by (pod) *100"
	ClusterCpuUsagePercentage    = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-(masters|replicas)-.*\"}[1m]))/sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests{pod=~\"%s-(masters|replicas)-.*\"})*100"
	MasterCpuUsagePercentage     = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-(masters)-.*\"}[1m]))/sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests{pod=~\"%s-(masters)-.*\"})*100"
	// WorkerCpuUsagePercentage	 = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-(replicas)-.*\", image!=\"\", container!=\"POD\"}[1m]))  / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{pod=~\"%s-(replicas)-.*\"})*100"
	WorkerCpuUsagePercentage 	 = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-(replicas)-.*\", image!=\"\", container!=\"POD\"}[1m])) / (sum(container_spec_cpu_quota{image!=\"\", pod=~\"%s-replicas-.*\", node=\"worker211\"})/100000) * 100"
	ClusterMemoryUsagePercentage = "sum(container_memory_rss{pod=~\"%s-(masters|replicas)-.*\"})/sum(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests{pod=~\"%s-(masters|replicas)-.*\"})*100"
	ClusterNum                   = "sum(kube_pod_container_info{pod=~\"%s-(masters|replicas)-.\"})"
	ReplicaMidCount				 = "sum(kube_statefulset_replicas{statefulset=~\"%s-replicas-mid\"})"
	ReplicaMidTwoCpuUsage		 = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-replicas-mid-(%s|%s).*\", image!=\"\", container!=\"POD\"}[1m]))  / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{pod=~\"%s-replicas-mid-(%s|%s).*\"})*100"
	ReplicaSmallCpuUsage		 = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-replicas-small-.*\", image!=\"\", container!=\"POD\"}[1m]))  / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{pod=~\"%s-replicas-small-.*\"})*100"
	ReplicaMidCpuUsage			 = "sum(rate(container_cpu_usage_seconds_total{pod=~\"%s-replicas-mid-.*\", image!=\"\", container!=\"POD\"}[1m]))  / sum(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits{pod=~\"%s-replicas-mid-.*\"})*100"
)
