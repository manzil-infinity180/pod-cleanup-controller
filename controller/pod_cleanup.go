package controller

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	coreInformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	coreListers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

func HelloController() {
	fmt.Println("demo")
}

type controller struct {
	clientset      kubernetes.Interface
	podLister      coreListers.PodLister
	podCacheSynced cache.InformerSynced
	//queue          workqueue.TypedRateLimitingInterface[any]
	queue workqueue.RateLimitingInterface // deprecated one
}

func NewController(clientset kubernetes.Interface, podInformer coreInformer.PodInformer) *controller {
	c := &controller{
		clientset:      clientset,
		podLister:      podInformer.Lister(),
		podCacheSynced: podInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod-cleaner"),
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("Hello jiii - AddFunc")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Println("Hello jiii2 - UpdateFunc")
			},
		})
	return c
}

func (c *controller) Run(ch <-chan struct{}) {
	fmt.Sprintf("starting controller")
	if !cache.WaitForCacheSync(ch, c.podCacheSynced) {
		fmt.Println("waiting for cache to be synced")
	}

	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *controller) worker() {
	for c.processItem() {

	}
}

func (c *controller) processItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("key: %s and err, %s\n", key, err.Error())
	}
	//namespace, err := cache.SplitMetaNamespaceKey(key)
	//if err != nil {
	//	fmt.Printf("spliting namespace and name, %s\n", err.Error())
	//}
	return true
}
