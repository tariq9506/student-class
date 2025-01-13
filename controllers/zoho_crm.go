package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"tutree/student-apis/models"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"
)

// ConvertUserStructToMap converts a User struct into a map[string]interface{} for flexible data handling.
//
// The function performs the following:
// 1. Initializes an empty map to store user details.
// 2. Sets default values like "null" for "First_Name" and "Last_Name" to ensure keys are always present.
// 3. Maps fields like ID, Phone, and QueryParams (lead source) to corresponding keys.
// 4. Embeds a nested "Location" map containing City and Country details.
//
// This conversion is useful for dynamically manipulating user data, especially for APIs or logging systems.

func convertUserStructToMap(user models.User) map[string]interface{} {
	userMap := make(map[string]interface{})
	environment := "prod"
	host := utility.GetHostURL()
	if strings.Contains(strings.ToLower(host), "test") {
		environment = "test"
	}
	userMap["environment"] = environment
	userMap["Id"] = user.ID
	userMap["First_Name"] = "null"
	userMap["Last_Name"] = "null"
	phone := user.DialingCode + user.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}
	userMap["Phone"] = phone
	userMap["Location"] = map[string]interface{}{
		"City":    user.Location.City,
		"Country": user.Location.Country,
	}
	source := utility.GetSourceOfSignup(user.QueryParams)
	userMap["query_param"] = source
	userMap["created_at"] = time.Now().Format("2006-01-02")

	return userMap
}

// ConvertStudentStructToMap converts a ZohoCRMLeads struct into a map[string]interface{}.
//
// This function dynamically processes the input struct and only inserts non-empty fields into the resulting map.
// Key functionalities include:
// 1. Phone number and email fields are validated and added if non-empty.
// 2. CreatedDate is checked for validity using IsZero() before insertion.
// 3. Name processing splits the full name into First_Name and Last_Name.
// 4. Boolean fields like Plane_Status and Demo_Booked are added only if true.
// This function ensures clean and efficient data transformation for CRM-related operations.

func convertLeadDataStructToMap(leads models.ZohoCRMLeads) map[string]interface{} {
	studentMap := make(map[string]interface{})

	// Only insert the non-empty fields into the map
	if strings.TrimSpace(leads.PhoneNumber) != "" {
		studentMap["Phone"] = leads.PhoneNumber
	}
	if leads.PhoneNumberVerified {
		studentMap["phone_verified"] = leads.PhoneNumberVerified
	}
	// Name fields processing
	name := strings.TrimSpace(leads.Name)
	nameParts := strings.Fields(name)

	if len(nameParts) > 1 {
		studentMap["First_Name"] = strings.Join(nameParts[:len(nameParts)-1], " ")
		studentMap["Last_Name"] = nameParts[len(nameParts)-1]
	} else if len(nameParts) == 1 {
		studentMap["First_Name"] = nameParts[0]
		studentMap["Last_Name"] = ""
	}
	if strings.TrimSpace(leads.EmailAddres) != "" {
		studentMap["Email"] = leads.EmailAddres
	}
	if strings.TrimSpace(leads.PlanName) != "" {
		studentMap["plan_name"] = leads.PlanName
	}

	// Boolean fields (add only if true)
	if leads.PlaneStatus {
		studentMap["plan_status"] = leads.PlaneStatus
	}
	if leads.IsDemoBooked {
		studentMap["Demo_Booked"] = leads.IsDemoBooked
	}

	return studentMap
}

// CreateLeadsOnZohoCRM sends user data to Zoho CRM and saves the generated lead ID.
//
// Steps performed:
// 1. Generates an access token for Zoho CRM using CreateZohoCRMAccessToken.
// 2. Converts the User struct into a map using ConvertUserStructToMap for posting.
// 3. Posts the user data to Zoho CRM using PostDataToZohoCRM and retrieves the lead ID.
// 4. Saves the lead ID locally using SaveCRMLeadsID for reference.
// Handles errors at each step with appropriate logs to ensure smooth debugging.

func CreateLeadsOnZohoCRM(user models.User) error {
	// Step 1: Generate Zoho CRM access token
	acessToken, err := services.CreateZohoCRMAccessToken()
	if err != nil {
		log.Println("[ERROR] CreateLeadsOnZohoCRM : Failed to get Zoho access token with error :", err)
		return err
	}
	log.Println("acessToken", acessToken)
	return nil
	var leadID int
	// Step 2: Convert user details to map format for Zoho CRM
	userMap := convertUserStructToMap(user)
	// Step 3: Create user lead data on Zoho CRM
	leadID, err = services.PostDataToZohoCRM(acessToken, userMap)

	if err != nil {
		log.Println("[ZOHO-ERROR] CreateLeadsOnZohoCRM -> PostDataToZohoCRM: Failed to post the user's data on zohoCRM with", err)
		return err
	}
	// Step 4: Save user lead data on Zoho CRM table in database
	err = models.SaveCRMLeadsID(int(user.ID), int(leadID))
	if err != nil {
		log.Println("[ZOHO-ERROR] CreateLeadsOnZohoCRM -> SaveCRMLeadsID: Failed to save the user's lead id with error :", err)
		return err
	}
	return nil
}

