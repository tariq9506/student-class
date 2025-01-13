package twiliowebhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/constants"
	"tutree/student-apis/controllers"
	"tutree/student-apis/models"
	sessionmodel "tutree/student-apis/models/sessionModel"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"
	"github.com/ttacon/libphonenumber"
)

type TwilioSMSDetails struct {
	SMSBody       string
	Phone         string
	ValidResponse bool
	CountryCode   string
}

type SessionConfirmationDetails struct {
	TwilioSMSDetails    TwilioSMSDetails
	Session             sessionmodel.Session
	User                models.User
	AttendanceConfirmed bool
}

const TriggerTimeFormat = "2006-01-02 15:04:05.999999-07:00"

func parseTwilioSmsDetails(c *gin.Context) (TwilioSMSDetails, error) {
	// Get the message body sent to our Twilio number
	body := c.PostForm("Body")

	// Get the phone number of the sender
	from := c.PostForm("From")

	// Log the received message body
	log.Println("parseTwilioSmsDetails: Received SMS Body:", body)

	// Parse the phone number using libphonenumber
	parsedNumber, err := libphonenumber.Parse(from, "US")
	if err != nil {
		log.Println("parseTwilioSmsDetails: [ERROR] Failed to parse phone number:", err)
		return TwilioSMSDetails{}, err
	}

	// Extract country code and national number from the parsed phone number
	countryCodeInt := parsedNumber.GetCountryCode()
	phoneInt := parsedNumber.GetNationalNumber()

	countryCode := fmt.Sprintf("%v", countryCodeInt)

	// Convert the national number to string format
	phoneNumber := fmt.Sprintf("%v", phoneInt)

	log.Printf("CountryCode: %v, PhoneNumber: %v\n", countryCode, phoneNumber)

	twilioSMSDetails := TwilioSMSDetails{
		SMSBody:     body,
		Phone:       phoneNumber,
		CountryCode: countryCode,
	}
	return twilioSMSDetails, nil
}

func TwilioWebhook(c *gin.Context) {

	twilioSMSDetails, err := parseTwilioSmsDetails(c)
	if err != nil {
		log.Println("TwilioWebhook: [ERROR] Failed to get user details:", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, err, c)
		return
	}

	HandleTwilioResponse(c, twilioSMSDetails)
}

