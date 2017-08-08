package lib

import "time"

// ClusterEvent is a message used by kwet
type ClusterEvent struct {
	Source    string      `json:"source"`
	Message   interface{} `json:"message"`
	Tags      []string    `json:"tags,omitempty"`
	Processed bool        `json:"processed"`
	Host      string      `json:"host,omitempty"`
	Identity  string      `json:"ident,omitempty"`
	PID       string      `json:"pid,omitempty"`
	Priority  string      `json:"priority,omitempty"`
	Facility  string      `json:"facility,omitempty"`
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
