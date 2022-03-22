package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"
	"database/sql"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	opengaussv1 "github.com/waterme7on/openGauss-operator/pkg/apis/opengausscontroller/v1"
	clientset "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned"
	ogscheme "github.com/waterme7on/openGauss-operator/pkg/generated/clientset/versioned/scheme"
	informers "github.com/waterme7on/openGauss-operator/pkg/generated/informers/externalversions/opengausscontroller/v1"
	listers "github.com/waterme7on/openGauss-operator/pkg/generated/listers/opengausscontroller/v1"
	_ "github.com/lib/pq"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const controllerAgentName = "openGauss-operator"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// Messages
	//
	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by OpenGauss"
	// MessageResourceSynced is the message used for an Event fired when a OpenGauss
	// is synced successfully
	MessageResourceSynced = "OpenGauss synced successfully"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeClientset kubernetes.Interface
	// openGaussClientset is a clientset generated for OpenGauss Objects
	openGaussClientset clientset.Interface
	dynamicClient      dynamic.Interface

	// openGauss controller manage service, configmap and statefulset of OpenGauss object
	// thus needing listers of according resources
	openGaussLister   listers.OpenGaussLister
	openGaussSynced   cache.InformerSynced
	deploymentLister  appslisters.DeploymentLister
	deploymentSynced  cache.InformerSynced
	statefulsetLister appslisters.StatefulSetLister
	statefulsetSynced cache.InformerSynced
	serviceLister     corelisters.ServiceLister
	serviceSynced     cache.InformerSynced
	configMapLister   corelisters.ConfigMapLister
	configMapSynced   cache.InformerSynced

	clusterList map[string]bool // existing cluster list (key format: namespace/name)
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	cfg       *rest.Config
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new OpenGauss controller
func NewController(
	cfg *rest.Config,
	kubeclientset kubernetes.Interface,
	openGaussClientset clientset.Interface,
	dynamicClient dynamic.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	statefulsetInformer appsinformers.StatefulSetInformer,
	serviceInformer coreinformers.ServiceInformer,
	configmapInformer coreinformers.ConfigMapInformer,
	openGaussInformer informers.OpenGaussInformer) *Controller {

	// Create new event broadcaster
	// Add OpenGauss controller types to the default kubernetes scheme
	// so events can be logged for OpenGauss controller types
	utilruntime.Must(ogscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event Broadcaster")
	eventBroadCaster := record.NewBroadcaster()
	eventBroadCaster.StartStructuredLogging(0)
	// starts sending events received from the specified eventBroadcaster to the given sink
	// EventSink knows how to store events.
	eventBroadCaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	// EventRecorder that records events with the given event source.
	recorder := eventBroadCaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		cfg:                cfg,
		kubeClientset:      kubeclientset,
		dynamicClient:      dynamicClient,
		openGaussClientset: openGaussClientset,
		openGaussLister:    openGaussInformer.Lister(),
		openGaussSynced:    openGaussInformer.Informer().HasSynced,
		deploymentLister:   deploymentInformer.Lister(),
		deploymentSynced:   deploymentInformer.Informer().HasSynced,
		statefulsetLister:  statefulsetInformer.Lister(),
		statefulsetSynced:  statefulsetInformer.Informer().HasSynced,
		serviceLister:      serviceInformer.Lister(),
		serviceSynced:      serviceInformer.Informer().HasSynced,
		configMapLister:    configmapInformer.Lister(),
		configMapSynced:    configmapInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "OpenGausses"),
		recorder:           recorder,
		clusterList:        map[string]bool{},
	}

	klog.Infoln("Setting up event handlers")
	// Set up event handler for OpenGauss
	openGaussInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueOpenGauss,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueOpenGauss(new)
		},
		DeleteFunc: controller.cleanConfig,
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObjects,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObjects(new)
		},
		DeleteFunc: controller.handleObjects,
	})

	statefulsetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObjects,
		UpdateFunc: func(old, new interface{}) {
			newSts := new.(*appsv1.StatefulSet)
			oldSts := old.(*appsv1.StatefulSet)
			if newSts.ResourceVersion == oldSts.APIVersion {
				return
			}
			controller.handleObjects(new)
		},
		DeleteFunc: controller.handleObjects,
	})

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObjects,
		UpdateFunc: func(old, new interface{}) {
			newSvc := new.(*corev1.Service)
			oldSvc := old.(*corev1.Service)
			if newSvc.ResourceVersion == oldSvc.APIVersion {
				return
			}
			controller.handleObjects(new)
		},
		DeleteFunc: controller.handleObjects,
	})

	return controller
}

