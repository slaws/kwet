package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type infos struct {
	NatsStatus bool
}

// Index welcomes
func Index(w http.ResponseWriter, r *http.Request) {
	info := &infos{}
	if nc.Conn == nil {
		info.NatsStatus = false
	} else {
		info.NatsStatus = (*nc.Conn).IsConnected()
	}
	t, err := template.ParseFiles("templates/index.html", "templates/queuelist.html")
	if err != nil {
		log.Errorf("Unable to parse index template : %s", err)
	}

	err = t.Execute(w, info)
	if err != nil {
		log.Errorf("Error executing template : %s ", err)
	}

}

// Settings allows to configure kwet-* modules
func Settings(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/settings.html")
	if err != nil {
		log.Errorf("Unable to parse index template : %s", err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Errorf("Error executing template : %s ", err)
	}
}

// ListQueues lists queues
func ListQueues(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(queueLists)
	if err != nil {
		log.Errorf("Unable to marshall queue list : %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
