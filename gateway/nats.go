package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	nats "github.com/nats-io/go-nats"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

type msg struct {
	message string
}

// PostEvent allows to post an event to nats
func PostEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error while reading body : %s", err)
		return
	}
	msg, err := json.Marshal(lib.ClusterEvent{Source: vars["application"], Message: string(body), Tags: []string{"application", "gateway"}})
	err = nc.Conn.Publish(vars["application"], msg)
	if err != nil {
		log.Error(err)
	}
}

// SocketEvent send events through a websocket
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

	if nc.Conn == nil || !(*nc.Conn).IsConnected() {
		//		http.Error(w, "No message bus available", http.StatusBadRequest)
		conn.Close()
		return
	}
	// nc.Subscribe(">", func(msg *nats.Msg) {
	// 	eventLogs <- string(msg.Data)
	// })
	u4, err := uuid.NewV4()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	var currentQueue = u4.String()
	var sub *nats.Subscription

	sub, err = (*nc.Conn).ChanSubscribe(currentQueue, eventLogs)
	if err != nil {
		log.Errorf("Unable to subscribe to %s : %s", currentQueue, err)
	}
	defer func(sub **nats.Subscription) {
		err := (*sub).Unsubscribe()
		if err != nil {
			log.Errorf("Unable to unsubcribe from : %s", err)
		}
		err = backend.DeleteHubRule(currentQueue)
		if err != nil {
			log.Errorf("Unable to delete rule %s : %s", currentQueue, err)
		}
	}(&sub)

	go echo(conn, eventLogs)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// if string(message) == currentQueue {
		// 	continue
		// }
		// if currentQueue == ">" && string(message) == "" {
		// 	continue
		// }
		// if string(message) == "" {
		// 	currentQueue = ">"
		// } else {
		// 	currentQueue = string(message)
		// }
		//
		// err = sub.Unsubscribe()
		// if err != nil {
		// 	log.Errorf("Unable to unsubscribe from %s", currentQueue)
		// }
		// sub, err = (*nc.Conn).ChanSubscribe(currentQueue, eventLogs)
		// if err != nil {
		// 	log.Errorf("Unable to subscribe from %s", currentQueue)
		// }
		r := lib.HubRule{
			Name:      currentQueue,
			Queue:     ".*",
			Condition: string(message),
			Action:    fmt.Sprintf("copy(event,'%s')", currentQueue),
		}
		jsonstr, err := json.Marshal(r)
		if err != nil {
			log.Errorf("Unable to marshal %+v : %s", r, err)
		}
		err = backend.SetHubRule(currentQueue, string(jsonstr))
		if err != nil {
			log.Errorf("Unable to set rule %s : %s", string(jsonstr), err)
		}
	}
}

func echo(conn *websocket.Conn, c chan *nats.Msg) {
	m := msg{}
	for {
		nm := <-c
		if !lib.ContainsString(queueLists, nm.Subject) {
			queueLists = append(queueLists, nm.Subject)
		}
		m.message = string(nm.Data)
		// fmt.Printf("Got message: %#v\n", m)
		conn.WriteMessage(websocket.TextMessage, []byte(m.message))
	}
}
