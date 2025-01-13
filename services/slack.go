package services

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// SlackRequestBody is Slack request structure
type SlackRequestBody struct {
	Text string `json:"text"`
}

// SendSlackMessage sends a message to a Slack channel using a Bot User OAuth Token.
// It retrieves the token from the environment, constructs the request, and sends it to Slack's API.
func SendSlackMessage(channelID string, msg string) {
	// Get Slack token from environment
	slackAuthToken := os.Getenv("SLACK_AUTH_TOKEN")
	if len(slackAuthToken) == 0 {
		log.Println("[ERROR] Failed to get SLACK_AUTH_TOKEN")
		return
	}

	// Create the request payload
	slackBody, _ := json.Marshal(map[string]string{
		"channel": channelID,
		"text":    msg,
	})

	// Create a new HTTP request to Slack's chat.postMessage API
	req, err := http.NewRequest(http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewBuffer(slackBody))
	if err != nil {
		log.Println("[ERROR] failed to create HTTP request: ", err)
		return
	}

	// Set the required headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+slackAuthToken)

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] failed to send request to Slack: ", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		log.Println("[ERROR] failed to read Slack response: ", err)
		return
	}

	// Check for a successful response
	var responseMap map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &responseMap); err != nil {
		log.Println("[ERROR] failed to parse Slack response: ", err)
		return
	}

	if ok, found := responseMap["ok"].(bool); !found || !ok {
		log.Println("[ERROR] error from Slack API: ", responseMap)
		return
	}
}
