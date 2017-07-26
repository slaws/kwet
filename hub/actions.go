package main

import (
	"encoding/json"
	"fmt"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

type actionFunc func(lib.ClusterEvent, string, *nats.Conn) error

var actions = map[string]actionFunc{"print": ShowIt, "none": DoNothing, "copy": MoveIt}

// ShowIt shows things
func ShowIt(evt lib.ClusterEvent, params string, nc *nats.Conn) error {
	log.Infof("I'm showing that : %s ", params)
	return nil
}

// DoNothing does ... well nothing :)
func DoNothing(evt lib.ClusterEvent, params string, nc *nats.Conn) error {
	log.Infof("Look ! I'm doing nothing !")
	return nil
}

// MoveIt when you like to
func MoveIt(evt lib.ClusterEvent, params string, nc *nats.Conn) error {
	log.Infof("Moving event to %s", params)
	evt.Tags = append(evt.Tags, "Notification")
	jstr, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("Unable to publish message : %s", err)
	}
	msg := &nats.Msg{Subject: params, Data: jstr}
	nc.PublishMsg(msg)
	return nil
}
