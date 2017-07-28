package lib

// ClusterEvent is a message used by kwet
type ClusterEvent struct {
	Source    string      `json:"source"`
	Message   interface{} `json:"message"`
	Tags      []string    `json:"tags,omitempty"`
	Processed bool        `json:"processed"`
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
	Title     string
	TitleLink string `toml:"title_link"`
	Text      string
	ImageURL  string `toml:"image_url"`
	Vars      map[string]string
}