// UpdateLeadsOnZohoCRM updates a user's lead information on Zoho CRM.
//
// Steps performed:
// 1. Generates an access token for Zoho CRM using CreateZohoCRMAccessToken.
// 2. Retrieves user details from the database using GetUserDetailsForCRMLeads.
// 3. Converts the user data into a map format using ConvertStudentStructToMap.
// 4. Fetches the Zoho lead ID associated with the user via GetZohoCRMLeadIDByUserID.
// 5. Updates the user's lead data on Zoho CRM using UpdateLeadsDataOnCRM.
// Logs errors at each stage and ensures smooth error handling for traceability.

func UpdateLeadsOnZohoCRM(userID int) error {
	// Step 1: Generate Zoho CRM access token
	acessToken, err := services.CreateZohoCRMAccessToken()
	if err != nil {
		log.Println("[ERROR] UpdateLeadsOnZohoCRM : Failed to get Zoho access token with error :", err)
		return err
	}
	// Step 2: Retrieve user details for CRM leads from the database
	leadsData, err := models.GetUserDetailsForCRMLeads(userID)
	if err != nil {
		log.Println(" UpdateLeadsOnZohoCRM : Failed to get Zoho access token with error :", err)
		return err
	}
	// Step 3: Convert user details to map format for Zoho CRM
	userMap := convertLeadDataStructToMap(leadsData)
	var zohoLeadId int
	if len(acessToken) != 0 {
		// Step 4: Fetch the Zoho CRM lead ID for the given user
		zohoLeadId, err = models.GetZohoCRMLeadIDByUserID(userID)
		if err != nil {
			log.Println("[ZOHO-ERROR]  UpdateLeadsOnZohoCRM: Failed to fetch the user's zoho lead id on zohoCRM with", err)
			return err
		}
	}
	// Step 5: Update user lead data on Zoho CRM
	err = services.UpdateLeadsDataOnCRM(acessToken, zohoLeadId, userMap)
	if err != nil {
		log.Println("UpdateLeadsOnZohoCRM: Failed to post the user's data on zohoCRM with", err)
		return err
	}
	return nil
}
func GetPhoneOfColdLeads(c *gin.Context) {
	// Step 1: Generate Zoho CRM access token
	acessToken, err := services.CreateZohoCRMAccessToken()
	if err != nil {
		log.Println("[ERROR] UpdateLeadsOnZohoCRM : Failed to get Zoho access token with error :", err)
		return
	}
	// Get cold leads phone numbers from the database
	// Step 5: Update user lead data on Zoho CRM
	leads, err := services.GetPhoneOfColdLeads(acessToken)
	if err != nil {
		log.Println("UpdateLeadsOnZohoCRM: Failed to post the user's data on zohoCRM with", err)
		return
	}
	// Step 2: Iterate over each lead and initiate the call
	// for _, lead := range leads {
	// 	phoneNumber := lead.Phone // Assuming the lead struct has PhoneNumber field

	// 	// Make the call to the current lead's phone number
	// 	statusCode, err := services.MakeACallOfZohoColdLeads(phoneNumber)
	// 	if err != nil {
	// 		// Log error in case of failure
	// 		log.Printf("Failed to initiate call to %s: %v\n", phoneNumber, err)
	// 	} else {
	// 		// Handle the response based on the status code
	// 		if statusCode == http.StatusOK {
	// 			log.Printf("Successfully initiated the call to %s\n", phoneNumber)
	// 		} else {
	// 			log.Printf("Call initiation failed for %s with status code: %d\n", phoneNumber, statusCode)
	// 		}
	// 	}
	// }
	var phoneNumber []string
	// phoneNumber = append(phoneNumber, "+16508808950")
	phoneNumber = append(phoneNumber, "+919506193753")
	statusCode := 0
	for _, phone := range phoneNumber {
		statusCode, err = services.MakeACallOfZohoColdLeads(phone)
		if err != nil {
			// Log error in case of failure
			log.Printf("Failed to initiate call to %s: %v\n", phone, err)
		} else {
			// Handle the response based on the status code
			if statusCode == http.StatusOK {
				log.Printf("Successfully initiated the call to %s\n", phone)
			} else {
				log.Printf("Call initiation failed for %s with status code: %d\n", phone, statusCode)
			}
		}
	}
	//statusCode, err := services.MakeACallOfZohoColdLeads(phoneNumber)

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"leads":      leads,
		"statusCode": statusCode,
	})
}
