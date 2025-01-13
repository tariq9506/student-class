package email

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/constants"
	"tutree/student-apis/controllers"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/utility"

	"tutree/student-apis/services"

	messageservices "tutree/student-apis/controllers/message_service"

	"github.com/gin-gonic/gin"
)

// test1
// SendReminderEmailForDemoSession handles the process of sending reminder emails for demo sessions.
//  1. Calls the SendReminderEmailForDemoSession function from the students package to retrieve a list of students who need reminder emails for demo sessions.
//  2. Logs an error message and exits if the retrieval fails.
//  3. Iterates over the list of students to send reminder emails individually:
//     a. Creates a map (pData) containing the session start time.
//     b. Prepares the email request data (requestData) with the student's parent email, dynamic data (session start time), and the template ID for the reminder email.
//     c. Calls ComposeDynamicTemplateEmailsForBrevoService from the services package to send the email using the prepared request data.
func SendReminderEmailForSession(c *gin.Context) {
	// Retrieve the list of students needing reminder emails for demo sessions
	student, err := students.SendReminderEmailForSession()
	if err != nil {
		log.Println("SendReminderEmailForSession: Failed to retrieve student information for sending reminder emails/SMS before the session starts in the next 15 minutes.", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	// Iterate over each student in the retrieved list
	for _, students := range student {
		message := fmt.Sprintf("Hey %s!\nYour class with Tutree is scheduled to start in 15 minutes.\n\nJoin here: %s\n\nSee you in the class,\nTeam Tutree", students.StudentName, students.MeetingLink)
		err = services.SendMessage(students.User.Phone, message, students.User.DialingCode)
		if err != nil {
			log.Println("SendReminderEmailForSession: sending session reminder message failed: ", err)
		}
		formatedDate, formatedTime := utility.FormatToReadableDateTime(students.SessionStartTime, students.SessionTimezone)
		// Prepare dynamic data for the email template

		templateID := constants.ReminderEmailOfDemoSession

		if students.SessionType == "paid" {
			templateID = constants.ReminderEmailOfPaidSession
		}
		pData := map[string]interface{}{
			"session_date": formatedDate,
			"session_time": formatedTime,
			"name":         students.StudentName,
			"meeting_link": students.MeetingLink,
		}
		// Prepare the request data for sending the email
		requestData := services.DynamicEmailDetailsForBrevo{
			To:          []string{students.ParentEmail},
			DynamicData: pData,
			TemplateID:  templateID,
		}
		// Send the reminder email using the dynamic template
		services.ComposeDynamicTemplateEmailsForBrevoService(requestData)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully send reminder SMS and Email",
	})
}

// SendReminderEmailForBookRegularSession handles the process of sending reminder emails to students
// who have purchased a subscription plan but have not yet booked a regular session.
//
// The email content is customized with dynamic data, including the plan name and maximum session limit,
// and is sent to the student's email address.
//
// If any error occurs during the process, such as failing to retrieve student information, an appropriate
// JSON error response is sent back to the client.
//
// Parameters:
// - c: *gin.Context - The context for the current HTTP request, used for JSON responses and request handling.
//
// Returns:
// - A JSON response indicating the success of sending the reminder emails or an error message if something went wrong.

func SendReminderEmailForBookRegularSession(c *gin.Context) {
	data, err := students.SendReminderEmailForBookRegularSession()
	if err != nil {
		log.Println("[ERROR] SendReminderEmailForBookRegularSession: Failed to retrieve student information for reminde to book regular session after payment.", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	dashboardLink := fmt.Sprintf("%s/student/dashboard", utility.GetHostURL())
	for _, student := range data {
		pData := map[string]interface{}{
			"plan_name":      student["plan_name"].(string),
			"max_session":    student["max_session"].(int64),
			"dashboard_link": dashboardLink,
		}
		requestData := services.DynamicEmailDetailsForBrevo{
			To:          []string{student["email"].(string)},
			DynamicData: pData,
			TemplateID:  constants.SendReminderEmailForBookRegularSession,
		}
		services.ComposeDynamicTemplateEmailsForBrevoService(requestData)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully send reminder Email",
	})
}

// Helper function to send reminders based on the specified time interval
func sendReminderNotifications(c *gin.Context, reminderBeforeHour int, emailTemplateID int64, smsMessageFormat string) {
	// Retrieve the list of students needing reminder emails for the specified time interval
	studentlist, err := students.SessionReminderNotificationBeforeHour(reminderBeforeHour, "", "")
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve student information for sending reminder emails/SMS before the session starts: %v", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Iterate over each student in the retrieved list
	for _, student := range studentlist {
		rescheduleURL := fmt.Sprintf("%s/student/dashboard/reschedule?sessionId=%d", utility.GetHostURL(), student.SessionID)
		formatedDate, formatedTime := utility.FormatToReadableDateTime(student.SessionStartTime, student.SessionTimezone)

		// Send SMS message
		smsMessage := fmt.Sprintf(smsMessageFormat, student.StudentName, student.SessionType, formatedTime, formatedDate, rescheduleURL)
		err = services.SendMessage(student.User.Phone, smsMessage, student.User.DialingCode)
		if err != nil {
			log.Printf("[ERROR] Sending session reminder message failed: %v", err)
		}

		// Prepare dynamic data for the email template
		emailData := map[string]interface{}{
			"session_type":   student.SessionType,
			"session_date":   formatedDate,
			"session_time":   formatedTime,
			"reschedule_url": rescheduleURL,
		}

		// Prepare the request data for sending the email
		emailRequest := services.DynamicEmailDetailsForBrevo{
			To:          []string{student.ParentEmail},
			DynamicData: emailData,
			TemplateID:  emailTemplateID,
		}

		// Send the reminder email using the dynamic template
		services.ComposeDynamicTemplateEmailsForBrevoService(emailRequest)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully sent reminder SMS and Email",
	})
}

// SessionReminderNotificationBefore2Hour sends reminder emails and SMS messages
// to students for sessions that will start within the next 2 hours.
func SessionReminderNotificationBefore2Hour(c *gin.Context) {
	// Retrieve the reminder hour value from environment variables
	sesssionReminderBeforeStr := os.Getenv("SESSION_REMIND_BEFORE_HOUR")
	if len(sesssionReminderBeforeStr) == 0 {
		log.Println("[NOT FOUND] Failed to retrieve the 'REMINDER_HOUR' value from environment variables.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		return
	}
	// Convert the reminder hour string to an integer
	sessionReminderBefore, err := strconv.Atoi(sesssionReminderBeforeStr)
	if err != nil {
		log.Printf("[ERROR] Failed to convert 'REMINDER_HOUR' to an integer: %v", err)
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, err, c)
		return
	}

	// Define SMS message format and email template ID for 2-hour reminder
	smsMessageFormat := "Hi %s, your %s class starts in 2 hours at %s on %s. Please join on time. \n\nYou can reschedule here: %s."
	emailTemplateID := constants.SessionReminderEmailBeforeTwoHour

	// Use the helper function to send reminders
	sendReminderNotifications(c, sessionReminderBefore, emailTemplateID, smsMessageFormat)
}

// SessionReminderNotificationBefore24Hour sends reminder emails and SMS messages
// to students for sessions that will start within the next 24 hours.
func SessionReminderNotificationBefore24Hour(c *gin.Context) {
	// Retrieve the reminder hour value from environment variables
	sesssionReminderBeforeStr := os.Getenv("SESSION_REMIND_BEFORE_24HOUR")
	if len(sesssionReminderBeforeStr) == 0 {
		log.Println("[NOT FOUND] Failed to retrieve the 'REMINDER_HOUR' value from environment variables.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		return
	}
	// Convert the reminder hour string to an integer
	sessionReminderBefore, err := strconv.Atoi(sesssionReminderBeforeStr)
	if err != nil {
		log.Printf("[ERROR] Failed to convert 'sesssionReminderBeforeStr' to an integer: %v", err)
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, err, c)
		return
	}

	// Define SMS message format and email template ID for 24-hour reminder
	smsMessageFormat := "Hi %s, your %s class starts in 24 hours at %s on %s. Please join on time. \n\nYou can reschedule here: %s."
	emailTemplateID := constants.SessionReminderEmailBefore24Hour

	// Use the helper function to send reminders
	sendReminderNotifications(c, sessionReminderBefore, emailTemplateID, smsMessageFormat)
}
func ScheduleOneDayBeforeReminderForPaidSession(c *gin.Context) {

	var (
		sessionReminderBefore int
		err                   error
	)
	reminderDate := c.PostForm("reminder_date")
	if len(reminderDate) == 0 {
		// log.Println("ScheduleOneDayBeforeReminderForPaidSession: Failed to fetch date and time of student's demo session.")
		// sesssionReminderBeforeStr := os.Getenv("SESSION_REMIND_BEFORE_24HOUR")
		// if len(sesssionReminderBeforeStr) == 0 {
		// 	log.Println("[NOT FOUND] Failed to retrieve the 'REMINDER_HOUR' value from environment variables.")
		// 	controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		// 	return
		// }
		// // Convert the reminder hour string to an integer
		// sessionReminderBefore, err = strconv.Atoi(sesssionReminderBeforeStr)
		// if err != nil {
		// 	log.Printf("[ERROR] ScheduleOneDayBeforeReminderForPaidSession :Failed to convert 'sesssionReminderBeforeStr' to an integer: %v", err)
		// 	controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, err, c)
		// 	return
		// }
		date := time.Now().Add(24 * time.Hour)
		reminderDate = date.Format(utility.DateLayout)
	}

	// Retrieve the list of students needing reminder emails for the specified time interval
	studentlist, err := students.SessionReminderNotificationBeforeHour(sessionReminderBefore, constants.SessionTypePaid, reminderDate)
	if err != nil {
		log.Printf("[ERROR] ScheduleOneDayBeforeReminderForPaidSession : Failed to retrieve student information for sending reminder emails/SMS before the session starts: %v", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	var data []messageservices.DataForScheduleMsgs
	// Iterate over each student in the retrieved list

	for _, student := range studentlist {
		phone := student.User.DialingCode + student.User.Phone
		if !strings.HasPrefix(phone, "+") {
			phone = fmt.Sprintf("+%s", phone)
		}
		data = append(data, messageservices.DataForScheduleMsgs{
			SessionData: messageservices.SessionData{
				ID:          int64(student.SessionID),
				StartTime:   student.SessionStartTime,
				Timezone:    student.SessionTimezone,
				SessionType: student.SessionType,
				MeetingLink: student.Tutor.ZoomLink,
			},
			TeacherData: messageservices.TeacherData{
				ID:       int64(student.Tutor.ID),
				Email:    student.Tutor.TutorEmail,
				Timezone: student.Tutor.Timezone,
				Name:     student.Tutor.TutorName,
			},
			StudentData: messageservices.StudentData{
				ID:          int64(student.User.ID),
				Email:       student.ParentEmail,
				Name:        student.StudentName,
				PhoneNumber: phone,
			},
		})

	}

	go messageservices.ScheduleMsgServices(data)
}

// SendReminderToScheduleClass fetches a list of students
// who need to be reminded to schedule a class and schedules reminder emails for them.
// It responds with a success message upon successful scheduling.
func SendReminderToScheduleClass(c *gin.Context) {
	// Fetch the list of students who need to receive a reminder to schedule a class
	studentList, err := students.GetStudentToSendReminderToSchedule()
	if err != nil {
		log.Println("[ERROR] SendReminderToScheduleClass: Failed to fetch student list to send reminder: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Schedule emails for the fetched list of students
	messageservices.ScheduleEmailForReminderToScheduleClass(studentList)

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Scheduled reminder successfully.",
	})
}
