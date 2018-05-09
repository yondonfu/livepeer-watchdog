package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TwilioConfig struct {
	AcctSid   string `json:"acctSid"`
	AuthToken string `json:"authToken"`
	From      string `json:"from"`
}

type TwilioService struct {
	config TwilioConfig
	url    string
	client *http.Client
}

func NewTwilioService(config TwilioConfig) *TwilioService {
	return &TwilioService{
		config: config,
		url:    fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%v/Messages.json", config.AcctSid),
		client: &http.Client{},
	}
}

func (ts *TwilioService) Notify(to string, msg string) error {
	msgData := url.Values{}
	msgData.Set("To", to)
	msgData.Set("From", ts.config.From)
	msgData.Set("Body", msg)
	reader := *strings.NewReader(msgData.Encode())

	req, err := http.NewRequest("POST", ts.url, &reader)
	if err != nil {
		return err
	}

	req.SetBasicAuth(ts.config.AcctSid, ts.config.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)

		return decoder.Decode(&data)
	} else {
		return fmt.Errorf("Twilio request failed with status code: %v", resp.Status)
	}
}
