package main

import (
	"encoding/json"
	"fmt"

	"github.com/Knetic/govaluate"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

var actionFunctions = map[string]govaluate.ExpressionFunction{
	// print prints the message to stdout
	"print": func(args ...interface{}) (interface{}, error) {
		message, found := args[0].(string)
		if found == false {
			return false, fmt.Errorf("print() expects a string as first argument")
		}
		log.Infof("[print] %s", message)
		return true, nil
	},
	// nothing does ... nothing :)
	"nothing": func(args ...interface{}) (interface{}, error) {
		return true, nil
	},
	// copy copies a message to another queue
	"copy": func(args ...interface{}) (interface{}, error) {
		evt := args[0].(lib.ClusterEvent)
		queue := args[1].(string)
		evt.Tags = append(evt.Tags, "Notification")
		jstr, err := json.Marshal(evt)
		if err != nil {
			return false, fmt.Errorf("Unable to marshal message : %s", err)
		}
		msg := &nats.Msg{Subject: queue, Data: jstr}
		err = (*nc.Conn).PublishMsg(msg)
		if err != nil {
			return false, fmt.Errorf("Unable to publish message to queue %s : %v", queue, err)
		}
		return true, nil
	},
}
