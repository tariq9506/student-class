package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type CallRequest struct {
	To           string `json:"phoneNumber"`       // Corrected key name
	Text         string `json:"text"`              // Corrected key name
	From         string `json:"from"`              // Corrected key name
	AssistanceID string `json:"your_assistant_id"` // Corrected key name
}

// MakeACallOfZohoColdLeads makes a call to a phone number using the vapi.ai API.
// Parameters:
// - phoneNumber: The phone number to call.
// Returns:
// - int: The status code of the API request.
// - error: Any error encountered during the process.

func MakeACallOfZohoColdLeads(phoneNumber string) (int, error) {
	// vapi.ai API URL
	url := os.Getenv("VAPI_AI_CALL_PHONE_URL")
	method := "POST"
	// Payload for the API request
	// Corrected the payload to match the API documentation
	assistanceId := os.Getenv("YOUR_ASSISTANT_ID")
	if len(assistanceId) == 0 {
		log.Println("Assistant ID not found in the environment variables")
		return 0, fmt.Errorf("assistant ID not found in the environment variables")
	}
	phoneId := os.Getenv("YOUR_PHONE_NUMBER_ID")
	if len(phoneId) == 0 {
		log.Println("Phone number ID not found in the environment variables")

		return 0, fmt.Errorf("phone number ID not found in the environment variables")
	}
	payload := fmt.Sprintf(`{
        "assistantId": "%s",
        "customer": {
            "number": "%s"
        },
        "phoneNumberId": "%s"
    }`, assistanceId, phoneNumber, phoneId)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	apikey := os.Getenv("VAPI_AI_API_KEY")
	if len(apikey) == 0 {
		log.Println("API key not found in the environment variables")
		return 0, fmt.Errorf("API key not found in the environment variables")
	}
	// Add headers
	req.Header.Set("Authorization", "Bearer "+apikey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer resp.Body.Close()
	// Check the response status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success: Log and return the status code
		fmt.Println("Call initiated successfully!")
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Client error (e.g., bad request)
		fmt.Printf("Client error: %d - Please check your request.\n", resp.StatusCode)
	} else if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		// Server error (e.g., API issue)
		fmt.Printf("Server error: %d - There was an issue with the API server.\n", resp.StatusCode)
	} else {
		// Other status codes
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
	}

	// Read and parse the response body for more info (optional)
	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		fmt.Println("Error decoding response:", err)
	} else {
		// Log or process the responseBody for further insights
		fmt.Printf("Response body: %+v\n", responseBody)
	}

	// Step 2: Fetch the phoneCallProviderId directly from the decoded map
	callerID, ok := responseBody["id"].(string)
	if !ok {
		log.Fatalf("Error: phoneCallProviderId not found or not a string in response")
	}
	log.Printf("Phone call initiated successfully with ID: %s\n", callerID)

	// Get the call details
	status, resone := GetCallDetails(callerID)
	// Return the status code for further handling
	log.Println("Call Status:", status)
	log.Println("Ended Reason:", resone)
	return resp.StatusCode, nil
}
func GetCallDetails(callID string) (string, string) {
	// Replace with the actual call ID
	apiKey := os.Getenv("VAPI_AI_API_KEY") // Replace with your actual API key
	vapiURL := os.Getenv("VAPI_AI_CALL_DETAILS_URL")

	// Construct the URL to fetch the call details
	url := fmt.Sprintf("%s%s", vapiURL, callID)
	log.Println("URL:", url)

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return "", ""
	}

	// Set headers for the request
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return "", ""
	}
	defer resp.Body.Close()

	// Handle non-OK status codes
	if resp.StatusCode != http.StatusOK {
		log.Println("Error: received status code", resp.StatusCode)
		return "", ""
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return "", ""
	}

	// Parse the JSON response body into a map
	var responseBody map[string]interface{}
	if err := json.Unmarshal(body, &responseBody); err != nil {
		log.Println("Error decoding JSON:", err)
		return "", ""
	}

	// Log the response body for debugging
	log.Printf("Response body: %+v\n", responseBody)

	// Fetch the call status and ended reason, checking if they exist
	status, ok := responseBody["status"].(string)
	if !ok {
		log.Println("Error: Unable to fetch 'status' from the response")
		return "", ""
	}
	log.Println("Call Status:", status)

	endedReason, ok := responseBody["endedReason"].(string)
	if !ok {
		log.Println("Error: Unable to fetch 'endedReason' from the response")
		return "", ""
	}
	log.Println("Ended Reason:", endedReason)

	// Return the status and endedReason
	return status, endedReason
}
