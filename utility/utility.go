package utility

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetHostURL get host url
func GetHostURL() string {
	return os.Getenv("HOST_URL")
}

// GetTutorHostURL get tutor host url
func GetTutorHostURL() string {
	return os.Getenv("TUTOR_HOST_URL")
}

// generateOTP
// input :
// Output: OTP
// Desc  : This controller will generate OTP.
// OTP Generation
func GenerateOTP() string {
	charSet := "1234567890"
	otp := randomStringGenerator(charSet, 4)
	return otp
}

func randomStringGenerator(charSet string, codeLength int32) string {
	code := ""
	charSetLength := int32(len(charSet))

	// Seed the random number generator to ensure different results each time
	randomNumber := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := int32(0); i < codeLength; i++ {
		// Generate a random index within the bounds of charSetLength
		index := randomNumber.Intn(int(charSetLength))

		// Append the character at the generated index to the code
		code += string(charSet[index])
	}

	// Return the generated random string
	return code
}
func GetClientIP(c *gin.Context) string {
	clientIP := c.ClientIP()
	if os.Getenv("ENV") == "local" {
		clientIP = os.Getenv("LOCAL_IP")
	}
	return clientIP
}

// input: no parameter
// output: bool
// func GetTwilioAccountID will check whether voip numbers on the server is allowed or not.
func AllowVoipNumbers() bool {
	isAllowed, err := strconv.ParseBool(os.Getenv("ALLOW_VOIP_NUMBERS"))
	if err != nil {
		log.Println("[ENV-MISSING] AllowVoipNumbers: failed while fetching .env variable for voip number")
		return true
	}
	return isAllowed
}

// input: no parameter
// output: string
// func GetTwilioAccountID will return twilio account sid which will be used for calling twilio api service
func GetTwilioAccountID() string {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	return accountSid
}

// input: no parameter
// output: string
// func GetTwilioAccountID will return twilio account authorization token which will be used for calling twilio api service
func GetTwilioAuthorizationToken() string {
	authToken := os.Getenv("TWILIO_ACCOUNT_AUTH_TOKEN")
	return authToken
}
func SQLNullStringToString(s sql.NullString) string {
	if s.Valid {
		return strings.TrimSpace(s.String)
	}

	return ""

}

// SQLNullIntToInt convert sql null int to int
func SQLNullIntToInt(i sql.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}

	return 0

}

// SQLNullFloatToFloat convert sql null float to float
func SQLNullFloatToFloat(f sql.NullFloat64) float64 {
	if f.Valid {
		return f.Float64
	}

	return 0.0

}

// SQLNullTimeToTime convert sql null time to time
func SQLNullTimeToTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}

	return time.Now()

}

func SQLNullBoolToBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}

// GetDomainName get company name
func GetDomainName() string {
	return os.Getenv("DOMAIN_NAME")
}

// ValidateEmailPattern checks if the given email address matches a predefined regular expression pattern for a valid email address.
// It uses a regular expression to validate the email format and returns whether the email is valid or not.
//
// Input Parameters:
// - email: string - The email address to be validated.
//
// Output:
// - (bool, error) - The function returns a boolean indicating whether the email is valid and an error if there is an issue compiling the regex pattern.
//   - bool: Returns true if the email matches the pattern, false otherwise.
//   - error: Returns an error if there is an issue compiling the regular expression; otherwise, nil.
//
// On success, the function returns true or false based on whether the email matches the pattern.
//
// On failure, it logs the error and returns false with the error message.
func ValidateEmailPattern(email string) (bool, error) {
	// Define the regular expression pattern for a valid email address.
	// This pattern checks that the email contains valid characters before and after the '@' symbol and has a domain with a dot.
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// Compile the regular expression pattern.
	re, err := regexp.Compile(emailPattern)
	// If there is an error compiling the regex pattern, log the error and return false and the error.
	if err != nil {
		log.Println("ValidateEmailPattern: Failed while compile regular expression for email pattern with error: ", err)
		return false, err
	}
	// Check if the email address matches the compiled regex pattern.
	if re.MatchString(email) {
		// If the email matches the pattern, log the success and return true with no error.
		log.Println("ValidateEmailPattern: validating email successfull.")
		return true, nil
	} else {
		// If the email does not match the pattern, log the failure and return false with no error.
		log.Println("ValidateEmailPattern: validating email failed.")
		return false, nil
	}
}

