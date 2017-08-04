package backends

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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

//SetNATSURL gets the url for NATS server
func (e *Etcd) SetNATSURL(natsurl string) error {
	err := e.set("/kwet/nats/url", natsurl)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

//SetHubRule set a hub rule
func (e *Etcd) SetHubRule(name, jsonRule string) error {
	err := e.set(fmt.Sprintf("/kwet/hub/rules/%s", name), jsonRule)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

//GetHubRules gets Hub routing rules
func (e *Etcd) GetHubRules() ([]lib.HubRule, error) {
	rules := make([]lib.HubRule, 0)
	list, err := e.getKV("/kwet/hub/rules", clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	if len(list) == 0 {
		return rules, nil
	}
	for _, value := range list {
		var rule lib.HubRule
		err := json.Unmarshal(value.Value, &rule)
		if err != nil {
			log.Errorf("Unable to unmarshal value %s : %s", string(value.Value), err)
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

//DeleteHubRule deletes a hub rule
func (e *Etcd) DeleteHubRule(name string) error {
	err := e.del(fmt.Sprintf("/kwet/hub/rules/%s", name))
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func (e *Etcd) GetSingleHubRule(key string) (*lib.HubRule, error) {
	list, err := e.getKV(fmt.Sprintf("/kwet/hub/rules/%s", key))
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	if len(list) == 0 {
		return nil, nil
	}
	var rule lib.HubRule
	err = json.Unmarshal(list[0].Value, &rule)
	if err != nil {
		log.Errorf("Unable to unmarshal value %s : %s", string(list[0].Value), err)
	}
	return &rule, nil
}

func (e *Etcd) set(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.RequestTimeout)
	_, err := e.Conn.Put(ctx, key, value)
	cancel()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func (e *Etcd) del(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.RequestTimeout)
	_, err := e.Conn.Delete(ctx, key)
	cancel()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func (e *Etcd) get(key string, options ...clientv3.OpOption) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.RequestTimeout)
	resp, err := e.Conn.Get(ctx, key, options...)
	cancel()
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("Record not found")
	}
	return string(resp.Kvs[0].Value), nil
}

func (e *Etcd) getKV(key string, options ...clientv3.OpOption) ([]*mvccpb.KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.RequestTimeout)
	resp, err := e.Conn.Get(ctx, key, options...)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("Record not found")
	}
	return resp.Kvs, nil
}

func (e *Etcd) WatchForNATSChanges(nc lib.Nats) {
	changes := e.Conn.Watch(context.Background(), "/kwet/nats", clientv3.WithPrefix())
	for wresp := range changes {
		for _, ev := range wresp.Events {
			if string(ev.Kv.Key) == "/kwet/nats/url" && string(ev.Kv.Key) != nc.Conn.ConnectedUrl() {
				log.Infof("NATS URL changed in etcd from %s to %s : reconnecting", nc.Conn.ConnectedUrl(), ev.Kv.Value)
				if nc.Conn != nil && nc.Conn.IsConnected() {
					nc.Disconnect()
				}
				err := nc.Connect(string(ev.Kv.Value))
				if err != nil {
					log.Warnf("Unable to connect to NATS at %s", string(ev.Kv.Value))
				}
			}
		}
	}
}
