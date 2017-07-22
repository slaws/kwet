package main

import (
	"encoding/json"
	"net/http"

	nats "github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Source      string `json:"source"`
	Message     string `json:"message"`
	Destination string `json:"destination"`
}

func NatsConnect(cluster, clientID string) (nats.Conn, error) {
	var err error
	nc, err = nats.Connect(cluster, clientID, nats.NatsURL("nats://nats.svc.k8s:4222"))
	if err != nil {
		log.Errorf("Error while connecting to nats url (%s) : %s", "nats://nats.svc.k8s:4222", err)
		return nil, err
	}
	return nc, nil
}

func ListQueue(w http.ResponseWriter, r *http.Request) {
	log.Info("Listing...")
	err := nc.Publish("foo", []byte("Someone is listing..."))
	if err != nil {
		log.Error(err)
	}
}

func PostEvent(w http.ResponseWriter, r *http.Request) {
	parser := json.NewDecoder(r.Body)
	var e Event
	err := parser.Decode(&e)
	if err != nil {
		log.Error(err)
	}
	msg, err := json.Marshal(e)
	if err != nil {
		log.Error(err)
	}
	err = nc.Publish(e.Destination, msg)
	if err != nil {
		log.Error(err)
	}
}
