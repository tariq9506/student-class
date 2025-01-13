package messageservices

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"tutree/student-apis/constants"
	twiliowebhook "tutree/student-apis/controllers/twilio_webhook"
	sessionmodel "tutree/student-apis/models/sessionModel"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"
)

const TriggerTimeFormat = "2006-01-02 15:04:05.999999-07:00"

type StudentData struct {
	ID                int64
	Email             string
	PhoneNumber       string
	Name              string
	RawMessageStudent []byte
	RawSMSStudent     []byte
	Timezone          string
}

type TeacherData struct {
	ID              int64
	Email           string
	PhoneNumber     string
	Timezone        string
	Name            string
	RawMessageTutor []byte
	RawSMSTutor     []byte
}

type SessionData struct {
	ID               int64
	StartTime        time.Time
	Timezone         string
	SessionType      string
	MeetingLink      string
	RemainingSession int64
}
type RecurringSessionData struct {
	SessionEarliestTime string
	SessionEarliestDate string
}
type PaymentDataAndPlanDetail struct {
	CardBrand      string
	CardLast4      string
	CardExpYear    string
	CardExpMonth   string
	InvoiceLink    string
	PurchasedAt    string
	PlanName       string
	AmountPaid     string
	PaymentMethod  string
	PlanMaxSession int
	PerWeekLimit   int
}

type CancelSubscriptionData struct {
	UserName         string
	CancellationDate string
	CancellationTime string
	PlanName         string
}

type DataForScheduleMsgs struct {
	SessionData SessionData
	TeacherData TeacherData
	StudentData StudentData
}
type PendingSessionData struct {
	ID               int64
	StartTime        time.Time
	Timezone         string
	RemainingSession int64
}

func ScheduleMsgsForSessionBook(studentData StudentData, teacherData TeacherData, sessionData SessionData) {

	// Prepare dynamic email data for students and tutors
	pDataStudent := prepareStudentEmailData(studentData, sessionData)
	studentData.RawMessageStudent = pDataStudent

	pDataTutor := prepareTutorEmailData(studentData, teacherData, sessionData)
	teacherData.RawMessageTutor = pDataTutor

	// Prepare dynamic sms data for students and tutors
	pDataStudentSMS := prepareStudentSMSData(studentData, sessionData)
	studentData.RawSMSStudent = pDataStudentSMS

	var sendMessageEvents []services.MessageInfo

	//set message raw data

	if sessionData.SessionType == constants.SessionTypeDemo {
		demoScheduleData := composeDemoScheduleEmails(studentData, teacherData, sessionData)
		sendMessageEvents = append(sendMessageEvents, demoScheduleData...)

		demoScheduleSMS := composeDemoScheduleSMS(studentData, teacherData, sessionData)
		sendMessageEvents = append(sendMessageEvents, demoScheduleSMS...)
	}

	reminderScheduleData := composeReminderEmails(studentData, teacherData, sessionData)
	sendMessageEvents = append(sendMessageEvents, reminderScheduleData...)

	reminderScheduleSMS := composeReminderSMS(studentData, teacherData, sessionData)
	sendMessageEvents = append(sendMessageEvents, reminderScheduleSMS...)

	services.CallMessageServiceWebhook(sendMessageEvents)

}

func ScheduleMsgsForSessionReschedule(studentData StudentData, teacherData TeacherData, sessionData SessionData) {

	// Prepare dynamic email data for students and tutors
	pDataStudent := prepareStudentEmailData(studentData, sessionData)
	studentData.RawMessageStudent = pDataStudent

	pDataTutor := prepareTutorEmailData(studentData, teacherData, sessionData)
	teacherData.RawMessageTutor = pDataTutor

	// Prepare dynamic sms data for students and tutors
	pDataStudentSMS := prepareStudentSMSData(studentData, sessionData)
	studentData.RawSMSStudent = pDataStudentSMS

	// pDataTutor := prepareTutorEmailData(studentData, teacherData, sessionData)
	// teacherData.RawMessageTutor = pDataTutor

	var sendMessageEvents []services.MessageInfo

	demoScheduleData := composeSessionReScheduleEmails(studentData, teacherData, sessionData)
	sendMessageEvents = append(sendMessageEvents, demoScheduleData...)

	reminderScheduleData := composeReminderEmails(studentData, teacherData, sessionData)
	sendMessageEvents = append(sendMessageEvents, reminderScheduleData...)

	//TODO
	demoScheduleSMS := composeDemoReScheduleSMS(studentData, teacherData, sessionData)
	sendMessageEvents = append(sendMessageEvents, demoScheduleSMS...)

	reminderScheduleSMS := composeReminderSMS(studentData, teacherData, sessionData)
	sendMessageEvents = append(sendMessageEvents, reminderScheduleSMS...)

	services.CallMessageServiceWebhook(sendMessageEvents)

}

// func composeStudentEmail(studentData StudentData, sessionData SessionData) services.MessageInfo {

//		message := services.MessageInfo{
//			UserID: studentData.ID,
//			MessageType:  constants.EmailMessageType,
//			EventType:    constants.DefaultEventType,
//			EmailContent: studentData.RawMessageStudent,
//			Status:       constants.PendingMessageStatus,
//			Interval:     constants.DefaultInterval,
//			Email:        studentData.Email,
//			SessionID:    sessionData.ID,
//		}
//		return message
//	}
func composeStudentSMS(studentData StudentData, sessionData SessionData, eventType string) services.MessageInfo {

	message := services.MessageInfo{
		UserID:      studentData.ID,
		MessageType: constants.SmsMessageType,
		EventType:   eventType,
		Status:      constants.PendingMessageStatus,
		Interval:    constants.DefaultInterval,
		PhoneNumber: studentData.PhoneNumber,
		SessionID:   sessionData.ID,
		SMSContent:  studentData.RawSMSStudent,
	}

	return message
}
func composeDemoScheduleSMSForStudent(studentData StudentData, sessionData SessionData) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	message := services.MessageInfo{
		UserID:      studentData.ID,
		MessageType: constants.SmsMessageType,
		EventType:   constants.DefaultEventType,
		TriggerTime: triggerTime,
		Status:      constants.PendingMessageStatus,
		Interval:    constants.DefaultInterval,
		PhoneNumber: studentData.PhoneNumber,
		SessionID:   sessionData.ID,
		SMSContent:  studentData.RawSMSStudent,
	}

	return message
}

