package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
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
	natsurl, err := backend.GetNATSURL()
	if err != nil {
		if err.Error() != "Record not found" {
			http.Error(w, "Connection with backend failed", http.StatusInternalServerError)
			log.Errorf("Connection with backend failed : %s", err)
			return
		}
		natsurl = ""
	}
	hubRules, err := backend.GetHubRules()
	if err != nil {
		log.Errorf("Unable to get hub rules : %s", err)
	}
	data := struct {
		Natsurl  string
		HubRules []lib.HubRule
	}{
		natsurl,
		hubRules,
	}
	t, err := template.ParseFiles("templates/settings.html")
	if err != nil {
		log.Errorf("Unable to parse index template : %s", err)
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Errorf("Error executing template : %s ", err)
	}
}

func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch vars["module"] {
	case "general":
		updateGeneral(r.Form, w, r)
		break
	case "hub":
		updateHub(r.Form, w, r)
		break
	}

}

func updateGeneral(post url.Values, w http.ResponseWriter, r *http.Request) {
	nats := r.Form["natsurl"][0]
	match, err := regexp.MatchString("nats://[a-zA-Z0-9\\.\\-]+:[0-9]+", nats)
	if err != nil {
		log.Errorf("Regexp error : %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !match {
		http.Error(w, "Invalid pattern", http.StatusInternalServerError)
		return
	}
	err = backend.SetNATSURL(r.Form["natsurl"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings#general-panel", http.StatusSeeOther)
}

func updateHub(post url.Values, w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "/settings#hub-panel", http.StatusSeeOther)
}

func AddNewHubRule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		t, err := template.ParseFiles("templates/newhubrule.html")
		if err != nil {
			log.Errorf("Unable to parse index template : %s", err)
		}
		err = t.Execute(w, nil)
		if err != nil {
			log.Errorf("Error executing template : %s ", err)
		}
		break
	case "POST":
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		name := r.Form["newRuleName"][0]
		queue := r.Form["newRuleQueue"][0]
		cond := r.Form["newRuleCondition"][0]
		act := r.Form["newRuleAction"][0]
		rule := lib.HubRule{
			Name:      strings.ToUpper(name),
			Queue:     queue,
			Condition: cond,
			Action:    act,
		}
		err = saveHubRule(rule, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings#hub-panel", http.StatusSeeOther)
	}
}

// EditHubRule allows to edit a specific rule
func EditHubRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	switch r.Method {
	case "GET":
		r, err := backend.GetSingleHubRule(vars["rulename"])
		if err != nil || r == nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t, err := template.ParseFiles("templates/newhubrule.html")
		if err != nil {
			log.Errorf("Unable to parse index template : %s", err)
		}
		err = t.Execute(w, r)
		if err != nil {
			log.Errorf("Error executing template : %s ", err)
		}
		break
	case "POST":
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		name := r.Form["newRuleName"][0]
		queue := r.Form["newRuleQueue"][0]
		cond := r.Form["newRuleCondition"][0]
		act := r.Form["newRuleAction"][0]
		rule := lib.HubRule{
			Name:      strings.ToUpper(name),
			Queue:     queue,
			Condition: cond,
			Action:    act,
		}
		err = saveHubRule(rule, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings#hub-panel", http.StatusSeeOther)
	}
}

func DeleteHubRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := backend.DeleteHubRule(vars["rulename"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/settings#hub-panel", http.StatusOK)
}

func saveHubRule(rule lib.HubRule, w http.ResponseWriter, r *http.Request) error {
	jsonstr, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	log.Infof("%+v", string(jsonstr))
	err = backend.SetHubRule(rule.Name, string(jsonstr))
	if err != nil {
		return err
	}
	return nil
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
