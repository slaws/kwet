package backends

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

// Notifier is an interface for medias
type Notifier interface {
	GetName() string
	Send(lib.ClusterEvent) error
}

// NotifierFactory creates Notifier
type NotifierFactory func(conf lib.Config) (Notifier, error)

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
func SetupNotifier(conf lib.Config) (Notifier, error) {
	factory, ok := backendNotifier[strings.ToLower(conf.Provider.Name)]
	if !ok {
		return nil, fmt.Errorf("No provider %s declared", strings.ToLower(conf.Provider.Name))
	}
	return factory(conf)
}
