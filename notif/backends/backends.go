package backends

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	"github.com/spf13/viper"
)

// Notifier is an interface for medias
type Notifier interface {
	GetName() string
	Send(lib.ClusterEvent) error
}

// NotifierFactory creates Notifier
type NotifierFactory func(conf *viper.Viper) (Notifier, error)

var backendNotifier = make(map[string]NotifierFactory)

// Register adds a new notifier
func Register(name string, service NotifierFactory) {
	if service == nil {
		log.Panicf("Backend notifier %s does not exist.", name)
	}
	_, registered := backendNotifier[name]
	if registered {
		log.Errorf("Backend notifier %s already registered. Ignoring.", name)
	}
	backendNotifier[name] = service
}

// ListNotifier lists available Notifier
func ListNotifier() {
	availableDatastores := make([]string, len(backendNotifier)-1)
	for k := range backendNotifier {
		availableDatastores = append(availableDatastores, k)
	}
	log.Infof("Available notifier are : %v", availableDatastores)
}

// SetupNotifier builds a notifier
func SetupNotifier(conf *viper.Viper) (Notifier, error) {
	factory, ok := backendNotifier[strings.ToLower(conf.GetString("provider"))]
	if !ok {
		return nil, fmt.Errorf("No provider %s declared", strings.ToLower(conf.GetString("provider")))
	}
	return factory(conf)
}
