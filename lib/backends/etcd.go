package backends

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/slaws/kwet/lib"

	log "github.com/sirupsen/logrus"
)

// Etcd defines etcd informations.
type Etcd struct {
	Endpoint       []string
	Conn           *clientv3.Client
	RequestTimeout time.Duration
}

func init() {
	Register("etcd", ConfigureEtcd)
}

// ConfigureEtcd creates a provider
func ConfigureEtcd(conf BackendConfig) (Backend, error) {
	return &Etcd{
		Endpoint:       conf.Endpoint,
		RequestTimeout: 5 * time.Second,
	}, nil
}

// Connect connects to the etcd cluster
func (e *Etcd) Connect() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   e.Endpoint,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("Unable to connect to etcd endpoints [%s] : %s", strings.Join(e.Endpoint, ","), err)

	}
	e.Conn = cli
	return nil
}

//GetNATSURL gets the url for NATS server
func (e *Etcd) GetNATSURL() (string, error) {
	url, err := e.get("/kwet/nats/url")
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	return url, nil
}

func (e *Etcd) get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.RequestTimeout)
	resp, err := e.Conn.Get(ctx, key)
	cancel()
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("Record not found")
	}
	// for _, ev := range resp.Kvs {
	// 	fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	// }
	// fmt.Printf("---- end of print kvs ----\n")
	// return "", fmt.Errorf("Todo")
	return string(resp.Kvs[0].Value), nil
}

func (e *Etcd) WatchForNATSChanges(nc lib.Nats) {
	changes := e.Conn.Watch(context.Background(), "/kwet/nats", clientv3.WithPrefix())
	for wresp := range changes {
		for _, ev := range wresp.Events {
			if string(ev.Kv.Key) == "/kwet/nats/url" && string(ev.Kv.Key) != nc.Conn.ConnectedUrl() {
				log.Infof("NATS URL changed in etcd from %s to %s : reconnecting", nc.Conn.ConnectedUrl(), ev.Kv.Value)
				nc.Disconnect()
				err := nc.Connect(string(ev.Kv.Value))
				if err != nil {
					log.Warnf("Unable to connect to NATS at %s", string(ev.Kv.Value))
				}
			}
		}
	}
}
