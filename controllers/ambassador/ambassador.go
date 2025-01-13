package ambassador

import (
	"database/sql"
	"log"
	"net/http"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	"tutree/student-apis/controllers/user"
	"tutree/student-apis/models"
	"tutree/student-apis/models/ambassador"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"
)

// AmbassadorRegistration handles the registration process for ambassador.
// It expects a POST request with a phone number in the form data.
// It validates the phone number, checks if it exists, verifies its deliverability,
// checks if VOIP numbers are allowed, and sends a verification code if all checks pass.
// Upon successful verification code sending, it returns a JSON response with success status,
// a success message, and the ambassador ID associated with the registration.

// AmbassadorRegistration godoc
// @Summary This controller will handles the registration process for ambassador.
// @description The AmbassadorRegistration function handles the process of signing up
// @description students on a website. It expects the ambassador to submit their phone number through a form.
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param phone-number  formData  string true "Phone"
// @Produce json
// @Success 200
// @Router /ambassador/authenticate [post]
func AmbassadorRegistration(c *gin.Context) {
	ambassador := user.RegisterUser(c)
	// All errors are handled in the 'RegisterUser' function.
	// If any error occurs, 'RegisterUser' returns an error in c.JSON and an empty 'User' struct.
	// We check if the struct is empty then return from the function to prevent sending both an error and a success response.
	if ambassador.ID == 0 {
		return
	}
	// Save the user's role as an ambassador asynchronously.
	go user.SaveUserRole(int(ambassador.ID), models.AmbassadorRole)

	myReferalLink := utility.GetHostURL() + "?inviter=" + ambassador.ReferralCode
	c.JSON(http.StatusOK, gin.H{"status": "success",
		"message":          "One time message has been sent to you phone number.",
		"user_id":          ambassador.ID,
		"phone_verified":   ambassador.IsPhoneVerified,
		"email_verified":   ambassador.IsEmailVerified,
		"my_refferal_code": ambassador.ReferralCode,
		"my_refferal_link": myReferalLink,
		"role":             "ambassador",
	})
}

// GetAmbassadorStats godoc
// @Summary This API is used to fetch ambassador statistics.
// @description The GetAmbassadorStats api handles the process of fetching up
// @description the statistics details of an ambassador based on thier user-id.
// @Tags Ambassador
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200
// @Router /ambassador/stats [get]
func GetAmbassadorStats(c *gin.Context) {
	// Fetch user toten from cookies
	token := c.GetHeader("Authorization")
	log.Println("Token ", token)
	if len(token) == 0 {
		log.Println("GetAmbassadorStats: Failed to check user session with error.")
		controllers.HandleJSONErrorResponse(apierror.ErrorOnCheckingSession, nil, c)
		return
	}

	userID, err := models.GetUserIdByToken(token)
	if err != nil {
		log.Println("GetAmbassadorStats: Failed to fetch user details by token with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, err, c)
		return
	}

	ambassadorStats, err := ambassador.GetAmbassadorStats(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("[ERROR] GetAmbassadorStats: Ambassador not found with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, err, c)
			return
		} else {
			log.Println("[ERROR] GetAmbassadorStats: Failed to fetch the ambassador statistics with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"stats":  ambassadorStats,
	})
}
