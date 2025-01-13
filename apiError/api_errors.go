package apierror

import (
	"fmt"
	"net/http"
)

type CustomAPIError struct {
	Code           int
	Message        string
	HttpStatusCode int
}

func (e *CustomAPIError) Error() string {
	return fmt.Sprintf("[ERROR] : %d: %s", e.Code, e.Message)
}

// Common Errors (1000 - 1100)
var (
	SomethingWentWrong         = CustomAPIError{Message: "Something went wrong", Code: 1000, HttpStatusCode: http.StatusInternalServerError}
	ErrorUserNotFound          = CustomAPIError{Message: "User not found", Code: 1001, HttpStatusCode: http.StatusNotFound}
	ErrorOnCheckingSession     = CustomAPIError{Message: "Session not found from cookie.", Code: 1002, HttpStatusCode: http.StatusBadGateway}
	InvailidUserID             = CustomAPIError{Message: "Please enter valid user ID.", Code: 1003, HttpStatusCode: http.StatusNotAcceptable}
	ErrorNoDataFound           = CustomAPIError{Message: "No data found.", Code: 1004, HttpStatusCode: http.StatusNotFound}
	ErrorDataNotProvided       = CustomAPIError{Message: "Missing mandatory field.", Code: 1005, HttpStatusCode: http.StatusBadRequest}
	ErrorUnAuthorizedAccess    = CustomAPIError{Message: "User/client are not authorized.", Code: 1006, HttpStatusCode: http.StatusUnauthorized}
	FieldMustBePositiveInteger = CustomAPIError{Message: "Invalid input, Please insert valid positive integer value.", Code: 1007, HttpStatusCode: http.StatusBadRequest}

	// /v1/student/register (1101 - 1110)

	//ErrorPhoneAlreadyExist = CustomAPIError{Message: "This phone number is already in use. Please use a different phone number for signing up.", Code: 1103, HttpStatusCode: http.StatusBadRequest}
	MissingPhoneNumber  = CustomAPIError{Message: "Phone Number Required: Please enter your phone number to proceed.", Code: 1101, HttpStatusCode: http.StatusBadRequest}
	InvailidPhoneNumber = CustomAPIError{Message: "Provided phone number is invalid, enter a correct phone number in order to continue with the process", Code: 1102, HttpStatusCode: http.StatusBadRequest}
	InvailidStudentID   = CustomAPIError{Message: "Please enter valid Student ID.", Code: 1103, HttpStatusCode: http.StatusBadRequest}
	FailedToSendOTP     = CustomAPIError{Message: "One time message has been not sent to your phone number.", Code: 1104, HttpStatusCode: http.StatusInternalServerError}

	// /v1/student/verifycode (1111 - 1120)
	FailedToCreateSession  = CustomAPIError{Message: "Failed while generating session for student.", Code: 1111, HttpStatusCode: http.StatusBadGateway}
	VerifyCode_OTPExpired  = CustomAPIError{Message: "The code you entered is expired. Try resend new code.", Code: 1112, HttpStatusCode: http.StatusBadRequest}
	VerifyCode_OTPNotMatch = CustomAPIError{Message: "The code you entered is not matched, please try again.", Code: 1113, HttpStatusCode: http.StatusBadRequest}
	VerifyCode_BlankOTP    = CustomAPIError{Message: "Please enter OTP, OTP can not be blank.", Code: 1114, HttpStatusCode: http.StatusNotFound}

	// /v1/session/book-demo (1121 - 1130)
	FailedWhileParsingDateTime = CustomAPIError{Message: "Failed while parsing time.", Code: 1121, HttpStatusCode: http.StatusBadRequest}
	FailedToValidateEmail      = CustomAPIError{Message: "Failed to validate email, Please enter valid email id.", Code: 1122, HttpStatusCode: http.StatusBadRequest}
	SessionDateTimeNotFound    = CustomAPIError{Message: "Invalid slot date time.", Code: 1123, HttpStatusCode: http.StatusBadRequest}
	FailedToGetGradeOfStudent  = CustomAPIError{Message: "Garde can not be empty, Please enter valid Grade of student.", Code: 1124, HttpStatusCode: http.StatusBadRequest}
	FailedToGetStudentName     = CustomAPIError{Message: "Student name can not be empty, Please enter valid name.", Code: 1125, HttpStatusCode: http.StatusBadRequest}
	SessionTimezoneNotFound    = CustomAPIError{Message: "Timezone can not be empty.", Code: 1126, HttpStatusCode: http.StatusBadRequest}
	//NoTutorFoundAtEnteredTime  = CustomAPIError{Message: "Not tutor is available on session-datetime. Please choose some other time slot", Code: 1126, HttpStatusCode: http.StatusNotFound}
	EmptyParentsEmailErr              = CustomAPIError{Message: "Failed to update email of parent,parent's email can not be empty.", Code: 1127, HttpStatusCode: http.StatusBadRequest}
	FailedToParseTutorIDIntoInteger   = CustomAPIError{Message: "Failed to parse tutor ID into integer.", Code: 1128, HttpStatusCode: http.StatusBadRequest}
	FailedToParseStudentIDIntoInteger = CustomAPIError{Message: "Failed to parse student ID into integer.", Code: 1129, HttpStatusCode: http.StatusBadRequest}
	FailedToParseSubjectIDIntoInteger = CustomAPIError{Message: "Failed to parse subject ID into integer.", Code: 1130, HttpStatusCode: http.StatusBadRequest}

	// /v1/session/list (1131 - 1140)
	FailedToConvertIntoInteger = CustomAPIError{Message: "Failed to convert into integer.", Code: 1131, HttpStatusCode: http.StatusBadRequest}

	// /v1/session/feedback (1141 - 1150)
	FailedToConvertIntoBoolean = CustomAPIError{Message: "Failed to convert into boolean.", Code: 1141, HttpStatusCode: http.StatusBadRequest}
	// /v1/session/demo-reschedule (1141 - 1150)
	InvailidSessionID                              = CustomAPIError{Message: "Failed to convert session id into integer, please enter valid session ID.", Code: 1141, HttpStatusCode: http.StatusBadRequest}
	RescheduleDemoSessionNotAllowed                = CustomAPIError{Message: "Rescheduling failed. Sessions cannot be rescheduled less than 60 minutes before their start time.", Code: 1142, HttpStatusCode: http.StatusBadRequest}
	RescheduleNotAllowedBeyondSubscriptionValidity = CustomAPIError{Message: "Rescheduling not allowed. The session start date and time are beyond the subscription plan's validity period.", Code: 1143, HttpStatusCode: http.StatusBadRequest}

	// /v1/session/student/profile [put] (1151 - 1160)
	FailedToParseSchoolID       = CustomAPIError{Message: "Failed to parse school id, please insert valid school id.", Code: 1151, HttpStatusCode: http.StatusBadRequest}
	FailedToFetchParentsName    = CustomAPIError{Message: "Parent's name can not be empty, please provide valid input.", Code: 1152, HttpStatusCode: http.StatusBadRequest}
	FailedToFetchStudentPicture = CustomAPIError{Message: "Student picture can not be empty, please provide valid input.", Code: 1153, HttpStatusCode: http.StatusBadRequest}
	FailedToGetSchoolID         = CustomAPIError{Message: "School ID can not be empty, please provide valid input.", Code: 1154, HttpStatusCode: http.StatusBadRequest}

	// /v1/session/book [POST] (1161 - 1170)
	FailedToBindJsonRespoce                = CustomAPIError{Message: "Failed while trying to bind JSON responce.", Code: 1161, HttpStatusCode: http.StatusBadRequest}
	FailedTBookDueToUnavailableSessionTime = CustomAPIError{Message: "Failed to book regular session due to unavailable session time.", Code: 1162, HttpStatusCode: http.StatusBadRequest}
	FailedDueToMismatchNumberSessions      = CustomAPIError{Message: "The number of selected sessions does not match the maximum sessions allowed by your plan. Please adjust the number of sessions to align with your session limit.", Code: 1163, HttpStatusCode: http.StatusBadRequest}

	// /v1/subscription (1171-1190)
	InvalidPlanID = CustomAPIError{Message: "Failed to convert subscription plan ID into integer, please enter valid subscription plan ID.", Code: 1171, HttpStatusCode: http.StatusBadRequest}

	// /v1/session/reschedule-later (1191-1200)
	FailedToRescheduleDemoSessionForLaterTime = CustomAPIError{Message: "Can not reschedule demo session for later time.", Code: 1191, HttpStatusCode: http.StatusBadRequest}
)
