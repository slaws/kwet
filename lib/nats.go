package lib

import (
	"fmt"

	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
)

// Event is a thing
type Event struct {
	Source      string `json:"source"`
	Message     string `json:"message"`
	Destination string `json:"destination"`
}

type Nats struct {
	Conn                 *nats.Conn
	SubjectSubscriptions []*nats.Subscription
}

// NatsConnect connects to the specified URL
func NatsConnect(url string, opt ...*nats.Options) (*nats.Conn, error) {
	// var err error
	nc, err := nats.Connect(url)
	if err != nil {
		log.Errorf("Error while connecting to nats url (%s) : %s", url, err)
		return nil, err
	}
	return nc, nil
}

func (n *Nats) Connect(url string, options ...nats.Option) error {
	nc, err := nats.Connect(url, options...)
	if err != nil {
		return fmt.Errorf("Error while connecting to nats url (%s) : %s", url, err)
	}
	n.Conn = nc
	return nil
}

func (n *Nats) Disconnect() {
	n.Conn.Close()
}
