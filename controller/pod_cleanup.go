// Package controller handles core logic for managing resources.
//
// Contributed by: Rahul Vishwakarma
// GitHub: https://github.com/manzil-infinity180
// Date: July 2025
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	slackFn "github.com/manzil-infinity180/pod-cleanup-controller/utils"
	"github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	coreInformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	coreListers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"os"
	"sync"
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
	queue       workqueue.RateLimitingInterface // deprecated one
	channelID   string
	clientSlack *slack.Client
}

func NewController(clientset kubernetes.Interface, podInformer coreInformer.PodInformer) *controller {
	godotenv.Load(".env")
	token := os.Getenv("SLACK_AUTH_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	clientSlack := slack.New(token, slack.OptionDebug(true))

	c := &controller{
		clientset:      clientset,
		podLister:      podInformer.Lister(),
		podCacheSynced: podInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod-cleaner"),
		clientSlack:    clientSlack,
		channelID:      channelID,
	}

	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAddFunc,
			UpdateFunc: c.handleUpdateFunc,
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

func (c *controller) handleAddFunc(obj interface{}) {
	// add func logic
	podObj, ok := obj.(*corev1.Pod)
	if !ok {
		fmt.Println("\n Not a Pod")
		return
	}
	c.onAddUpdateController(podObj)
	// extract name & namespace
	// look for the pod.Status.Phase - "Running", "Pending", "Succeeded", "Failed", "Unknown"
	// look for "CrashLoopBackOff" containers and RestartCount(>= 5 then delete it)
}
func (c *controller) handleUpdateFunc(obj interface{}, new interface{}) {
	// update logic
	podObj, ok := new.(*corev1.Pod)
	if !ok {
		fmt.Println("\n Not a Pod")
		return
	}
	/**
	state := p.Status.ContainerStatuses[0].State
	    if state.Terminated != nil {
	        exitCode = state.Terminated.ExitCode
	        return true, nil
	    }
	*/

	//fmt.Println(podObj.Name)
	//fmt.Println(podObj.Status.Conditions)
	//fmt.Println(podObj.Status.Conditions[0])
	//fmt.Println(podObj.Status.ContainerStatuses[0].RestartCount)
	//fmt.Println(podObj.Status.ContainerStatuses[0].State)
	b, err := json.Marshal(podObj.Status)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	fmt.Println(string(b))

	//b, err = json.Marshal(podObj.Status.ContainerStatuses)
	//if err != nil {
	//	fmt.Println("Error marshalling JSON:", err)
	//	return
	//}
	//fmt.Println(string(b))
	//
	//fmt.Println("\n \n ######## \n")
	//b, err = json.Marshal(podObj.Status.ContainerStatuses[0])
	//if err != nil {
	//	fmt.Println("Error marshalling JSON:", err)
	//	return
	//}
	//fmt.Println(string(b))

}

func (c *controller) onAddUpdateController(pod *corev1.Pod) {
	if isSeenBefore(pod.UID) {
		return
	}
	// case1: Failed or Evicted
	if pod.Status.Phase == corev1.PodFailed || pod.Status.Reason == "Evicted" {
		makeSeen(pod.UID)
		fmt.Printf("🔥 Failed/Evicted pod detected: %s/%s\n",
			pod.Namespace, pod.Name)
		// sending to slack
		fmt.Printf("Sending message to slack \n")
		attachment := slackFn.BuildSlackAttachment("FailedOrEvicted", pod, 21)
		c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))

		go c.deletePodFunc(pod)
		return
	}

	// case2: CrashLoopBackOff + restart count
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" && cs.RestartCount >= 5 {
			makeSeen(pod.UID)
			fmt.Printf("🔥 CrashLoopBackOff pod detected: %s/%s (restarts: %d)\n",
				pod.Namespace, pod.Name, cs.RestartCount)

			// sending to slack
			fmt.Printf("Sending message to slack \n")
			attachment := slackFn.BuildSlackAttachment("CrashLoopBackOff", pod, 21)
			c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))

			go c.deletePodFunc(pod)
			return
		}
	}
	obj := ExtractPodDetails{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		StartTime: pod.Status.StartTime,
	}
	b, err := json.MarshalIndent(obj, "", "  ")
	if err == nil {
		fmt.Printf("📦 Pod Status (Tracked):\n%s\n", b)
	}
}

var seenPods sync.Map

func isSeenBefore(uid types.UID) bool {
	_, ok := seenPods.Load(uid)
	return ok
}

func makeSeen(uid types.UID) {
	seenPods.Store(uid, struct {
	}{})
}

type ExtractPodDetails struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Phase     string       `json:"phase"`
	StartTime *metav1.Time `json:"startTime"`
}

func (c *controller) deletePodFunc(pod *corev1.Pod) {
	obj := ExtractPodDetails{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Phase:     string(pod.Status.Phase),
		StartTime: pod.Status.StartTime,
		//Conditions: pod.Status.Conditions,
	}
	//ctx := context.Background()
	time.Sleep(20 * time.Second)
	//time.Sleep(5 * time.Minute) // reducing for demo purpose
	err := c.clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err != nil {
		// sending to slack
		fmt.Printf("Sending message to slack \n")
		attachment := slackFn.BuildSlackAttachment("FailedToDelete", pod, 21)
		c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))

		fmt.Printf("Failed to delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
	} else {
		if b, err := json.MarshalIndent(obj, "", "  "); err == nil {
			// sending to slack
			fmt.Printf("Sending message to slack \n")
			attachment := slackFn.BuildSlackAttachment("Deleted", pod, 21)
			c.clientSlack.PostMessage(c.channelID, slack.MsgOptionAttachments(attachment))
			fmt.Printf("✅ Deleted pod:\n%s\n", b)
		}
	}
	seenPods.Delete(pod.UID)
}

/**
// other way to handle
c := &controller{
	clientset:   clientset,
	deleteQueue: make(chan *corev1.Pod, 100), // buffer size as needed
}
---

if pod.Status.Phase == corev1.PodFailed || pod.Status.Reason == "Evicted" {
	if isSeenBefore(pod.UID) {
		return
	}
	makeSeen(pod.UID)

	go func(pod *corev1.Pod) {
		time.Sleep(5 * time.Minute)
		c.deleteQueue <- pod.DeepCopy() // deepcopy to avoid mutation issues
	}(pod)
}
---
func (c *controller) startDeletionWorker(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("🛑 Deletion worker shutting down")
				return
			case pod := <-c.deleteQueue:
				err := c.clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Printf("❌ Failed to delete pod %s/%s: %v\n", pod.Namespace, pod.Name, err)
				} else {
					fmt.Printf("✅ Deleted pod: %s/%s\n", pod.Namespace, pod.Name)
				}
				// Allow reprocessing if needed
				seenPods.Delete(pod.UID)
			}
		}
	}()
}
*/
