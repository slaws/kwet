package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

// func NatsConnect(cluster, clientID string) (nats.Conn, error) {
// 	var err error
// 	nc, err = nats.Connect(cluster, clientID, nats.NatsURL("nats://nats.svc.k8s:4222"))
// 	if err != nil {
// 		log.Errorf("Error while connecting to nats url (%s) : %s", "nats://nats.svc.k8s:4222", err)
// 		return nil, err
// 	}
// 	return nc, nil
// }

func PostEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error while reading body : %s", err)
		return
	}
	msg, err := json.Marshal(lib.ClusterEvent{Source: vars["application"], Message: string(body), Tags: []string{"application", "gateway"}})
	err = nc.Publish(vars["application"], msg)
	if err != nil {
		log.Error(err)
	}
}
