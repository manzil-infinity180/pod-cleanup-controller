// Contributed by: Rahul Vishwakarma
// GitHub: https://github.com/manzil-infinity180
// Date: July 2025
package main

import (
	"fmt"

	"github.com/manzil-infinity180/pod-cleanup-controller/client"
	"github.com/manzil-infinity180/pod-cleanup-controller/controller"
	"k8s.io/client-go/informers"
	"os"
	"time"
)

func main() {
	fmt.Println("### pod cleanup controller ###")
	context := os.Getenv("CONTEXT")
	fmt.Println(context)
	clientset, err := client.GetClientSetWithContext(context)
	if err != nil {
		fmt.Println()
		fmt.Errorf("%s", err.Error())
	}
	ch := make(chan struct{})
	factory := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	c := controller.NewController(clientset, factory.Core().V1().Pods())
	factory.Start(ch)
	c.Run(ch)
}
