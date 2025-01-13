package session

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/common"
	"tutree/student-apis/controllers"
	messageservices "tutree/student-apis/controllers/message_service"
	studentscontroller "tutree/student-apis/controllers/studentsController"
	twiliowebhook "tutree/student-apis/controllers/twilio_webhook"
	"tutree/student-apis/models/metadata"
	sessionmodel "tutree/student-apis/models/sessionModel"
	"tutree/student-apis/models/stripe"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/models/tutor"
	"tutree/student-apis/utility"

	"tutree/student-apis/constants"
	"tutree/student-apis/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetRequestDataForSessionBooking(c *gin.Context) (sessionmodel.RequestDataForDemoSessionBooking, error) {
	var requestData sessionmodel.RequestDataForDemoSessionBooking
	session := c.PostForm("session_datetime")
	if len(session) == 0 {
		log.Println("GetRequestDataForSessionBooking: Failed to fetch date and time of student's demo session.")
		controllers.HandleJSONErrorResponse(apierror.SessionDateTimeNotFound, nil, c)
		return requestData, nil
	}
	startSession, err := utility.ConvertToUTCTime(session)
	if err != nil {
		log.Println("[ERROR] GetRequestDataForSessionBooking: Failed to parse session time with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.FailedWhileParsingDateTime, err, c)
		return requestData, err
	}
	requestData.SessionDateTime = startSession
	tutorID, err := strconv.Atoi(c.PostForm("tutor_id"))
	if err != nil {
		log.Println("[ERROR] GetRequestDataForSessionBooking: Failed to book demo session due to invalid tutor ID with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToParseTutorIDIntoInteger, err, c)
		return requestData, err
	}
	if tutorID <= 0 {
		log.Println("GetRequestDataForSessionBooking: Failed to book demo session due to invalid tutor ID, tutor should be positive integer.")
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
		return requestData, nil
	}
	requestData.TutorID = tutorID
	timezone := c.PostForm("timezone")
	if len(timezone) == 0 {
		log.Println("GetRequestDataForSessionBooking: Failed to fetch timezone of student's demo session.")
		controllers.HandleJSONErrorResponse(apierror.SessionTimezoneNotFound, nil, c)
		return requestData, nil
	}
	requestData.TimeZone = timezone
	subjectIDStr := c.PostForm("subject_id")
	if len(subjectIDStr) != 0 {
		subjectID, err := strconv.Atoi(subjectIDStr)
		if err != nil {
			log.Println("[ERROR] GetRequestDataForSessionBooking: Failed to book demo session due to invalid subject ID with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, err, c)
			return requestData, err
		}
		if subjectID <= 0 {
			log.Println("GetRequestDataForSessionBooking: Failed to book demo session due to invalid subject ID, subject id should be positive integer.")
			controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
			return requestData, nil
		}
		requestData.SubjectID = subjectID
	} else {
		subjectsList, err := metadata.GetSubjects()
		if err != nil {
			log.Println("[ERROR] GetSubbjectId : Failed while getting subject id from databbase with error :", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return requestData, err
		}
		requestData.SubjectID = int(subjectsList[0].ID)

	}
	return requestData, nil

}

// BookDemoSession handles the booking of a demo session for a student.
// It retrieves the student's profile from the token, validates input fields,
// checks for available tutors, and saves the session information to the database.
//
// Input Parameters:
// - c: *gin.Context - The context object from the Gin framework containing
//   request data, response methods, and other contextual information.
//
// The function extracts the following form parameters from the request:
// - "session-datetime": A string representing the date and time for the demo session.
// - "student-name": The name of the student booking the demo session.
// - "grade": The grade of the student.
// - "parent-email": The email address of the student's parent.
//
// Outputs:
// - A JSON response with the following fields on success:
//   - "status": "Success" - Indicates that the booking was successful.
//   - "message": "Your free session is scheduled." - Confirmation message for the student.
//   - "session_id": The unique identifier for the newly created demo session.
//

// BookDemoSession godoc
// @Summary This controller funnction handles the booking of a demo session for a student
// @description This controller will book the demo session for student.
// @description This api is taking student name, parent's email, grade and session date time as postform.
// @Tags DemoSessionBooking
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param session-datetime  formData  string true "Session Datetime"
// @Param student-name  formData  string true "Student Name"
// @Param grade  formData  string true "Student Grade"
// @Param parent-email  formData  string true "parent's Email`"
// @Produce json
// @Success 200
// @Router /session/book-demo [POST]
func BookDemoSession(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("BookDemoSession: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	studentID, updateProfileErr := studentscontroller.UpdateStudentProfile(c, userID.(int))
	if updateProfileErr != nil {
		log.Println("BookDemoSession: Failed to update student profile with error: ", updateProfileErr)
		return
	}

	studentProfile, fetchStudentProfileErr := students.GetStudentProfile(userID.(int))
	if fetchStudentProfileErr != nil {
		log.Println("BookDemoSession: Failed to fetch student profile with error: ", fetchStudentProfileErr)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, fetchStudentProfileErr, c)
		return
	}
	requestData, err := GetRequestDataForSessionBooking(c)
	if err != nil {
		log.Println("BookDemoSession: Failed to fetch mandatory fields for booking student's demo session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		return
	}
	requestData.SessionType = constants.SessionTypeDemo
	requestData.StudentID = studentID
	tutorDetails, err := tutor.GetTutorByID(requestData.TutorID)
	if err != nil {
		log.Println("BookDemoSession: Failed to get tutor details with error: ", err)
	}
	// Here we check user can schedule thier demo again.
	isAllowedToBookDemo, err := sessionmodel.CheckStudentAllowedToBookDemo(userID.(int))
	if err != nil {
		log.Println("BookDemoSession: Failed to check session information forr allowing to book demo with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	// here we call model function to store information of demo session with perticular student and tutor.
	var sessionID int
	if isAllowedToBookDemo {
		err, sessionID = sessionmodel.SaveDemoSessionInfoOfStudent(userID.(int), requestData)
		if err != nil {
			log.Println("BookDemoSession: Failed to save student information with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
	}
	phone := studentProfile.User.DialingCode + studentProfile.User.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}

	studentData := messageservices.StudentData{
		ID:          int64(userID.(int)),
		Email:       studentProfile.ParentEmail,
		Name:        studentProfile.StudentName,
		PhoneNumber: phone,
	}

	teacherData := messageservices.TeacherData{
		ID:       int64(requestData.TutorID),
		Email:    tutorDetails.TutorEmail,
		Timezone: tutorDetails.Timezone,
		Name:     tutorDetails.TutorName,
	}

	sessionData := messageservices.SessionData{
		ID:          int64(sessionID),
		StartTime:   requestData.SessionDateTime,
		Timezone:    requestData.TimeZone,
		SessionType: "demo",
		MeetingLink: tutorDetails.ZoomLink,
	}
	go messageservices.ScheduleMsgsForSessionBook(studentData, teacherData, sessionData)

	formatedDate, formatedTime := utility.FormatToReadableDateTime(requestData.SessionDateTime, requestData.TimeZone)

	formatedDateTutor, formatedTimeTutor := utility.FormatToReadableDateTime(requestData.SessionDateTime, tutorDetails.Timezone)
	//message := fmt.Sprintf("Hi! Your child's demo class for building amazing games is confirmed for %v. The class link will be sent 24 hours before the session.", formatedDateTime)
	// message := fmt.Sprintf(
	// 	"Hey %s!\n\n"+
	// 		"Your registration for the Coding Demo Class is confirmed. You are about to embark on an exciting coding journey!\n\n"+
	// 		"Time: %s at %s\n\n"+
	// 		"Please ensure to join the class within the first 5 minutes of the scheduled time.\n\n"+
	// 		"All the best,\n"+
	// 		"Team Tutree",
	// 	studentProfile.StudentName, formatedDate, formatedTime,
	// )
	// err = services.SendMessage(studentProfile.User.Phone, message, studentProfile.User.DialingCode)
	// if err != nil {
	// 	log.Println("BookDemoSession: sending confirmation message of demo session failed: ", err)
	// }
	// rescheduleURL := utility.GetHostURL() + "/student/dashboard/demo-reschedule"

	// pDataStudent := map[string]interface{}{
	// 	"session_date":     formatedDate,
	// 	"session_time":     formatedTime,
	// 	"name":             studentProfile.StudentName,
	// 	"event_start_time": requestData.SessionDateTime,
	// 	"reschedule_url":   rescheduleURL,
	// }

	// pDataTutor := map[string]interface{}{
	// 	"session_date":     formatedDateTutor,
	// 	"session_time":     formatedTimeTutor,
	// 	"kid_name":         studentProfile.StudentName,
	// 	"session_type":     "demo",
	// 	"tutor_name":       tutorDetails.TutorName,
	// 	"event_start_time": requestData.SessionDateTime,
	// 	"tutor_dashboard":  utility.GetTutorHostURL() + "/dashboard",
	// }

	// fmt.Println("***dynamic data tutor", pDataTutor)
	// requestDataStudent := services.DynamicEmailDetailsForBrevo{
	// 	To:          []string{studentProfile.ParentEmail},
	// 	DynamicData: pDataStudent,
	// 	// TemplateID:  constants.CongratulationEmailOnDemoBooking,
	// }

	// requestDataTutor := services.DynamicEmailDetailsForBrevo{
	// 	To:          []string{tutorDetails.TutorEmail},
	// 	DynamicData: pDataTutor,
	// 	// TemplateID:  constants.CongratulationTutorEmailOnDemoBooking,
	// }

	// Send notification on slack whenever new demo booked.
	countryCode := studentProfile.User.Location.CountryCode
	if countryCode == "US" {
		channelID := os.Getenv("DEMO_BOOKED_CHANNEL")
		// Check if the channel ID is empty; if so, log the error and skip sending the message
		if len(channelID) == 0 {
			log.Println("[ERROR] Failed to get DEMO_BOOKED_CHANNEL")
		} else {
			studentTime := formatedDate + " at " + formatedTime + ", " + requestData.TimeZone
			teacherTime := formatedDateTutor + " at " + formatedTimeTutor + ", " + tutorDetails.Timezone
			source := utility.GetSourceOfSignup(studentProfile.User.QueryParams)
			location := studentProfile.User.Location.City + ", " + studentProfile.User.Location.State + ", " + countryCode
			message := fmt.Sprintf("*New Demo Booked on Tutree* \nPhone Number: %s \nLocation: %s \nKid Name: %s\nKid Grade: %s"+
				"\nParent Email: %s\nSession Time: %s\nTimezone: %s\nSource: %s\nStudent Time: %s\nTeacher Time: %s\nTeacher Name: %s", studentProfile.User.Phone, location,
				studentProfile.StudentName, studentProfile.Grade, studentProfile.ParentEmail, requestData.SessionDateTime, requestData.TimeZone, source, studentTime, teacherTime, tutorDetails.TutorName)
			go services.SendSlackMessage(channelID, message)
		}
	}
	// update the leads on ZOHO CRM lead board
	err = controllers.UpdateLeadsOnZohoCRM(userID.(int))
	if err != nil {
		log.Println("BookDemoSession: Failed to update student information in zoho crm lead board with error: ", err)
		//	controllers.HandleJSONErrorResponse(apierror.FailedToCreateLeadOnZohoCRM, err, c)
		//return
	}
	// go services.ComposeDynamicTemplateEmailsForBrevoServiceWithICS(requestDataStudent)

	// go services.ComposeDynamicTemplateEmailsForBrevoServiceWithICS(requestDataTutor)

	c.JSON(http.StatusOK, gin.H{
		"status":       "Success",
		"message":      "Your free session is scheduled.",
		"session_time": requestData.SessionDateTime,
		"session_id":   sessionID,
		"subject_id":   requestData.SubjectID,
		"student_id":   requestData.StudentID,
		"meeting_link": tutorDetails.ZoomLink,
		"timezone":     requestData.TimeZone,
	})
}

// StudentRegistration godoc
// @Summary This Api will retrieve session list.
// @description This function retrieve a list of tutor sessions based on query parameters(If not provided then it will all list).
// @Tags Tutor
// @Accept application/x-www-form-urlencoded
// @Param date-from query string false "start-date"
// @Param date-to query  string false "end-date"
// @Param session-type query string false "session-type"
// @Param status query string false "session-status"
// @Param tutor-id query string false "tutor-id"
// @Produce json
// @Success 200
// @Router /v1/session/list [get]
func GetSessionlist(c *gin.Context) {
	/*
		This function handles the HTTP GET request to retrieve a list of tutor sessions based on query parameters(If not provided then it will all list).
		It uses the Gin framework to manage the request context and returns a JSON response with the list of sessions.

		Default case:
		- If "date-from not provided, it added current date".
		- If "date-to" is not provided, it defaults to 30 days from the current date.
	*/

	var (
		dateFrom, dateTo string
		err              error
		tutorID          int64
	)

	userID, exists := c.Get("userID")
	if !exists {
		log.Println("BookDemoSession: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	// plan, err := stripe.GetActivePlanDetails(userID.(int))

	dateFromParam := c.Query("date_from")
	// this will provide default date(current-date)
	dateFrom, err = utility.GetFormattedDate(dateFromParam)
	if err != nil {
		controllers.HandleJSONErrorResponse(apierror.FailedWhileParsingDateTime, err, c)
		return
	}

	dateToParam := c.Query("date_to")
	// this will provide default date(current-date)
	dateTo, err = utility.GetFormattedDate(dateToParam)
	if err != nil {
		controllers.HandleJSONErrorResponse(apierror.FailedWhileParsingDateTime, err, c)
		return
	}
	// adding the 30 days to current date if it is null.
	// if dateToParam == "" || len(dateToParam) == 0 {
	// 	// dateTo = time.Now().AddDate(0, 0, 30).Format(utility.DateLayout)
	// }

	status := strings.TrimSpace(c.Query("status"))
	tutorId := strings.TrimSpace(c.Query("tutor_id"))

	if tutorId != "" {
		tutorID, err = strconv.ParseInt(tutorId, 10, 64)
		if err != nil {
			log.Println("[ERROR] GetSessionlist: Failed to convert tutor-id in integer with ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
			return
		}
	}
	sessionType := strings.TrimSpace(c.Query("session_type"))
	grade := strings.ToLower(strings.TrimSpace(c.Query("grade")))
	studentName := strings.ToLower(strings.TrimSpace(c.Query("s_name")))
	subjectName := strings.ToLower(strings.TrimSpace(c.Query("sub_name")))

	// A `filters` struct is created with the parsed parameters to hold the filter criteria for the tutor sessions.
	filters := sessionmodel.SessionListFilters{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Status:      status,
		TutorID:     tutorID,
		SessionType: sessionType,
		SubjectName: subjectName,
		StudentName: studentName,
		Grade:       grade,
		UserID:      userID.(int),
	}

	sessionList, err := sessionmodel.GetSessionlist(filters)
	if err != nil {
		log.Println("[ERROR] GetSessionlist: Failed to get session list with ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "Success",
		"sessionList": sessionList,
	})
}

func GetSession(c *gin.Context) {

	var (
		sessionID int64
		err       error
	)

	sessionFromParam := strings.TrimSpace(c.Query("session_id"))

	if sessionFromParam == "" || len(sessionFromParam) == 0 {
		log.Println("[ERROR] GetSession: session-id not provided ", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, err, c)
		return

	} else {
		sessionID, err = strconv.ParseInt(sessionFromParam, 10, 64)
		if err != nil || sessionID <= 0 {
			log.Println("[ERROR] GetSession: Failed to convert session-id in integer with ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
			return
		}
	}

	session, err := sessionmodel.GetSession(sessionID, 0)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("[NO_DATA_FOUND] GetSession: no data found for provided session_id")
			controllers.HandleJSONErrorResponse(apierror.ErrorNoDataFound, err, c)
			return
		} else {
			log.Println("[ERROR] GetSession: Failed to get session list with ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Successfully fetch session details",
		"session": session,
	})
}

func CancelSession(c *gin.Context) {
	var (
		sessionID int64
		err       error
	)
	sessionFromParam := c.Query("session_id")
	if sessionFromParam == "" || len(sessionFromParam) == 0 {
		log.Println("[ERROR] CancelSession: session-id not provided ", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, err, c)
		return

	} else {
		sessionID, err = strconv.ParseInt(sessionFromParam, 10, 64)
		if err != nil || sessionID <= 0 {
			log.Println("[ERROR] CancelSession: Failed to convert session-id in integer with ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
			return
		}
	}
	// default minutes will be 60 mintues. It is manged by .env varibale.
	cancelSessionInMinutes, err := strconv.ParseInt(os.Getenv("CANCEL_SESSION_BEFORE"), 10, 64)
	if err != nil {
		cancelSessionInMinutes = 60
	}

	err = sessionmodel.CancelSession(sessionID, int(cancelSessionInMinutes))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("[NO_DATA_FOUND] CancelSession: no data found for provided session_id")
			controllers.HandleJSONErrorResponse(apierror.ErrorNoDataFound, err, c)
			return
		} else {
			log.Println("[ERROR] CancelSession: Failed to get session list with ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully canceled the scheduled session",
	})

}

func AddSessionFeedback(c *gin.Context) {
	var (
		sessionID int64
		rating    float64
		liked     bool
		err       error
	)

	sessionFromParam := c.PostForm("session_id")
	if sessionFromParam == "" || len(sessionFromParam) == 0 {
		log.Println("[ERROR] AddSessionFeedback: session-id not provided ", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, err, c)
		return

	} else {
		sessionID, err = strconv.ParseInt(sessionFromParam, 10, 64)
		if err != nil || sessionID <= 0 {
			log.Println("[ERROR] AddSessionFeedback: Failed to convert session-id in integer with ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
			return
		}
	}

	likedStr := c.PostForm("liked")
	if likedStr == "" {
		log.Println("[ERROR] AddSessionFeeback missing liked from user side.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, err, c)
		return
	} else {
		// Parse liked to boolean
		liked, err = strconv.ParseBool(likedStr)
		if err != nil {
			log.Println("[ERROR] AddSessionFeeback: Failed to parse liked into boolean with ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoBoolean, err, c)
			return
		}
	}

	ratingStr := c.PostForm("rating")
	comment := c.PostForm("comment")

	if ratingStr != "" || len(ratingStr) != 0 {
		rating, err = strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			log.Println("[ERROR] AddSessionFeeback: Failed to convert liked into integer with ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
			return
		}
	}

	feedback := sessionmodel.SessionFeedback{
		SessionID: int(sessionID),
		Liked:     liked,
		Rating:    rating,
		Comment:   comment,
	}
	err = sessionmodel.AddSessionFeedback(feedback)
	if err != nil {
		log.Println("[ERROR] AddSessionFeedback: Failed to get add session feedback with ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully added feedback",
		"status":  "Success",
	})
}

func RequestDataForRescheduleSession(c *gin.Context) (sessionmodel.RequestDataForReschedule, error) {
	var rescheduleSessionInfo sessionmodel.RequestDataForReschedule
	sessionID, err := strconv.Atoi(c.PostForm("session_id"))
	if err != nil || sessionID <= 0 {
		log.Println("[ERROR] RequestDataForRescheduleSession: Failed to convert seession ID from string to number with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidSessionID, err, c)
		return rescheduleSessionInfo, err
	}
	rescheduleSessionInfo.SessionID = sessionID
	startSession := c.PostForm("session_datetime")
	if len(startSession) == 0 {
		log.Println("RequestDataForRescheduleSession: Failed to fetch date and time of student's demo session.")
		controllers.HandleJSONErrorResponse(apierror.SessionDateTimeNotFound, err, c)
		return rescheduleSessionInfo, err
	}
	startSessionInUTC, err := utility.ConvertToUTCTime(startSession)
	if err != nil {
		log.Println("[ERROR] RequestDataForRescheduleSession: Failed to parse session time with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.FailedWhileParsingDateTime, err, c)
		return rescheduleSessionInfo, err
	}
	rescheduleSessionInfo.SessionStartTime = startSessionInUTC
	// ensuring at given session date time by student any tutor is available or not if available, fetched id of tutor.
	tutorID, err := strconv.Atoi(c.PostForm("tutor_id"))
	if err != nil {
		log.Println("[ERROR] RequestDataForRescheduleSession: Failed to fetch available tutor from database with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToParseTutorIDIntoInteger, err, c)
		return rescheduleSessionInfo, err
	}
	if tutorID <= 0 {
		log.Println("[INVAILID] RequestDataForRescheduleSession: Failed to book demo session due to invalid tutor ID, tutor should be positive integer.")
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
		return rescheduleSessionInfo, nil
	}
	rescheduleSessionInfo.TutorID = tutorID

	rescheduleSessionInfo.TimeZone = c.PostForm("timezone")

	prevSession, err := sessionmodel.GetPreviousSessionDetails(sessionID)
	if err != nil {
		log.Println("[ERROR] RequestDataForRescheduleSession : Failed to get session details of previous session before rescheduling session with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return rescheduleSessionInfo, err
	}
	rescheduleSessionInfo.Type = prevSession.SessionType
	rescheduleSessionInfo.SubjectName = prevSession.SubjectName
	if len(rescheduleSessionInfo.TimeZone) == 0 {
		rescheduleSessionInfo.TimeZone = prevSession.TimeZone
	}
	rescheduleSessionInfo.SessionEndTime = utility.SessionEndDateTime(rescheduleSessionInfo.SessionStartTime)
	return rescheduleSessionInfo, nil
}

// RescheduleDemoSession handles the rescheduling of a demo session for a student.
// It validates the authorization token, parses and validates the session ID and datetime,
// checks for an available tutor, and then reschedules the session if all checks pass.
//
// Parameters:
// - c: *gin.Context - The context for the HTTP request, which provides access to the request and response objects.
// session ID :- type integer
// session date-time :- type string
//
// Response:
// - JSON response with status and message, and the new session time if successful.
// - JSON error response with an appropriate error message if any validation or processing step fails.

// RescheduleDemoSession godoc
// @Summary This function handles the rescheduling of a demo session for a students.
// @description function reschedules a student's demo session by validating the authorization token,
// @description converting session details, finding an available tutor, updating the session, and responding with a success message
// @Tags Session
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param session-datetime  formData  string true "Session Datetime"
// @Param session-id formData number true "Session ID"
// @Produce json
// @Success  200
// @Router /session/demo-reschedule [put]
func RescheduleDemoSession(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("RescheduleDemoSession: Failed to validate user session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	requestData, err := RequestDataForRescheduleSession(c)

	if err != nil {
		log.Println("[ERROR] RescheduleDemoSession : Failed to get mandetory data from handeler function RequestDataForRescheduleSession with error :", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, err, c)
		return
	}
	if requestData.Type == utility.SessionTypePaid {
		endDateStr, err := students.GetActiveSubscriptionDetailsRenewalDate(userID.(int))
		if err != nil {
			log.Println("[ERROR] RescheduleDemoSession: Failed while try to get end date of subscription of student with error:", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
		if len(endDateStr) != 0 {
			log.Println("endDate", endDateStr)
			var endDate time.Time
			endDate, err = utility.ParseDateWithEndOfDayTime(endDateStr)
			if err != nil {
				log.Println("[ERROR] RescheduleDemoSession: Failed while try to parse string into time.Time with error: ", err)
				controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
				return
			}
			if requestData.SessionStartTime.After(endDate) {
				log.Println("[ERROR] RescheduleDemoSession : Failed to reschedule session due to start date time exceeded from active plan.")
				controllers.HandleJSONErrorResponse(apierror.RescheduleNotAllowedBeyondSubscriptionValidity, nil, c)
				return
			}
		}
	}
	var afterFormatedDate, afterFormatedTime string
	afterFormatedDate, afterFormatedTime = utility.FormatToReadableDateTime(requestData.SessionStartTime, requestData.TimeZone)

	studentProfile, err := students.GetStudentProfile(userID.(int))
	if err != nil {
		log.Println("[ERROR] RescheduleDemoSession : Failed to fetch student profile with  error:", err)
	}
	tutorDetails, err := tutor.GetTutorByID(requestData.TutorID)
	if err != nil {
		log.Println("[ERROR] RescheduleDemoSession: Failed to get tutor details with error: ", err)
	}
	tutorSessionDate, tutorSessionTime := utility.FormatToReadableDateTime(requestData.SessionStartTime, tutorDetails.Timezone)
	// default minutes will be 60 mintues. It is manged by .env varibale.
	rescheduleSessionInMinutes, err := strconv.ParseInt(os.Getenv("RESCHEDULE_SESSION_BEFORE"), 10, 64)
	if err != nil {
		rescheduleSessionInMinutes = utility.RescheduleSessionBeforeInMinutes
	}
	reschedulingStatus, err := sessionmodel.CheckSessionReschedulingStatus(rescheduleSessionInMinutes, int64(requestData.SessionID))
	if reschedulingStatus == "not allowed" {
		log.Println("RescheduleDemoSession: Rescheduling failed. Sessions cannot be rescheduled less than 60 minutes before their start time.")
		controllers.HandleJSONErrorResponse(apierror.RescheduleDemoSessionNotAllowed, nil, c)
		return
	}
	if err != nil {
		log.Println("RescheduleDemoSession: Failed to reschedule student demo session with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	// here we call model function to store information of demo session with perticular student and tutor.
	err = sessionmodel.ResheduleDemoSession(userID.(int), requestData)
	if err != nil {
		log.Println("[ERROR] RescheduleDemoSession: Failed to reschedule student demo session with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	if studentProfile.User.Location.CountryCode == "US" {
		channelID := os.Getenv("DEMO_BOOKED_CHANNEL")
		// Check if the channel ID is empty; if so, log the error and skip sending the message
		if len(channelID) == 0 {
			log.Println("[ERROR] Failed to get DEMO_BOOKED_CHANNEL")
		} else {
			studentTime := afterFormatedDate + " at " + afterFormatedTime + ", " + requestData.TimeZone
			teacherTime := tutorSessionDate + " at " + tutorSessionTime + ", " + tutorDetails.Timezone
			source := utility.GetSourceOfSignup(studentProfile.User.QueryParams)
			location := studentProfile.User.Location.City + ", " + studentProfile.User.Location.State + ", " + studentProfile.User.Location.CountryCode
			message := fmt.Sprintf("%s has rescheduled their demo class. Here are the new timings: \n\nPhone Number: %s \nLocation: %s \nKid Name: %s\nKid Grade: %s"+
				"\nParent Email: %s\nSession Time: %s\nSource: %s\nStudent Time: %s\nTeacher Time: %s\nTeacher Name: %s", studentProfile.ParentEmail, studentProfile.User.Phone, location,
				studentProfile.StudentName, studentProfile.Grade, studentProfile.ParentEmail, requestData.SessionStartTime, source, studentTime, teacherTime, tutorDetails.TutorName)
			go services.SendSlackMessage(channelID, message)
		}
	}

	// if requestData.SessionType == "paid" {
	// 	requestData.SessionType = constants.SessionTypeIfPaid
	// } else {
	// 	requestData.SessionType = constants.SessionTypeIfDemo
	// }

	phone := studentProfile.User.DialingCode + studentProfile.User.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}
	var data []messageservices.DataForScheduleMsgs
	data = append(data, messageservices.DataForScheduleMsgs{
		SessionData: messageservices.SessionData{
			ID:          int64(requestData.SessionID),
			StartTime:   requestData.SessionStartTime,
			Timezone:    requestData.TimeZone,
			SessionType: requestData.Type,
			MeetingLink: tutorDetails.ZoomLink,
		},
		TeacherData: messageservices.TeacherData{
			ID:       int64(requestData.TutorID),
			Email:    tutorDetails.TutorEmail,
			Timezone: tutorDetails.Timezone,
			Name:     tutorDetails.TutorName,
		},
		StudentData: messageservices.StudentData{
			ID:          int64(userID.(int)),
			Email:       studentProfile.ParentEmail,
			Name:        studentProfile.StudentName,
			PhoneNumber: phone,
		},
	})

	go messageservices.RescheduleMsgServices(data)

	// emailSubject := fmt.Sprintf("%s's Tutree Coding Class", studentProfile.StudentName)
	// studentDynamicData := map[string]interface{}{
	// 	"kid_name":         studentProfile.StudentName,
	// 	"session_date":     afterFormatedDate,
	// 	"session_time":     afterFormatedTime,
	// 	"session_type":     requestData.SessionType,
	// 	"event_start_time": requestData.StartSession,
	// 	"event_summary":    emailSubject,
	// }
	// requestDate := services.DynamicEmailDetailsForBrevo{
	// 	To:          []string{studentProfile.ParentEmail},
	// 	DynamicData: studentDynamicData,
	// TemplateID:  constants.SessionRescheduleSuccessfully,
	// }
	// go services.ComposeDynamicTemplateEmailsForBrevoServiceWithICS(requestDate)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Your free session is re-scheduled.",
		"session_time": requestData.SessionStartTime,
		"timezone":     requestData.TimeZone,
		"subject_name": requestData.SubjectName,
	})
}

// BookRegularSession handles the booking of a regular session for a student. It validates the user's session,
// processes the request data to identify available time slots, and saves the booking information.
//
// Steps:
// 1. Validate User Session
// 2. Bind JSON Data
// 3. Fetch Available Slots
// 4. Save Session Information
//
//  5. Send Success Response: If all operations are successful, it responds with a success message indicating that
//     the session has been booked successfully.
//
// Parameters:
//   - c *gin.Context: The Gin context object, which contains request data, session information, and provides methods
//     for sending responses.
//
// Returns:
// -  Sends a JSON response indicating success or failure based on the execution of the function.
func BookRegularSession(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("BookRegularSession: Failed to validate user session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	var session sessionmodel.SessionBookRequest
	err := c.ShouldBindJSON(&session)
	if err != nil {
		log.Println("BookRegularSession: Failed to bind json responce with error :", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToBindJsonRespoce, err, c)
		return
	}
	if len(session.TimeZone) == 0 {
		log.Println("BookRegularSession: Failed to fetch timezone of student's regular session.")
		controllers.HandleJSONErrorResponse(apierror.SessionTimezoneNotFound, nil, c)
		return
	}
	if len(session.Slots) == 0 {
		log.Println("BookRegularSession:Failed to fetch day and time of student's regular session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		return
	}
	if len(session.SessionType) == 0 {
		log.Println("BookRegularSession: Failed to fetch session type of student's regular session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		return
	}
	if session.TutorID <= 0 {
		log.Println("BookRegularSession: Failed to fetch tutur ID of student's regular session.")
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
		return
	}
	if len(session.StartFrom) == 0 {
		currentTime := time.Now()

		session.StartFrom = currentTime.Format("2006-01-02")
		// This function is added to change the billing cycle for users who scheduled their first class today.

	}
	if session.SubjectID <= 0 {
		log.Println("BookRegularSession: Failed to fetch subject ID of student's regular session.")
		controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
		return
	}
	if session.StudentID <= 0 {
		log.Println("BookRegularSession: Failed to fetch student ID of student's regular session.")
		controllers.HandleJSONErrorResponse(apierror.FailedToParseStudentIDIntoInteger, nil, c)
		return
	}
	sessionTimeUnavailable, err := sessionmodel.VarifyRequestedSessionDateTime(session.Slots, session.TutorID)

	if err != nil {
		log.Println("[ERROR] BookRegularSession : Failed to varify session date time are available or not with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	if len(sessionTimeUnavailable) != 0 {
		if len(sessionTimeUnavailable) == 1 {
			err = fmt.Errorf("missing: %s, session time partially unavailable", sessionTimeUnavailable[0])
			log.Println("[UNAVAILABLE TIME] BookRegularSession : Failed due to only one unavailable time provided.", err)
			controllers.HandleJSONErrorResponse(apierror.FailedTBookDueToUnavailableSessionTime, err, c)
			return
		} else {
			err = fmt.Errorf("missing: %s to %s", sessionTimeUnavailable[0], sessionTimeUnavailable[1])
			log.Println("[UNAVAILABLE TIME] BookRegularSession : Failed due to given time and date of session unavailable.", err)
			controllers.HandleJSONErrorResponse(apierror.FailedTBookDueToUnavailableSessionTime, err, c)
			return
		}
	}

	SessiondayTime, err := GetDistinctDayNamesAndTimes(session.Slots, nil)
	if err != nil {
		log.Println("VarifyRequestedSessionDateTime : Failed with error", err)
	}

	// get available tutor slot using session day and time choosing student.
	plan, err := stripe.GetActivePlanDetails(userID.(int))
	if err != nil {
		log.Println("BookRegularSession: Failed to fetch active subscription plan slots in given user id with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	if len(session.Slots) != plan.MaxSession {
		log.Printf("[SESSION NOT BOOKED] BookRegularSession : The number of selected sessions does not match the maximum sessions ---> %d allowed by your plan.", plan.MaxSession)
		controllers.HandleJSONErrorResponse(apierror.FailedDueToMismatchNumberSessions, nil, c)
		return
	}
	endDateStr, err := students.GetActiveSubscriptionDetailsRenewalDate(userID.(int))
	if err != nil {
		log.Println("BookRegularSession [ERROR]: Failed while try to get end date of subscription of student with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	log.Println("endDate", endDateStr)
	var endDate time.Time
	if len(endDateStr) != 0 {
		layout := time.RFC3339
		parsedTime, err := time.Parse(layout, endDateStr)
		if err != nil {
			log.Println("BookRegularSession [ERROR]: Failed while try to parse string into time.Time with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}

		// Extract only the date portion (Year, Month, Day)
		//endDate := parsedTime.Format("2006-01-02")
		endDate = time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 23, 59, 59, 0, parsedTime.Location())
	}

	if endDate.IsZero() {
		startTime, err := time.Parse("2006-01-02", session.StartFrom)
		if err != nil {
			log.Println("BookRegularSession : Failed to parse start from time string to time.Time  with error :", err)
			return
		}
		endDate = startTime.AddDate(0, 0, 27)

	}
	//
	// Save Session in billing period with credit = true
	//
	// var sessionDateTime []time.Time
	// slots, err := sessionmodel.GetAvailableTutorSlotsForRegularSession(session.SessionDayTime, session.TutorID, plan.MaxSession, endDate, session.StartFrom)
	// if err != nil {
	// 	log.Println("BookRegularSession: Failed to fetch available tutor slots in given day and time with error:", err)
	// 	controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
	// 	return
	// }

	// sessionDateTime = append(sessionDateTime, slots...)

	// save student's regular session informations in session table.

	err = sessionmodel.SessionSaveAndCredit(userID.(int), session.TutorID, session.SubjectID, session.StudentID, session.Slots, session.TimeZone)
	if err != nil {
		log.Println("BookRegularSession: Failed to save student's regular session in database with error  :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	//Default Grace Period
	gracePeriod, err := strconv.Atoi(os.Getenv("GRACE_PERIOD"))
	if err != nil || gracePeriod < 0 {
		gracePeriod = 27
	}

	//
	// Save Next session with credit = false for Grace Period
	//
	sessionStartDate, _ := getMinSessionStart(session.Slots)

	billingPeriodStartDate := sessionStartDate.AddDate(0, 0, 27)
	nextMonthStartDate := billingPeriodStartDate //endDate.Add(time.Second)
	nextMonthStartFromStr := nextMonthStartDate.Format("2006-01-02")
	nextMonthEndOfDay := nextMonthStartDate.AddDate(0, 0, gracePeriod)

	fmt.Println("NEXT MONTH START AND END DATE-->", nextMonthStartDate, "NEXT Month START from string-->", nextMonthStartFromStr, "NEXT MONTH END DAY Of day-->", nextMonthEndOfDay)

	slots, err := sessionmodel.GetAvailableTutorSlotsForRegularSession(SessiondayTime, session.TutorID, plan.MaxSession, nextMonthEndOfDay, nextMonthStartFromStr)
	if err != nil {
		log.Println("BookRegularSession: Failed to fetch available tutor slots for [GRACE-PERIOD] to add with error:", err)
		//return
	}
	var sessionSlot []sessionmodel.SessionSlots
	for _, startTime := range slots {
		slot := sessionmodel.SessionSlots{
			SessionStart: startTime,
		}
		sessionSlot = append(sessionSlot, slot)
	}
	// Save student's regular session information in session table

	err = sessionmodel.SaveSessions(userID.(int), session.TutorID, session.SubjectID, session.StudentID, sessionSlot, session.TimeZone)
	userSessionPreferenceDayNTime, err := GetDistinctDayNamesAndTimes(session.Slots, &session.TimeZone)
	if err != nil {
		log.Println("GetDistinctDayNamesAndTimes : Failed with error", err)
	}

	var sessionPreferencesfDateTime []string
	timezone := session.TimeZone
	// Save Prefence of user.
	preferences := []sessionmodel.StudentSessionPreference{}

	for _, pref := range userSessionPreferenceDayNTime {
		preference := sessionmodel.StudentSessionPreference{
			UserID:    int64(userID.(int)),
			TutorID:   int64(session.TutorID),
			DayOfWeek: pref.Day,
			StartTime: pref.Time,
			StartDate: sessionStartDate.Format(utility.DateLayout),
			Timezone:  session.TimeZone,
			SubjectID: int64(session.SubjectID),
			StudentID: int64(session.StudentID),
		}

		preferences = append(preferences, preference)
	}
	fmt.Println("UPSERTING STUDENT PREFERENCE-->>>", preferences)
	_, err = sessionmodel.AddOrUpdateSessionPreference(preferences)
	if err != nil {
		log.Println("[ERROR]:BookRegularSession: InsertStudentPreferences: Failed to insert session preferences :", err)
	}
	go func() {
		for _, dayTime := range userSessionPreferenceDayNTime {
			dayTime.Time = utility.GetTimeWithAMOrPM(dayTime.Time)
			//
			caser := cases.Title(language.Und)            // Initializes a Title case transformer with unspecified language (`Und`), allowing it to handle general title case formatting.
			formattedDayName := caser.String(dayTime.Day) // Converts `dayTime.Day` to title case (e.g., "monday" becomes "Monday").
			sessionPreferencesfDateTime = append(sessionPreferencesfDateTime, fmt.Sprintf("%s, at %s %s", formattedDayName, dayTime.Time, utility.GetTimeZoneAbbreviation(session.TimeZone)))

		}
		sessionEarliestDate, sessionEarliestTime := FindEarliestDate(session.Slots)
		go PrepareDataAndSendEmailForRecuringPreference(userID.(int), sessionEarliestDate, sessionEarliestTime)
	}()

	// Setting up the subscription (valid until) when user schedule their first class.
	firstPaidSessionEndDateAndTime, err := sessionmodel.UpdateSubscriptionStartDateForUser(userID.(int))
	if err != nil {
		log.Println("[ERROR] BookRegularSession: Failed to UpdateSubscriptionStartDateForUser with err ", err)
	}
	loc, _ := time.LoadLocation(timezone)

	// NOW WE ARE UPDATING THE 'trial end' by (default it is 28 days after) after after user booked their first session
	// we update the trial end to their first session_end date and time with buffer time after which it get deducted.
	err = common.UpdateSubscriptionTrialEndToDate(userID.(int), firstPaidSessionEndDateAndTime, loc)
	if err != nil {
		log.Println("[ERROR] BookRegularSession -> UpdateTrialEndToSessionEndDate: Failed to update the trialENDDateAndTime with ", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Session booked successfully",
		"timezone": session.TimeZone,
	})
}

// PrepareDataAndSendEmailForRecuringPreference prepare the data for session preference and send email to student.
func PrepareDataAndSendEmailForRecuringPreference(userID int, sessionEarliestDate string, sessionEarliestTime string) {
	studentProfile, err := students.GetStudentProfile(userID)
	if err != nil {
		log.Println("[ERROR] PrepareDataAndSendEmailForRecuringPreference : Failed to fetch student details with error :", err)
		return
	}
	studentData := messageservices.StudentData{
		ID:    int64(userID),
		Email: studentProfile.ParentEmail,
		Name:  studentProfile.StudentName,
	}
	sessionData := messageservices.RecurringSessionData{
		SessionEarliestTime: sessionEarliestTime,
		SessionEarliestDate: sessionEarliestDate,
	}
	go messageservices.SendEmailForRecuringSessionPreference(studentData, sessionData)
}

// FindEarliestDate takes an array of date strings, parses them into time.Time values,
// and returns the earliest date. If parsing fails, it logs an error.
func FindEarliestDate(dates []sessionmodel.SessionSlots) (string, string) {
	var earliestDate time.Time

	for i, dateStr := range dates {
		// Parse the date string

		// On the first iteration, set the initial earliest date
		if i == 0 || dateStr.SessionStart.Before(earliestDate) {
			earliestDate = dateStr.SessionStart
		}
	}

	return earliestDate.Format(utility.DateLayout), earliestDate.Format(utility.TimeLayout)
}
func getMinSessionStart(slots []sessionmodel.SessionSlots) (time.Time, error) {
	if len(slots) == 0 {
		return time.Time{}, fmt.Errorf("no session slots available")
	}

	minDate := slots[0].SessionStart

	for _, slot := range slots[1:] {
		if slot.SessionStart.Before(minDate) {
			minDate = slot.SessionStart
		}
	}

	return minDate, nil
}

func CreditSessionsForUser(userID int) {
	log.Println("CreditSessionsForUser")
	//Default Grace Period
	gracePeriod, err := strconv.Atoi(os.Getenv("GRACE_PERIOD"))
	if err != nil || gracePeriod < 0 {
		gracePeriod = 27
	}

	// Get available tutor slot using session day and time chosen by the student.
	plan, err := stripe.GetActivePlanDetails(userID)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to fetch active subscription plan slots for given user id with error:", err)
		return
	}

	if len(plan.StartDate.String) == 0 || len(plan.ValidUntil.String) == 0 {
		log.Println("[NO-DATA-FOUND] CreditSessionsForUse: there is no valid date and plan start date found so, we are not adding session")
		return
	}

	fmt.Println("INTO THE CreditSessionFOrUSER", plan)

	startDate := plan.StartDate.String
	validUntilStr := plan.ValidUntil.String
	var slots []time.Time

	validUntilTime, err := time.Parse(time.RFC3339, validUntilStr)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to parse validUntil with error:", err)

	}

	// Get total count of credited sessions
	totalCount, err := sessionmodel.CreditSessionsForUser(userID, startDate, validUntilStr)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to update the credit session", err)
		return
	}

	fmt.Println("UPDATED CreditSessionsForUser of gracePeriod 'TRUE' SESSION COUNT---->>>>>>", totalCount)

	// If no sessions have been credited, book regular sessions

	// Get the student's session day/time preferences
	sessionDayTime, err := sessionmodel.GetStudentSessionPreference(userID)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to get student session preferences with error:", err)
		return
	}

	var (
		tutorID, subjectID, studentID int
		timeZone                      string
		sessionPreferenceDayTime      []sessionmodel.SessionDayTime
	)
	for _, s := range sessionDayTime {
		utcDay, utcTime, err := utility.ConvertTimeInUTC(s.Day, s.Time, s.TimeZone)
		if err != nil {
			log.Println("[ERROR] CreditSessionsForUser : Failed to convert time in UTC with error :", err)
		}
		sessionPreferenceDayTime = append(sessionPreferenceDayTime, sessionmodel.SessionDayTime{
			Day:  strings.ToLower(utcDay),
			Time: utcTime,
		})
		tutorID = s.TutorID
		timeZone = s.TimeZone
		subjectID = s.SubjectID
		studentID = s.StudentID

	}
	var sessionSlot []sessionmodel.SessionSlots
	if totalCount == 0 {
		// Fetch available tutor slots
		slots, err = sessionmodel.GetAvailableTutorSlotsForRegularSession(sessionPreferenceDayTime, tutorID, plan.MaxSession, validUntilTime, startDate)
		if err != nil {
			log.Println("CreditSessionsForUser: Failed to fetch available tutor slots with error:", err)
			return
		}
		fmt.Println("CREDITSESSIONFORUSER-->>>", slots)

		// Save student's regular session information in session table

		for _, startTime := range slots {
			slot := sessionmodel.SessionSlots{
				SessionStart: startTime,
			}
			sessionSlot = append(sessionSlot, slot)
		}
		err := sessionmodel.SessionSaveAndCredit(userID, tutorID, subjectID, studentID, sessionSlot, timeZone)
		if err != nil {
			log.Println("CreditSessionsForUser: Failed to save student's regular session in database with error:", err)
			return
		}

	}

	// added grace time period calculate will be fetched from .env file.
	nextMonthStart := validUntilStr
	nextMonthEndOfDay := validUntilTime.AddDate(0, 0, gracePeriod)

	slots, err = sessionmodel.GetAvailableTutorSlotsForRegularSession(sessionPreferenceDayTime, tutorID, plan.MaxSession, nextMonthEndOfDay, nextMonthStart)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to fetch available tutor slots with error:", err)
		return
	}

	// Save student's regular session information in session table
	for _, startTime := range slots {
		slot := sessionmodel.SessionSlots{
			SessionStart: startTime,
		}
		sessionSlot = append(sessionSlot, slot)
	}
	err = sessionmodel.SaveSessions(userID, tutorID, subjectID, studentID, sessionSlot, timeZone)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to save student's regular session in database with error:", err)
		return
	}

}

// GetUsersWithTodaySession is an API endpoint that runs daily via a cron scheduler.
// The purpose of this API is to retrieve users who have scheduled their first
// regular session (regular classes) today. The list of user IDs is then passed
// to the billing cycle change function to adjust the billing cycle accordingly.

func GetUsersWithTodaySession(c *gin.Context) {
	date := ""
	environment := os.Getenv("ENV")
	if environment == "test" {
		date = c.PostForm("date")

	}
	// Retrieve users who have scheduled their first regular session today.
	userIDs, err := sessionmodel.GetUsersWithTodaySession(date)
	if err != nil {
		// Log the error and send a JSON response with the error message.
		log.Println("[ERROR] GetUsersWithTodaySession: Failed to get users with today's session : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve users with today's session.",
			"error":   err.Error(),
		})
		return
	}

	// Update the billing cycle and get the updated subscription details.
	subscriptionDetails, err := stripe.MarkTrialEnd(userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update billing cycles.",
			"error":   err.Error(),
		})
		return
	}

	// Send the subscription details in the JSON response.
	c.JSON(http.StatusOK, gin.H{
		"status":                "success",
		"updated_subscriptions": subscriptionDetails,
	})
}

// DeleteSession handles the process of deleting a session based on the provided session ID.
// It performs the following steps:
//
//  1. Retrieves the `session_id` from the form data. If the `session_id` is missing, it logs an error
//     and returns a "No Data Found" response.
//
//  2. Converts the `session_id` from string to integer. If the conversion fails, it logs an error
//     and returns a "Failed to Convert" response.
//
//  3. Checks if the `session_id` is valid (non-zero). If invalid, it logs an error and returns a
//     "Field Must Be Positive Integer" response.
//
//  4. Calls the `DeleteSession` method from the `sessionmodel` to delete the session from the database.
//     If deletion fails, it logs an error and returns a generic "Something Went Wrong" response.
//
// 5. If the session is successfully deleted, it returns a JSON response with a success message.

// DeleteSession godoc
// @Summary This function handles to delete a session of a students by session id.
// @description Function provide a feature to delete a student's session by validating the authorization token,
// @Tags Session
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param session-id formData number true "Session ID"
// @Produce json
// @Success  200
// @Router /session [DELETE]
func DeleteSession(c *gin.Context) {

	// Get the userID from the session context and validate it.
	// If userID doesn't exist, log an error and return a "User Not Found" response.
	_, exist := c.Get("userID")
	if !exist {
		log.Println("DeleteSession: Failed to validate user session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	// Retrieve session ID from form data
	sessionIdStr := c.PostForm("session_id")
	if len(sessionIdStr) == 0 {
		log.Println("[NOT FOUND] DeleteSession: Failed to fetch session ID through form data.")
		controllers.HandleJSONErrorResponse(apierror.ErrorNoDataFound, nil, c)
		return
	}

	// Convert session ID from string to integer
	sessionID, err := strconv.Atoi(sessionIdStr)
	if err != nil {
		log.Println("[ERROR] DeleteSession: Failed to convert session ID from string to integer with error:", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
		return
	}

	// Validate session ID
	if sessionID == 0 {
		log.Println("[INVALID DATA] DeleteSession: Failed due to invalid session ID.")
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
		return
	}

	// Delete session from the database
	err = sessionmodel.DeleteSession(sessionID)
	if err != nil {
		log.Println("[ERROR] DeleteSession: Failed to delete session from database with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Session deleted successfully",
	})
}

// SessionRescheduleLater handles the process of rescheduling later a student's session.
// This feature provides a chance to student to reschedule thier class in future, if student unable to attend thier class or misssed and if
// teacher cancelled his class.
//
// 1. Retrieves and validates the `userID` from the session context. If `userID` is not found, it logs an error
//    and returns a "User Not Found" response.
//
// 2. Retrieves the `session_id` from the form data. If the `session_id` is missing, it logs an error
//    and returns a "No Data Found" response.
// 	  Converts the `session_id` from string to integer. If the conversion fails, it logs an error
//    and returns a "Failed to Convert" response.
//    Checks if the `session_id` is valid (non-zero). If invalid, it logs an error and returns a "Field Must Be Positive Integer" response.
//
// 3. Fetches the session details from the database using the `session_id`. If the session cannot be retrieved,
//    it logs an error and returns a generic "Something Went Wrong" response.
//
// 4. Assigns the current `userID` to the session's student user field and saves the current session details
//    to the session_exceptions table. If saving fails, it logs an error and returns a generic "Something Went Wrong" response.
//
// 5. Deletes the current session from the session table. If deletion fails, it logs an error and returns
//    a generic "Something Went Wrong" response.

// SessionRescheduleLater godoc
// @Summary This function handles the rescheduling session in later of a paid session for a students.
// @description Function provide a feature to reschedules later  a student's paid session by validating the authorization token,
// @description session id, and delete old record of that session.
// @Tags Session
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param session_id formData number true "Session ID"
// @Produce json
// @Success  200
// @Router /session/reschedule-later [POST]
func SessionRescheduleLater(c *gin.Context) {
	// Get the userID from the session context and validate it.
	// If userID doesn't exist, log an error and return a "User Not Found" response.
	userId, exist := c.Get("userID")
	if !exist {
		log.Println("SessionRescheduleLater: Failed to validate user session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	// Fetch the session ID from the form data and validate its presence.
	// Log an error and return "No Data Found" if session_id is missing.
	sessionIdStr := c.PostForm("session_id")
	if len(sessionIdStr) == 0 {
		log.Println("[NOT FOUND] SessionRescheduleLater :Failed to fetch session id through form data.")
		controllers.HandleJSONErrorResponse(apierror.ErrorNoDataFound, nil, c)
		return
	}

	// Convert session ID from string to integer.
	// If conversion fails, log the error and return "Failed to Convert" response.
	sessionID, err := strconv.Atoi(sessionIdStr)
	if err != nil {
		log.Println("[ERROR] SessionRescheduleLater :Failed to convert session id from string to integer with error :", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToConvertIntoInteger, err, c)
		return
	}

	// Check if session ID is valid (non-zero). Log an error if invalid.
	if sessionID == 0 {
		log.Println("[INVALID DATA] SessionRescheduleLater :Failed due to invalid session id.")
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
		return
	}
	Id := userId.(int)
	// Retrieve the session details from the database using the session ID.
	// Log an error if the session could not be found.
	session, err := sessionmodel.GetSession(int64(sessionID), int64(Id))

	if err != nil {
		log.Println("[ERROR] SessionRescheduleLater : Failed to get session from database with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	// Assign the current userID to the session's student user field.
	session.Student.User.ID = int64(Id)
	if session.SessionType == utility.SessionTypeDemo {
		log.Println("[INVALID] SessionRescheduleLater : Failed, can not reschedule demo session  for a  later time.")
		controllers.HandleJSONErrorResponse(apierror.FailedToRescheduleDemoSessionForLaterTime, nil, c)
		return
	}

	// Save the current session details to the exceptions table.
	// If saving fails, log the error and return a generic error response.
	err = sessionmodel.SaveExceptionSessionOfStudent(session)
	if err != nil {
		log.Println("[ERROR] SessionRescheduleLater : Failed to save previous session details in database with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Delete the current session from the session table.
	// If deletion fails, log the error and return a generic error response.
	err = sessionmodel.DeleteSession(sessionID)
	if err != nil {
		log.Println("[ERROR] SessionRescheduleLater : Failed to delete previous session details from database with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	var sessionData messageservices.SessionData
	sessionData.ID = int64(sessionID)
	// Automatically cancels session reminders when a student mark reschedules later the session.
	go messageservices.CancelSessionReminders(messageservices.StudentData{}, messageservices.TeacherData{}, sessionData)
	go messageservices.ScheduleEmailForSessionReschedule(session)
	c.JSON(http.StatusOK, gin.H{
		"message": "Success! You can now reschedule your class for a later time.",
	})
}

// GetExpectedSession godoc
// @Summary This function give the pending classes of a students which he want to reschedule later.
// @description Function provide a feature to list the pending classes by validating the authorization token,
// @Tags Session
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Produce json
// @Success  200
// @Router /session/pending-book [GET]
func GetExpectedSession(c *gin.Context) {
	// Retrieve the userID from the context
	userID, exists := c.Get("userID")
	if !exists {
		// If userID is not present in the context, log an error and return a "User Not Found" response
		log.Println("GetExpectedSession: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	// Fetch the expected sessions for the user
	expectedSessions, err := sessionmodel.GetExpectedSession(userID.(int))
	if err != nil {
		// If there's an error fetching the sessions, log the error and return a generic error response
		log.Printf("[ERROR] GetExpectedSession: Error fetching session list from database: %v", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Return a successful response with the session list
	c.JSON(http.StatusOK, gin.H{
		"message":      "Successfully fetched session list.",
		"session_list": expectedSessions,
	})
}
func GetDistinctDayNamesAndTimes(dates []sessionmodel.SessionSlots, timezone *string) ([]sessionmodel.SessionDayTime, error) {
	dayNamesAndTime := make(map[string]string) // Use a map to store distinct day name and time combination

	for _, dateSlot := range dates {
		if timezone != nil {
			localTime, err := utility.ConvertTimeZone(dateSlot.SessionStart, *timezone)
			if err != nil {
				log.Println("[ERROR] GetDistinctDayNamesAndTimes : Failed to convert time in student's local timezone with error :", err)
			}
			dayName := strings.ToLower(localTime.Weekday().String())

			// Extract the time part (HH:mm:ss)
			timePart := localTime.Format("15:04:05")

			// Use a combination of dayName and timePart as key
			dayNamesAndTime[dayName] = timePart
		} else {
			dayName := strings.ToLower(dateSlot.SessionStart.Weekday().String())

			// Extract the time part (HH:mm:ss)
			timePart := dateSlot.SessionStart.Format("15:04:05")

			// Use a combination of dayName and timePart as key
			dayNamesAndTime[dayName] = timePart
		}
	}

	// Create the result list of distinct day names and times
	var distinctDayTimes []sessionmodel.SessionDayTime
	for dayName, timePart := range dayNamesAndTime {
		distinctDayTimes = append(distinctDayTimes, sessionmodel.SessionDayTime{
			Day:  dayName,
			Time: timePart,
		})
	}

	return distinctDayTimes, nil
}

// BookPendingSession godoc
// @Summary This controller function handles the booking of a pending session for a student
// @description This controller will book the pending session for student.
// @description This api is taking student session date time,timezone and tutor id as postform.
// @Tags DemoSessionBooking
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param session-datetime  formData  string true "Session Datetime"
// @Param timezone formData string true "Time Zone"
// @Param tutor_id formData number true "Tutor ID"
// @Produce json
// @Success 200
// @Router /session/pending-book [POST]
func BookPendingSession(c *gin.Context) {
	// Retrieve the userID from the context
	userID, exists := c.Get("userID")
	if !exists {
		// If userID is not present in the context, log an error and return a "User Not Found" response
		log.Println("BookPendingSession: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	requestData, err := GetRequestDataForSessionBooking(c)
	if err != nil {
		log.Println("BookPendingSession: Failed to fetch mandatory fields for booking student's pending sessions with err :", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
	}
	requestData.SessionType = constants.SessionTypePaid
	// Fetch student's details using parent's id (user id).
	studentsList, err := students.GetChildOfUser(userID.(int))
	if err != nil {
		log.Println("[ERROR] BookPendingSession : Failed to get student id from database with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	var (
		studentEmail, studentName, studentPhone string
	)
	for _, student := range studentsList {
		phone := student.ParentPhoneDialingCode + student.ParentPhone
		if !strings.HasPrefix(phone, "+") {
			phone = fmt.Sprintf("+%s", phone)
		}
		requestData.StudentID = int(student.ID)
		studentName = student.StudentName
		studentEmail = student.ParentEmail
		studentPhone = phone
	}
	// Fetch tutor's details using tutor's id.
	tutorDetails, err := tutor.GetTutorByID(requestData.TutorID)
	if err != nil {
		log.Println("[ERROR] BookPendingSession: Failed to get tutor details with error: ", err)
	}
	endDateStr, err := students.GetActiveSubscriptionDetailsRenewalDate(userID.(int))
	if err != nil {
		log.Println("[ERROR] BookPendingSession : Failed while try to get end date of subscription of student with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	var sessionID int
	log.Println("endDate", endDateStr)
	var endDate time.Time
	if len(endDateStr) != 0 {
		layout := time.RFC3339
		endDate, err = time.Parse(layout, endDateStr)
		if err != nil {
			log.Println("BookPendingSession [ERROR]: Failed while try to parse string into time.Time with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
	}
	if time.Now().After(endDate) {
		// If user's subscription plan expire then student can book only cancelled session which marks by tutor.
		log.Println("[Subscription paln Expire] BookPendingSession : Your subscription paln expire.")
		err, sessionID = sessionmodel.SaveDemoSessionInfoOfStudent(userID.(int), requestData)
		if err != nil {
			log.Println("[ERROR] BookPendingSession: Failed to save student information with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		} else {
			err := sessionmodel.DeleteCancelledSession(userID.(int))
			if err != nil {
				log.Println("[ERROR] BookPendingSession: Failed to delete student's cancelled pending session information with error: ", err)
				controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
				return
			}
		}

	} else {
		err, sessionID = sessionmodel.SaveDemoSessionInfoOfStudent(userID.(int), requestData)
		if err != nil {
			log.Println("[ERROR] BookPendingSession: Failed to save student information with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		} else {
			err := sessionmodel.DeletePendingSession(userID.(int))
			if err != nil {
				log.Println("[ERROR] BookPendingSession: Failed to delete student's active or absent pending session information with error: ", err)
				controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
				return
			}
		}
	}
	var remainingPendingSession int64
	remainingPendingSession, err = sessionmodel.CountRemainingPendingSessionOfStudent(userID.(int))
	if err != nil {
		log.Println("[ERROR] BookPendingSession: Failed to fetch count of student's pending session with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	subject, err := metadata.GetSubjectById(requestData.SubjectID)
	if err != nil {
		log.Println("[ERROR BookPendingSession : Failed to fetch subject details from database", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	studentData := messageservices.StudentData{
		ID:          int64(userID.(int)),
		Email:       studentEmail,
		Name:        studentName,
		PhoneNumber: studentPhone,
	}

	teacherData := messageservices.TeacherData{
		ID:       int64(requestData.TutorID),
		Email:    tutorDetails.TutorEmail,
		Timezone: tutorDetails.Timezone,
		Name:     tutorDetails.TutorName,
	}

	sessionData := messageservices.SessionData{
		ID:               int64(sessionID),
		StartTime:        requestData.SessionDateTime,
		Timezone:         requestData.TimeZone,
		SessionType:      constants.SessionTypePaid,
		MeetingLink:      tutorDetails.ZoomLink,
		RemainingSession: remainingPendingSession,
	}
	// sending pending session booking notification
	go messageservices.SendEmailForPendingSessionBooking(studentData, sessionData)

	// Get the current time
	currentTime := time.Now()
	// Calculate the difference
	timeDifference := requestData.SessionDateTime.Sub(currentTime)
	// Check if the session is within the next 24 hours
	if timeDifference > 0 && timeDifference <= 24*time.Hour {
		log.Println("The session is within the next 24 hours.")
		// schedule session reminder for paid session booking
		go messageservices.ScheduleMsgsForSessionBook(studentData, teacherData, sessionData)

	} else {
		log.Println("[NOTE] Reminder scheduled via cron: The session's scheduled time is not within the next 24 hours.")
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "Your pending session booked successfully.",
		"session_id":   sessionID,
		"session_time": requestData.SessionDateTime,
		"timezone":     requestData.TimeZone,
		"subject_name": subject.Name,
	})

}

func SessionCancelIfNotConfirmed(c *gin.Context) {

	dateTime := c.PostForm("date_time")

	var sessionsConfirmationDetails []twiliowebhook.SessionConfirmationDetails
	sessionCancelMessageInfo, err := sessionmodel.GetSessionsToCancelIfNotConfirmed(dateTime)
	if err != nil {
		log.Println("[ERROR] SessionCancelIfNotConfirmed: Failed to get sessions to cancel with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	var sessionIDArr []string

	for sessionID, session := range sessionCancelMessageInfo {
		sessionIDArr = append(sessionIDArr, strconv.FormatInt(sessionID, 10))

		phone := session.Student.User.DialingCode + session.Student.User.Phone
		if !strings.HasPrefix(phone, "+") {
			phone = fmt.Sprintf("+%s", phone)
		}

		sessionsConfirmationDetails = append(sessionsConfirmationDetails, twiliowebhook.SessionConfirmationDetails{
			TwilioSMSDetails: twiliowebhook.TwilioSMSDetails{
				Phone: phone,
			},
			User:                session.Student.User,
			Session:             session,
			AttendanceConfirmed: false,
		})

		// If session is paid, save the current session details in the exceptions table
		if session.SessionType == constants.SessionTypePaid {
			err = sessionmodel.SaveExceptionSessionOfStudent(session)
			if err != nil {
				log.Println("[ERROR] TwilioWebhook: Failed to save session details in the database:", err)
				controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
				return
			}
		}

	}
	sessionIDs := strings.Join(sessionIDArr, ",")
	if len(sessionIDArr) != 0 {
		err = sessionmodel.UpdateSessionStatusIfNotConfirmed(sessionIDs, constants.SessionStatusCancel)
		if err != nil {
			log.Println("[ERROR] SessionCancelIfNotConfirmed: Failed to update sessions as cancelled with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}

		go messageservices.RescheduleMsgsToSendSessionNotConfirmedSMS(sessionsConfirmationDetails)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "successfully updated session status",
	})
}

// Mock structure for SessionPreference; replace with your actual model.
type StudentSessionPreferenceResponse struct {
	ID        int    `json:"id"`
	UserID    int64  `json:"user_id"`
	TutorID   int64  `json:"tutor_id"`
	DayOfWeek string `json:"day_of_week"`
	StartTime string `json:"start_time"`
	StartDate string `json:"start_date"`
	Timezone  string `json:"timezone"`
	StudentID int64  `json:"student_id"`
	SubjectID int64  `json:"subject_id"`
}

type StudentSessionPreferenceRequest struct {
	Timezone    string       `json:"timezone" binding:"required"`
	TutorID     int64        `json:"tutor_id"`
	StudentID   int64        `json:"student_id"`
	SubjectID   int64        `json:"subject_id" binding:"required"`
	Preferences []Preference `json:"preferences" binding:"required,dive"`
}

type Preference struct {
	DayOfWeek string `json:"day_of_week"`
	StartTime string `json:"start_time"`
	StartDate string `json:"start_date"`
}

// GetSessionPreferenceById retrieves a single session preference by its ID.
func GetSessionPreferenceById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		log.Println("[ERROR] GetSessionPreferenceById: Failed to convert UserID to int64. Error:", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidSessionID, err, c)
		return
	}
	// Fetching session preference from the database by ID
	preference, err := sessionmodel.GetSessionPreferenceById(id, 0)
	if err != nil {
		log.Println("[ERROR] GetSessionPreferenceById: Failed to get session preference with ID:", id, "Error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	if len(preference) == 0 {
		log.Println("[ERROR] GetSessionPreferenceById: Failed to get session preference with ID:", id, "Error:", err)
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}
	formattedStartTime, _ := utility.GetTime(preference[0].StartTime)

	// If found, send the response
	c.JSON(http.StatusOK, gin.H{
		"data": StudentSessionPreferenceResponse{
			ID:        preference[0].ID,
			UserID:    preference[0].UserID,
			TutorID:   preference[0].TutorID,
			DayOfWeek: preference[0].DayOfWeek,
			StartTime: formattedStartTime,
			StartDate: preference[0].StartDate,
			Timezone:  preference[0].Timezone,
			StudentID: preference[0].StudentID,
			SubjectID: preference[0].SubjectID,
		},
	})
}

// GetAllSessionPreferences retrieves all session preferences of a USER.
func GetAllSessionPreferences(c *gin.Context) {

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		log.Println("[ERROR] GetAllSessionPreferences: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	uid := userID.(int)
	allpreferences, err := sessionmodel.GetAllSessionPreferences(uid)
	if err != nil {
		log.Println("[ERROR] GetAllSessionPreferences: Failed to get all sessions with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Mapping all preferences to response struct
	var sessionPreferences []StudentSessionPreferenceResponse
	for _, preference := range allpreferences {

		formattedStartTime, _ := utility.GetTime(preference.StartTime)

		sessionPreferences = append(sessionPreferences, StudentSessionPreferenceResponse{
			ID:        preference.ID,
			UserID:    preference.UserID,
			TutorID:   preference.TutorID,
			DayOfWeek: preference.DayOfWeek,
			StartTime: formattedStartTime,
			StartDate: preference.StartDate,
			Timezone:  preference.Timezone,
			StudentID: preference.StudentID,
			SubjectID: preference.SubjectID,
		})
	}

	// Return the session preferences as JSON response
	c.JSON(http.StatusOK, gin.H{
		"data": sessionPreferences,
	})
}

func AddSessionPreference(c *gin.Context) {
	var request StudentSessionPreferenceRequest

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		log.Println("[ERROR] AddSessionPreference: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	uid := userID.(int)

	// Bind JSON body to the request struct
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("[ERROR] AddSessionPreference: Failed to bind JSON data:", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToBindJsonRespoce, err, c)
		return
	}

	// Initialize a slice to hold the session preferences in the correct struct format
	var sessionPreferences []sessionmodel.StudentSessionPreference

	// Convert each incoming preference into a session preference struct
	for _, pref := range request.Preferences {
		sessionPreference := sessionmodel.StudentSessionPreference{
			UserID:    int64(uid),
			TutorID:   request.TutorID,
			DayOfWeek: pref.DayOfWeek,
			StartTime: pref.StartTime,
			StartDate: pref.StartDate,
			Timezone:  request.Timezone,
			StudentID: request.StudentID,
			SubjectID: request.SubjectID,
		}

		// Append the session preference to the slice
		sessionPreferences = append(sessionPreferences, sessionPreference)
	}

	// Call the model to insert the new session preferences into the database
	ids, err := sessionmodel.AddSessionPreference(sessionPreferences)
	if err != nil {
		log.Println("[ERROR] AddSessionPreference: Failed to add session preferences with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// If successful, return the created IDs
	c.JSON(http.StatusOK, gin.H{
		"message": "Session preferences added successfully",
		"ids":     ids,
	})
}

func UpdateSessionPreference(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		log.Println("[ERROR] UpdateSessionPreference: Failed to convert UserID to int64. Error:", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidSessionID, err, c)
		return
	}

	// Bind the request body to the StudentSessionPreferenceRequest struct
	var request StudentSessionPreferenceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		// If binding fails, return a bad request response with the error message
		log.Println("[ERROR] UpdateSessionPreference: Failed to bindJSON data with ", err)
		controllers.HandleJSONErrorResponse(apierror.FailedToBindJsonRespoce, err, c)
		return
	}

	// Create a new StudentSessionPreference struct with the updated data
	sessionPreference := sessionmodel.StudentSessionPreference{
		ID:        int(id),
		TutorID:   request.TutorID,
		DayOfWeek: request.Preferences[0].DayOfWeek,
		StartTime: request.Preferences[0].StartTime,
		StartDate: request.Preferences[0].StartDate,
		Timezone:  request.Timezone,
		StudentID: request.StudentID,
		SubjectID: request.SubjectID,
	}

	// Call the model function to update the session preference in the database
	updatedID, err := sessionmodel.UpdateSessionPreference(sessionPreference)
	if err != nil {
		log.Println("[ERROR] UpdateSessionPreference: Failed to add session preference with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// If successful, return the updated ID and a success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Session preference updated successfully",
		"id":      updatedID,
	})
}

func DeleteSessionPreference(c *gin.Context) {

	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		log.Println("[ERROR] DeleteSessionPreference: Failed to convert UserID to int64. Error:", err)
		controllers.HandleJSONErrorResponse(apierror.InvailidSessionID, err, c)
		return
	}

	// Call the delete function and handle errors
	deletedID, err := sessionmodel.DeleteSessionPreferenceById(id)
	if err != nil {
		log.Println("[ERROR] DeleteSessionPreference: Failed to add session preference with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return

	}

	// Respond with success (if deletion succeeded)
	c.JSON(http.StatusOK, gin.H{"message": "Session preference deleted successfully",
		"deletedId": deletedID,
	})

}

// BookSessionAfterPaymentSetup books sessions for a user based on their preferences after payment is set up.
// The function finds available slots for the student's chosen tutor within a specific grace period, defined by the .env file.
// If an error occurs during any step, the function logs it and returns the error.

func BookSessionAfterPaymentSetup(userID int) error {

	// Retrieve grace period from environment variable, setting to 27 days if not found or invalid.
	gracePeriod, err := strconv.Atoi(os.Getenv("GRACE_PERIOD"))
	if err != nil || gracePeriod < 0 {
		gracePeriod = 27 // Default to 27 days grace period if invalid
	}

	// Get the active subscription plan details for the user.
	plan, err := stripe.GetActivePlanDetails(userID)
	if err != nil {
		log.Println("BookSessionAfterPaymentSetup: Failed to fetch active subscription plan slots for given user id with error:", err)
		return err
	}

	// Get the student's preferred day and time for sessions.
	sessionDayTime, err := sessionmodel.GetAllSessionPreferences(userID)
	if err != nil {
		log.Println("BookSessionAfterPaymentSetup: Failed to get student session preferences with error:", err)
		return err
	}
	if len(sessionDayTime) == 0 && (sessionDayTime[0].TutorID == 0 || sessionDayTime[0].DayOfWeek == "" || sessionDayTime[0].StartTime == "") {
		err = errors.New("[SESSION_PREFERENCE_NOT_SET] BookSessionAfterPaymentSetup: Cannot proceed with booking regular sessions. The user has not provided any session day or time preferences, which are required to allocate a session slot. Please advise the user to set their preferred day and time for sessions in their account settings.")
		log.Println(err)
		return err
	}

	// Load student's timezone for date calculations.
	loc, err := time.LoadLocation(sessionDayTime[0].Timezone)

	// Parse the StartDate string into a time.Time object
	parsedStartDate, err := time.Parse("2006-01-02T15:04:05Z", sessionDayTime[0].StartDate)
	// If "start_date" is not saved in our application then we are adding default date i.e. todays date.
	if err != nil {

		parsedStartDate = time.Now().In(loc).UTC()
	}

	// Convert to the specified timezone and then to UTC
	startDate := parsedStartDate.In(loc).UTC()
	startDatestr := startDate.Format("2006-01-02")

	var (
		tutorID, subjectID, studentID int
		timeZone                      string
		validUntilTime                time.Time
	)

	// Calculate the date until which the sessions are valid, based on the grace period.
	validUntilTime = startDate.AddDate(0, 0, gracePeriod) // Example: If today is Nov 1, and gracePeriod is 27, this becomes Nov 28.

	var sessionPreferenceDayNTime []sessionmodel.SessionDayTime
	for _, s := range sessionDayTime {
		parsedTime, _ := time.Parse("2006-01-02T15:04:05Z", s.StartTime)

		// Format the time as "01:30:00"
		formattedTime := parsedTime.Format("15:04:05")
		fmt.Println("Formatted time:", formattedTime)
		// Convert each session preference time to UTC.
		UTCDay, UTCTime, err := utility.ConvertTimeInUTC(s.DayOfWeek, formattedTime, s.Timezone)
		if err != nil {
			log.Println("BookSessionAfterPaymentSetup : Failed to convert session preference time in UTC with error :", err)
		}

		// Assign session-related details.
		tutorID = int(s.TutorID)
		timeZone = s.Timezone
		subjectID = int(s.SubjectID)
		studentID = int(s.StudentID)

		// Append each session preference in UTC to the session preference list.
		sessionPreferenceDayNTime = append(sessionPreferenceDayNTime, sessionmodel.SessionDayTime{
			Time: UTCTime,
			Day:  strings.ToLower(UTCDay),
		})
	}

	var sessionSlot []sessionmodel.SessionSlots

	// Fetch available tutor slots based on session preferences, maximum allowed sessions, and the calculated validity period.
	slots, err := sessionmodel.GetAvailableTutorSlotsForRegularSession(sessionPreferenceDayNTime, tutorID, plan.MaxSession, validUntilTime, startDatestr)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to fetch available tutor slots with error:", err)
		return err
	}

	// Prepare the session slot list for database saving.
	for _, startTime := range slots {
		slot := sessionmodel.SessionSlots{
			SessionStart: startTime,
		}
		sessionSlot = append(sessionSlot, slot)
	}

	// Save the regular session slots in the database.
	err = sessionmodel.SessionSaveAndCredit(userID, tutorID, subjectID, studentID, sessionSlot, timeZone)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to save student's regular session in database with error:", err)
		return err
	}

	var sessionSlotFalse []sessionmodel.SessionSlots

	// Calculate next month's start and end dates for session availability to add next month session with [FALSE] Creadit.
	nextMonthStart := validUntilTime.Format("2006-01-02")          // Start date for the next set of sessions
	nextMonthEndOfDay := validUntilTime.AddDate(0, 0, gracePeriod) // Validity end date for next month sessions

	// Fetch next month's available tutor slots based on updated period.
	creditedFalseslots, err := sessionmodel.GetAvailableTutorSlotsForRegularSession(sessionPreferenceDayNTime, tutorID, plan.MaxSession, nextMonthEndOfDay, nextMonthStart)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to fetch available tutor slots with error:", err)
		return err
	}

	// Prepare the next month's session slots for saving in the database.
	for _, startTime := range creditedFalseslots {
		slot := sessionmodel.SessionSlots{
			SessionStart: startTime,
		}
		sessionSlotFalse = append(sessionSlotFalse, slot)
	}

	// Save the next month’s regular session information in the database.
	err = sessionmodel.SaveSessions(userID, tutorID, subjectID, studentID, sessionSlotFalse, timeZone)
	if err != nil {
		log.Println("CreditSessionsForUser: Failed to save student's regular session in database with error:", err)
		return err
	}
	// Send email for session preference while booking session from ppc page
	var sessionPreferencesfDateTime []string
	for _, dayTime := range sessionDayTime {
		parsedTime, _ := time.Parse(time.RFC3339, dayTime.StartTime)
		dayTime.StartTime = parsedTime.Format("15:04:05")
		startTime := utility.GetTimeWithAMOrPM(dayTime.StartTime)
		caser := cases.Title(language.Und)                  // Initializes a Title case transformer with unspecified language (`Und`), allowing it to handle general title case formatting.
		formattedDayName := caser.String(dayTime.DayOfWeek) // Converts `dayTime.Day` to title case (e.g., "monday" becomes "Monday").
		sessionPreferencesfDateTime = append(sessionPreferencesfDateTime, fmt.Sprintf("%s, at %s %s", formattedDayName, startTime, utility.GetTimeZoneAbbreviation(timeZone)))

	}
	sessionEarliestDate, sessionEarliestTime := FindEarliestDate(sessionSlot)
	go PrepareDataAndSendEmailForRecuringPreference(userID, sessionEarliestDate, sessionEarliestTime)
	//sessionmodel.UpdateSubscriptionStartDateForUser(userID)

	return nil
}
