package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

type msg struct {
	message string
}

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

func SocketEvent(w http.ResponseWriter, r *http.Request) {
	// if r.Header.Get("Origin") != "http://"+r.Host {
	// 	http.Error(w, "Origin not allowed", 403)
	// 	log.Errorf("%s Refused", r.RemoteAddr)
	// 	return
	// }
	var eventLogs = make(chan *nats.Msg, 100)
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	// nc.Subscribe(">", func(msg *nats.Msg) {
	// 	eventLogs <- string(msg.Data)
	// })
	_, err = nc.ChanSubscribe(">", eventLogs)
	if err != nil {
		log.Errorf("Unable to subscribe to '>' : %s", err)
	}

	go echo(conn, eventLogs)
}

func SocketSpecificEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var eventLogs = make(chan *nats.Msg, 100)
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	_, err = nc.ChanSubscribe(vars["queue"], eventLogs)
	if err != nil {
		log.Errorf("Unable to subscribe to '>' : %s", err)
	}

	go echo(conn, eventLogs)
}

func echo(conn *websocket.Conn, c chan *nats.Msg) {
	m := msg{}
	for {
		nm := <-c
		if !lib.ContainsString(queueLists, nm.Subject) {
			queueLists = append(queueLists, nm.Subject)
		}
		m.message = string(nm.Data)
		fmt.Printf("Got message: %#v\n", m)
		conn.WriteMessage(websocket.TextMessage, []byte(m.message))
	}
}
