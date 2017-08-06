package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/pflag"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	store "github.com/slaws/kwet/lib/backends"
	"github.com/slaws/kwet/notif/backends"
	"github.com/tidwall/gjson"
)

var debug bool
var nc = lib.Nats{}
var backend store.Backend

// Variabilize replaces vars in string
func Variabilize(format string, params map[string]string, data string) string {
	for key, val := range params {
		log.Infof("Variabilizing %s with %+v", key, gjson.Get(data, val))
		format = strings.Replace(format, "%{"+key+"}", fmt.Sprintf("%s", gjson.Get(data, val)), -1)
	}
	log.Infof("==================== end of params")
	return format
}

func main() {
	var natsURL string
	var err error
	notifyQueue := pflag.StringP("queue", "q", "notify", "Queue to subscribe to")
	backendType := pflag.StringP("backend", "b", "etcd", "Backend type")
	backendURL := pflag.StringArrayP("endpoint", "e", []string{"kwet-etcd-cluster-client:2379"}, "backend URL")
	pflag.StringVarP(&natsURL, "nats", "s", "", "NATS server URL")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	backend, err = store.SetupBackend(store.BackendConfig{Type: *backendType, Endpoint: *backendURL})
	if err != nil {
		log.Errorf("Unable to configure Backend %s : %s", *backendType, err)
	}
	err = backend.Connect()
	if err != nil {
		log.Errorf("Unable to connect to backend %s : %s ", *backendType, err)
		os.Exit(1)
	}
	if natsURL == "" {
		var error error
		natsURL, error = backend.GetNATSURL()
		if error != nil {
			log.Errorf("Unable to get NATS URL :%s", error)
		}
	}
	if natsURL != "" {
		err = nc.Connect(natsURL)
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Warnf("No url provided for NATS. Not connected.")
	}

	provider, err := backends.SetupNotifier(backend)
	if err != nil {
		log.Errorf("Unable to configure notifier : %s", err)
	}

	log.Info("Starting kwet Notifier...")
	backends.ListNotifier()
	if nc.Conn != nil && (*nc.Conn).IsConnected() {
		sub, err := (*nc.Conn).Subscribe(*notifyQueue, func(msg *nats.Msg) {
			messageHandler(msg, provider)
		})
		if err != nil {
			log.Warnf("Unable to subscribe to %s", *notifyQueue)
		} else {
			nc.SubjectSubscriptions = append(nc.SubjectSubscriptions, sub)
		}
	}
	go WatchForConfigChanges(provider)
	runtime.Goexit()
}

func WatchForConfigChanges(p backends.Notifier) {
	events := make(chan lib.ConfigChangeEvent, 10)
	go backend.WatchForNATSChanges(&nc, &events)
	for {
		evt := <-events
		switch evt.Type {
		case "NATSURLChange":
			if nc.Conn != nil && nc.Conn.IsConnected() {
				nc.Disconnect()
			}
			err := nc.Connect(evt.Params)
			if err != nil {
				log.Warnf("Unable to connect to NATS at %s", evt.Params)
				continue
			}
			for i, sub := range nc.SubjectSubscriptions {
				subj := sub.Subject
				sub.Unsubscribe()
				if nc.Conn != nil && (*nc.Conn).IsConnected() {
					s, err := (*nc.Conn).Subscribe(subj, func(msg *nats.Msg) {
						messageHandler(msg, p)
					})
					if err != nil {
						log.Warnf("Unable to subscribe to %s", subj)
					} else {
						nc.SubjectSubscriptions[i] = s
					}
				}

			}
		}

		log.Warnf("Event ! %+v", evt)
	}
}

func messageHandler(msg *nats.Msg, provider backends.Notifier) {
	var smsg lib.ClusterEvent
	log.Infof("msg : %s", string(msg.Data))
	err := json.Unmarshal(msg.Data, &smsg)
	if err != nil || smsg.Source == "" || smsg.Message == "" {
		err = provider.Send(lib.ClusterEvent{
			Source:  "kwet-notif",
			Message: fmt.Sprintf("Malformed message received.\n```%+v```", string(msg.Data)),
		})
		if err != nil {
			log.Errorf("Error while sending notification : %s", err)
		}
		return
	}
	//rules, found := conf.Format[smsg.Source]
	rule, err := backend.GetSingleNotifFormatRule(smsg.Source)
	if err != nil {
		log.Errorf("Error while retrieving rule for %s : %s", smsg.Source, err)
		return
	}
	log.Infof("value, %+v, %s", rule, smsg.Source)
	var mesg map[string]interface{}
	err = json.Unmarshal([]byte(smsg.Message.(string)), &mesg)
	if err != nil {
		log.Warnf("Error while processing message !")
	} else {
		data := smsg.Message.(string)
		log.Warnf("Title : %s", Variabilize(rule.Title, rule.Vars, data))
		if Variabilize(rule.Title, rule.Vars, data) != "" {
			dataMsg := fmt.Sprintf(`{
				"title": "%s",
				"title_link": "%s",
				"text": "%s",
				"image_url": "%s"
				}`,
				Variabilize(rule.Title, rule.Vars, data),
				Variabilize(rule.TitleLink, rule.Vars, data),
				Variabilize(rule.Text, rule.Vars, data),
				Variabilize(rule.ImageURL, rule.Vars, data))
			smsg.Message = dataMsg
		}
	}

	log.Infof("Sending :%+v", smsg)
	err = provider.Send(smsg)
	if err != nil {
		log.Errorf("Error while sending notification : %s", err)
	}
}
