package lib

import (
	"time"
)

// ClusterEvent is a message used by kwet
type ClusterEvent struct {
	Kind string
	*SyslogMessage
	*K8SMessage
	Source    string   `json:"source"`
	Tags      []string `json:"tags,omitempty"`
	Processed bool     `json:"processed"`
}

//HubRule defines a routing rule for kwet-hub
type HubRule struct {
	Name      string
	Queue     string
	Condition string
	Action    string
	Params    string
}

type SyslogMessage struct {
	Tag       string    `json:"tag"`
	Priority  string    `json:"pri"`
	Time      time.Time `json:"time"`
	Host      string    `json:"host"`
	Ident     string    `json:"ident"`
	PID       string    `json:"pid"`
	MsgID     string    `json:"msgid"`
	ExtraData string    `json:"extradata"`
	Message   string    `json:"message"`
}

type K8SMessage struct {
	Tag        string            `json:"tag"`
	Log        string            `json:"log"`
	Stream     string            `json:"stream"`
	Time       string            `json:"time"`
	Docker     map[string]string `json:"docker"`
	Kubernetes *K8SMetadata      `json:"kubernetes"`
	Host       string            `json:"host"`
	MasterURL  string            `json:"master_url"`
}

type K8SMetadata struct {
	PodName       string            `json:"pod_name"`
	Namespace     string            `json:"namespace_name"`
	PodID         string            `json:"pod_id"`
	Labels        map[string]string `json:"labels"`
	Annotations   map[string]string `json:"annotations"`
	Host          string            `json:"host"`
	ContainerName string            `json:"container_name"`
	DockerID      string            `json:"docker_id"`
}

type Config struct {
	Provider ProviderInfo
	Format   map[string]FormatRule
}

type ProviderInfo struct {
	Name string
	URL  string
}

type FormatRule struct {
	Name      string
	Title     string
	TitleLink string `toml:"title_link"`
	Text      string
	ImageURL  string `toml:"image_url"`
	Vars      map[string]string
}

type ConfigChangeEvent struct {
	Type   string
	Params string
}
