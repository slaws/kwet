package lib

// ClusterEvent is a message used by kwet
type ClusterEvent struct {
	Source    string      `json:"source"`
	Message   interface{} `json:"message"`
	Tags      []string    `json:"tags,omitempty"`
	Processed bool        `json:"processed"`
}
