package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
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
			Ingress []struct {
				Hostname string `json:"hostname"`
			} `json:"ingress"`
		} `json:"loadBalancer"`
	} `json:"status"`
}

func checkEmptyVariable(name, variable string) {
	if len(strings.TrimSpace(variable)) == 0 {
		fmt.Printf("Please, set %s", name)

		os.Exit(2)
	}
}

var (
	kubeconfig = flag.String("kubeconfig", "/home/hugo.carvalho/.kube/config", "absolute path to the kubeconfig file")
)

func main() {
	var addJSON addJSONs
	var host, name string

	// Envs parameters
	ttl, err := strconv.ParseInt(os.Getenv("ANUS_TTL"), 10, 64)
	if err != nil {
		fmt.Println("Error to convert string to int64 (ANUS_TTL)")
		panic(err.Error())
	}
	domain := os.Getenv("ANUS_DOMAIN")
	zoneID := os.Getenv("ANUS_ZONEID")

	// Test empty confs variables
	checkEmptyVariable("Env: ANUS_TTL", strconv.FormatInt(ttl, 10))
	checkEmptyVariable("Env: ANUS_DOMAIN", domain)
	checkEmptyVariable("Env: ANUS_ZONEID", zoneID)

	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := route53.New(sess)

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

				if addJSON.Spec.Type == "LoadBalancer" {
					host = addJSON.Status.LoadBalancer.Ingress[0].Hostname
					name = addJSON.Metadata.Name + "-" + addJSON.Metadata.Namespace + "." + domain

					fmt.Println("ADD -", "Name:", addJSON.Metadata.Name, "- NameSpace:", addJSON.Metadata.Namespace, "- Type:", addJSON.Spec.Type, "- Host:", host)

					createRecord(svc, name, host, zoneID, ttl)
				}
			},
			DeleteFunc: func(obj interface{}) {
				objMarshal, _ := json.Marshal(obj)
				_ = json.Unmarshal(objMarshal, &addJSON)

				if addJSON.Spec.Type == "LoadBalancer" {
					host = addJSON.Status.LoadBalancer.Ingress[0].Hostname
					name = addJSON.Metadata.Name + "-" + addJSON.Metadata.Namespace + "." + domain

					fmt.Println("DELETE -", "Name:", addJSON.Metadata.Name, "- NameSpace:", addJSON.Metadata.Namespace, "- Type:", addJSON.Spec.Type, "- Host:", host)

					deleteRecord(svc, name, host, zoneID, ttl)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				objMarshal, _ := json.Marshal(newObj)
				_ = json.Unmarshal(objMarshal, &addJSON)

				if addJSON.Spec.Type == "LoadBalancer" {
					host = addJSON.Status.LoadBalancer.Ingress[0].Hostname
					name = addJSON.Metadata.Name + "-" + addJSON.Metadata.Namespace + "." + domain

					fmt.Println("EDIT -", "Name:", addJSON.Metadata.Name, "- NameSpace:", addJSON.Metadata.Namespace, "- Type:", addJSON.Spec.Type, "- Host:", host)

					createRecord(svc, name, host, zoneID, ttl)
				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

func createRecord(svc *route53.Route53, name, target, zoneID string, ttl int64) {
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(name),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(target),
							},
						},
						TTL: aws.Int64(ttl),
					},
				},
			},
			Comment: aws.String("Sample update."),
		},
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ChangeResourceRecordSets(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Change Response:")
	fmt.Println(resp)
}

func deleteRecord(svc *route53.Route53, name, host, zoneID string, ttl int64) {
	request := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(name),
						Type: aws.String("CNAME"),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(host),
							},
						},
						TTL: aws.Int64(ttl),
					},
				},
			},
		},
		HostedZoneId: aws.String(zoneID),
	}
	resp, err := svc.ChangeResourceRecordSets(request)
	if err != nil {
		fmt.Println("Unable to delete DNS Record", err)
	}

	fmt.Println("Delete Response:")
	fmt.Println(resp)
}