func composeDemoScheduleEmailsForStudent(studentData StudentData, sessionData SessionData) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	message := services.MessageInfo{
		UserID: studentData.ID,

		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailContent: studentData.RawMessageStudent,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        studentData.Email,
		SessionID:    sessionData.ID,
	}

	return message
}

func composeDemoScheduleEmailsForTutor(teacherData TeacherData, sessionData SessionData) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	teacherMessageData := services.MessageInfo{
		TutorID:      teacherData.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailType:    constants.CongratulationTutorEmailOnDemoBookingMsgServ,
		EmailContent: teacherData.RawMessageTutor,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        teacherData.Email,
		SessionID:    sessionData.ID,
	}

	return teacherMessageData
}

func composeDemoScheduleEmails(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo

	studentMessage := composeDemoScheduleEmailsForStudent(studentData, sessionData)
	studentMessage.EmailType = constants.CongratulationEmailOnDemoBookingMsgServ

	tutorMessage := composeDemoScheduleEmailsForTutor(teacherData, sessionData)
	tutorMessage.EmailType = constants.CongratulationTutorEmailOnDemoBookingMsgServ

	sendMessageEvents = append(sendMessageEvents, studentMessage)
	sendMessageEvents = append(sendMessageEvents, tutorMessage)

	return sendMessageEvents
}

func composeSessionReScheduleEmails(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo

	studentMessage := composeDemoScheduleEmailsForStudent(studentData, sessionData)
	studentMessage.EmailType = constants.SessionRescheduleSuccessfullyMsgServ

	tutorMessage := composeDemoScheduleEmailsForTutor(teacherData, sessionData)
	tutorMessage.EmailType = constants.CongratulationTutorEmailOnDemoBookingMsgServ

	sendMessageEvents = append(sendMessageEvents, studentMessage)
	sendMessageEvents = append(sendMessageEvents, tutorMessage)

	return sendMessageEvents
}

// func composeDemoScheduleEmails(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

// 	var sendMessageEvents []services.MessageInfo

// 	utcTime := time.Now().UTC()
// 	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

// 	if studentData.ID != 0 || studentData.Email != "" {
// 		// composing CONGRATULATION MAIL

// 		message := services.MessageInfo{
// 			UserID:       studentData.ID,
// 			TutorID:      teacherData.ID,
// 			MessageType:  constants.EmailMessageType,
// 			EventType:    constants.DefaultEventType,
// 			EmailContent: studentData.RawMessageStudent,
// 			TriggerTime:  triggerTime,
// 			Status:       constants.PendingMessageStatus,
// 			Interval:     constants.DefaultInterval,
// 			Email:        studentData.Email,
// 			SessionID:    sessionData.ID,
// 		}
// 		if sessionData.SessionType == "demo" {
// 			message.EmailType = constants.CongratulationEmailOnDemoBookingMsgServ
// 			sendMessageEvents = append(sendMessageEvents, message)
// 		} else if sessionData.SessionType == constants.SessionTypeIfDemo || sessionData.SessionType == constants.SessionTypeIfPaid {
// 			fmt.Println("composed reschedule mail")
// 			message.EmailType = constants.SessionRescheduleSuccessfullyMsgServ
// 			sendMessageEvents = append(sendMessageEvents, message)
// 		}
// 	}

// 	if teacherData.ID != 0 || teacherData.Email != "" {
// 		// composing CONGRATULATION MAIL

// 		teacherMessageData := services.MessageInfo{
// 			UserID:       studentData.ID,
// 			TutorID:      teacherData.ID,
// 			MessageType:  constants.EmailMessageType,
// 			EventType:    constants.DefaultEventType,
// 			EmailType:    constants.CongratulationTutorEmailOnDemoBookingMsgServ,
// 			EmailContent: teacherData.RawMessageTutor,
// 			TriggerTime:  triggerTime,
// 			Status:       constants.PendingMessageStatus,
// 			Interval:     constants.DefaultInterval,
// 			Email:        teacherData.Email,
// 			SessionID:    sessionData.ID,
// 		}
// 		sendMessageEvents = append(sendMessageEvents, teacherMessageData)
// 	}
// 	return sendMessageEvents
// }

