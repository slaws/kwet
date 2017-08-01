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
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	nc.Subscribe(">", func(msg *nats.Msg) {
		eventLogs <- string(msg.Data)
	})

	go echo(conn)
}

func echo(conn *websocket.Conn) {
	m := msg{}
	for {
		m.message = <-eventLogs

		fmt.Printf("Got message: %#v\n", m)
		conn.WriteMessage(websocket.TextMessage, []byte(m.message))

	}
}
