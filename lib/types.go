package lib

// ClusterEvent is a message used by kwet
type ClusterEvent struct {
	Source  string `json:"source"`
	Message string `json:"message"`
}
