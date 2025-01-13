package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/models"
	"tutree/student-apis/models/apitracker"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"
)

// GetDetailsOfSupportedCountryByIP retrieves the details of a country that is supported by the application
// based on the IP address of the client making the request.
func GetDetailsOfSupportedCountryByIP(c *gin.Context) (models.CountryDetail, error) {
	// Get the client's IP address from the context.
	clientIP := utility.GetClientIP(c)

	// Determine the location of the client using their IP address.
	location := GetLocationFromIP(clientIP)

	// Get the details of the supported country by code.
	details, err := models.GetDetailsOfSupportedCountryByCode(location.CountryCode)
	if err != nil {
		// If an error occurs while retrieving the details of the supported country, log the error.
		log.Printf("GetDetailsOfSupportedCountryByIP: failed to get details of supported country with: %v", err)
		return models.CountryDetail{}, err
	}
	// If no error occurs, return the details of the supported country.
	return details, nil
}

// GetLocationFromIP convert ip to location
func GetLocationFromIP(ip string) services.Location {
	location := services.GetLocationLocally(ip)

	if location.CountryCode != "US" {
		return location
	}

	currentCity, currentState, currentCountry, _ := models.GetLocationByZipcode(location.PostalCode)

	if currentCity == "" {
		currentCity = "New York City"
	}

	if currentState == "" {
		currentState = "NY"
	}

	if currentCountry == "" {
		currentCountry = "US"
	}

	location.City = currentCity
	location.State = currentState
	location.Country = currentCountry

	return location
}

// HandleJSONErrorResponse handles sending a standardized JSON error response to the client.
// Parameters:
// - customError: An instance of CustomAPIError containing the error details to be sent.
// - detailError: An optional detailed error that provides additional context.
// - c: The Gin context used to send the response.
func HandleJSONErrorResponse(customError apierror.CustomAPIError, detailError error, c *gin.Context) {
	log.Println(customError.Error())
	response := gin.H{
		"message": customError.Message,
		"status":  "failed",
		"code":    customError.Code,
	}
	if detailError != nil {
		response["detailed_msg"] = detailError.Error()
	}

	sendJSONResponse(customError.HttpStatusCode, response, c)
}

// sendJSONResponse helper function sends a JSON response with the specified status code.
// Parameters:
// code: The HTTP status code to send.
// json: The JSON response body to send. If nil, an empty JSON object is created.
// c: The Gin context used to send the response.
func sendJSONResponse(code int, json gin.H, c *gin.Context) {
	if json == nil {
		json = gin.H{}
	}
	json["status"] = http.StatusText(code)
	c.JSON(code, json)
}

// RandStringForAgentID create random string of length 5-8
// We need to auto created agent_id, to track our agent
func RandStringForAgentID(phone string) string {

	var letterRunes = []rune("1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	seeding := time.Now().UnixNano() + StringToBin(utility.GetDomainName()) + StringToBin(phone)
	rand.Seed(seeding)

	// Set n to be a random number between 5 and 8 inclusive
	n := 5 + rand.Intn(4) // 5 + random number between 0 and 3

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// stringToBin return binary for a string.
func StringToBin(s string) (abc int64) {
	var binString string
	for _, c := range s {
		binString = fmt.Sprintf("%s%b", binString, c)
	}

	i, err := strconv.ParseInt(binString[:10], 10, 64)
	if err != nil {
		log.Println("stringToBin: failed:", err)
	}
	return i
}

func GetLocationDetailsFromIP(c *gin.Context) {
	// Get the client's IP address from the context.
	clientIP := c.Query("ip")
	log.Println("ip: ", clientIP)

	// Determine the location of the client using their IP address.
	location := GetLocationFromIP(clientIP)

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"location": location,
	})
}

// input: *gin.Context
// ApiTracker is a function triggered by an incoming HTTP request. It checks for bot-related user agents and skips storing
// data if it detects a bot. If not, it constructs a complete URL using request details, captures necessary data like IP,
// path, query parameters, and user session information, then saves this API call's details into a database asynchronously
// using the StoreAPIHistory function.
func ApiTracker(c *gin.Context) {
	go func() {

		userAgent := c.GetHeader("User-Agent")
		if strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "Bot") {
			return
		}

		url := utility.GetHostURL() + c.Request.URL.Path
		ip := utility.GetClientIP(c)
		query := c.Request.URL.RawQuery
		apiMethod := c.Request.Method
		// Parse the multipart form. The argument specifies the max memory for parsing.
		// c.Request.ParseMultipartForm is used to parse form file fields if multipart.File is present
		err := c.Request.ParseMultipartForm(5 * 1024 * 1024)
		if err != nil {
			log.Println("[NOTE]: apiTracker: failed to parse multipart form with error: ", err)
		}
		// c.Request.PostForm would store form text values and not form file values
		formValues := c.Request.PostForm
		encodedString := formValues.Encode()
		tokenString := c.GetHeader("Authorization")

		apiTracker := apitracker.APITracker{
			URL:         url,
			IP:          ip,
			Query:       query,
			UserSession: tokenString,
			APIMethod:   apiMethod,
			BodyParams:  encodedString,
			UserAgent:   userAgent,
		}
		apitracker.StoreAPIHistory(apiTracker)
	}()
}
func GetZohoAccesstoken(c *gin.Context) {
	accessToken, err := services.CreateZohoCRMAccessToken()
	if err != nil {
		log.Println("[ERROR] GetZohoAccesstoken : Failed to get Zoho access token with error :", err)
		//controllers.HandleJSONErrorResponse(apierror.ErrorZohoAccessToken, err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"access_token": accessToken,
	})
}
