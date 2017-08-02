package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Data struct {
	Queues []string
}

// Index welcomes
func Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/queuelist.html")
	if err != nil {
		log.Errorf("Unable to parse index template : %s", err)
	}

	// t.ExecuteTemplate(w, "templates/index.html", "foo")
	log.Infof("%+v", queueLists)
	//	d := Data{Queues: queueLists}
	err = t.Execute(w, queueLists)
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
