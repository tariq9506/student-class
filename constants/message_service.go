package constants

var (
	CongratulationEmailOnDemoBookingMsgServ       = "CongratulationEmailOnDemoBooking"
	ReminderEmailOfDemoSessionMsgServ             = "ReminderEmailOfDemoSession"
	CongratulationTutorEmailOnDemoBookingMsgServ  = "CongratulationTutorEmailOnDemoBooking"
	ReminderEmailOfPaidSessionMsgServ             = "ReminderEmailOfPaidSession"
	SuccessfulPaymentEmailMsgServ                 = "SuccessfulPaymentEmail"
	SendReminderEmailForBookRegularSessionMsgServ = "SendReminderEmailForBookRegularSession"
	SessionRescheduleSuccessfullyMsgServ          = "SessionRescheduleSuccessfully"
	SessionReminderEmailBeforeTwoHourMsgServ      = "SessionReminderEmailBeforeTwoHour"
	SessionReminderEmailBefore24HourMsgServ       = "SessionReminderEmailBefore24Hour"
	OnboardingTutorEmailMsgServ                   = "OnboardingTutorEmail"
	ReminderEmailOfDemoSessionTutorMsgServ        = "ReminderEmailOfDemoSessionTutor"
	DemoSessionCancelEmailMsgServ                 = "DemoSessionCancelEmail"
	DemoSessionStartedEmailMsgServ                = "DemoSessionStartedEmail"
	StudentAbsentDemoEmailMsgServ                 = "StudentAbsentDemoEmail"
	DemoSessionCompletedEmailMsgServ              = "DemoSessionCompletedEmail"
	SessionReminderEmailAT9AMMsgServ              = "SessionReminderEmailAT9AM"
	NotifySessionPrefEmailServ                    = "NotifySessionPrefEmail"
	ReminderToScheduleClassEmailServ              = "ReminderToScheduleClassEmail"
	ReminderToScheduleClassSMS                    = "ReminderToScheduleClassSMS"
	SendCurriculumViaSMSOnNewSignUP               = "SendCurriculumViaSMSOnNewSignUP"
	SuccessfullyBookPendingSessionEmail           = "SuccessfullyBookPendingSessionEmail"
)

var (
	CongratulationSMSOnDemoBooking         = "CongratulationSMSOnDemoBooking"
	SessionRescheduleSuccessfullySMS       = "SessionRescheduleSuccessfully"
	SessionReminderSMSBefore24Hour         = "SessionReminderSMSBefore24Hour"
	ReminderSMSOfDemoSession               = "ReminderSMSOfDemoSession"
	SessionReminderSMSBeforeTwoHour        = "SessionReminderSMSBeforeTwoHour"
	SessionReminderSMSBeforeFourHour       = "SessionReminderSMSBeforeFourHour"
	SessionReminderSMSAT9AM                = "SessionReminderSMSAT9AM"
	TwilioSMSInvalidResponse               = "TwilioSMSInvalidResponse"
	TwilioSMSAttendanceConfirmed           = "TwilioSMSAttendanceConfirmed"
	TwilioSMSSessionReschedule             = "TwilioSMSSessionReschedule"
	TwilioSMSSessionRescheduleIfNoResponse = "TwilioSMSSessionRescheduleIfNoResponse"
	SuccessfullPaymentEmail                = "SuccessfulPaymentEmail"
	SessionRescheduleEmail                 = "SessionRescheduleEmail"
	SubscriptionCancelEmail                = "SubscriptionCancelEmail"
	SessionCancelEmailForTutor             = "SessionCancelEmailForTutor"
)

var (
	EmailMessageType = "email"
	SmsMessageType   = "sms"
)

var (
	DefaultEventType   = "now"
	ScheduledEventType = "scheduled"
)

var (
	PendingMessageStatus = "pending"
)

var (
	DefaultInterval = "once"
	DailyInterval   = "daily"
)
