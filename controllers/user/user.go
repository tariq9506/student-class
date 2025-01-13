package user

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	phonecontroller "tutree/student-apis/controllers/phoneController"
	"tutree/student-apis/models"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"

	messageservices "tutree/student-apis/controllers/message_service"

	"github.com/gin-gonic/gin"
)

func RegisterUser(c *gin.Context) models.User {

	userIP := utility.GetClientIP(c)
	//userIP = "138.199.35.199"
	queryParams := c.PostForm("query-params")
	phoneNumber := c.PostForm("phone-number")
	// Check if the phone number is empty.
	if len(phoneNumber) == 0 {
		log.Println("Registration Failed: Phone number not found. Please enter a valid phone number.")
		controllers.HandleJSONErrorResponse(apierror.MissingPhoneNumber, nil, c)
		return models.User{}
	}
	// Use a regular expression to validate the phone number format (exactly 10 digits).
	match, err := regexp.MatchString("^[0-9]{10}$", phoneNumber)
	if !match || err != nil {
		log.Println("RegisterUser: Failed, invalid phone number with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidPhoneNumber, err, c)
		return models.User{}
	}

	referralCode := strings.TrimSpace(c.PostForm("referral-code"))
	if len(referralCode) == 0 {
		log.Println("RegisterUser: referral code not found")
	}
	// this key is insure that user wants to download curriculum pdf or not.
	curriculumDownloadStr := c.PostForm("curriculum_download")
	curriculumDownload := false
	if len(curriculumDownloadStr) != 0 {
		curriculumDownload, err = strconv.ParseBool(curriculumDownloadStr)
		if err != nil {
			log.Println("[ERROR] Registration : Failed to parse in bool to check curriculum download or not , with error :", err)
			controllers.HandleJSONErrorResponse(apierror.ErrorNoDataFound, nil, c)
			return models.User{}
		}
	}
	location := controllers.GetLocationFromIP(userIP)
	details, err := models.GetDetailsOfSupportedCountryByCode(location.CountryCode)
	if err != nil {
		log.Println("RegisterUser: GetDetailsOfSupportedCountryByCode failed to get location information with error: ", err)
	}

	myReferalCode := controllers.RandStringForAgentID(phoneNumber)
	otp := utility.GenerateOTP()
	user := models.User{
		Phone:        phoneNumber,
		DialingCode:  details.CountryPhoneCode,
		UserIP:       userIP,
		ReferralCode: myReferalCode,
		Location:     location,
		OTP:          otp,
		QueryParams:  queryParams,
		Timezone:     location.TimeZone,
	}

	currentEnvironment := os.Getenv("ENV")
	requestFrom := c.Request.URL.Path
	if strings.Contains(strings.ToLower(currentEnvironment), "prod") && (strings.Contains(strings.ToLower(requestFrom), "/v1")) {
		// This func will check is tht given phone number deliverable or not, if not deliverable will return an error message
		isDeliverable := phonecontroller.IsPhNumberDeliverable(phoneNumber, details.CountryCode)
		if !isDeliverable {
			log.Println("RegisterUser: failed phone number lookup with flag :", isDeliverable)
			c.JSON(http.StatusOK, gin.H{
				"status":  "Failed",
				"message": "Please enter a valid number",
			})
			return models.User{}
		}
		// This func will check is tht given phone number voip or not, if not deliverable will return an error message
		isVoipNumberAllowed := utility.AllowVoipNumbers()
		if !isVoipNumberAllowed {
			isVoip := phonecontroller.IsPhoneNumberVoip(phoneNumber, details.CountryCode)
			if isVoip {
				log.Println("RegisterUser: failed while checking phone number is voip :", isVoip)
				c.JSON(http.StatusOK, gin.H{
					"status":  "Failed",
					"message": "Please enter a valid number",
				})
				return models.User{}
			}
		}

		// func to send the verification code to the student's phone number
		err = phonecontroller.SendPhoneNumberVerificationCode(user)
		if err != nil {
			log.Println("Registration Failed: Unable to send verification code. Please try again later", err)
			//controllers.HandleJSONErrorResponse(apierror.FailedToSendOTP, err, c)
			//return models.User{}
		}
	}

	action, _ := models.SaveNewUser(&user)
	// Check the action type to determine the user's activity:
	// - 'insert': User has signed up.
	// - 'update': User has signed in.
	//
	// If `curriculumDownload` is true, send an SMS with the curriculum download link to the user's number,
	// regardless of whether the user signed up or signed in.
	if action == "insert" || curriculumDownload {
		phone := user.DialingCode + user.Phone
		if !strings.HasPrefix(phone, "+") {
			phone = fmt.Sprintf("+%s", phone)
		}
		studentData := students.StudentToScheduleReminder{
			ID:    user.ID,
			Phone: phone,
		}

		go messageservices.SendCurriculumViaSMSOnNewSignUP(studentData)

	}
	if len(referralCode) > 0 {
		go saveInviter(referralCode, user)
	}
	// lead list on zoho CRM
	if action == "insert" {
		err = controllers.CreateLeadsOnZohoCRM(user)
		if err != nil {
			log.Println("Registration Failed: Failed to create record on zoho CRM lead board with error :", err)

		}
	}
	return user
}
func saveInviter(referralCode string, user models.User) {

	inviterId, err := models.GetUserByReferralCode(referralCode)
	if err != nil {
		log.Println("AmbassadorRegistration: SaveInviterDetails: Failed to get inviter details with error: ", err)
	}
	if inviterId > 0 {
		err = models.SaveInviterDetails(int(user.ID), inviterId)
		if err != nil {
			log.Println("AmbassadorRegistration: SaveInviterDetails: Failed to save inviter details with error: ", err)
		}
	}
}

func SaveUserRole(userId int, role string) {
	roleID, err := models.GetRoleID(role)
	if err != nil {
		log.Println("saveUserRole: Failed to get role id with error: ", err)
	}

	if roleID > 0 {
		err = models.SaveUserRole(userId, roleID)
		if err != nil {
			log.Println("saveUserRole: Failed to save user role with error: ", err)
		}
	}
}

// DeactivateUserSession handles the user logout process by invalidating the user's active session.
// Parameters:
// - c *gin.Context: The context for the current HTTP request.
//
// JSON Response:
// - On success: 200
// - On failure: An appropriate error message indicating the reason for failure.

// DeactivateUserSession godoc
// @Summary User logout process by invalidating the user's active session
// @description This fuctionality is used for user logout action.
// @Tags user
// @Accept application/x-www-from-urlencoded
// @Params Authorization header string true "Authorization token (Bearer token)"
// @Produce json
// @Success 200
// @Router /logout [DELETE]
func DeactivateUserSession(c *gin.Context) {
	userID, exist := c.Get("userID")
	token := c.GetHeader("Authorization")
	if !exist {
		log.Println("DeactivateUserSession: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	err := models.DeactivateUserSession(userID.(int), token)
	if err != nil {
		log.Println("DeactivateUserSession: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, err, c)
		return
	}
	services.RemoveCookies(c, "token")
	c.JSON(http.StatusOK, gin.H{
		"message": "Session expired successfully.",
	})
}
