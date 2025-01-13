package phonecontroller

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
	"tutree/student-apis/controllers/jwt"
	"tutree/student-apis/models"

	"tutree/student-apis/models/phone"

	"tutree/student-apis/services"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"
)

// SendPhoneNumberVerificationCode generates an OTP, saves it to the database, and sends it via SMS to the specified phone number.
// Parameters:
// - phoneNo: The phone number to which the OTP will be sent.
// - dialingCode: The international dialing code for the phone number.
// - userIP: Information of user's IP.
// - location: Location contain user's location information.
// Returns:
// - error: Any error encountered during the process.
// - student struct: The student's details associated with the phone number, or 0 if an error occurred.
func SendPhoneNumberVerificationCode(user models.User) error {

	var message string
	if user.IsPhoneVerified {
		message = user.OTP + " is the verification code to log in to your Tutree account. Please DO NOT SHARE this code with anyone.\n@" + os.Getenv("DOMAIN_NAME") + " #" + user.OTP

	} else {
		// content for the otp message
		message = user.OTP + " is the verification code to sign up to your Tutree account. Please DO NOT SHARE this code with anyone.\n@" + os.Getenv("DOMAIN_NAME") + " #" + user.OTP
	}

	log.Println("OTPmessage", message)
	// This is the twilio service to send the otp to the given phone number.
	// function use to send message given phone having message and otp
	err := services.SendMessage(user.Phone, message, user.DialingCode)
	if err != nil {
		log.Println("SendPhoneNumberVerificationCode: sending otp failed: ", err)
		return err
	}

	return nil
}

// IsPhNumberDeliverable
// input : phone number
// ouput : true/ false, ie, phone number deliverable or not
// This controller will func to do the phone number lookup
func IsPhNumberDeliverable(phone, countrycode string) bool {

	// func to check the phone number deliverable or not
	isDeliverable, err := services.IsPhNumberDeliverable(phone, countrycode)
	if err != nil {
		log.Println("IsPhNumberDeliverable : Phone number lookup failed with error : ", err)
		return false
	}
	return isDeliverable
}

// IsPhoneNumberVoip
// input : phone number, countryCode
// ouput : true/ false, ie, phone number voip or not
// This controller will func to do the phone number lookup
func IsPhoneNumberVoip(phone, countryCode string) bool {
	// func to check the phone number voip or not
	isVoip, err := services.IsPhoneNumberVoip(phone, countryCode)
	if err != nil {
		log.Println("IsPhoneNumberVoip : Phone number lookup failed with error : ", err)
		return true
	}

	return isVoip
}

// VerifyCode
// input : OTP code from query parameter, studentId from token string
// Output: success/failed
// Desc  : This controller will verify the given code same as OTP. It will also check if OTP is expired.

