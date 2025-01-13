package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"tutree/student-apis/utility"
)

// CreateZohoCRMAccessToken generates an access token for Zoho CRM using the provided refresh token, client ID, and client secret.
//
// This function performs the following steps:
// 1. Retrieves the refresh token, client ID, client secret, and Zoho website URL from environment variables.
// 2. Constructs a POST request to the Zoho API to refresh the access token.
// 3. Sends the request to the Zoho server and handles the response.
// 4. Parses the response JSON to extract the access token.
// 5. Returns the generated access token if successful, or an error if any issues occur during the process.
//
// Parameters:
// - None (all necessary information is fetched from environment variables)
//
// Returns:
// - The generated access token as a string if successful, or an error if any issue occurs.
func CreateZohoCRMAccessToken() (string, error) {
	var accessToken string
	data := url.Values{}
	// fetch refresh token from env
	refreshToken := utility.GetZohoCRMRefreshToken()
	if len(refreshToken) == 0 {
		log.Println("[NOT FOUND] CreateZohoCRMAccessToken : Failed to get refresh token from ENV.")
		return accessToken, errors.New("refresh token not found")
	}
	data.Set("refresh_token", refreshToken)
	fmt.Println("REFERESH TOKEN=================", refreshToken)
	// fetch client id from env
	clientID := utility.GetZohoCRMClientID()
	if len(clientID) == 0 {
		log.Println("[NOT FOUND] CreateZohoCRMAccessToken : Failed to get client id from ENV.")
		return accessToken, errors.New("client id not found")
	}
	data.Set("client_id", clientID)
	// fetch client secret from env
	clientSecret := utility.GetZohoCRMClientSecret()
	if len(clientSecret) == 0 {
		log.Println("[NOT FOUND] CreateZohoCRMAccessToken : Failed to get client secret from ENV.")
		return accessToken, errors.New("client secret not found")
	}
	data.Set("client_secret", clientSecret)

	data.Set("grant_type", "refresh_token")
	zohoWebsite := utility.GetZohoCRMWebsiteLinkToGenerateAccessToken()
	if len(zohoWebsite) == 0 {
		log.Println("[NOT FOUND] CreateZohoCRMAccessToken : Failed to get client secret from ENV.")
		return accessToken, errors.New("zoho website not found")
	}

	body := bytes.NewBufferString(data.Encode())
	// Create the HTTP request
	req, err := http.NewRequest("POST", zohoWebsite, body)
	if err != nil {
		log.Println("[ERROR] CreateZohoCRMAccessToken : Failed to make a request to generate access token with error :", err)
		return accessToken, err
	}
	// Set the Content-Type header to x-www-form-urlencoded
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] CreateZohoCRMAccessToken : Failed to send request to generate access token with error :", err)
		return accessToken, err
	}
	defer resp.Body.Close()
	// Read and print the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[ERROR] CreateZohoCRMAccessToken : Failed to read response to generate access token with error :", err)
		return accessToken, err
	}
	// Parse the JSON response to extract access_token
	var requestResult map[string]interface{}
	err = json.Unmarshal(respBody, &requestResult)
	if err != nil {
		fmt.Println("[ERROR] CreateZohoCRMAccessToken : Failed to parse response into json to generate access token with error :", err)
		return accessToken, err
	}
	// Extract and print the access_token
	if token, ok := requestResult["access_token"].(string); ok {
		accessToken = token
		fmt.Println("Access Token:", accessToken)
	} else {
		fmt.Println("Access Token not found in the response")
	}
	return accessToken, nil
}

// PostDataToZohoCRM sends lead data to Zoho CRM to create a new lead on the CRM system using the provided access token.
//
// The function performs the following steps:
// 1. It constructs a payload with the provided lead data and converts it into a JSON format.
// 2. It creates an HTTP POST request with the JSON payload and sets the required authorization headers (access token).
// 3. It sends the request to the Zoho CRM API endpoint to create the lead.
// 4. It checks for any errors in the response and parses the body to extract the lead ID if the request is successful.
// 5. If successful, the lead ID is returned; otherwise, an error message is returned.
//
// Parameters:
// - accessToken: OAuth access token for Zoho CRM.
// - data: The lead data to be sent to Zoho CRM as a map of key-value pairs.
//
// Returns:
// - The lead ID as an integer if the lead creation is successful.
// - An error if any issues occur during the request or response processing.
func PostDataToZohoCRM(accessToken string, data map[string]interface{}) (int, error) {
	// Get the URL for Zoho API
	url := utility.GetZohoApiToCreateLeadOnZohoCRMLeadBoard()

	// Prepare payload
	payload := map[string]interface{}{
		"data": []map[string]interface{}{data},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("error marshalling payload: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %v", err)
	}

	// Check for 2xx HTTP status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("error response from Zoho CRM: %s", string(body))
	}

	// Parse response body
	var response struct {
		Data []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details struct {
				ID string `json:"id"`
			} `json:"details"`
			Status string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("error parsing response: %v", err)
	}

	// Validate response status
	if len(response.Data) > 0 && response.Data[0].Status == "success" {
		leadIDStr := response.Data[0].Details.ID
		log.Printf("Lead successfully added with ID: %s", leadIDStr)
		leadID, err := strconv.Atoi(leadIDStr)
		if err != nil {

		}
		return leadID, nil
	}

	return 0, fmt.Errorf("unexpected response from Zoho CRM: %s", string(body))
}

// UpdateLeadsDataOnCRM updates the information of an existing lead in Zoho CRM.
//
// This function performs the following steps:
// 1. It creates a JSON payload with the provided lead data.
// 2. It constructs an HTTP PUT request to update the lead using the lead's ID.
// 3. It sets the required authorization header (Bearer token) and content type.
// 4. It sends the request to Zoho CRM and checks for errors in the response.
// 5. It prints the response status and body for debugging purposes.
//
// Parameters:
// - accessToken: OAuth access token for Zoho CRM.
// - leadID: The ID of the lead to be updated.
// - data: A map of key-value pairs containing the updated lead data.
//
// Returns:
// - An error if the update operation fails; otherwise, it returns nil.

func UpdateLeadsDataOnCRM(accessToken string, leadID int, data map[string]interface{}) error {
	// Prepare payload
	payload := map[string]interface{}{
		"data": []map[string]interface{}{data},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %v", err)
	}

	// Create the HTTP request
	url := fmt.Sprintf("https://www.zohoapis.in/crm/v2/Leads/%d", leadID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the headers
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Print the response
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(responseBody))
	return nil
}

type Lead struct {
	Phone      string `json:"Phone"`
	LeadSource string `json:"Lead_Source"`
	District   string `json:"district"`
}

type Response struct {
	Data []Lead `json:"data"`
}

func GetPhoneOfColdLeads(accessToken string) ([]Lead, error) {
	// ZOHO API URL
	url := "https://www.zohoapis.in/crm/v2/Leads?fields=Phone,Lead_Source,district"

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return []Lead{}, err
	}

	// Set headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return []Lead{}, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return []Lead{}, err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return []Lead{}, err
	}
	leads := []Lead{}
	// Extract and print phone numbers where Lead_Source is "Cold_data"
	for _, lead := range response.Data {
		if lead.LeadSource == "Cold_data" {
			lead.Phone = fmt.Sprintf("+1%s", lead.Phone)
			leads = append(leads, lead)
		}
	}
	return leads, nil
}
