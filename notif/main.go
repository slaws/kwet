package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/slaws/kwet/notif/backends"
	"github.com/tidwall/gjson"
)

var nc *nats.Conn
var debug bool

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
	natsURL := pflag.StringP("nats", "s", "nats://nats:4222", "NATS server URL")
	notifyQueue := pflag.StringP("queue", "q", "notify", "Queue to subscribe to")
	c := pflag.StringP("config", "c", "/etc/kwet-notif.toml", "Config file")
	pflag.BoolVarP(&debug, "debug", "d", false, "Enable debugging")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	var conf lib.Config
	if _, err := toml.DecodeFile(*c, &conf); err != nil {
		log.Errorf("Error while parsing config file : %s", err)
	}

	provider, err := backends.SetupNotifier(conf)
	if err != nil {
		log.Errorf("Unable to configure notifier : %s", err)
	}

	log.Info("Starting kwet Notifier...")
	backends.ListNotifier()
	nc, err := lib.NatsConnect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}
	nc.Subscribe(*notifyQueue, func(msg *nats.Msg) {
		var smsg lib.ClusterEvent
		if debug {
			log.Infof("Message received from [%s] : %s", *notifyQueue, string(msg.Data))
		}
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
		rules, found := conf.Format[smsg.Source]
		log.Infof("%b value, %+v, %s", found, rules, smsg.Source)
		if found {
			var mesg map[string]interface{}
			err = json.Unmarshal([]byte(smsg.Message.(string)), &mesg)
			if err != nil {
				log.Warnf("Error while processing message !")
			}
			data := smsg.Message.(string)
			if Variabilize(rules.Title, rules.Vars, data) != "" {
				dataMsg := fmt.Sprintf(`{
					"title": "%s",
					"title_link": "%s",
					"text": "%s",
					"image_url": "%s"
					}`,
					Variabilize(rules.Title, rules.Vars, data),
					Variabilize(rules.TitleLink, rules.Vars, data),
					Variabilize(rules.Text, rules.Vars, data),
					Variabilize(rules.ImageURL, rules.Vars, data))
				smsg.Message = dataMsg
			}
		}

		log.Infof("Sending :%+v", smsg)
		err = provider.Send(smsg)
		if err != nil {
			log.Errorf("Error while sending notification : %s", err)
		}

	})

	runtime.Goexit()
}
