package main

import (
	"encoding/json"
	"flag"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/Knetic/govaluate"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/spf13/pflag"
)

// Config defines a configuration
type Config struct {
	TopicRules map[string][]rule `toml:"rules"`
}

type topicRule struct {
	Rules []rule `toml:"when"`
}
type rule struct {
	Condition string
	Action    string
	Params    string
}

func main() {

	natsURL := pflag.StringP("nats", "s", "nats://nats:4222", "NATS server URL")
	c := pflag.StringP("config", "c", "/etc/kwet-notif.json", "Config file, valid format are json, yaml, toml and hcl")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	var conf Config
	if _, err := toml.DecodeFile(*c, &conf); err != nil {
		log.Errorf("Error while parsing config file : %s", err)
	}

	log.Info("Starting kwet-hub")
	nc, err := lib.NatsConnect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}

	nc.Subscribe("*", func(msg *nats.Msg) {
		var evt lib.ClusterEvent
		err := json.Unmarshal(msg.Data, &evt)
		if err != nil {
			log.Warningf("Error while processing message '%s': %s. Skipping", string(msg.Data), err)
			return
		}
		if lib.ContainsString(evt.Tags, "Notification") {
			return
		}

		tlist := conf.TopicRules["all"]
		for _, rule := range tlist {
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
			log.Infof("When %s, then %s(%s)", rule.Condition, rule.Action, rule.Params)
			act, found := actions[rule.Action]
			if found == false {
				log.Warnf("Action '%s' unknown. Skipping", rule.Action)
				continue
			}
			err = act(evt, rule.Params, nc)
			if err != nil {
				log.Warningf("err : %s", err)
			}

		}
	})

	runtime.Goexit()
}