// GetStandardTimeZone returns the standard time zone abbreviation for a given time zone identifier.
// Parameters:
// - identifier: A string time zone identifier (e.g., "Asia/Kolkata").
// Returns:
// - abbreviation: A string containing the standard time zone abbreviation (e.g., "IST" for "Asia/Kolkata").
// - error: An error if the time zone identifier is invalid or cannot be loaded.
func GetStandardTimeZone(identifier string) (string, error) {
	loc, err := time.LoadLocation(identifier)
	if err != nil {
		log.Printf("Failed to load location for time zone identifier '%s': %v\n", identifier, err)
		return "", err
	}
	now := time.Now().In(loc)
	// _, offsetOffset := now.Zone() // Zone() returns the abbreviation and the offset in seconds from UTC
	// offset := fmt.Sprintf("UTC%+03d:%02d", offsetOffset/3600, (offsetOffset%3600)/60)
	abbreviation := now.Format("MST")
	return abbreviation, nil
}

// input: no parameter
// output: AWS_URL
// GetAWSURL will return url where file would be uploaded on aws.
func GetAWSURL(bucketName string) string {
	url := "https://%s.s3.amazonaws.com"
	url = fmt.Sprintf(url, bucketName)
	return url
}

// input: no parameter
// output: cloudfront_url
// GetCloudFrontURL will return cloud front url where file would be uploaded on CDN.
func GetCloudFrontURL() string {
	return os.Getenv("BUCKET_FOR_CLOUDFRONT")
}

// input: no parameter
// output: string
// GetBrevoSenderName will return sender name which is used to send email from brevo service.
func GetBrevoSenderName() string {
	return os.Getenv("BREVO_SENDER_NAME")
}

// input: no parameter
// output: string
// GetBrevoSenderEmail will return sender email which is used to send email from brevo service.
func GetBrevoSenderEmail() string {
	return os.Getenv("BREVO_SENDER_EMAIL")
}

// input: no parameter
// output: string
// GetBrevoAPIKey will return api key which is used to send email from brevo service.
func GetBrevoAPIKey() string {
	return os.Getenv("BREVO_API_KEY")
}

// GetSourceOfSignup returns the signup source based on query parameters.
// It identifies sources like "Google ad," "Whatsapp Group," "Craigslist flyer,"
// "Craigslist Link," "Flyer1 - Pink," and "Flyer2" by analyzing specific parameters
// such as "gclid," "utm_medium," "utm_source," and "inviter."
// func GetSourceOfSignup(queryParam string) string {
// 	// Parse the query string into a URL.Values map
// 	query, err := url.ParseQuery(queryParam)
// 	if err != nil {
// 		return queryParam
// 	}

// 	// Check for Google ad
// 	if _, exists := query["gclid"]; exists {
// 		return "Google Ads"
// 	}
// 	if _, exists := query["gbraid"]; exists {
// 		return "Google Ads"
// 	}
// 	if _, exists := query["wbraid"]; exists {
// 		return "Google Ads"
// 	}

// 	// Check for Bing Ads
// 	if _, exists := query["msclkid"]; exists {
// 		return "Bing Ads"
// 	}

// 	// Check for Bing Ads
// 	if _, exists := query["flyer"]; exists {
// 		return "Flyer"
// 	}

// 	utmMedium := query.Get("utm_medium")
// 	utmSource := query.Get("utm_source")
// 	utmCampaign := query.Get("utm_campaign")
// 	inviter := query.Get("inviter")

