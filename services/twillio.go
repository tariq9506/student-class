package services

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/twilio/twilio-go"

	"tutree/student-apis/utility"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"

	LookupsV1 "github.com/twilio/twilio-go/rest/lookups/v1"
	lookupsV2 "github.com/twilio/twilio-go/rest/lookups/v2"
)

// SendMessage sends an SMS message to a specified phone number using the Twilio API.
// Parameters:
// - phone: The recipient's phone number (without the country code).
// - message: The content of the SMS message to be sent.
// - dialingCode: The international dialing code for the recipient's country (e.g., "+1" for the US).
// Returns:
// - error: Returns an error if the message sending fails, otherwise returns nil.
func SendMessage(phone string, message string, dialingCode string) error {

	phone = fmt.Sprintf("%s%s", dialingCode, phone)
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_ACCOUNT_AUTH_TOKEN")
	twilioFrom := os.Getenv("TWILIO_FROM_NUMBER")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &openapi.CreateMessageParams{}

	params.SetTo(phone)
	params.SetFrom(twilioFrom)
	params.SetBody(message)
	_, err := client.Api.CreateMessage(params)

	if err != nil {
		log.Println("SendMessage : failed while sending message to the tutree user, with error:", err)
		return err
	}

	return nil
}

// IsPhNumberDeliverable checks if a given phone number is deliverable by using Twilio's Lookup API.
// Parameters:
// - phone: The phone number to be checked, in E.164 format (e.g., "+14155552671").
// - countrycode: The country code of the phone number, in ISO 3166-1 alpha-2 format (e.g., "US").
// Returns:
// - bool: True if the phone number is deliverable, false otherwise.
// - error: An error if the lookup fails or if there is an issue with the Twilio API.
func IsPhNumberDeliverable(phone, countrycode string) (bool, error) {

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_ACCOUNT_AUTH_TOKEN")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &LookupsV1.FetchPhoneNumberParams{}
	params.SetCountryCode(countrycode)

	params.SetType([]string{"carrier"})
	_, err := client.LookupsV1.FetchPhoneNumber(phone, params)
	if err != nil {
		log.Println("IsPhNumberDeliverable : The phone number lookup failed, with error:", err)
		return false, err
	}

	return true, nil
}

// IsPhoneNumberVoip checks if the given phone number is a VoIP number or not using Twilio's Lookup API.
// Parameters:
// - phone: The phone number to be checked, in E.164 format (e.g., "+14155552671").
// - countrycode: The country code of the phone number, in ISO 3166-1 alpha-2 format (e.g., "US").
// Returns:
// - bool: True if the phone number is a VoIP number, false otherwise.
// - error: An error if the lookup fails or if there is an issue with the Twilio API.
func IsPhoneNumberVoip(phone, countrycode string) (bool, error) {
	// Retrieve Twilio Account SID and Auth Token from utility functions
	accountSid := utility.GetTwilioAccountID()
	authToken := utility.GetTwilioAuthorizationToken()
	// Create a new Twilio Rest Client with the provided Account SID and Auth Token

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})
	// Initialize parameters for the phone number lookup

	params := &lookupsV2.FetchPhoneNumberParams{}
	params.SetCountryCode(countrycode)
	params.SetFields("line_type_intelligence")
	// Fetch phone number details using the Twilio Lookup API

	resp, err := client.LookupsV2.FetchPhoneNumber(phone, params)
	if err != nil {
		log.Println("IsPhoneNumberVoip : The phone number lookup failed, with error:", err)
		return true, err
	}
	if resp.LineTypeIntelligence != nil {
		// Type assertion to extract line type intelligence data as a map
		lineTypeIntelligenceValue, ok := (*resp.LineTypeIntelligence).(map[string]interface{})
		if !ok {
			log.Println("IsPhoneNumberVoip: failed due to Invalid LineTypeIntelligence format")
			return true, err
		}
		// Extract the line type from the line type intelligence data

		if lineType, ok := lineTypeIntelligenceValue["type"].(string); ok {
			log.Println("lineType", lineType)
			if lineType != "nonFixedVoip" {
				return false, nil
			}
		}
	} else {
		return false, nil
	}
	return true, nil
}