func HandleTwilioResponse(c *gin.Context, twilioSMSDetails TwilioSMSDetails) {

	var err error

	sessionConfirmationDetails := SessionConfirmationDetails{
		TwilioSMSDetails: twilioSMSDetails,
	}

	smsBodyLower := strings.ToLower(twilioSMSDetails.SMSBody)

	if _, validResponse := constants.TwilioSMSSessionConfirmReply[smsBodyLower]; !validResponse {

		if _, validResponse := constants.TwilioSMSSessionRescheduleORCancelReply[smsBodyLower]; !validResponse {
			// If Invalid response comes schedule msg for invalid sms response and return from function call.
			twilioSMSDetails.ValidResponse = validResponse
			go ScheduleMsgsForTwilioResponse(sessionConfirmationDetails)

			log.Println("TwilioWebhook: Invalid Twilio Response")
			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": "response is invalid",
				"body":    twilioSMSDetails.SMSBody,
			})
			return
		}
	}
	twilioSMSDetails.ValidResponse = true

	// Retrieve user details based on the phone number.
	user, err := models.GetUserDetailsWithPhone(twilioSMSDetails.Phone)
	if err != nil {
		log.Println("TwilioWebhook: [ERROR] Failed to get user details:", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, err, c)
		return
	}

	// Get the user's first upcoming session details.
	session, err := sessionmodel.GetNextScheduledSession(user.ID)
	if err != nil {
		log.Println("TwilioWebhook: [ERROR] Failed to get session details:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	sessionConfirmationDetails = SessionConfirmationDetails{
		TwilioSMSDetails: twilioSMSDetails,
		User:             user,
		Session:          session,
	}

	// Handle a positive confirmation response from Twilio
	if exists := constants.TwilioSMSSessionConfirmReply[smsBodyLower]; exists {

		// Update Session confirmation value as true in Database
		err = sessionmodel.MarkSessionConfirmationStatus(int64(session.SessionID), true)
		if err != nil {
			log.Println("TwilioWebhook: [ERROR] Failed to mark session confirmation status as TRUE with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}

		// Cancel Reminder sms before 4 hour from an upcoming session if user confirms his attendance.
		cancelMsgEvents := []services.CancelMsgEvent{}
		cancelMsgEvents = append(cancelMsgEvents, services.CancelMsgEvent{
			SessionID: int64(session.SessionID),
			SMSType:   constants.SessionReminderSMSBeforeFourHour,
		})
		go services.CancelMessageServiceWebhook(cancelMsgEvents)

		// Schedule Twilio response sms for attendance confirmation.
		sessionConfirmationDetails.AttendanceConfirmed = true
		go ScheduleMsgsForTwilioResponse(sessionConfirmationDetails)

		log.Println("TwilioWebhook: SESSION has been marked as CONFIRMED")
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "User has confirmed attendance for the session",
		})
		return
	}

	// // Handle reschedule or cancellation response from Twilio
	// if exists := constants.TwilioSMSSessionRescheduleORCancelReply[smsBodyLower]; exists {

	// 	// If session is paid, save the current session details in the exceptions table
	// 	if session.SessionType == constants.SessionTypePaid {
	// 		err = sessionmodel.SaveExceptionSessionOfStudent(session)
	// 		if err != nil {
	// 			log.Println("[ERROR] TwilioWebhook: Failed to save session details in the database:", err)
	// 			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
	// 			return
	// 		}
	// 	}

	// 	// Delete the current session from the session table
	// 	err = sessionmodel.DeleteSession(session.SessionID)
	// 	if err != nil {
	// 		log.Println("[ERROR] TwilioWebhook: Failed to delete session details from the database:", err)
	// 		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
	// 		return
	// 	}

	// 	sessionConfirmationDetails.AttendanceConfirmed = false
	// 	go RescheduleMsgsForTwilioResponse(sessionConfirmationDetails)

	// 	// Log the successful cancellation and respond
	// 	log.Printf("TwilioWebhook: Session for user %v has been cancelled and added to the Reschedule later bucket\n", user.ID)
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status":  "success",
	// 		"message": "Successfully cancelled the session",
	// 	})
	// 	return
	// }

}

func ScheduleMsgsForTwilioResponse(sessionConfirmationDetails SessionConfirmationDetails) {

	var sendMessageEvents []services.MessageInfo

	if !sessionConfirmationDetails.TwilioSMSDetails.ValidResponse { // Compose sms for invalid response

		sendMessageEvents = ComposeInvalidTwilioResponseSMS(sessionConfirmationDetails.TwilioSMSDetails)

	} else if sessionConfirmationDetails.AttendanceConfirmed { // compose sms for attendance confirmation.

		sendMessageEvents = ComposeAttendanceConfirmTwilioResponseSMS(sessionConfirmationDetails)

	} else { // compose sms for session reschedule.

		sendMessageEvents = ComposeSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails)

	}

	services.CallMessageServiceWebhook(sendMessageEvents)

}

func RescheduleMsgsForTwilioResponse(sessionConfirmationDetails SessionConfirmationDetails) {
	// Cancel Reminder sms for this session if user confirms his attendance.
	cancelMsgEvents := []services.CancelMsgEvent{}
	cancelMsgEvents = append(cancelMsgEvents, services.CancelMsgEvent{
		SessionID: int64(sessionConfirmationDetails.Session.SessionID),
	})
	services.CancelMessageServiceWebhook(cancelMsgEvents)

	var sendMessageEvents []services.MessageInfo

	if !sessionConfirmationDetails.TwilioSMSDetails.ValidResponse { // Compose sms for invalid response

		sendMessageEvents = ComposeInvalidTwilioResponseSMS(sessionConfirmationDetails.TwilioSMSDetails)

	} else if sessionConfirmationDetails.AttendanceConfirmed { // compose sms for attendance confirmation.

		sendMessageEvents = ComposeAttendanceConfirmTwilioResponseSMS(sessionConfirmationDetails)

	} else { // compose sms for session reschedule.

		// Add results to the slice without overwriting
		sendMessageEvents = append(sendMessageEvents, ComposeSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails)...)
		// send email to student on reschedule the session
		sendMessageEvents = append(sendMessageEvents, ComposeStudentEmailForSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails)...)
		// send email to tutor on reschedule the session
		sendMessageEvents = append(sendMessageEvents, ComposeTutorEmailForSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails)...)

	}

	services.CallMessageServiceWebhook(sendMessageEvents)
}