//		switch {
//		case utmMedium == "whatsapp" && utmSource == "groups":
//			return "Whatsapp Group"
//		case utmMedium == "CL" && utmCampaign == "free_demo" && utmSource == "":
//			return "Craigslist flyer"
//		case utmMedium == "CL" && utmSource != "":
//			return "Craigslist Link"
//		case inviter == "flyer1":
//			return "Flyer1 - Pink"
//		case inviter == "flyer2":
//			return "Flyer2"
//		default:
//			if len(queryParam) != 0 && queryParam != "/" {
//				return "tutree.com/" + queryParam
//			} else {
//				return "tutree.com"
//			}
//		}
//	}
func GetSourceOfSignup(queryParam string) string {
	// Parse the query string into a URL.Values map
	// query, err := url.ParseQuery(queryParam)
	// if err != nil {
	// 	return queryParam
	// }

	// Check for Google ad
	if strings.Contains(queryParam, "gclid") || strings.Contains(queryParam, "gbraid") || strings.Contains(queryParam, "wbraid") {
		if strings.Contains(queryParam, "lpo/coding-classes-for-kids") {
			return "Google Ads LPO"
		}
		return "Google Ads"
	}

	// Check for Bing Ads
	if strings.Contains(queryParam, "msclkid") {
		if strings.Contains(queryParam, "lpo/coding-classes-for-kids") {
			return "Bing Ads LPO"
		}
		return "Bing Ads"
	}

	// Check for Bing Ads
	if strings.Contains(queryParam, "flyer") {
		if strings.Contains(queryParam, "lpo/coding-classes-for-kids") {
			return "Flyer LPO"
		}
		return "Flyer"
	}

	if strings.Contains(strings.ToLower(queryParam), "vapi") {
		return "vapi"
	}

	return "Tutree Home"

	// utmMedium := query.Get("utm_medium")
	// utmSource := query.Get("utm_source")
	// utmCampaign := query.Get("utm_campaign")
	// inviter := query.Get("inviter")

	// switch {
	// case utmMedium == "whatsapp" && utmSource == "groups":
	// 	return "Whatsapp Group"
	// case utmMedium == "CL" && utmCampaign == "free_demo" && utmSource == "":
	// 	return "Craigslist flyer"
	// case utmMedium == "CL" && utmSource != "":
	// 	return "Craigslist Link"
	// case inviter == "flyer1":
	// 	return "Flyer1 - Pink"
	// case inviter == "flyer2":
	// 	return "Flyer2"
	// default:
	// 	if len(queryParam) != 0 && queryParam != "/" {
	// 		return "tutree.com/" + queryParam
	// 	} else {
	// 		return "tutree.com"
	// 	}
	// }
}

func GetTimeDurationUntilConfirmationAllowed() string {
	return os.Getenv("TIME_DURATION_UNTIL_CONFIRMATION_ALLOWED")
}
func GetCurriculumPDFLink() string {
	return os.Getenv("CURRICULUM_PDF_LINK")
}

// GetZohoCRMRefreshToken retrieves the Zoho CRM refresh token from the environment variables.
// This token is used to generate a new access token for Zoho CRM API calls.
func GetZohoCRMRefreshToken() string {
	return os.Getenv("ZOHO_CRM_REFRESH_TOKEN")
}

// GetZohoCRMClientID retrieves the Zoho CRM Client ID from the environment variables.
// The Client ID is required for authentication with Zoho's OAuth2 API.
func GetZohoCRMClientID() string {
	return os.Getenv("ZOHO_CLIENT_ID")
}

// GetZohoCRMClientSecret retrieves the Zoho CRM Client Secret from the environment variables.
// The Client Secret, along with the Client ID, is used to authenticate with Zoho's OAuth2 API.
func GetZohoCRMClientSecret() string {
	return os.Getenv("ZOHO_CLIENT_SECRET")
}

// GetZohoCRMWebsiteLinkToGenerateAccessToken retrieves the website link to generate the Zoho CRM access token.
// This link is used during the OAuth2 flow to obtain a new access token.
func GetZohoCRMWebsiteLinkToGenerateAccessToken() string {
	return os.Getenv("ZOHO_WEBSITE_LINK_TO_GENERATE_ACCESS_TOKEN")
}

// GetZohoApiToCreateLeadOnZohoCRMLeadBoard retrieves the API endpoint URL for creating leads on Zoho CRM.
// This endpoint allows posting new lead data to Zoho's CRM system.
func GetZohoApiToCreateLeadOnZohoCRMLeadBoard() string {
	return os.Getenv("ZOHO_CRM_WEBSITE_TO_SEND_LEADS")
}

// GetZohoApiToUpdateLeadOnZohoCRMLeadBoard retrieves the API endpoint URL for updating leads on Zoho CRM.
// This endpoint allows updating existing lead data on Zoho's CRM system.
func GetZohoApiToUpdateLeadOnZohoCRMLeadBoard() string {
	return os.Getenv("ZOHO_CRM_WEBSITE_TO_UPDATE_LEADS")
}
