package main

import (
	"fmt"
	"net/http"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
)

func ListQueue(w http.ResponseWriter, r *http.Request) {
	log.Info("Listing...")
	nc, _ := nats.Connect("nats://nats.svc.k8s:4222")
	nc.Subscribe("foo", func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})
	nc.Close()
}