// VerifyCode godoc
// @Summary This controller will verify the given code same as OTP. It will also check if OTP is expired
// @description This controller will verify the given code same as OTP. It will also check if OTP is expired.
// @description This api is taking code and studentId as postform
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param code  formData  string true "Code"
// @Param user-id  formData  string true "user-id"
// @Param User-Agent header string true "User Agent"
// @Produce json
// @Success 200
// @Router /otp/verify [POST]
func VerifyCode(c *gin.Context) {

	code := c.PostForm("code")
	if len(code) == 0 {
		log.Println("VerifyCode: Verification Code Required, Please enter the OTP (One-Time Password) to proceed.")
		controllers.HandleJSONErrorResponse(apierror.VerifyCode_BlankOTP, nil, c)
		return

	}
	log.Println("CODE ENTER BY STUDENT: ", code)

	userID, err := strconv.Atoi(c.PostForm("user-id"))
	if err != nil {
		log.Println("Verification Code Error: Failed to convert student ID into a valid number.", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidStudentID, err, c)
		return
	}

	// function use to fetch student details by student id
	user, err := models.GetUserByID(userID)
	if err != nil {
		log.Println("VerifyCode: failed to fetch student data :", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, err, c)
		return
	}
	OTP, err := phone.GetValidVerificationCode(int64(userID))
	if err != nil {
		log.Println("VerifyCode: Error occurred while fetching OTP or checking if it is expired for user ID:", err)
		controllers.HandleJSONErrorResponse(apierror.VerifyCode_OTPExpired, err, c)
		return
	}
	// currentEnvironment represents the current operational environment of the server, indicating whether it is running
	// locally, in a testing environment, or in production.
	currentEnvironment := os.Getenv("ENV")
	// for local, testing environment this line of code would accept "1234" as valid OTP.
	if code == OTP || (strings.Contains("local, test, beta", strings.ToLower(currentEnvironment)) && code == "1234") {
		err := phone.SetPhoneVerified(int64(userID))
		if err != nil {
			log.Println("VerifyCode: failed to verify Phone number:", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
		// Send notification on slack whenever new student verifiy their phone-number
		if !user.IsPhoneVerified && user.Location.CountryCode == "US" {
			channelID := os.Getenv("VERIFIED_SIGNUP_CHANNEL")
			if len(channelID) == 0 {
				log.Println("[ERROR] Failed to get VERIFIED_SIGNUP_CHANNEL")
			} else {
				source := utility.GetSourceOfSignup(user.QueryParams)
				location := user.Location.City + ", " + user.Location.State + ", " + user.Location.CountryCode
				message := fmt.Sprintf("*New Student Verified Signup* \nPhone: %s \nLocation: %s \nSource: %s \nPhone Verification: YES",
					user.Phone, location, source)
				go services.SendSlackMessage(channelID, message)
			}
		}

		token := jwt.CreateUserAuth(c, user)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Sucessfully verified phone number.",
			"token":   token,
			"user_id": user.ID,
		})
		return
	} else {
		//handleJSONErrorResponse(apierror.VerifyCode_OTPMatched, nil, c)
		log.Println("Verification Failed: The OTP entered is incorrect. Please retry with the correct OTP.")
		controllers.HandleJSONErrorResponse(apierror.VerifyCode_OTPNotMatch, nil, c)
		return

	}
}

// ResendVerificationCode handles the process of resending a verification code to a student's phone number.
// It validates the phone number, checks deliverability and VOIP status, retrieves the student's location,
// and sends the verification code if all conditions are met. Responses are provided based on success or failure.

// ResendVerificationCode godoc
// @Summary This controller will resend the given code same as OTP.
// @description This controller will resend the OTP.
// @description This api is taking phone number as postform and user ip from header.
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param phone-number  formData  string true "Phone Number"
// @Produce json
// @Success 200
// @Router /otp/send [POST]
func ResendVerificationCode(c *gin.Context) {

	userIP := utility.GetClientIP(c)
	phoneNumber := c.PostForm("phone-number")
	if len(phoneNumber) == 0 {
		log.Println("ResendVerificationCode: failed, phone nnumber can not be empty.")
		controllers.HandleJSONErrorResponse(apierror.MissingPhoneNumber, nil, c)
		return
	}
	match, err := regexp.MatchString("^[0-9]{10}$", phoneNumber)
	if !match || err != nil {
		log.Println("ResendVerificationCode: Failed, invalid phone number with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidPhoneNumber, err, c)
		return
	}
	location := controllers.GetLocationFromIP(userIP)
	details, err := models.GetDetailsOfSupportedCountryByCode(location.CountryCode)
	if err != nil {
		log.Println("ResendVerificationCode: GetDetailsOfSupportedCountryByCode failed to get location iformation with error: ", err)
	}
	otp := utility.GenerateOTP()

	user := models.User{
		Phone:       phoneNumber,
		DialingCode: details.CountryPhoneCode,
		UserIP:      userIP,
		Location:    location,
		OTP:         otp,
	}
	_, err = phone.SaveOTP(user)
	if err != nil {
		log.Println("ResendVerificationCode Failed: Unable to send verification code. Please try again later", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	// func to send the verification code to the student's phone number
	err = SendPhoneNumberVerificationCode(user)
	if err != nil {
		log.Println("ResendVerificationCode Failed: Unable to send verification code. Please try again later", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToSendOTP, err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"mesage": "One time message has been sent to you phone number",
	})
}
