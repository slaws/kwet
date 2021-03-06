package main

import (
	"flag"
	"net/http"
	"os"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/slaws/kwet/lib/backends"
	"github.com/spf13/pflag"
)

var nc = lib.Nats{}
var backend backends.Backend
var queueLists = make([]string, 0)

func main() {
	var err error
	backendType := pflag.StringP("backend", "b", "etcd", "Backend type")
	backendURL := pflag.StringArrayP("endpoint", "e", []string{"kwet-etcd-cluster-client:2379"}, "backend URL")
	var natsURL string
	pflag.StringVarP(&natsURL, "nats", "s", "", "NATS server URL")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	log.Info("Starting kwet...")
	router := NewRouter()

	backend, err = backends.SetupBackend(backends.BackendConfig{Type: *backendType, Endpoint: *backendURL})
	if err != nil {
		log.Errorf("Unable to configure Backend %s : %s", *backendType, err)
	}
	err = backend.Connect()
	if err != nil {
		log.Errorf("Unable to connect to backend %s : %s ", *backendType, err)
		os.Exit(1)
	}

	if natsURL == "" {
		var error error
		natsURL, error = backend.GetNATSURL()
		if error != nil {
			log.Errorf("Unable to get NATS URL :%s", error)
		}
	}
	if natsURL != "" {
		err = nc.Connect(natsURL, nats.Name("kwet-gateway"))
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Warnf("No url provided for NATS. Not connected.")
	}
	go WatchForConfigChanges()
	log.Fatal(http.ListenAndServe(":8080", router))
}

func WatchForConfigChanges() {
	events := make(chan lib.ConfigChangeEvent, 10)
	go backend.WatchForNATSChanges(&nc, &events)
	for {
		evt := <-events
		switch evt.Type {
		case "NATSURLChange":
			if nc.Conn != nil && nc.Conn.IsConnected() {
				nc.Disconnect()
			}
			err := nc.Connect(evt.Params)
			if err != nil {
				log.Warnf("Unable to connect to NATS at %s", evt.Params)
				continue
			}
		}

		log.Warnf("Event ! %+v", evt)
	}
}
