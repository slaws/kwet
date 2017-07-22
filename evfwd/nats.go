package main

import (
	"encoding/json"
	"net/http"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
)

// Event is a thing
type Event struct {
	Source      string `json:"source"`
	Message     string `json:"message"`
	Destination string `json:"destination"`
}

func NatsConnect() (*nats.Conn, error) {
	var err error
	nc, err = nats.Connect("nats://nats.svc.k8s:4222")
	if err != nil {
		log.Errorf("Error while connecting to nats url (%s) : %s", "nats://nats.svc.k8s:4222", err)
		return nil, err
	}
	return nc, nil
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
