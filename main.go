package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
)

var (
	kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
)

func main() {
	// for {
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "services", v1.NamespaceAll, fields.Everything())
	watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "services", "testehugo", fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Service{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("service added: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("service deleted: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("service changed \n")
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}

	// // Use shared informers to listen for add/update/delete of services in the specified namespace.
	// // Set resync period to 0, to prevent processing when nothing has changed
	// // informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clientset, 0)
	// informerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(clientset, 0, kubeinformers.WithNamespace("testehugo"))
	// serviceInformer := informerFactory.Core().V1().Services()

	// // Add default resource event handlers to properly initialize informer.
	// serviceInformer.Informer().AddEventHandler(
	// 	cache.ResourceEventHandlerFuncs{
	// 		AddFunc: func(obj interface{}) {
	// 			// fmt.Printf("service added: %s \n", obj)

	// 			fmt.Println(obj)
	// 		},
	// 		DeleteFunc: func(obj interface{}) {
	// 			fmt.Printf("service deleted: %s \n", obj)
	// 		},
	// 		UpdateFunc: func(oldObj, newObj interface{}) {
	// 			fmt.Printf("service changed: %s \n", newObj)
	// 		},
	// 	},
	// )

	// // TODO informer is not explicitly stopped since controller is not passing in its channel.
	// informerFactory.Start(wait.NeverStop)

	// // wait for the local cache to be populated.
	// err = wait.Poll(time.Second, 60*time.Second, func() (bool, error) {
	// 	return serviceInformer.Informer().HasSynced() == true, nil
	// })
	// if err != nil {
	// 	fmt.Errorf("failed to sync cache: %v", err)
	// }

	// // // Transform the slice into a map so it will
	// // // be way much easier and fast to filter later
	// // serviceTypes := make(map[string]struct{})
	// // for _, serviceType := range serviceTypeFilter {
	// // 	serviceTypes[serviceType] = struct{}{}
	// // }

	// stop := make(chan struct{})
	// go controller.Run(stop)
	// for {
	// 	time.Sleep(time.Second)
	// }
	// // }
}