func composeReminderEmails(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo
	triggerTime := ""
	sessionBeginTime := sessionData.StartTime
	currentTimeInUTC := time.Now().UTC()
	var message services.MessageInfo

	if studentData.ID != 0 || studentData.Email != "" {
		// composing REMINDER EMAIL 24hrs

		if currentTimeInUTC.Add(24 * time.Hour).Before(sessionBeginTime) {
			t := sessionData.StartTime.Add(-24 * time.Hour)
			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			message = services.MessageInfo{
				UserID:       studentData.ID,
				MessageType:  constants.EmailMessageType,
				EventType:    constants.ScheduledEventType,
				EmailContent: studentData.RawMessageStudent,
				TriggerTime:  triggerTime,
				Status:       constants.PendingMessageStatus,
				Interval:     constants.DefaultInterval,
				Email:        studentData.Email,
				EmailType:    constants.SessionReminderEmailBefore24HourMsgServ,
				SessionID:    sessionData.ID,
			}
			sendMessageEvents = append(sendMessageEvents, message)
		}

		// composing REMINDER EMAIL 2hrs
		if currentTimeInUTC.Add(2 * time.Hour).Before(sessionBeginTime) {
			t := sessionData.StartTime.Add(-2 * time.Hour)

			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			message = services.MessageInfo{
				UserID:       studentData.ID,
				MessageType:  constants.EmailMessageType,
				EventType:    constants.ScheduledEventType,
				EmailContent: studentData.RawMessageStudent,
				TriggerTime:  triggerTime,
				Status:       constants.PendingMessageStatus,
				Interval:     constants.DefaultInterval,
				Email:        studentData.Email,
				EmailType:    constants.SessionReminderEmailBeforeTwoHourMsgServ,
				SessionID:    sessionData.ID,
			}
			sendMessageEvents = append(sendMessageEvents, message)
		}

		// composing 15 MIN BEFORE REMINDER EMAIL
		if currentTimeInUTC.Add(15 * time.Minute).Before(sessionBeginTime) {
			t := sessionData.StartTime.Add(-15 * time.Minute)
			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			message = services.MessageInfo{
				UserID:       studentData.ID,
				MessageType:  constants.EmailMessageType,
				EventType:    constants.ScheduledEventType,
				EmailContent: studentData.RawMessageStudent,
				TriggerTime:  triggerTime,
				Status:       constants.PendingMessageStatus,
				Interval:     constants.DefaultInterval,
				Email:        studentData.Email,
				EmailType:    constants.ReminderEmailOfDemoSessionMsgServ,
				SessionID:    sessionData.ID,
			}

			sendMessageEvents = append(sendMessageEvents, message)
		}

		// composing mail at 9:00 AM
		// currentDate := currentTimeInUTC.Truncate(24 * time.Hour)

		// convert startTime in users timezone

		timeInUserTimezone, err := utility.ConvertTimeZone(sessionData.StartTime, sessionData.Timezone)
		if err != nil {
			log.Println("composeReminderEmails: failed to convert time in user location with error: ", err)
		}
		fmt.Println("timeInUserTimezone", timeInUserTimezone)

		timeAt9AMInUserTimezone := time.Date(timeInUserTimezone.Year(), timeInUserTimezone.Month(), timeInUserTimezone.Day(), 9, 0, 0, 0, timeInUserTimezone.Location())
		fmt.Println("timeAt9AMInUserTimezone", timeAt9AMInUserTimezone)
		triggerTime = timeAt9AMInUserTimezone.In(time.UTC).Format("2006-01-02 15:04:05.999999-07:00")

		fmt.Println("triggerTime", triggerTime)
		timeAt9AMInUTC := timeAt9AMInUserTimezone.In(time.UTC)

		fmt.Println("timeAt9AMInUTC", timeAt9AMInUTC)

		if currentTimeInUTC.Before(timeAt9AMInUTC) && sessionBeginTime.After(timeAt9AMInUTC) {
			fmt.Println("inside if statement")
			message = services.MessageInfo{
				UserID:       studentData.ID,
				MessageType:  constants.EmailMessageType,
				EventType:    constants.ScheduledEventType,
				EmailContent: studentData.RawMessageStudent,
				TriggerTime:  triggerTime,
				Status:       constants.PendingMessageStatus,
				Interval:     constants.DefaultInterval,
				Email:        studentData.Email,
				EmailType:    constants.SessionReminderEmailAT9AMMsgServ,
				SessionID:    sessionData.ID,
			}
			sendMessageEvents = append(sendMessageEvents, message)
		}

	}
	return sendMessageEvents
}

