package backends

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/slaws/kwet/lib"
)

// RocketChat is a config struct
type RocketChat struct {
	URL string
}

type RcMsg struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Title     string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Text      string `json:"text,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
	Color     string `json:"color,omitempty"`
}

func init() {
	Register("rocketchat", CreateRocketChat)
}

// GetName returns provider name
func (rc *RocketChat) GetName() string {
	return "RocketChat"
}

// Send sends a notification
func (rc *RocketChat) Send(message lib.ClusterEvent) error {
	attach := Attachment{}
	err := json.Unmarshal([]byte(message.Message.(string)), &attach)
	log.Infof("---> %+v", attach)
	var smsg RcMsg
	if err != nil || attach.Title == "" {
		smsg = RcMsg{
			Text: fmt.Sprintf("Notification from %s", message.Source),
			Attachments: []Attachment{
				Attachment{
					Text: message.Message.(string),
				},
			},
		}
	} else {
		smsg = RcMsg{
			Text: fmt.Sprintf("Notification from %s", message.Source),
			Attachments: []Attachment{
				attach,
			},
		}
	}
	m, err := json.Marshal(smsg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", rc.URL, bytes.NewBuffer(m))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected statuscode %d. Return : %s", resp.StatusCode, string(body))
	}
	return nil
}

// CreateRocketChat creates a provider
func CreateRocketChat(conf lib.Config) (Notifier, error) {
	return &RocketChat{
		URL: conf.Provider.URL,
	}, nil
}