// Run will set up the event handlers for types monitored.
// It will block until stopCh is closed, at which point it will shutdown the workqueue and
// wait for workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// start the informer factories to begin populating the informer caches
	klog.Infoln("Starting openGauss controller")

	// wait for the caches to be synced before starting workers
	klog.Infoln("Syncing informers' caches")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentSynced, c.statefulsetSynced, c.serviceSynced, c.configMapSynced, c.openGaussSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// starting workers
	klog.Infoln("Starting workers")
	// Launch workers to process OpenGauss Resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Infoln("Started workers")
	<-stopCh
	klog.Infoln("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
		time.Sleep(time.Second * 5)
	}
}

// processNextWorkItem will read a single work item from workqueue and attempt to process it by calling syncHandler
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// wrap this block in a func so we can defer c.workqueue.Done
	err := func(obj interface{}) error {
		// call Done here so that workqueue knows that the item have been processed
		defer c.workqueue.Done(obj)
		// We expect strings to come off the workqueue. These are of the form namespace/name.
		// We do this as the delayed nature of the workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the workqueue.
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// run syncHandler, passing the string  "namespace/name" of opengauss to be synced
		// TODO: syncHandler
		// here simply print out the object
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		// if no error occurs, we Forget the items as it has been processed successfully
		c.workqueue.Forget(obj)
		klog.Infoln("Successfully synced", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) UpdateStatusIPs(og *opengaussv1.OpenGauss, masterSts *appsv1.StatefulSet, replicaSts *appsv1.StatefulSet) error {
	if og.Status == nil {
		og.Status = &opengaussv1.OpenGaussStatus{}
	}

	og.Status.MasterIPs = []string{}
	klog.Info("Update opengauss IPs")
	for i := 0; i < int(*og.Spec.OpenGauss.Master.Replicas); i++ {
		masterName := fmt.Sprintf("%v-%d", masterSts.Name, i)
		for {
			master, err := c.kubeClientset.CoreV1().Pods(og.Namespace).Get(context.TODO(), masterName, v1.GetOptions{})
			if err != nil {
				klog.Error("Error when update Opengauss Master IPs")
			}
			if master != nil && master.Status.ContainerStatuses != nil {
				if len(master.Status.PodIP) == 0 {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				klog.Info("master ip: " + master.Status.PodIP)
				og.Status.MasterIPs = append(og.Status.MasterIPs, master.Status.PodIP)
				break
			}
		}
	}	
	og.Status.ReplicasIPs = []string{}
	for i := 0; i < int(*og.Spec.OpenGauss.Worker.Replicas); i++ {
		replicasName := fmt.Sprintf("%v-%d", replicaSts.Name, i)
		for {
			replicas, err := c.kubeClientset.CoreV1().Pods(og.Namespace).Get(context.TODO(), replicasName, v1.GetOptions{})
			if err != nil {
				klog.Error("Error when update Opengauss Replica IPs")
			}
			if replicas != nil && replicas.Status.ContainerStatuses != nil {
				if len(replicas.Status.PodIP) == 0 {
					time.Sleep(time.Millisecond * 500)
					continue
				}
				klog.Info("replica ip: " + replicas.Status.PodIP)
				og.Status.ReplicasIPs = append(og.Status.ReplicasIPs, replicas.Status.PodIP)
				break
			}
		}
	}
	return nil
}

// syncHandler compares the actual state with the desired and attempt to coverge the two.
// It then updates the status of OpenGauss
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the openGauss resource with the namespace and name
	og, err := c.openGaussLister.OpenGausses(namespace).Get(name)

	if err != nil {
		// The openGauss object may not exist.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("openGauss '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	klog.Info("Syncing status of OpenGauss ", og.Name)

	// 1. check if all components are deployed, includes service, configmap, master and worker statefulsets
	// create or get pvc
	var pvc *corev1.PersistentVolumeClaim = nil
	var pvcConfig *corev1.PersistentVolumeClaim = nil
	pvcConfig = NewPersistentVolumeClaim(og)
	pvc, err = c.createOrGetPVC(og.Namespace, pvcConfig)
	if err != nil {
		return err
	}

	if !c.clusterList[key] && og.Spec.OpenGauss.Origin != nil {
		// a new cluster is created
		// 1. root cluster: set true
		// 2. leaf cluster: addNewMaster then set true
		klog.Infof("Add a new cluster: %s", key)
		c.addNewMaster(og)
	}

	// create or update master configmap
	masterConfigMap, masterConfigMapRes := NewMasterConfigMap(og)
	err = c.createOrUpdateDynamicConfigMap(og.Namespace, masterConfigMap, masterConfigMapRes)
	if err != nil {
		return err
	}

	// create or get master statefulset
	var masterStsConfig *appsv1.StatefulSet = nil
	var masterStatefulset *appsv1.StatefulSet = nil
	masterStsConfig = NewMasterStatefulsets(og)
	masterStatefulset, err = c.createOrGetStatefulset(og.Namespace, masterStsConfig)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// create or get master service
	masterSvcConfig := NewMasterService(og)
	_, err = c.createOrGetService(og.Namespace, masterSvcConfig)
	if err != nil {
		return err
	}

	// create or update replica configmap
	replicaConfigMap, relicaConfigMapRes := NewReplicaConfigMap(og)
	err = c.createOrUpdateDynamicConfigMap(og.Namespace, replicaConfigMap, relicaConfigMapRes)
	if err != nil {
		return err
	}

	// create or get replica statefulset
	var replicaStsConfig *appsv1.StatefulSet = nil
	var replicasStatefulset *appsv1.StatefulSet = nil
	replicaStsConfig = NewReplicaStatefulsets(og)
	replicasStatefulset, err = c.createOrGetStatefulset(og.Namespace, replicaStsConfig)
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}
	// create or get master service
	replicaSvcConfig := NewReplicasService(og)
	_, err = c.createOrGetService(og.Namespace, replicaSvcConfig)
	if err != nil {
		return err
	}

	shardingsphereSvcConfig := NewShardingsphereService(og)
	shardingsphereSvc, err  := c.createOrGetService(og.Namespace, shardingsphereSvcConfig)
	if err != nil {
		return err
	}

	// create sharding-sphere configmap at the time cluster is established
	klog.Infof("Update shardingsphere config")
	if og.Status == nil {
		_, err := c.createShardingsphereConfigmap(og, masterStatefulset, replicasStatefulset)
		if err != nil {
			return err
		}
	} else {
		err = c.UpdateShardingsphereReadwriteConfig(og, masterStatefulset, replicasStatefulset, shardingsphereSvc)
		if err != nil {
			klog.Error("Create or update shardingsphere read-write config: error: ", err)
			return err
		}
	
	}
	
	// create or get shardingsphere statefulset
	var shardingsphereStsConfig *appsv1.StatefulSet = nil
	var shardingsphereStatefulset *appsv1.StatefulSet = nil
	shardingsphereStsConfig = NewShardingSphereStatefulset(og)
	shardingsphereStatefulset, err = c.createOrGetStatefulset(og.Namespace, shardingsphereStsConfig)
	if err != nil {
		return err
	}

	// create or get shardingsphere servie


	// 2. check if all components are controlled by opengauss
	// checked if statefulsets are controlled by this og resource
	if !v1.IsControlledBy(masterStatefulset, og) {
		msg := fmt.Sprintf(MessageResourceExists, masterStatefulset.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	if !v1.IsControlledBy(replicasStatefulset, og) {
		msg := fmt.Sprintf(MessageResourceExists, replicasStatefulset.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	if !v1.IsControlledBy(shardingsphereSvc, og) {
		msg := fmt.Sprintf(MessageResourceExists, shardingsphereSvc.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	if !v1.IsControlledBy(shardingsphereStatefulset, og) {
		msg := fmt.Sprintf(MessageResourceExists, shardingsphereStatefulset.Name)
		c.recorder.Event(og, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// 3. check if the status of all components satisfy(don't need to check status of service)
	// checked if replicas number are correct
	if *og.Spec.OpenGauss.Master.Replicas != (*masterStatefulset.Spec.Replicas) ||
		*og.Spec.OpenGauss.Worker.Replicas != (*replicasStatefulset.Spec.Replicas) {
		// update configmap
		masterConfigMap, masterConfigMapRes := NewMasterConfigMap(og)
		err = c.createOrUpdateDynamicConfigMap(og.Namespace, masterConfigMap, masterConfigMapRes)
		if err != nil {
			return err
		}
		replicaConfigMap, replicaConfigMapRes := NewReplicaConfigMap(og)
		err = c.createOrUpdateDynamicConfigMap(og.Namespace, replicaConfigMap, replicaConfigMapRes)
		if err != nil {
			return err
		}
		// update statefulset
		klog.V(4).Infof("OpenGauss '%s' specified master replicas: %d, master statefulset Replicas %d", name, *og.Spec.OpenGauss.Master.Replicas, *masterStatefulset.Spec.Replicas)
		masterStatefulset, err = c.kubeClientset.AppsV1().StatefulSets(og.Namespace).Update(context.TODO(), NewMasterStatefulsets(og), v1.UpdateOptions{})
		if err != nil {
			return err
		}
		klog.V(4).Infof("OpenGauss '%s' specified master replicas: %d, master statefulset Replicas %d", name, *og.Spec.OpenGauss.Worker.Replicas, *replicasStatefulset.Spec.Replicas)
		replicasStatefulset, err = c.kubeClientset.AppsV1().StatefulSets(og.Namespace).Update(context.TODO(), NewReplicaStatefulsets(og), v1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// update shardingsphere Image if needed
	if og.Spec.OpenGauss.Origin == nil && shardingsphereStatefulset != nil && og.Spec.OpenGauss.Shardingsphere.Image != shardingsphereStatefulset.Spec.Template.Spec.Containers[0].Image {
		newTs := int(time.Now().Unix())
		oldTs, err := strconv.Atoi(shardingsphereStatefulset.Spec.Template.Annotations["version/config"])
		if err != nil || newTs-oldTs >= 60 {
			shardingsphereStsConfig.Spec.Template.Annotations = map[string]string{
				"version/config": strconv.Itoa(int(time.Now().Unix())),
			}
			shardingsphereStatefulset, err = c.kubeClientset.AppsV1().StatefulSets(og.Namespace).Update(context.TODO(), shardingsphereStsConfig, v1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}
	// checked if persistent volume claims are correct
	if og.Spec.OpenGauss.Origin == nil && pvc != nil && (*og.Spec.Resources.Requests.Storage() != *pvc.Spec.Resources.Requests.Storage() || og.Spec.StorageClassName != *pvc.Spec.StorageClassName) {
		klog.V(4).Infof("Update OpenGauss pvc storage")
		pvc, err = c.kubeClientset.CoreV1().PersistentVolumeClaims(og.Namespace).Update(context.TODO(), NewPersistentVolumeClaim(og), v1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	// check if shardingsphere statefulset is correct
	if og.Spec.OpenGauss.Origin == nil && shardingsphereStatefulset != nil && *og.Spec.OpenGauss.Shardingsphere.Replicas != *shardingsphereStatefulset.Spec.Replicas {
		klog.V(4).Infof("Openguass %s shardingsphere deployments, expected replicas: %d, actual replicas: %d", og.Name, *og.Spec.OpenGauss.Shardingsphere.Replicas, *shardingsphereStatefulset.Spec.Replicas)
		shardingsphereStatefulset, err = c.kubeClientset.AppsV1().StatefulSets(og.Namespace).Update(context.TODO(), NewShardingSphereStatefulset(og), v1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	// finally update opengauss resource status
	err = c.updateOpenGaussStatus(og, masterStatefulset, replicasStatefulset, shardingsphereStatefulset, shardingsphereSvc, pvc)
	if err != nil {
		return err
	}

	// record normal event
	c.recorder.Event(og, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	if !c.clusterList[key] {
		c.clusterList[key] = true
	}
	return nil
}

// update opengauss status
func (c *Controller) updateOpenGaussStatus(
	og *opengaussv1.OpenGauss,
	masterStatefulset  		  *appsv1.StatefulSet,
	replicasStatefulset       *appsv1.StatefulSet,
	shardingsphereStatefulSet *appsv1.StatefulSet,
	shardingsphereSvc 		  *corev1.Service,
	pvc *corev1.PersistentVolumeClaim) error {
	var err error
	ogCopy := og.DeepCopy()
	if ogCopy.Status == nil {
		ogCopy.Status = &opengaussv1.OpenGaussStatus{}
	}
	ogCopy.Status.MasterStatefulset = masterStatefulset.Name
	ogCopy.Status.ReplicasStatefulset = replicasStatefulset.Name
	ogCopy.Status.ReadyMaster = (strconv.Itoa(int(masterStatefulset.Status.ReadyReplicas)))
	ogCopy.Status.ReadyReplicas = (strconv.Itoa(int(replicasStatefulset.Status.ReadyReplicas)))

	if shardingsphereStatefulSet != nil {
		ogCopy.Status.ReadyShardingsphere = (strconv.Itoa(int(shardingsphereStatefulSet.Status.ReadyReplicas)))
	}
	ogCopy.Status.PersistentVolumeClaimName = pvc.Name
	if (masterStatefulset.Status.ReadyReplicas) == *ogCopy.Spec.OpenGauss.Master.Replicas &&
		(replicasStatefulset.Status.ReadyReplicas) == *ogCopy.Spec.OpenGauss.Worker.Replicas {
		ogCopy.Status.OpenGaussStatus = "READY"
	}

	// }
	// if (!c.clusterList[og.Namespace+"/"+og.Name] || (og.Status != nil && (og.Status.ReadyReplicas != ogCopy.Status.ReadyReplicas || og.Status.ReadyMaster != ogCopy.Status.ReadyMaster))) && replicasStatefulset.Status.ReadyReplicas == *og.Spec.OpenGauss.Worker.Replicas {
	// 	klog.Infof("Update mycat config: %s", og.Name)
	// 	time.Sleep(SyncInterval)
	// 	klog.Infof("Reload mycat: %s", og.Name)
	// 	err = c.restartMycat(og)
	// 	if err != nil {
	// 		klog.Infof("Reload mycat error:%s", err)
	// 	}
	// }
	ogCopy, err = c.openGaussClientset.ControllerV1().OpenGausses(ogCopy.Namespace).UpdateStatus(context.TODO(), ogCopy, v1.UpdateOptions{})
	if err != nil {
		klog.Infoln("Failed to update opengauss status:", ogCopy.Name, " error:", err)
	}
	return err
}

func ExecCmd(db *sql.DB, ctx context.Context, cmd string) error {
	_, err := db.ExecContext(ctx,
		cmd,
	)
	if err != nil {
		klog.Error(fmt.Sprintf("Error when executing cmd:\n %s", cmd))
		return err
	}		
	return nil
}

func (c *Controller) createShardingsphereConfigmap(og *opengaussv1.OpenGauss, masterSts *appsv1.StatefulSet, replicaSts *appsv1.StatefulSet) (*corev1.ConfigMap, error) {
	c.UpdateStatusIPs(og, masterSts, replicaSts)
	sphereConfigMap := NewShardingSphereConfigMap(og)
	var err error
	if og.Spec.OpenGauss.Origin == nil {
		sphereConfigMap, err = c.createOrGetConfigMap(og.Namespace, sphereConfigMap)
		if err != nil {
			return nil, err
		}
	}
	return sphereConfigMap, nil
}

func (c *Controller) UpdateShardingsphereReadwriteConfig(og *opengaussv1.OpenGauss, masterSts *appsv1.StatefulSet, replicaSts *appsv1.StatefulSet, svc *corev1.Service) error {
	i := 0
	
	connStr := fmt.Sprintf("user=root dbname=postgres password=root host=%s port=5432 sslmode=disable", svc.Spec.ClusterIP)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	masterIPs := []string{}
	for i := 0; i < int(*og.Spec.OpenGauss.Master.Replicas); i++ {
		m_replicas_name := fmt.Sprintf("%v-%d", masterSts.Name, i)
		m_replicas, _ := c.kubeClientset.CoreV1().Pods(og.Namespace).Get(context.TODO(), m_replicas_name, v1.GetOptions{})
		if m_replicas != nil && m_replicas.Status.ContainerStatuses != nil {
			masterIPs = append(masterIPs, m_replicas.Status.PodIP)
		}
	}
	replicasIPs := []string{}
	for i := 0; i < int(*og.Spec.OpenGauss.Worker.Replicas); i++ {
		w_replicas_name := fmt.Sprintf("%v-%d", replicaSts.Name, i)
		w_replicas, _ := c.kubeClientset.CoreV1().Pods(og.Namespace).Get(context.TODO(), w_replicas_name, v1.GetOptions{})
		if w_replicas != nil && w_replicas.Status.ContainerStatuses != nil {
			replicasIPs = append(replicasIPs, w_replicas.Status.PodIP)
		}
	}
	klog.Infof("masterIPs %v, replicasIPs %v", masterIPs, replicasIPs)
	klog.Infof("og masterIPs %v, replicasIPs %v", og.Status.MasterIPs, og.Status.ReplicasIPs)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	// update new, rebooted Master's IP
	if len(og.Status.MasterIPs) == 1 && len(masterIPs) == 1 && og.Status.MasterIPs[0] != masterIPs[0] {
		cmd := fmt.Sprintf("alter resource primary_ds (HOST=%s, PORT=5432, DB=postgres, USER=gaussdb, PASSWORD=Enmo@123);", masterIPs[0])
		err := ExecCmd(db, ctx, cmd)
		if err != nil {
			return err
		}
	}

	// update rebooted, deleted, new Replica's IP
	for ; i < len(og.Status.ReplicasIPs) && i < len(replicasIPs); i++ {
		if (og.Status.ReplicasIPs[i] != replicasIPs[i]) {
			cmd := fmt.Sprintf("alter resource primary_ds_%d (HOST=%s, PORT=5432, DB=postgres, USER=gaussdb, PASSWORD=Enmo@123);", i, replicasIPs[i])
			err := ExecCmd(db, ctx, cmd)
			if err != nil {
				return err
			}
		}
	}
	if i < len(og.Status.ReplicasIPs) {
		// Delete deleted resources
		cmd := "ALTER READWRITE_SPLITTING RULE readwrite_ds(WRITE_RESOURCE=primary_ds, READ_RESOURCES("
		for j := 0; j < i; j++ {
			cmd += "replica_ds_" + strconv.Itoa(j) + ","
		}
		cmd = cmd[:len(cmd)-1] + "));"
		klog.V(4).Info("Alter read-write rule before delete resources")
		ExecCmd(db, ctx, cmd)

		cmd = "drop resource "
		for j := i; j < len(og.Status.ReplicasIPs); j++ {
			cmd += "replica_ds_" + strconv.Itoa(j) + ","
		}
		cmd = cmd[:len(cmd)-1] + ";"
		klog.V(4).Info("Delete Resource")
		err = ExecCmd(db, ctx, cmd)
		if err != nil {
			return err
		}
	} else if i < len(replicasIPs) { // Add new Replica's IP
		cmd := "add resource "
		for j := i; j < len(replicasIPs); j++ {
			cmd += "replica_ds_" + strconv.Itoa(j) + fmt.Sprintf("(HOST=%s, PORT=5432, DB=postgres, USER=gaussdb, PASSWORD=Enmo@123),", replicasIPs[j])
		}
		cmd = cmd[:len(cmd)-1] + ";"
		klog.V(4).Info("Add Resource")
		err := ExecCmd(db, ctx, cmd)
		if err != nil {
			return err
		}

		cmd = "ALTER READWRITE_SPLITTING RULE readwrite_ds(WRITE_RESOURCE=primary_ds, READ_RESOURCES("
		for j := 0; j < len(replicasIPs); j++ {
			cmd += "replica_ds_" + strconv.Itoa(j) + ","
		}
		cmd = cmd[:len(cmd)-1] + "));"
		err = ExecCmd(db, ctx, cmd)
		if err != nil {
			return err
		}			
	}
	
	// update read-write splitting rule
	// oldlen := len(og.Status.MasterIPs) + len(og.Status.ReplicasIPs)
	// newlen := len(masterIPs) + len(replicasIPs)
	// if  oldlen != newlen {
	// 	cmd := "READWRITE_SPLITTING RULE readwrite_ds(WRITE_RESOURCE=primary_ds, READ_RESOURCES("
	// 	for i := 0; i < newlen - 1; i++ {
	// 		cmd += "replica_ds_" + strconv.Itoa(i) + ","
	// 	}
	// 	cmd = cmd[:len(cmd)-1] + "));"
	// 	if oldlen == 0 || oldlen > newlen{
	// 		cmd = "ADD " + cmd
	// 	} else {
	// 		cmd = "ALTER " + cmd
	// 	}
	// 	err := ExecCmd(db, ctx, cmd)
	// 	if err != nil {
	// 		return err
	// 	}		
	// }

	og.Status.MasterIPs = masterIPs
	og.Status.ReplicasIPs = replicasIPs
	return nil
}

func (c *Controller) execCmd(ns string, pod string, cmd *[]string) error {
	req := c.kubeClientset.CoreV1().RESTClient().Post().Resource("pods").Namespace(ns).Name(pod).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: *cmd,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}
	req.VersionedParams(option, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(c.cfg, "POST", req.URL())
	if err != nil {
		return err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	klog.V(4).Info("execommand:", stdout.String())
	return err
}

// doCheckpoint
func (c *Controller) doCheckpoint(og *opengaussv1.OpenGauss) error {
	masterSts := og.Spec.OpenGauss.Origin.Master + "-masters"
	for i := 0; i < 1; i++ { // TODO
		masterPod := fmt.Sprintf("%s-%d", masterSts, i)
		command := ("/checkpoint.sh")
		cmd := []string{
			"bash",
			"-c",
			command,
		}
		err := c.execCmd(og.Namespace, masterPod, &cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) cleanConfig(obj interface{}) {
	var object opengaussv1.OpenGauss
	var ok bool
	if object, ok = obj.(opengaussv1.OpenGauss); !ok {
		return
	}
	mycat := object.Name + "-mycat-cm"
	if object.Spec.OpenGauss.Origin != nil {
		mycat = object.Spec.OpenGauss.Origin.Master + "-mycat-cm"
	}
	cm, err := c.kubeClientset.CoreV1().ConfigMaps(object.Namespace).Get(context.TODO(), mycat, v1.GetOptions{})
	if err != nil {
		return
	}
	c.createOrUpdateConfigMap(object.Namespace, cm)
	return
}

func (c *Controller) addNewMaster(og *opengaussv1.OpenGauss) error {
	// 1. remove root cluster route
	klog.Infof("Reload mycat host config:%s", og.Name)

	// 2. check point
	klog.Infof("Origin master do checkpoint:%s", og.Name)
	err := c.doCheckpoint(og)
	if err != nil {
		return err
	}
	// 3. start new cluster and reload config
	// this is done in the normal syncHandler
	return nil
}

// createOrUpdatePVC creates or get pvc of opengauss
func (c *Controller) createOrGetPVC(ns string, config *corev1.PersistentVolumeClaim) (pvc *corev1.PersistentVolumeClaim, err error) {
	// get pvc
	klog.V(4).Infoln("try to get pvc for opengauss:", config.Name)
	pvc, err = c.kubeClientset.CoreV1().PersistentVolumeClaims(ns).Get(context.TODO(), config.Name, v1.GetOptions{})
	if err != nil {
		// (try to) create pvc
		klog.V(4).Infoln("try to create pvc for opengauss:", config.Name)
		pvc, err = c.kubeClientset.CoreV1().PersistentVolumeClaims(ns).Create(context.TODO(), config, v1.CreateOptions{})
	}
	return
}

// createOrGetService creates or gets service of opengauss
func (c *Controller) createOrGetService(ns string, config *corev1.Service) (svc *corev1.Service, err error) {
	// get svc
	klog.V(4).Infoln("try to get svc for opengauss: ", config.Name)
	svc, err = c.kubeClientset.CoreV1().Services(ns).Get(context.TODO(), config.Name, v1.GetOptions{})
	if err != nil {
		// (try to) create Service
		klog.V(4).Infoln("try to create svc for opengauss:", config.Name)
		svc, err = c.kubeClientset.CoreV1().Services(ns).Create(context.TODO(), config, v1.CreateOptions{})
	}
	return
}

// createOrGetStatefulset creates or get statefulset of opengauss
func (c *Controller) createOrGetStatefulset(ns string, config *appsv1.StatefulSet) (sts *appsv1.StatefulSet, err error) {
	// get pvc
	klog.V(4).Infoln("try to get statefulset for opengauss:", config.Name)
	sts, err = c.kubeClientset.AppsV1().StatefulSets(ns).Get(context.TODO(), config.Name, v1.GetOptions{})
	if err != nil {
		// (try to) create pvc
		klog.V(4).Infoln("try to create statefulset for opengauss:", config.Name)
		sts, err = c.kubeClientset.AppsV1().StatefulSets(ns).Create(context.TODO(), config, v1.CreateOptions{})
		if err != nil {
			klog.V(4).Infoln(config.Spec)
		}
	}
	return
}

// createOrGetDeployment creates or get deployment of mycat
func (c *Controller) createOrGetDeployment(ns string, config *appsv1.Deployment) (deployment *appsv1.Deployment, err error) {
	// get deployment
	klog.V(4).Infoln("try to get deployment for opengauss:", config.Name)
	deployment, err = c.deploymentLister.Deployments(ns).Get(config.Name)
	if err != nil {
		klog.V(4).Infoln("try to create deployment for opengauss:", config.Name)
		deployment, err = c.kubeClientset.AppsV1().Deployments(ns).Create(context.TODO(), config, v1.CreateOptions{})
	}
	if err != nil {
		klog.V(4).Infoln(config.Spec)
	}
	return
}

// createOrUpdateDynamicConfigMap creates or update configmap for opengauss
func (c *Controller) createOrUpdateDynamicConfigMap(ns string, cm *unstructured.Unstructured, cmRes schema.GroupVersionResource) error {
	klog.V(4).Infoln("try to create configmap:", cm.GetName())
	_, err := c.dynamicClient.Resource(cmRes).Namespace(ns).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		klog.V(4).Infoln("failed to create, try to update configmap:", cm.GetName())
		_, err = c.dynamicClient.Resource(cmRes).Namespace(ns).Update(context.TODO(), cm, v1.UpdateOptions{})
	}
	if err != nil {
		klog.Infoln("failed to create or update configmap:", cm.GetName())
	}
	return err
}

// createOrUpdateConfigMap creates or update configmap for opengauss
func (c *Controller) createOrUpdateConfigMap(ns string, cm *corev1.ConfigMap) error {
	klog.V(4).Infoln("try to create configmap:", cm.GetName())
	_, err := c.kubeClientset.CoreV1().ConfigMaps(ns).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		klog.V(4).Infoln("failed to create, try to update configmap:", cm.GetName())
		_, err = c.kubeClientset.CoreV1().ConfigMaps(ns).Update(context.TODO(), cm, v1.UpdateOptions{})
	}
	if err != nil {
		klog.Infoln("failed to create or update configmap:", cm.GetName())
	}
	return err
}

// createOrGetConfigMap create or get configmap for og
func (c *Controller) createOrGetConfigMap(ns string, cmConfig *corev1.ConfigMap) (cm *corev1.ConfigMap, err error) {
	klog.V(4).Infoln("try to get configmap:", cmConfig.Name)
	cm, err = c.kubeClientset.CoreV1().ConfigMaps(ns).Get(context.TODO(), cmConfig.Name, v1.GetOptions{})
	if err != nil {
		klog.V(4).Infoln("try to create configmap:", cm.Name)
		cm, err = c.kubeClientset.CoreV1().ConfigMaps(ns).Create(context.TODO(), cmConfig, v1.CreateOptions{})
		if err != nil {
			klog.Infoln("failed to create or get configmap:", cmConfig.Name)
		}
	}
	return
}

// enqueueFoo takes a OpenGauss resource and converts it into a namespace/name
// string which is then put onto the work queue.
// This method should **not** be passed resources of any type other than OpenGauss.
func (c *Controller) enqueueOpenGauss(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// handdleObjects will take any resource implementing metav1.Object and attempt
// to find the opengauss resource that owns it.
// It does this by looking at the objects metadata.ownerReferences field for an appropriate OwnerReference
// It then enqueues that opengauss resource to be processed.
// If the resource does not have a ownerReference, it will be skipped.
func (c *Controller) handleObjects(obj interface{}) {
	var object v1.Object
	var ok bool
	if object, ok = obj.(v1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(v1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := v1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "OpenGauss" {
			return
		}

		og, err := c.openGaussLister.OpenGausses(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of og '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueOpenGauss(og)
		return
	}
}