func prepareStudentSMSData(sessionConfirmationDetails SessionConfirmationDetails) []byte {

	formatedDate, formatedTime := utility.FormatToReadableDateTime(sessionConfirmationDetails.Session.StartSession, sessionConfirmationDetails.Session.TimeZone)

	content := map[string]interface{}{
		"{student_name}":   sessionConfirmationDetails.Session.Student.StudentName,
		"{date}":           formatedDate,
		"{time}":           formatedTime,
		"{reschedule_url}": fmt.Sprintf("%s/student/dashboard/reschedule?sessionId=%d", utility.GetHostURL(), sessionConfirmationDetails.Session.SessionID),
	}

	smsContent, err := json.Marshal(content)
	if err != nil {
		log.Println("composeReminderSMS: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return smsContent

}

func ComposeTwilioResponseSMS(phoneNumber string) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	message := services.MessageInfo{
		MessageType: constants.SmsMessageType,
		EventType:   constants.DefaultEventType,
		TriggerTime: triggerTime,
		Status:      constants.PendingMessageStatus,
		Interval:    constants.DefaultInterval,
		PhoneNumber: phoneNumber,
	}

	return message
}

func ComposeInvalidTwilioResponseSMS(twilioSMSDetails TwilioSMSDetails) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo

	phone := twilioSMSDetails.CountryCode + twilioSMSDetails.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}

	message := ComposeTwilioResponseSMS(phone)
	message.SMSType = constants.TwilioSMSInvalidResponse

	sendMessageEvents = append(sendMessageEvents, message)

	return sendMessageEvents
}

func ComposeAttendanceConfirmTwilioResponseSMS(sessionConfirmationDetails SessionConfirmationDetails) []services.MessageInfo {
	var sendMessageEvents []services.MessageInfo

	phone := sessionConfirmationDetails.TwilioSMSDetails.CountryCode + sessionConfirmationDetails.TwilioSMSDetails.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}

	pDataStudentSMS := prepareStudentSMSData(sessionConfirmationDetails)

	message := ComposeTwilioResponseSMS(phone)
	message.SMSType = constants.TwilioSMSAttendanceConfirmed
	message.SMSContent = pDataStudentSMS
	message.UserID = sessionConfirmationDetails.User.ID
	message.SessionID = int64(sessionConfirmationDetails.Session.SessionID)

	sendMessageEvents = append(sendMessageEvents, message)

	return sendMessageEvents
}

func ComposeSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails SessionConfirmationDetails) []services.MessageInfo {
	var sendMessageEvents []services.MessageInfo

	phone := sessionConfirmationDetails.TwilioSMSDetails.CountryCode + sessionConfirmationDetails.TwilioSMSDetails.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}

	pDataStudentSMS := prepareStudentSMSData(sessionConfirmationDetails)

	message := ComposeTwilioResponseSMS(phone)
	message.SMSType = constants.TwilioSMSSessionReschedule
	message.SMSContent = pDataStudentSMS
	message.UserID = sessionConfirmationDetails.User.ID
	message.SessionID = int64(sessionConfirmationDetails.Session.SessionID)

	sendMessageEvents = append(sendMessageEvents, message)

	return sendMessageEvents
}
func prepareStudentEmailData(sessionConfirmationDetails SessionConfirmationDetails) []byte {
	// Generate the link to the student's dashboard
	pendingLink := utility.GetHostURL() + "/student/dashboard#classes"

	// Format session start date and time to a readable format
	formatedDate, formatedTime := utility.FormatToReadableDateTime(sessionConfirmationDetails.Session.StartSession, sessionConfirmationDetails.Session.TimeZone)

	// Create a map containing the dynamic data for the email
	content := map[string]interface{}{
		"kid_name":     sessionConfirmationDetails.Session.Student.StudentName,
		"session_type": sessionConfirmationDetails.Session.SessionType,
		"session_date": formatedDate,
		"session_time": formatedTime,
		"pending_link": pendingLink,
	}

	// Marshal the map into a JSON byte array
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareDataForSessionReschedule: Failed to marshal dynamic data into JSON with error:", err)
	}

	return emailContent

}
func ComposeDataForSessionRescheduleForStudent(sessionData SessionConfirmationDetails) services.MessageInfo {
	// Get the current UTC time to set the email trigger time
	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format(TriggerTimeFormat)

	// Create a MessageInfo object with all the necessary fields
	message := services.MessageInfo{
		UserID:       sessionData.User.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailContent: prepareStudentEmailData(sessionData),
		EmailType:    constants.SessionRescheduleEmail,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        sessionData.Session.Student.ParentEmail,
	}

	return message
}
func ComposeStudentEmailForSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails SessionConfirmationDetails) []services.MessageInfo {
	var sendMessageEvents []services.MessageInfo
	message := ComposeDataForSessionRescheduleForStudent(sessionConfirmationDetails)
	sendMessageEvents = append(sendMessageEvents, message)

	return sendMessageEvents
}
func prepareTutorEmailData(sessionConfirmationDetails SessionConfirmationDetails) []byte {
	// Generate the link to the student's dashboard
	tutorDashboardLink := utility.GetTutorHostURL() + "/dashboard"

	// Format session start date and time to a readable format
	formatedDate, formatedTime := utility.FormatToReadableDateTime(sessionConfirmationDetails.Session.StartSession, sessionConfirmationDetails.Session.Tutor.Timezone)

	// Create a map containing the dynamic data for the email
	content := map[string]interface{}{
		"tutor_name":   sessionConfirmationDetails.Session.Tutor.TutorName,
		"session_type": sessionConfirmationDetails.Session.SessionType,
		"session_date": formatedDate,
		"session_time": formatedTime,
		"view_details": tutorDashboardLink,
	}

	// Marshal the map into a JSON byte array
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareDataForSessionReschedule: Failed to marshal dynamic data into JSON with error:", err)
	}

	return emailContent

}
func ComposeDataForSessionRescheduleForTutor(sessionData SessionConfirmationDetails) services.MessageInfo {
	// Get the current UTC time to set the email trigger time
	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format(TriggerTimeFormat)

	// Create a MessageInfo object with all the necessary fields
	message := services.MessageInfo{
		UserID:       sessionData.User.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailContent: prepareTutorEmailData(sessionData),
		EmailType:    constants.SessionCancelEmailForTutor,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        sessionData.Session.Tutor.TutorEmail,
	}

	return message
}
func ComposeTutorEmailForSessionRescheduleTwilioResponseSMS(sessionConfirmationDetails SessionConfirmationDetails) []services.MessageInfo {
	var sendMessageEvents []services.MessageInfo
	message := ComposeDataForSessionRescheduleForTutor(sessionConfirmationDetails)
	sendMessageEvents = append(sendMessageEvents, message)

	return sendMessageEvents
}

func ComposeSessionRescheduleIfNoResponse(sessionConfirmationDetails SessionConfirmationDetails) []services.MessageInfo {
	var sendMessageEvents []services.MessageInfo

	pDataStudentSMS := prepareStudentSMSData(sessionConfirmationDetails)

	message := ComposeTwilioResponseSMS(sessionConfirmationDetails.TwilioSMSDetails.Phone)
	message.SMSType = constants.TwilioSMSSessionRescheduleIfNoResponse
	message.SMSContent = pDataStudentSMS
	message.UserID = sessionConfirmationDetails.User.ID
	message.SessionID = int64(sessionConfirmationDetails.Session.SessionID)

	sendMessageEvents = append(sendMessageEvents, message)

	return sendMessageEvents
}
