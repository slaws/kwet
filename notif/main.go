package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/fsnotify/fsnotify"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	b "github.com/slaws/kwet/notif/backends"
	"github.com/spf13/viper"
)

var nc *nats.Conn

func main() {
	natsURL := flag.String("s", "nats://nats:4222", "NATS server URL ( default: nats://nats:4222 )")
	notifyQueue := flag.String("q", "notify", "Queue to subscribe to ( default : notify )")
	configFile := flag.String("c", "/etc/kwet-notif.json", "Config file, valid format are json, yaml, toml and hcl (default: /tmp/foo.json)")

	flag.Parse()
	config := viper.New()
	// config.SetConfigName("foo")
	// config.AddConfigPath("/tmp/")
	config.SetConfigFile(*configFile)
	err := config.ReadInConfig()
	if err != nil {
		log.Error("Config file not found...")
		os.Exit(-1)
	}
	if !config.IsSet("provider") {
		log.Error("No provider specified. Exiting.")
	}
	provider, err := b.SetupNotifier(config)
	if err != nil {
		log.Errorf("Unable to configure notifier : %s", err)
	}
	config.WatchConfig()
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s. Reloading.", e.Name)
	})

	log.Info("Starting kwet Notifier...")
	b.ListNotifier()
	nc, err := lib.NatsConnect(*natsURL)
	if err != nil {
		log.Fatal(err)
	}
	nc.Subscribe(*notifyQueue, func(msg *nats.Msg) {
		var smsg lib.ClusterEvent
		err := json.Unmarshal(msg.Data, &smsg)
		if err != nil || smsg.Source == "" || smsg.Message == "" {
			provider.Send(lib.ClusterEvent{
				Source:  "kwet-notif",
				Message: fmt.Sprintf("Malformed message received.\n```%+v```", string(msg.Data)),
			})
			return
		}
		provider.Send(smsg)

	})

	runtime.Goexit()
}
