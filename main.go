package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
)

type addJSONs struct {
	Metadata struct {
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		SelfLink          string    `json:"selfLink"`
		UID               string    `json:"uid"`
		ResourceVersion   string    `json:"resourceVersion"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Labels            struct {
			Release string `json:"release"`
			Run     string `json:"run"`
		} `json:"labels"`
		Annotations struct {
			KubectlKubernetesIoLastAppliedConfiguration string `json:"kubectl.kubernetes.io/last-applied-configuration"`
		} `json:"annotations"`
	} `json:"metadata"`
	Spec struct {
		Ports []struct {
			Name       string `json:"name"`
			Protocol   string `json:"protocol"`
			Port       int    `json:"port"`
			TargetPort int    `json:"targetPort"`
		} `json:"ports"`
		Selector struct {
			Run string `json:"run"`
		} `json:"selector"`
		ClusterIP       string `json:"clusterIP"`
		Type            string `json:"type"`
		SessionAffinity string `json:"sessionAffinity"`
	} `json:"spec"`
	Status struct {
		LoadBalancer struct {
		} `json:"loadBalancer"`
	} `json:"status"`
}

var (
	kubeconfig = flag.String("kubeconfig", "/home/hugo.carvalho/.kube/config", "absolute path to the kubeconfig file")
)

func main() {
	var addJSON addJSONs

	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "services", v1.NamespaceAll, fields.Everything())
	// watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "services", "testehugo", fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Service{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				objMarshal, _ := json.Marshal(obj)
				_ = json.Unmarshal(objMarshal, &addJSON)

				fmt.Println("ADD -", "Name:", addJSON.Metadata.Name, "NameSpace:", addJSON.Metadata.Namespace, "Type:", addJSON.Spec.Type)
			},
			DeleteFunc: func(obj interface{}) {
				objMarshal, _ := json.Marshal(obj)
				_ = json.Unmarshal(objMarshal, &addJSON)

				fmt.Println("DELETE -", "Name:", addJSON.Metadata.Name, "NameSpace:", addJSON.Metadata.Namespace, "Type:", addJSON.Spec.Type)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				objMarshal, _ := json.Marshal(newObj)
				_ = json.Unmarshal(objMarshal, &addJSON)

				fmt.Println("EDIT -", "Name:", addJSON.Metadata.Name, "NameSpace:", addJSON.Metadata.Namespace, "Type:", addJSON.Spec.Type)
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
