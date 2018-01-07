package backends

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
	store "github.com/slaws/kwet/lib/backends"
)

// Notifier is an interface for medias
type Notifier interface {
	GetName() string
	Send(lib.ClusterEvent) error
}

// NotifierFactory creates Notifier
type NotifierFactory func(conf store.Backend) (Notifier, error)

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
func SetupNotifier(conf store.Backend) (Notifier, error) {
	provider, err := conf.GetNotifProvider()
	if err != nil {
		return nil, err
	}
	factory, ok := backendNotifier[strings.ToLower(provider)]
	if !ok {
		return nil, fmt.Errorf("No provider %s declared", strings.ToLower(provider))
	}
	return factory(conf)
}
