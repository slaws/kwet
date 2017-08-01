package main

import (
	"encoding/json"
	"flag"
	"regexp"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Knetic/govaluate"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/spf13/pflag"
)

// Config defines a configuration
type Config struct {
	ListenQueues []string          `toml:"listen_queues,omitempty"`
	SyslogQueues []string          `toml:"syslog_queues"`
	TopicRules   map[string][]rule `toml:"rules"`
}

type topicRule struct {
	Rules []rule `toml:"when"`
}
type rule struct {
	Condition string
	Action    string
	Params    string
}

var nc *nats.Conn

func processSyslogMessages(message *nats.Msg, conf Config, nc *nats.Conn) {
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
		applyRules(syslogMessages, conf, message.Subject, nc)
	}
}

func processClusterEvent(msg *nats.Msg, conf Config, nc *nats.Conn) {
	var evt lib.ClusterEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		log.Warningf("Error while processing message '%s': %s. Skipping", string(msg.Data), err)
		return
	}
	if lib.ContainsString(evt.Tags, "Notification") {
		return
	}
	applyRules(evt, conf, msg.Subject, nc)
}

func applyRules(evt lib.ClusterEvent, conf Config, source string, nc *nats.Conn) {
	var ruleList []rule
	for key, value := range conf.TopicRules {
		match, _ := regexp.MatchString(key, source)
		if match {
			ruleList = append(ruleList, value...)
		}
	}

	for _, rule := range ruleList {
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
	natsURL := pflag.StringP("nats", "s", "nats://nats:4222", "NATS server URL")
	c := pflag.StringP("config", "c", "/etc/kwet-hub.toml", "Config file (toml format)")
	l := pflag.StringArrayP("syslogqueue", "l", nil, "Queues receiving syslog messages from fluentd")
	q := pflag.StringArrayP("queue", "q", nil, "Queue to listen on for events")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	var conf Config
	if _, err = toml.DecodeFile(*c, &conf); err != nil {
		log.Errorf("Error while parsing config file : %s", err)
	}
	if len(*q) != 0 {
		conf.ListenQueues = *q
	}
	if len(*l) != 0 {
		conf.SyslogQueues = *l
	}

	log.Info("Starting kwet-hub")
	log.Infof("Listening for events on %s.Syslog Queues : %s", strings.Join(conf.ListenQueues, ","), strings.Join(conf.SyslogQueues, ","))
	nc, err = lib.NatsConnect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}

	for _, queue := range conf.ListenQueues {
		nc.Subscribe(queue, func(msg *nats.Msg) {
			log.Infof("Message for : %s", msg.Subject)
			if lib.MatchStringInList(conf.SyslogQueues, msg.Subject) {
				processSyslogMessages(msg, conf, nc)
			} else {
				processClusterEvent(msg, conf, nc)
			}
		})
	}

	runtime.Goexit()
}
