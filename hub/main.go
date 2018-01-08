package main

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/Knetic/govaluate"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/slaws/kwet/lib/backends"
	"github.com/spf13/pflag"
)

var nc = lib.Nats{}
var backend backends.Backend
var listenQueues = []string{">"}
var syslogQueues = make([]string, 0)
var topicRules = make([]lib.HubRule, 0)

func processSyslogMessages(message *nats.Msg, nc lib.Nats) {
	log.Infof("Queue %s is a syslog queue ... using special format..", message.Subject)
	var list []interface{}
	var syslogMessages lib.ClusterEvent
	err := json.Unmarshal(message.Data, &list)
	if err != nil {
		log.Warnf("Unable to process log list %+v : %s ", string(message.Data), err)
	}
	for _, v := range list {
		listMsg := v.([]interface{})
		msg, err := json.Marshal(listMsg[1])
		if err != nil {
			log.Warningf("Unable to marshal %+v : %s", listMsg[1], err)
		}

		err = json.Unmarshal(msg, &syslogMessages)
		if err != nil {
			log.Warningf("Unable to unmarshal %v : %v", listMsg[1], err)
		}
		syslogMessages.Source = "fluent-syslog"
		syslogMessages.Tags = append(syslogMessages.Tags, "syslog", "fluent")
		applyRules(syslogMessages, message.Subject, nc)
	}
}

func processClusterEvent(msg *nats.Msg, nc lib.Nats) {
	var evt lib.ClusterEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		log.Warningf("Error while processing message '%s': %s. Skipping", string(msg.Data), err)
		return
	}
	if lib.ContainsString(evt.Tags, "Notification") {
		return
	}
	applyRules(evt, msg.Subject, nc)
}

func processMessage(msg lib.ClusterEvent, nc lib.Nats) {
	if lib.ContainsString(msg.Tags, "Notification") {
		return
	}
	applyRules(msg, msg.Source, nc)
}

func applyRules(evt lib.ClusterEvent, source string, nc lib.Nats) {
	var ruleList []lib.HubRule
	for _, value := range topicRules {
		match, _ := regexp.MatchString(value.Queue, source)
		if match {
			ruleList = append(ruleList, value)
		}
	}

	for _, rule := range ruleList {
		log.Debugf("Processing rule %s", rule.Name)
		expression, err := govaluate.NewEvaluableExpressionWithFunctions(rule.Condition, functions)
		if err != nil {
			log.Warningf("Error while creating condition : %s. Skipping", err)
			continue
		}
		params := make(map[string]interface{}, 0)
		params["event"] = evt
		result, err := expression.Evaluate(params)
		if err != nil {
			log.Errorf("Error while evaluating condition '%s' : %s. Skipping", rule.Condition, err)
			continue
		}
		if result == false {
			continue
		}
		action, err := govaluate.NewEvaluableExpressionWithFunctions(rule.Action, actionFunctions)
		if err != nil {
			log.Warningf("Error while evaluating action : %s. Skipping", err)
			continue
		}
		result, err = action.Evaluate(params)
		if err != nil {
			log.Errorf("Error while doing action '%s' : %s. Skipping", rule.Action, err)
			continue
		}
		if result == false {
			log.Warningf("action '%s' failed: %s. Skipping", rule.Action, err)
			continue
		}

	}
}

func main() {

	var err error
	backendType := pflag.StringP("backend", "b", "etcd", "Backend type")
	backendURL := pflag.StringArrayP("endpoint", "e", []string{"kwet-etcd-cluster-client:2379"}, "backend URL")
	var natsURL string
	pflag.StringVarP(&natsURL, "nats", "s", "", "NATS server URL")
	logLevelOpt := pflag.StringP("log-level", "l", "info", "Log Level (panic, fatal, error, warn, info, debug)")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	logLevel, err := log.ParseLevel(*logLevelOpt)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.SetLevel(logLevel)

	backend, err = backends.SetupBackend(backends.BackendConfig{Type: *backendType, Endpoint: *backendURL})
	if err != nil {
		log.Errorf("Unable to configure Backend %s : %s", *backendType, err)
	}
	err = backend.Connect()
	if err != nil {
		log.Errorf("Unable to connect to backend %s : %s ", *backendType, err)
		os.Exit(1)
	}
	syslogQueues, err = backend.GetSyslogQueues()
	if err != nil {
		log.Errorf("Unable to get Hub Rules :%s", err)
	}

	log.Info("Starting kwet-hub")
	log.Infof("Listening for events on %s.Syslog Queues : %s", strings.Join(listenQueues, ","), strings.Join(syslogQueues, ","))

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
	topicRules, err = backend.GetHubRules()
	if err != nil {
		log.Errorf("Unable to get Hub Rules :%s", err)
	}

	for _, queue := range listenQueues {
		(*nc.Conn).Subscribe(queue, func(msg *nats.Msg) {
			//log.Infof("Message for : %s", msg.Subject)
			var raw interface{}
			err := json.Unmarshal(msg.Data, &raw)
			if err != nil {
				log.Debugf("Unable to process raw msg %+v : %s ", string(msg.Data), err)
			}
			v := reflect.ValueOf(raw)
			switch v.Kind() {
			case reflect.Slice:
				var m [][]interface{}
				err := json.Unmarshal(msg.Data, &m)
				if err != nil {
					log.Warnf("Unable to process raw msg %+v : %s ", string(msg.Data), err)
				}
				for _, mg := range m {
					var t lib.ClusterEvent
					mesg, err := json.Marshal(mg[1])
					if err != nil {
						log.Warningf("Unable to marshal %+v : %s", mg[1], err)
						continue
					}
					err = json.Unmarshal(mesg, &t)
					if err != nil {
						log.Warnf("Unable to unmarshall %+v : %s", string(mesg), err)
						continue
					}
					t.Source = msg.Subject
					if t.K8SMessage != nil {
						t.Kind = "kubernetes"
					} else if t.SyslogMessage != nil {
						t.Kind = "syslog"
					} else {
						t.Kind = "unknown"
					}
					//log.Infof("%+v", t)
					processMessage(t, nc)
				}
			case reflect.Map:
				log.Debugf("Map received")
				log.Debugf("%+v", v)
				var t lib.ClusterEvent
				err = json.Unmarshal(msg.Data, &t)
				if err != nil {
					log.Warnf("Unable to unmarshall %+v : %s", string(msg.Data), err)
				} else {
					t.Kind = "syslog"
					t.Source = msg.Subject
					processMessage(t, nc)
				}
			default:
				log.Printf("Unknown type %T\n", v)
			}
		})
	}
	go WatchForConfigChanges()
	runtime.Goexit()
}

func WatchForConfigChanges() {
	var err error
	events := make(chan lib.ConfigChangeEvent, 10)
	go backend.WatchForHubRulesChanges(&nc, &events)
	for {
		evt := <-events
		switch evt.Type {
		case "HubRuleChange":
			topicRules, err = backend.GetHubRules()
			if err != nil {
				log.Errorf("Unable to get Hub Rules :%s", err)
			}
		}
	}
}
