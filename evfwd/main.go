package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

//var nc *nats.Conn

// SimpleMessage is a more simple version of v1.Event
type SimpleMessage struct {
	Count     int32  `json:"count"`
	Message   string `json:"message"`
	ObjectRef string `json:"objref"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	FirstSeen string `json:"firstseen"`
	LastSeen  string `json:"lastseen"`
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// Controller allows to track events
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

// NewController creates a Controller
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func makeMessage(evt v1.Event) []byte {
	m, err := json.Marshal(SimpleMessage{
		Count:     evt.Count,
		Message:   evt.Message,
		ObjectRef: fmt.Sprintf("%s/%s", evt.InvolvedObject.Namespace, evt.InvolvedObject.Name),
		Type:      evt.Type,
		Source:    fmt.Sprintf("%s/%s", evt.Source.Component, evt.Source.Host),
		FirstSeen: fmt.Sprintf("%d", evt.FirstTimestamp.Unix()),
		LastSeen:  fmt.Sprintf("%d", evt.LastTimestamp.Unix()),
	})
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}

	message, err := json.Marshal(lib.ClusterEvent{Source: "kubernetes", Message: string(m), Tags: []string{"application", "kubernetes"}})
	if err != nil {
		log.Errorf("%s", err)
		return nil
	}
	return message
}

func main() {
	//var err error
	var kubeconfig *string

	natsURL := flag.String("s", "nats://nats:4222", "NATS server URL ( default: nats://nats:4222 )")
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	log.Info("Starting kwet Event Forwarder...")
	// "nats://nats.svc.k8s:4222"
	nc, err := lib.NatsConnect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}

	//Trying in config first
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Info("Not in cluster... trying kubeconfig")

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	source := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"events",
		api.NamespaceAll,
		fields.Everything())
	_, controller := cache.NewInformer(
		source,
		// The object type.
		&v1.Event{},
		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		time.Second*10,

		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				evt := obj.(*v1.Event)
				nc.Publish("eventpool", makeMessage(*evt))
			},
			UpdateFunc: func(obj interface{}, obj2 interface{}) {
				if obj != obj2 {
					evt := obj.(*v1.Event)
					nc.Publish("eventpool", makeMessage(*evt))
				}
			},
			DeleteFunc: func(obj interface{}) {
				evt := obj.(*v1.Event)
				nc.Publish("eventpool", makeMessage(*evt))
			},
		})
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-signals:
			fmt.Printf("received signal %#v, exiting...\n", s)
			os.Exit(0)
		}
	}
}
