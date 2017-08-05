package backends

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/slaws/kwet/lib"
)

type Backend interface {
	Connect() error
	GetNATSURL() (string, error)
	SetNATSURL(string) error
	GetHubRules() ([]lib.HubRule, error)
	GetSingleHubRule(string) (*lib.HubRule, error)
	SetHubRule(string, string) error
	DeleteHubRule(string) error
	GetSyslogQueues() ([]string, error)
	SetSyslogQueues(string) error
	GetNotifProvider() (string, error)
	SetNotifProvider(string) error
	GetNotifProviderConfig(string) (*lib.ProviderInfo, error)
	SetNotifProviderConfig(string, string) error
	GetNotifFormatRules() ([]lib.FormatRule, error)
	GetSingleNotifFormatRule(string) (*lib.FormatRule, error)
	SetNotifFormatRule(string, string) error
	DeleteNotifFormatRule(string) error
	WatchForNATSChanges(lib.Nats)
}

type BackendConfig struct {
	Type     string
	Endpoint []string
}

// BackendFactory creates a backend
type BackendFactory func(conf BackendConfig) (Backend, error)

var backendList = make(map[string]BackendFactory)

// Register adds a new notifier
func Register(name string, service BackendFactory) {
	if service == nil {
		log.Panicf("Backend %s does not exist.", name)
	}
	_, registered := backendList[name]
	if registered {
		log.Errorf("Backend %s already registered. Ignoring.", name)
	}
	backendList[name] = service
}

// SetupNotifier builds a notifier
func SetupBackend(conf BackendConfig) (Backend, error) {
	factory, ok := backendList[strings.ToLower(conf.Type)]
	if !ok {
		return nil, fmt.Errorf("No backend %s declared", strings.ToLower(conf.Type))
	}
	return factory(conf)
}
