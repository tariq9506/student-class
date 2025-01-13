package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type MessageInfo struct {
	ID              int64           `json:"id"`
	UserID          int64           `json:"user_id"`
	TutorID         int64           `json:"tutor_id"`
	SessionID       int64           `json:"session_id"`
	EmailTemplateID int64           `json:"email_template_id"`
	EmailType       string          `json:"email_type"`
	EmailContent    json.RawMessage `json:"email_content"`
	EventType       string          `json:"event_type"`
	SMSContent      json.RawMessage `json:"sms_content"`
	TriggerTime     string          `json:"trigger_time"`
	MessageType     string          `json:"message_type"`
	Status          string          `json:"status"`
	Interval        string          `json:"interval"`
	Email           string          `json:"email"`
	PhoneNumber     string          `json:"phone_number"`
	SMSType         string          `json:"sms_type"`
}

type CancelMsgEvent struct {
	UserID          int64  `json:"user_id"`
	TutorID         int64  `json:"tutor_id"`
	SessionID       int64  `json:"session_id"`
	EmailTemplate   string `json:"email_template"`
	EmailTemplateID int64  `json:"email_template_id"`
	SMSType         string `json:"sms_type"`
}

func CallMessageServiceWebhook(sendMessageEvent []MessageInfo) {
	url := os.Getenv("MESSAGE_SERVICE_HOST") + "/v1/store-message"
	requestBody, err := json.Marshal(sendMessageEvent)
	if err != nil {
		log.Println("CallMessageServiceWebhook: failed to marshal payload with error: ", err)
		return
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("CallMessageServiceWebhook: Error creating request:", err)
		return
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the http.DefaultClient
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("CallMessageServiceWebhook: Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	// Print the response status and body
	log.Println("Response Status:", resp.Status)
	log.Println("Response Body:", string(body))

}

func CancelMessageServiceWebhook(cancelMsgEvent []CancelMsgEvent) error {
	url := os.Getenv("MESSAGE_SERVICE_HOST") + "/v1/cancel/daily-msg"
	requestBody, err := json.Marshal(cancelMsgEvent)
	if err != nil {
		log.Println("CallMessageServiceWebhook: failed to marshal payload with error: ", err)
		return err
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("CallMessageServiceWebhook: Error creating request:", err)
		return err
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the http.DefaultClient
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("CallMessageServiceWebhook: Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return err
	}

	// Print the response status and body
	log.Println("Response Status:", resp.Status)
	log.Println("Response Body:", string(body))

	return nil
}