func prepareStudentEmailData(studentData StudentData, sessionData SessionData) []byte {

	sessionTypeDisplayText := "Regular Class"
	if sessionData.SessionType == constants.SessionTypeDemo {
		sessionTypeDisplayText = "Trial-Class"

	}
	formatedDate, formatedTime := utility.FormatToReadableDateTime(sessionData.StartTime, sessionData.Timezone)

	content := map[string]interface{}{
		"session_date":     formatedDate,
		"session_time":     formatedTime,
		"name":             studentData.Name,
		"event_start_time": sessionData.StartTime,
		"reschedule_url":   utility.GetHostURL() + "/student/dashboard/demo-reschedule",
		"session_type":     sessionTypeDisplayText,
		"meeting_link":     sessionData.MeetingLink,
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("composeReminderSMS: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return emailContent
}
func prepareStudentSMSData(studentData StudentData, sessionData SessionData) []byte {

	sessionTypeDisplayText := "Regular Class"
	if sessionData.SessionType == constants.SessionTypeDemo {
		sessionTypeDisplayText = "Trial-Class"

	}

	formatedDate, formatedTime := utility.FormatToReadableDateTime(sessionData.StartTime, sessionData.Timezone)

	content := map[string]interface{}{
		"{student_name}":   studentData.Name,
		"{meeting_link}":   sessionData.MeetingLink,
		"{session_type}":   sessionTypeDisplayText,
		"{date}":           formatedDate,
		"{time}":           formatedTime,
		"{reschedule_url}": fmt.Sprintf("%s/student/dashboard/reschedule?sessionId=%d", utility.GetHostURL(), sessionData.ID),
	}

	smsContent, err := json.Marshal(content)
	if err != nil {
		log.Println("composeReminderSMS: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return smsContent

}

func prepareTutorEmailData(studentData StudentData, teacherData TeacherData, sessionData SessionData) []byte {

	sessionTypeDisplayText := "Regular Class"
	if sessionData.SessionType == constants.SessionTypeDemo {
		sessionTypeDisplayText = "Trial-Class"

	}
	formattedDateTutor, formattedTimeTutor := utility.FormatToReadableDateTime(sessionData.StartTime, teacherData.Timezone)

	content := map[string]interface{}{
		"session_date":     formattedDateTutor,
		"session_time":     formattedTimeTutor,
		"kid_name":         studentData.Name,
		"session_type":     sessionTypeDisplayText,
		"tutor_name":       teacherData.Name,
		"event_start_time": sessionData.StartTime,
		"tutor_dashboard":  utility.GetTutorHostURL() + "/dashboard",
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("composeReminderSMS: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return emailContent
}

func ScheduleReminderMessages(studentData StudentData, teacherData TeacherData, sessionData SessionData) {

}

func composeDemoReScheduleSMS(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo

	studentSMS := composeDemoScheduleSMSForStudent(studentData, sessionData)
	studentSMS.SMSType = constants.SessionRescheduleSuccessfullySMS
	sendMessageEvents = append(sendMessageEvents, studentSMS)

	//TODO: TUTOR SMS

	return sendMessageEvents
}

func composeDemoScheduleSMS(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo

	studentSMS := composeDemoScheduleSMSForStudent(studentData, sessionData)
	studentSMS.SMSType = constants.CongratulationSMSOnDemoBooking
	sendMessageEvents = append(sendMessageEvents, studentSMS)

	//TODO: TUTOR SMS

	return sendMessageEvents
}

func composeReminderSMS(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo
	triggerTime := ""
	sessionBeginTime := sessionData.StartTime
	currentTimeInUTC := time.Now().UTC()
	//var message services.MessageInfo

	if studentData.ID != 0 || studentData.PhoneNumber != "" {
		// composing REMINDER SMS 24hrs

		if currentTimeInUTC.Add(24 * time.Hour).Before(sessionBeginTime) {
			t := sessionData.StartTime.Add(-24 * time.Hour)
			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			msg := composeStudentSMS(studentData, sessionData, constants.ScheduledEventType)
			msg.SMSType = constants.SessionReminderSMSBefore24Hour
			msg.TriggerTime = triggerTime
			sendMessageEvents = append(sendMessageEvents, msg)
		}

		if currentTimeInUTC.Add(4 * time.Hour).Before(sessionBeginTime) {

			t := sessionData.StartTime.Add(-4 * time.Hour)
			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			msg := composeStudentSMS(studentData, sessionData, constants.ScheduledEventType)
			msg.SMSType = constants.SessionReminderSMSBeforeFourHour
			msg.TriggerTime = triggerTime
			sendMessageEvents = append(sendMessageEvents, msg)

		}

		// composing REMINDER EMAIL 2hrs
		if currentTimeInUTC.Add(2 * time.Hour).Before(sessionBeginTime) {

			t := sessionData.StartTime.Add(-2 * time.Hour)
			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			msg := composeStudentSMS(studentData, sessionData, constants.ScheduledEventType)
			msg.SMSType = constants.SessionReminderSMSBeforeTwoHour
			msg.TriggerTime = triggerTime
			sendMessageEvents = append(sendMessageEvents, msg)

		}

		// composing 15 MIN BEFORE REMINDER EMAIL
		if currentTimeInUTC.Add(15 * time.Minute).Before(sessionBeginTime) {
			t := sessionData.StartTime.Add(-15 * time.Minute)
			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
			msg := composeStudentSMS(studentData, sessionData, constants.ScheduledEventType)
			msg.SMSType = constants.ReminderSMSOfDemoSession
			msg.TriggerTime = triggerTime
			sendMessageEvents = append(sendMessageEvents, msg)

		}

		// composing mail at 9:00 AM
		// currentDate := currentTimeInUTC.Truncate(24 * time.Hour)

		// convert startTime in users timezone

		timeInUserTimezone, err := utility.ConvertTimeZone(sessionData.StartTime, sessionData.Timezone)
		if err != nil {
			log.Println("composeReminderEmails: failed to convert time in user location with error: ", err)
		}
		fmt.Println("timeInUserTimezone", timeInUserTimezone)

		timeAt9AMInUserTimezone := time.Date(timeInUserTimezone.Year(), timeInUserTimezone.Month(), timeInUserTimezone.Day(), 9, 0, 0, 0, timeInUserTimezone.Location())
		fmt.Println("timeAt9AMInUserTimezone", timeAt9AMInUserTimezone)
		//triggerTime = timeAt9AMInUserTimezone.In(time.UTC).Format("2006-01-02 15:04:05.999999-07:00")

		timeAt9AMInUTC := timeAt9AMInUserTimezone.In(time.UTC)

		fmt.Println("timeAt9AMInUTC", timeAt9AMInUTC)

		if currentTimeInUTC.Before(timeAt9AMInUTC) && sessionBeginTime.After(timeAt9AMInUTC) {
			triggerTime = timeAt9AMInUserTimezone.In(time.UTC).Format("2006-01-02 15:04:05.999999-07:00")
			fmt.Println("triggerTime", triggerTime)
			msg := composeStudentSMS(studentData, sessionData, constants.ScheduledEventType)
			msg.SMSType = constants.SessionReminderSMSAT9AM
			msg.TriggerTime = triggerTime
			sendMessageEvents = append(sendMessageEvents, msg)

		}

	}
	return sendMessageEvents
}

// func composeReminderSMS(studentData StudentData, teacherData TeacherData, sessionData SessionData) []services.MessageInfo {

// 	var sendMessageEvents []services.MessageInfo
// 	triggerTime := ""
// 	sessionBeginTime := sessionData.StartTime
// 	currentTimeInUTC := time.Now().UTC()
// 	var message services.MessageInfo

// 	if studentData.ID != 0 || studentData.PhoneNumber != "" {
// 		// composing REMINDER SMS 24hrs

// 		if currentTimeInUTC.Add(24 * time.Hour).Before(sessionBeginTime) {
// 			t := sessionData.StartTime.Add(-24 * time.Hour)
// 			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
// 			message = services.MessageInfo{
// 				UserID:      studentData.ID,
// 				TutorID:     teacherData.ID,
// 				MessageType: constants.SmsMessageType,
// 				EventType:   constants.DefaultEventType,
// 				TriggerTime: triggerTime,
// 				Status:      constants.PendingMessageStatus,
// 				Interval:    constants.DefaultInterval,
// 				PhoneNumber: studentData.PhoneNumber,
// 				SessionID:   sessionData.ID,
// 				SMSContent:  studentData.RawSMSStudent,
// 				SMSType:     constants.SessionReminderSMSBefore24Hour,
// 			}
// 			sendMessageEvents = append(sendMessageEvents, message)
// 		}

// 		// composing REMINDER EMAIL 2hrs
// 		if currentTimeInUTC.Add(2 * time.Hour).Before(sessionBeginTime) {

// 			t := sessionData.StartTime.Add(-2 * time.Hour)

// 			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
// 			message = services.MessageInfo{
// 				UserID:      studentData.ID,
// 				TutorID:     teacherData.ID,
// 				MessageType: constants.SmsMessageType,
// 				EventType:   constants.DefaultEventType,
// 				TriggerTime: triggerTime,
// 				Status:      constants.PendingMessageStatus,
// 				Interval:    constants.DefaultInterval,
// 				PhoneNumber: studentData.PhoneNumber,
// 				SessionID:   sessionData.ID,
// 				SMSContent:  studentData.RawSMSStudent,
// 				SMSType:     constants.SessionReminderSMSBeforeTwoHour,
// 			}
// 			sendMessageEvents = append(sendMessageEvents, message)
// 		}

// 		// composing 15 MIN BEFORE REMINDER EMAIL
// 		if currentTimeInUTC.Add(15 * time.Minute).Before(sessionBeginTime) {
// 			t := sessionData.StartTime.Add(-15 * time.Minute)
// 			triggerTime = t.Format("2006-01-02 15:04:05.999999-07:00")
// 			message = services.MessageInfo{
// 				UserID:      studentData.ID,
// 				TutorID:     teacherData.ID,
// 				MessageType: constants.SmsMessageType,
// 				EventType:   constants.DefaultEventType,
// 				TriggerTime: triggerTime,
// 				Status:      constants.PendingMessageStatus,
// 				Interval:    constants.DefaultInterval,
// 				PhoneNumber: studentData.PhoneNumber,
// 				SessionID:   sessionData.ID,
// 				SMSContent:  studentData.RawSMSStudent,
// 				SMSType:     constants.ReminderSMSOfDemoSession,
// 			}

// 			sendMessageEvents = append(sendMessageEvents, message)
// 		}

// 		// composing mail at 9:00 AM
// 		// currentDate := currentTimeInUTC.Truncate(24 * time.Hour)

// 		// convert startTime in users timezone

// 		timeInUserTimezone, err := utility.ConvertTimeZone(sessionData.StartTime, sessionData.Timezone)
// 		if err != nil {
// 			log.Println("composeReminderEmails: failed to convert time in user location with error: ", err)
// 		}
// 		fmt.Println("timeInUserTimezone", timeInUserTimezone)

// 		timeAt9AMInUserTimezone := time.Date(timeInUserTimezone.Year(), timeInUserTimezone.Month(), timeInUserTimezone.Day(), 9, 0, 0, 0, timeInUserTimezone.Location())
// 		fmt.Println("timeAt9AMInUserTimezone", timeAt9AMInUserTimezone)
// 		triggerTime = timeAt9AMInUserTimezone.In(time.UTC).Format("2006-01-02 15:04:05.999999-07:00")

// 		fmt.Println("triggerTime", triggerTime)
// 		timeAt9AMInUTC := timeAt9AMInUserTimezone.In(time.UTC)

// 		fmt.Println("timeAt9AMInUTC", timeAt9AMInUTC)

// 		if currentTimeInUTC.Before(timeAt9AMInUTC) && sessionBeginTime.After(timeAt9AMInUTC) {

// 			message = services.MessageInfo{
// 				UserID:      studentData.ID,
// 				TutorID:     teacherData.ID,
// 				MessageType: constants.SmsMessageType,
// 				EventType:   constants.DefaultEventType,
// 				TriggerTime: triggerTime,
// 				Status:      constants.PendingMessageStatus,
// 				Interval:    constants.DefaultInterval,
// 				PhoneNumber: studentData.PhoneNumber,
// 				SessionID:   sessionData.ID,
// 				SMSContent:  studentData.RawSMSStudent,
// 				SMSType:     constants.SessionReminderSMSAT9AM,
// 			}
// 			sendMessageEvents = append(sendMessageEvents, message)
// 		}

// 	}
// 	return sendMessageEvents
// }

func ConvertToUTCTime(timeStr string) (time.Time, error) {

	// Example time string in ISO 8601 forma
	// Parse the time string to a time.Time object
	// parsedTime, err := time.Parse(time.RFC3339, time)
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Println("Error parsing time:", err)
		return time.Time{}, err
	}

	// Ensure the time is in UTC
	utcTime := parsedTime.UTC()

	// Format the time in ISO 8601 format
	formattedUTC := utcTime.Format(time.RFC3339)

	log.Println("Original time string:", timeStr)
	log.Println("Parsed UTC time:", formattedUTC)
	return utcTime, nil

}

func CancelSessionReminders(studentData StudentData, teacherData TeacherData, sessionData SessionData) error {

	cancelMsgEvent := []services.CancelMsgEvent{
		{
			SessionID: sessionData.ID,
		},
	}

	err := services.CancelMessageServiceWebhook(cancelMsgEvent)
	if err != nil {
		log.Println("CancelSessionReminders: Failed to call cancel message service with error: ", err)
		return err
	}
	return nil
}

func ScheduleMsgForStatusChange() {

}

func RescheduleMsgServices(DataForScheduleMsgs []DataForScheduleMsgs) {
	for _, data := range DataForScheduleMsgs {
		err := CancelSessionReminders(data.StudentData, data.TeacherData, data.SessionData)
		if err != nil {
			log.Println("RescheduleMsgServices: Failed to call cancel message service with error: ", err)
			return
		}
		ScheduleMsgsForSessionReschedule(data.StudentData, data.TeacherData, data.SessionData)
	}

}

func ScheduleMsgServices(DataForScheduleMsgs []DataForScheduleMsgs) {
	for _, data := range DataForScheduleMsgs {
		err := CancelSessionReminders(data.StudentData, data.TeacherData, data.SessionData)
		if err != nil {
			log.Println("RescheduleMsgServices: Failed to call cancel message service with error: ", err)
			return
		}
		ScheduleMsgsForSessionBook(data.StudentData, data.TeacherData, data.SessionData)
	}

}

func ScheduleMsgsToSendSessionNotConfirmedSMS(sessionsConfirmationDetails []twiliowebhook.SessionConfirmationDetails) {

	sendMessageEvents := []services.MessageInfo{}
	for _, sessionConfirmationDetail := range sessionsConfirmationDetails {
		sendMessageEvents = twiliowebhook.ComposeSessionRescheduleIfNoResponse(sessionConfirmationDetail)
	}

	services.CallMessageServiceWebhook(sendMessageEvents)
}

func RescheduleMsgsToSendSessionNotConfirmedSMS(sessionsConfirmationDetails []twiliowebhook.SessionConfirmationDetails) {

	cancelMsgEvents := []services.CancelMsgEvent{}

	sendMessageEvents := []services.MessageInfo{}

	for _, sessionConfirmationDetail := range sessionsConfirmationDetails {
		cancelMsgEvents = append(cancelMsgEvents, services.CancelMsgEvent{
			SessionID: int64(sessionConfirmationDetail.Session.SessionID),
		})

		sendMessageEvents = twiliowebhook.ComposeSessionRescheduleIfNoResponse(sessionConfirmationDetail)
	}

	services.CancelMessageServiceWebhook(cancelMsgEvents)

	services.CallMessageServiceWebhook(sendMessageEvents)
}

func ScheduleMsgSuccessfullPaymentEmail(studentData StudentData, paymentData PaymentDataAndPlanDetail) {
	// Prepare dynamic email data for payment of students
	pDataPayment := prepareSuccessfullPaymentData(paymentData)
	studentData.RawMessageStudent = pDataPayment

	var sendMessageEvents []services.MessageInfo

	paymentEmailData := composePaymentEmailOfStudent(studentData)
	sendMessageEvents = append(sendMessageEvents, paymentEmailData)

	services.CallMessageServiceWebhook(sendMessageEvents)
}

func ScheduleSubscriptionCancellationEmail(studentData StudentData, cancelSubscriptionData CancelSubscriptionData) {
	// Prepare dynamic email data for payment of students
	pDataPayment := prepareCancellationSubscriptionData(cancelSubscriptionData)
	studentData.RawMessageStudent = pDataPayment

	var sendMessageEvents []services.MessageInfo

	paymentEmailData := composeSubscriptionCancelEmail(studentData)
	sendMessageEvents = append(sendMessageEvents, paymentEmailData)

	services.CallMessageServiceWebhook(sendMessageEvents)

}

func prepareCancellationSubscriptionData(cancelSubscriptionData CancelSubscriptionData) []byte {
	content := map[string]interface{}{
		"kid_name":    cancelSubscriptionData.UserName,
		"plan_name":   cancelSubscriptionData.PlanName,
		"cancel_date": cancelSubscriptionData.CancellationDate,
		"cancel_time": cancelSubscriptionData.CancellationTime,
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareSuccessfullPaymentData: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return emailContent

}

func composeSubscriptionCancelEmail(studentData StudentData) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	message := services.MessageInfo{
		UserID:       studentData.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailType:    constants.SubscriptionCancelEmail,
		EmailContent: studentData.RawMessageStudent,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        studentData.Email,
	}
	return message
}

func prepareSuccessfullPaymentData(paymentData PaymentDataAndPlanDetail) []byte {

	content := map[string]interface{}{
		"card_brand":          paymentData.CardBrand,
		"card_last4":          paymentData.CardLast4,
		"exp_year":            paymentData.CardExpYear,
		"exp_month":           paymentData.CardExpMonth,
		"invoice_link":        paymentData.InvoiceLink,
		"purchased_at":        paymentData.PurchasedAt,
		"plan_max_session":    paymentData.PlanMaxSession,
		"plan_per_week_limit": paymentData.PerWeekLimit,
		"plan_name":           paymentData.PlanName,
		"amount_paid":         paymentData.AmountPaid,
		"payment_method":      paymentData.PaymentMethod,
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareSuccessfullPaymentData: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return emailContent
}

func composePaymentEmailOfStudent(studentData StudentData) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	message := services.MessageInfo{
		UserID:       studentData.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailType:    constants.SuccessfullPaymentEmail,
		EmailContent: studentData.RawMessageStudent,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        studentData.Email,
	}
	return message
}
func SendEmailForRecuringSessionPreference(studentData StudentData, sessionData RecurringSessionData) {

	// Prepare dynamic email data for students and tutors
	pDataStudent := prepareStudentEmailDataForSessionPreference(studentData, sessionData)
	studentData.RawMessageStudent = pDataStudent

	var sendMessageEvents []services.MessageInfo

	//set message raw data

	reminderScheduleData := composeEmailsOfSessionPreference(studentData)
	sendMessageEvents = append(sendMessageEvents, reminderScheduleData...)

	services.CallMessageServiceWebhook(sendMessageEvents)

}
func prepareStudentEmailDataForSessionPreference(studentData StudentData, sessionData RecurringSessionData) []byte {

	content := map[string]interface{}{
		"name":         studentData.Name,
		"session_time": sessionData.SessionEarliestTime,
		"session_date": sessionData.SessionEarliestDate,
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareStudentEmailDataForSessionPreference: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return emailContent
}
func composeEmailsOfSessionPreference(studentData StudentData) []services.MessageInfo {

	var sendMessageEvents []services.MessageInfo
	triggerTime := ""
	currentTimeInUTC := time.Now().UTC()
	var message services.MessageInfo

	if studentData.ID != 0 || studentData.Email != "" {
		triggerTime = currentTimeInUTC.Format("2006-01-02 15:04:05.999999-07:00")
		message = services.MessageInfo{
			UserID:       studentData.ID,
			MessageType:  constants.EmailMessageType,
			EventType:    constants.DefaultEventType,
			EmailContent: studentData.RawMessageStudent,
			TriggerTime:  triggerTime,
			Status:       constants.PendingMessageStatus,
			Interval:     constants.DefaultInterval,
			Email:        studentData.Email,
			EmailType:    constants.NotifySessionPrefEmailServ,
		}
		sendMessageEvents = append(sendMessageEvents, message)
	}
	return sendMessageEvents
}

// ScheduleEmailForSessionReschedule schedules a reschedule notification email to be sent to a student's parent.
func ScheduleEmailForSessionReschedule(session sessionmodel.Session) {
	var studentData StudentData

	// Prepare email content for session reschedule notification
	studentData.RawMessageStudent = prepareDataForSessionReschedule(session)
	studentData.Email = session.Student.ParentEmail
	studentData.ID = int64(session.Student.User.ID)

	var sendMessageEvents []services.MessageInfo

	// Compose the message information for the email
	sessionRescheduleData := composeDataForSessionReschedule(studentData)
	sendMessageEvents = append(sendMessageEvents, sessionRescheduleData)

	// Trigger the message service webhook to send the email
	services.CallMessageServiceWebhook(sendMessageEvents)
}

// prepareDataForSessionReschedule prepares the email content for notifying a student (or their parent)
// about a rescheduled session. It formats the session details, such as date and time, and includes
// a link to the student's dashboard for additional session details.
func prepareDataForSessionReschedule(session sessionmodel.Session) []byte {

	// Generate the link to the student's dashboard
	pendingLink := utility.GetHostURL() + "/student/dashboard#classes"

	// Format session start date and time to a readable format
	formatedDate, formatedTime := utility.FormatToReadableDateTime(session.StartSession, session.TimeZone)

	// Create a map containing the dynamic data for the email
	content := map[string]interface{}{
		"kid_name":     session.StudentName,
		"session_type": session.SessionType,
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

// composeDataForSessionReschedule composes the message information required for sending a reschedule
// notification email to the student's parent. It includes the student's ID, email content, email type,
// and other metadata, such as the trigger time and email status.
func composeDataForSessionReschedule(studentData StudentData) services.MessageInfo {
	// Get the current UTC time to set the email trigger time
	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format(TriggerTimeFormat)

	// Create a MessageInfo object with all the necessary fields
	message := services.MessageInfo{
		UserID:       studentData.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailContent: studentData.RawMessageStudent,
		EmailType:    constants.SessionRescheduleEmail,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        studentData.Email,
	}

	return message
}

// ScheduleEmailForSessionReschedule schedules a reminder email or SMS to be sent to students.
func ScheduleEmailForReminderToScheduleClass(studentList []students.StudentToScheduleReminder) {
	var studentData StudentData
	var sendMessageEvents []services.MessageInfo

	for _, student := range studentList {
		studentData.ID = student.ID
		studentData.Timezone = student.Timezone

		// Check if student has an email, prepare and schedule an email reminder
		if len(student.Email) != 0 {
			studentData.RawMessageStudent = prepareDataForReminderToScheduleClass(student)
			studentData.Email = student.Email

			scheduleReminderData := composeDataForReminderToScheduleClassEmail(studentData)
			sendMessageEvents = append(sendMessageEvents, scheduleReminderData)
		}

		// Check if student has a phone number, prepare and schedule an SMS reminder
		if len(student.Phone) != 0 {
			studentData.RawSMSStudent = prepareDataForReminderToScheduleClassSMS()
			studentData.PhoneNumber = student.Phone

			sessionRescheduleData := composeDataForReminderToScheduleClassSMS(studentData)
			sendMessageEvents = append(sendMessageEvents, sessionRescheduleData)
		}
	}

	// Trigger the message service webhook to send the reminders
	services.CallMessageServiceWebhook(sendMessageEvents)
}

// prepareDataForReminderToScheduleClass generates the content for a reminder email.
func prepareDataForReminderToScheduleClass(student students.StudentToScheduleReminder) []byte {
	bookNowLink := utility.GetHostURL() + "/student/login"
	content := map[string]interface{}{
		"kid_name": student.Name,
		"book_now": bookNowLink,
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareDataForReminderToScheduleClass: Failed to marshal email content into JSON:", err)
	}

	return emailContent
}

// prepareDataForReminderToScheduleClassSMS generates the content for a reminder SMS.
func prepareDataForReminderToScheduleClassSMS() []byte {
	bookNowLink := utility.GetHostURL() + "/student/login"
	content := map[string]interface{}{
		"{book_now}": bookNowLink,
	}
	smsContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareDataForReminderToScheduleClassSMS: Failed to marshal SMS content into JSON:", err)
	}
	return smsContent
}

// composeDataForReminderToScheduleClassEmail prepares the email MessageInfo for scheduling.
func composeDataForReminderToScheduleClassEmail(studentData StudentData) services.MessageInfo {
	// Load student's timezone
	location, _ := time.LoadLocation(studentData.Timezone)
	now := time.Now().In(location)

	// Set the trigger time for tomorrow at 9:00 AM in the student's timezone
	tomorrow9AM := time.Date(
		now.Year(), now.Month(), now.Day()+1, // Tomorrow's date
		9, 0, 0, 0, // 9:00 AM
		location, // Student's timezone
	)

	// Convert the time to UTC to standardize the email trigger time
	triggerTime := tomorrow9AM.UTC().Format(TriggerTimeFormat)

	// Create a MessageInfo object with the email configuration
	message := services.MessageInfo{
		UserID:       studentData.ID,
		MessageType:  constants.EmailMessageType,
		EventType:    constants.ScheduledEventType,
		EmailContent: studentData.RawMessageStudent,
		EmailType:    constants.ReminderToScheduleClassEmailServ,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        studentData.Email,
	}

	return message
}

// composeDataForReminderToScheduleClassSMS prepares the SMS MessageInfo for scheduling.
func composeDataForReminderToScheduleClassSMS(studentData StudentData) services.MessageInfo {
	// Load student's timezone
	location, _ := time.LoadLocation(studentData.Timezone)
	now := time.Now().In(location)

	// Set the trigger time for tomorrow at 9:00 AM in the student's timezone
	tomorrow9AM := time.Date(
		now.Year(), now.Month(), now.Day()+1, // Tomorrow's date
		9, 0, 0, 0, // 9:00 AM
		location, // Student's timezone
	)

	// Convert the time to UTC to standardize the SMS trigger time
	triggerTime := tomorrow9AM.UTC().Format(TriggerTimeFormat)

	// Create a MessageInfo object with the SMS configuration
	message := services.MessageInfo{
		UserID:      studentData.ID,
		MessageType: constants.SmsMessageType,
		EventType:   constants.ScheduledEventType,
		SMSContent:  studentData.RawSMSStudent,
		SMSType:     constants.ReminderToScheduleClassSMS,
		TriggerTime: triggerTime,
		Status:      constants.PendingMessageStatus,
		Interval:    constants.DefaultInterval,
		PhoneNumber: studentData.PhoneNumber,
	}

	return message
}

// SendCurriculumViaSMSOnNewSignUP is responsible to send sms on every new signup which contain a link of downloading the pdf of
// curriculum.
func SendCurriculumViaSMSOnNewSignUP(student students.StudentToScheduleReminder) {
	var studentData StudentData
	var sendMessageEvents []services.MessageInfo
	studentData.RawSMSStudent = prepareDataForSendingCurriculum()
	studentData.PhoneNumber = student.Phone
	studentData.ID = student.ID

	sessionRescheduleData := composeSMSForSendingCurriculum(studentData)
	sendMessageEvents = append(sendMessageEvents, sessionRescheduleData)
	// Trigger the message service webhook to send the reminders
	services.CallMessageServiceWebhook(sendMessageEvents)
}

// Prepare SMS data of sending curriculum sms
func prepareDataForSendingCurriculum() []byte {
	curriculumPdfLink := utility.GetCurriculumPDFLink()
	content := map[string]interface{}{
		"{curriculum_url}": curriculumPdfLink,
	}
	smsContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareDataForSendingCurriculum: Failed to marshal SMS content into JSON:", err)
	}
	return smsContent
}
func composeSMSForSendingCurriculum(studentData StudentData) services.MessageInfo {
	now := time.Now().UTC().Add(30 * time.Minute)

	// Create a MessageInfo object with the SMS configuration
	message := services.MessageInfo{
		UserID:      studentData.ID,
		MessageType: constants.SmsMessageType,
		EventType:   constants.DefaultEventType,
		SMSContent:  studentData.RawSMSStudent,
		SMSType:     constants.SendCurriculumViaSMSOnNewSignUP,
		TriggerTime: now.Format("2006-01-02 15:04:05.999999-07:00"),
		Status:      constants.PendingMessageStatus,
		Interval:    constants.DefaultInterval,
		PhoneNumber: studentData.PhoneNumber,
	}

	return message
}

// SendEmailForPendingSessionBooking is responsible to send an email when student book any session from pending bucket.
func SendEmailForPendingSessionBooking(studentData StudentData, sessionData SessionData) {
	// Prepare dynamic email data for students and tutors
	pDataStudent := prepareStudentEmailDataForPendingSessionBooking(studentData, sessionData)
	studentData.RawMessageStudent = pDataStudent
	var sendMessageEvents []services.MessageInfo

	demoScheduleData := composePendingSessionBookingEmailsForStudent(studentData, sessionData)
	sendMessageEvents = append(sendMessageEvents, demoScheduleData)
	services.CallMessageServiceWebhook(sendMessageEvents)
}
func prepareStudentEmailDataForPendingSessionBooking(studentData StudentData, sessionData SessionData) []byte {

	formatedDate, formatedTime := utility.FormatToReadableDateTime(sessionData.StartTime, sessionData.Timezone)
	linkToRenderPendingSessionPage := fmt.Sprintf("%s/student/dashboard#classes", utility.GetHostURL())
	// clickableSessionCount := fmt.Sprintf(`<a href="%s" target="_blank">%d classes</a>`, linkToRenderPendingSessionPage, sessionData.RemainingSession)
	content := map[string]interface{}{
		"session_date":          formatedDate,
		"session_time":          formatedTime,
		"kid_name":              studentData.Name,
		"pending_link":          linkToRenderPendingSessionPage,
		"pending_session_count": sessionData.RemainingSession,
	}
	emailContent, err := json.Marshal(content)
	if err != nil {
		log.Println("prepareStudentEmailDataForPendingSessionBooking: Failed to marshal dynamic data into json.RawMessage with error: ", err)
	}
	return emailContent
}
func composePendingSessionBookingEmailsForStudent(studentData StudentData, sessionData SessionData) services.MessageInfo {

	utcTime := time.Now().UTC()
	triggerTime := utcTime.Format("2006-01-02 15:04:05.999999-07:00")

	message := services.MessageInfo{
		UserID: studentData.ID,

		MessageType:  constants.EmailMessageType,
		EventType:    constants.DefaultEventType,
		EmailType:    constants.SuccessfullyBookPendingSessionEmail,
		EmailContent: studentData.RawMessageStudent,
		TriggerTime:  triggerTime,
		Status:       constants.PendingMessageStatus,
		Interval:     constants.DefaultInterval,
		Email:        studentData.Email,
		SessionID:    sessionData.ID,
	}

	return message
}
