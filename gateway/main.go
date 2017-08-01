package main

import (
	"flag"
	"net/http"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/spf13/pflag"
)

var nc *nats.Conn
var eventLogs = make(chan string, 100)

func main() {
	var err error
	natsURL := pflag.StringP("nats", "s", "nats://nats:4222", "NATS server URL")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	log.Info("Starting kwet...")

	router := NewRouter()
	//	nc, err := NatsConnect("test-cluster", "1")
	nc, err = lib.NatsConnect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", router))
}
