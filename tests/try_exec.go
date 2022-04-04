package main

import (
	"bytes"
	"flag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "/root/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	err = reloadMycat(kubeClient, cfg)
	klog.Info(err)
}

func execCmd(kubeClientset kubernetes.Interface, cfg *rest.Config, ns string, pod string, cmd *[]string) error {
	req := kubeClientset.CoreV1().RESTClient().Post().Resource("pods").Namespace(ns).Name(pod).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: *cmd,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}
	klog.Info(*cmd)
	req.VersionedParams(option, scheme.ParameterCodec)
	klog.Info(req.URL())
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}
	klog.Info(req.URL())
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	klog.Info(stdout.String())
	return err
}

// reloadMycatConfig when add/remove master/worker
func reloadMycat(client kubernetes.Interface, cfg *rest.Config) error {
	// wait to sync configmap to mounted volume in pod
	mycatPod := "a-mycat-sts-0"
	command := ("/root/config/updateConfig.sh")
	cmd := []string{
		"bash",
		"-c",
		command,
	}
	err := execCmd(client, cfg, "test", mycatPod, &cmd)
	if err != nil {
		return err
	}
	return nil
}
