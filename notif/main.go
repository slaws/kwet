package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/slaws/kwet/notif/backends"
)

var nc *nats.Conn
var debug bool

func main() {
	natsURL := pflag.StringP("nats", "s", "nats://nats:4222", "NATS server URL")
	notifyQueue := pflag.StringP("queue", "q", "notify", "Queue to subscribe to")
	c := pflag.StringP("config", "c", "/etc/kwet-notif.json", "Config file, valid format are json, yaml, toml and hcl")
	pflag.BoolVarP(&debug, "debug", "d", false, "Enable debugging")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	configFile := *c
	config := viper.New()

	config.SetConfigFile(configFile)
	err := config.ReadInConfig()
	if err != nil {
		log.Errorf("Config file not found...: %s", err)
		os.Exit(-1)
	}
	if !config.IsSet("provider") {
		log.Error("No provider specified. Exiting.")
	}
	provider, err := backends.SetupNotifier(config)
	if err != nil {
		log.Errorf("Unable to configure notifier : %s", err)
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s. Reloading.", e.Name)
	})

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
		err = provider.Send(smsg)
		if err != nil {
			log.Errorf("Error while sending notification : %s", err)
		}

	})

	runtime.Goexit()
}
