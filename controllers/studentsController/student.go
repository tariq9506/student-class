package studentscontroller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	"tutree/student-apis/controllers/jwt"
	"tutree/student-apis/controllers/user"
	"tutree/student-apis/models"
	sessionmodel "tutree/student-apis/models/sessionModel"
	"tutree/student-apis/models/stripe"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"
)

// StudentRegistration handles the registration process for students.
// It expects a POST request with a phone number in the form data.
// It validates the phone number, checks if it exists, verifies its deliverability,
// checks if VOIP numbers are allowed, and sends a verification code if all checks pass.
// Upon successful verification code sending, it returns a JSON response with success status,
// a success message, and the student ID associated with the registration.

// StudentRegistration godoc
// @Summary This controller will handles the registration process for students.
// @description The StudentRegistration function handles the process of signing up
// @description students on a website. It expects the student to submit their phone number through a form.
// @Tags PhoneVerification
// @Accept application/x-www-form-urlencoded
// @Param phone-number  formData  string true "Phone"
// @Produce json
// @Success 200
// @Router /student/authenticate [post]
func StudentRegistration(c *gin.Context) {
	student := user.RegisterUser(c)
	// All errors are handled in the 'RegisterUser' function.
	// If any error occurs, 'RegisterUser' returns an error in c.JSON and an empty 'User' struct.
	// We check if the struct is empty then return from the function to prevent sending both an error and a success response.
	if student.ID == 0 {
		return
	}

	customer, err := stripe.FetchCustomerInfo(int(student.ID))
	if err != nil {
		log.Printf("CreateSubscriptionWithTrial: Error fetching customer info for userId %d: %v", student.ID, err)

	}
	if customer.StripeId == "" {
		log.Println("[ACCOUNT-CREATION-ON-STRIPE]: StudentRegistration: Creating user on stripe")
		go CreateUserOnStripe(student)
	}
	// Save the user's role as an student asynchronously.
	go user.SaveUserRole(int(student.ID), models.StudentRole)

	// Send notification on slack whenever new student signups
	if !student.IsPhoneVerified && student.Location.CountryCode == "US" {
		channelID := os.Getenv("STUDENT_SIGNUP_CHANNEL")
		if len(channelID) == 0 {
			log.Println("[ERROR] Failed to get STUDENT_SIGNUP_CHANNEL")
		} else {
			source := utility.GetSourceOfSignup(student.QueryParams)
			location := student.Location.City + ", " + student.Location.State + ", " + student.Location.CountryCode
			message := fmt.Sprintf("*New Student Signup* \nPhone: %s \nLocation: %s \nSource: %s", student.Phone, location, source)
			go services.SendSlackMessage(channelID, message)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "success",
		"message":          "One time message has been sent to you phone number.",
		"user_id":          student.ID,
		"phone_verified":   student.IsPhoneVerified,
		"email_verified":   student.IsEmailVerified,
		"my_refferal_code": student.ReferralCode,
		"role":             "student",
	})
}

func StudentRegistrationPPC(c *gin.Context) {
	student := user.RegisterUser(c)
	// All errors are handled in the 'RegisterUser' function.
	// If any error occurs, 'RegisterUser' returns an error in c.JSON and an empty 'User' struct.
	// We check if the struct is empty then return from the function to prevent sending both an error and a success response.
	if student.ID == 0 {
		return
	}

	customer, err := stripe.FetchCustomerInfo(int(student.ID))
	if err != nil {
		log.Printf("CreateSubscriptionWithTrial: Error fetching customer info for userId %d: %v", student.ID, err)

	}
	if customer.StripeId == "" {
		log.Println("[ACCOUNT-CREATION-ON-STRIPE]: StudentRegistrationPPC: Creating user on stripe")
		go CreateUserOnStripe(student)
	}

	// Save the user's role as an student asynchronously.
	go user.SaveUserRole(int(student.ID), models.StudentRole)

	// Send notification on slack whenever new student signups
	if !student.IsPhoneVerified && student.Location.CountryCode == "US" {
		channelID := os.Getenv("STUDENT_SIGNUP_CHANNEL")
		if len(channelID) == 0 {
			log.Println("[ERROR] Failed to get STUDENT_SIGNUP_CHANNEL")
		} else {
			source := utility.GetSourceOfSignup(student.QueryParams)
			location := student.Location.City + ", " + student.Location.State + ", " + student.Location.CountryCode
			message := fmt.Sprintf("*New Student Signup* \nPhone: %s \nLocation: %s \nSource: %s", student.Phone, location, source)
			go services.SendSlackMessage(channelID, message)
		}
	}

	token := jwt.CreateUserAuth(c, student)

	c.JSON(http.StatusOK, gin.H{"status": "success",
		"message":          "One time message has been sent to you phone number.",
		"user_id":          student.ID,
		"phone_verified":   student.IsPhoneVerified,
		"email_verified":   student.IsEmailVerified,
		"my_refferal_code": student.ReferralCode,
		"role":             "student",
		"token":            token,
	})
}

// UpdateStudentProfile handles the updating of a student's profile information based on the provided user ID.
// It retrieves the existing student details, validates and updates the new information provided via POST form data,
// and then updates the student's profile in the database.
//
// Parameters:
//   - c (*gin.Context): The Gin context for handling the HTTP request and response.
//   - userID (int): The unique identifier for the student whose profile is to be updated.
//
// Returns:
//   - error: An error object if there is any failure during retrieval, validation, or updating of the student's profile.
func UpdateStudentProfile(c *gin.Context, userID int) (int, error) {
	var studentDetails students.Student

	studentDetails, err := students.GetStudentProfile(userID)
	if err != nil {
		log.Println("[ERROR] UpdateStudentProfile: failed to get student details by user id with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return 0, err
	}
	// accept student name.
	name := c.PostForm("student-name")
	if len(name) == 0 && len(studentDetails.StudentName) == 0 {
		log.Println("UpdateStudentProfile: Failed to fetch Student name.")

		// NOTE:- Commented code for "main" branch.
		// controllers.HandleJSONErrorResponse(apierror.FailedToGetStudentName, nil, c)
		// return 0, errors.New("failed to fetch Student name")
	}
	if len(name) == 0 {
		name = studentDetails.StudentName
	}
	// accept student grade.
	grade := c.PostForm("grade_id")
	gradeID, err := strconv.Atoi(grade)
	if err != nil && (gradeID == 0 && studentDetails.GradeID == 0) {
		log.Println("UpdateStudentProfile: Failed to fetch student grade.")
		controllers.HandleJSONErrorResponse(apierror.FailedToGetGradeOfStudent, nil, c)
		return 0, errors.New("failed to fetch student grade")
	}
	if gradeID <= 0 {
		gradeID = studentDetails.GradeID
	}
	// accept parent's email.
	parentEmail := c.PostForm("parent-email")
	if len(parentEmail) == 0 && len(studentDetails.ParentEmail) == 0 {
		// here we validate parent's email is valid or not.
		valid, err := utility.ValidateEmailPattern(parentEmail)
		if !valid || err != nil {
			log.Println("UpdateStudentProfile: Failed to validate email of parent with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToValidateEmail, err, c)
			return 0, err
		}
	}
	if len(parentEmail) == 0 {
		parentEmail = studentDetails.ParentEmail
	}

	studentProfile := students.Student{
		StudentName: name,
		GradeID:     gradeID,
		ParentEmail: parentEmail,
	}
	// here we update student's information.
	studentID, stripeCustomerID, err := students.UpdateStudentProfile(userID, studentProfile)
	if err != nil {
		log.Println("UpdateStudentProfile: Failed to save student information with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return 0, err
	}

	err = services.UpdateUserInformationOnStripe(stripeCustomerID, studentProfile.ParentEmail, studentProfile.StudentName)
	if err != nil {
		log.Println("[STRIPE-ERROR] UpdateStudentProfile --> UpdateUserInformationOnStripe: Failed to update the user information on stripe with:", err)
	} else {
		err := stripe.UpdateUser2StripeUpdatedAt(stripeCustomerID)
		if err != nil {
			log.Println("[ERROR] UpdateStudentProfile --> UpdateUser2StripeUpdatedA: Failed to update the updated_at with :", err)

		}
	}
	// update the leads on ZOHO CRM lead board
	err = controllers.UpdateLeadsOnZohoCRM(userID)
	if err != nil {
		log.Println("UpdateStudentProfile: Failed to update student information in zoho crm lead board with error: ", err)
		//	controllers.HandleJSONErrorResponse(apierror.FailedToCreateLeadOnZohoCRM, err, c)
		//return
	}
	return studentID, nil
}

// GetStudentProfile handles the request to retrieve a student's profile and demo session information.
// Parameter -
//		 c - the Gin context for the request.
// This function retrieves the student's profile based on the provided Authorization token and
// fetches the student's demo session details. It responds with a JSON object containing the student's information
// and demo session details.

// GetStudentProfile godoc
// @Summary This controller will handles to fetch profile of students.
// @description This function retrieves the student's profile based on the provided Authorization token and
// @description fetches the student's demo session details. It responds with a JSON object containing the student's information
// @description and demo session details.
// @Tags Student
// @Accept application/x-www-form-urlencoded
// @Param Authorization header  string true "Token"
// @Produce json
// @Success 200
// @Router /student/profile [get]
func GetStudentProfile(c *gin.Context) {
	// Fetch the authorization token from the request header.
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("BookDemoSession: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	var wg sync.WaitGroup
	wg.Add(5)

	var studentProfile students.Student
	var subscription map[string]interface{}
	var demoSessions []map[string]interface{}
	var isPaidSessionBooked bool
	var signupSource string
	var sessionPrefereces []sessionmodel.StudentSessionPreference
	var fetchProfileErr, fetchSessionErr, fetchSubscriptionErr, fetchPaidSessionErr, fetchSessionPreferenceErr error

	go func() {
		defer wg.Done()
		// Fetch the student's profile details using user ID.
		studentProfile, fetchProfileErr = students.GetStudentProfile(userID.(int))
		if fetchProfileErr != nil {
			log.Println("[ERROR] GetStudentProfile: Failed to fetch student's details by using user ID with error: ", fetchProfileErr)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, fetchProfileErr, c)
			return
		}
		signupSource = utility.GetSourceOfSignup(studentProfile.User.QueryParams)
	}()

	go func() {
		defer wg.Done()
		// Fetch the student's demo sessions.
		demoSessions, fetchSessionErr = sessionmodel.GetRecentDemoSessions(userID.(int))
		if fetchSessionErr != nil {
			log.Println("[ERROR] GetStudentProfile: Failed to fetch student's demo session details from database with error: ", fetchSessionErr)
		}
	}()
	go func() {
		defer wg.Done()
		// Fetch the student's profile details using user ID.
		subscription, fetchSubscriptionErr = students.GetLatestSubscriptionPlan(userID.(int))
		if fetchSubscriptionErr != nil {
			log.Println("[ERROR] GetStudentProfile: Failed to fetch student's subscription details by using user ID with error: ", fetchSubscriptionErr)
		}
	}()
	go func() {
		defer wg.Done()
		// Fetch the student's profile details using user ID.
		isPaidSessionBooked, fetchPaidSessionErr = students.IsRegularSessionBooked(userID.(int))
		if fetchPaidSessionErr != nil {
			log.Println("[ERROR] GetStudentProfile: Failed to fetch student's subscription details by using user ID with error: ", fetchPaidSessionErr)
		}
	}()
	go func() {
		defer wg.Done()
		// Fetch the student's profile details using user ID.
		sessionPrefereces, fetchSessionPreferenceErr = sessionmodel.GetSessionPreferenceById(0, int64(userID.(int)))
		if fetchSessionPreferenceErr != nil {
			log.Println("[ERROR] GetStudentProfile: Failed to fetch student's session preference details by using user ID with error: ", fetchPaidSessionErr)
		}
	}()

	wg.Wait()

	// if fetchProfileErr != nil || fetchSessionErr != nil {
	// 	log.Println("[ERROR] GetStudentProfile: Failed to fetch student's demo session details or fetch student's details from database with error: ", fetchProfileErr, fetchSessionErr)
	// 	controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, nil, c)
	// 	return
	// }
	response := gin.H{
		"name":                      studentProfile.StudentName,
		"parent_email":              studentProfile.ParentEmail,
		"grade_id":                  studentProfile.GradeID,
		"garde_name":                studentProfile.Grade,
		"picture_url":               studentProfile.StudentPictureURL,
		"parent_name":               studentProfile.ParentName,
		"phone_verified":            studentProfile.IsPhoneVerified,
		"email_verified":            studentProfile.IsEmailVerified,
		"my_refferal_code":          studentProfile.RefferalCode,
		"role":                      studentProfile.Role,
		"phone_number":              studentProfile.User.Phone,
		"regular_session_scheduled": isPaidSessionBooked,
		"school_name":               studentProfile.SchoolName,
		"signup_source":             signupSource,
	}
	if fetchSessionErr == nil {
		response["demo_session"] = demoSessions
	}
	if fetchSubscriptionErr == nil {
		response["subscription"] = subscription
	}
	if fetchSessionPreferenceErr == nil {
		response["session_preference"] = sessionPrefereces
	}
	// Send the JSON response with the student profile and demo session details.
	c.JSON(http.StatusOK, response)

}

// This function handles the process of updating a student's profile information. It retrieves the student's
// current profile details, processes updates to fields such as name, grade, parent's email, school ID, profile picture,
// and parent's name, and then saves these updates.
//
//Context (c): Contains the HTTP request and response, including:
// userID (from context): The unique identifier for the user making the request.
// student_name (form field): The new name of the student (optional).
// grade (form field): The new grade of the student (optional).
// parent_email (form field): The new email address of the student's parent (optional).
// school_id (form field): The new school ID for the student (optional).
// student_picture (form file): The new profile picture of the student (optional).
// parent_name (form field): The new name of the student's parent (optional).
// Output:

// Success Response:

// Status Code: 200 OK
// Body: JSON object with a success message indicating the profile was updated successfully.

// UpdateStudentProfileDetails godoc
// @Summary This function handles the process of updating a student's profile information.
// @description This function handles the process of updating a student's profile information. It retrieves the student's
// @description current profile details, processes updates to fields such as name, grade, parent's email, school ID, profile picture,
// @description and parent's name, and then saves these updates.
// @Tags Student
// @Accept application/x-www-form-urlencoded
// @Param Authorization header  string true "Token"
// @Param grade  formData  string false "Grade"
// @Param parent_email formData  string false "Parent's Email"
// @Param student_name formData  string false "Student Name"
// @Param parent_name formData  string false "Parent's Name"
// @Param school_id formData  string false "School ID"
// @Param student_picture formData  string false "Student Picture"
// @Produce json
// @Success 200
// @Router /session/student/profile [put]
func UpdateStudentProfileDetails(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("UpdateStudentProfileDetails: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	studentDetails, err := students.GetStudentProfile(userID.(int))
	if err != nil {
		log.Println("UpdateStudentProfileDetails: failed to get student details by user id with error: ", err)
		return
	}
	var UpdatedSchoolID int
	// accept student name.
	name := c.PostForm("student_name")
	if len(name) == 0 && len(studentDetails.StudentName) == 0 {
		log.Println("UpdateStudentProfileDetails: Failed to fetch Student name.")
		controllers.HandleJSONErrorResponse(apierror.FailedToGetStudentName, nil, c)
		return
	}
	if len(name) == 0 {
		name = studentDetails.StudentName
	}
	// accept student grade.
	grade := c.PostForm("grade")
	if len(grade) == 0 && len(studentDetails.Grade) == 0 {
		controllers.HandleJSONErrorResponse(apierror.FailedToGetGradeOfStudent, nil, c)
		log.Println("UpdateStudentProfileDetails: Failed to fetch student grade.")
		return
	}
	if len(grade) == 0 {
		grade = studentDetails.Grade
	}
	// accept parent's email.
	parentEmail := c.PostForm("parent_email")
	if len(parentEmail) == 0 && len(studentDetails.ParentEmail) == 0 {
		log.Println("UpdateStudentProfileDetails: Failed to update email of parent,parents email can not be empty.")
		controllers.HandleJSONErrorResponse(apierror.EmptyParentsEmailErr, err, c)
		return
	}
	if len(parentEmail) != 0 {
		// here we validate parent's email is valid or not.
		valid, err := utility.ValidateEmailPattern(parentEmail)
		if !valid || err != nil {
			log.Println("UpdateStudentProfileDetails: Failed to validate email of parent with error: ", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToValidateEmail, err, c)
			return
		}
	}
	if len(parentEmail) == 0 {
		parentEmail = studentDetails.ParentEmail
	}
	schoolID := c.PostForm("school_id")
	if len(schoolID) == 0 && studentDetails.SchoolID == 0 {
		controllers.HandleJSONErrorResponse(apierror.FailedToGetSchoolID, nil, c)
		log.Println("UpdateStudentProfileDetails: Failed, school id can not be empty.")
		return
	}
	if len(schoolID) != 0 {
		UpdatedSchoolID, err = strconv.Atoi(schoolID)

		if err != nil {
			log.Println("UpdateStudentProfileDetails: Failed to parse student's school id with error :", err)
			controllers.HandleJSONErrorResponse(apierror.FailedToParseSchoolID, err, c)
			return
		}

		if UpdatedSchoolID <= 0 {
			log.Println("UpdateStudentProfileDetails: Failed, school ID must be positive integer.")
			controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
			return
		}
	} else {
		UpdatedSchoolID = studentDetails.SchoolID
	}
	// accept student parent's name.
	parentName := c.PostForm("parent_name")
	if len(parentName) == 0 && len(studentDetails.ParentName) == 0 {
		controllers.HandleJSONErrorResponse(apierror.FailedToFetchParentsName, nil, c)
		log.Println("UpdateStudentProfileDetails: Failed to fetch student's parent's name.")
		return
	}
	if len(parentName) == 0 {
		parentName = studentDetails.ParentName
	}
	// // accept student picture URL.
	var studentProfilepicture string
	err = c.Request.ParseMultipartForm(5 * 1024 * 1024)
	var bufferForPicture bytes.Buffer
	picture, fileHandler1, err := c.Request.FormFile("student_picture")

	if picture == nil && len(studentDetails.StudentPictureURL) == 0 {
		controllers.HandleJSONErrorResponse(apierror.FailedToFetchStudentPicture, nil, c)
		log.Println("UpdateStudentProfileDetails: Failed to fetch student's picture.")
		return
	}
	if err != nil {
		log.Println("UpdateStudentProfileDetails: failed while fetching student image from form with error ", err)
	}

	if err == nil {
		fileInfo := services.FileInfo{
			Bucket:        os.Getenv("BUCKET_FOR_CLOUDFRONT"),
			FilePath:      "student",
			CloudFrontCDN: true,
		}
		// Copy the contents of the 'picture' variable, which is of type 'multipart.File', into a 'bytes.Buffer' bufferForPicture.
		// This is done because we will be using the 'Read()' method of 'picture' while uploading it to AWS,
		// and the 'bytes.Buffer' also provides a 'Read()' method for compatibility.
		_, err = io.Copy(&bufferForPicture, picture)
		if err != nil {
			log.Println("UpdateStudentProfileDetails: failed while copying contents to buffer with error ", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}
		// UploadFileToS3AtLocation would be uploaded student picture inside student folder on AWS
		// sample url- https://synkduptest.s3.amazonaws.com/thumbnail/{image_link}
		studentProfilepicture = services.UploadFileToS3AtLocation(bufferForPicture, fileHandler1, fileInfo)
	}
	if picture == nil {
		studentProfilepicture = studentDetails.StudentPictureURL
	}

	studentProfile := students.Student{
		StudentName:       name,
		Grade:             grade,
		ParentEmail:       parentEmail,
		SchoolID:          UpdatedSchoolID,
		StudentPictureURL: studentProfilepicture,
		ParentName:        parentName,
	}
	// here we update student's information.
	_, stripeCustomerID, err := students.UpdateStudentProfile(userID.(int), studentProfile)
	if err != nil {
		log.Println("UpdateStudentProfileDetails: Failed to save student information with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	err = services.UpdateUserInformationOnStripe(stripeCustomerID, studentProfile.ParentEmail, studentProfile.StudentName)
	if err != nil {
		log.Println("[STRIPE-ERROR] UpdateStudentProfileDetails --> UpdateUserInformationOnStripe: Failed to update the user information on stripe with:", err)
	} else {
		err := stripe.UpdateUser2StripeUpdatedAt(stripeCustomerID)
		if err != nil {
			log.Println("[ERROR] UpdateStudentProfileDetails --> UpdateUser2StripeUpdatedA: Failed to update the updated_at with :", err)

		}
	}
	// update the leads on ZOHO CRM lead board
	err = controllers.UpdateLeadsOnZohoCRM(userID.(int))
	if err != nil {
		log.Println("AddChildForUser: Failed to update student information in zoho crm lead board with error: ", err)
		//	controllers.HandleJSONErrorResponse(apierror.FailedToCreateLeadOnZohoCRM, err, c)
		//return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully update student's profile."})
}

// SendReminderSMSForDemoBooking sends a reminder SMS to students, prompting them to schedule a free demo session.
// Handles errors
//    - Failing to retrieve phone numbers from the database.
//    - Failing to send the SMS message to any individual student.
// If any error occurs during the process, the function logs the error, sends an error response to the client, and terminates further execution.

func SendReminderSMSForDemoBooking(c *gin.Context) {
	// Retrieve the phone numbers of students to send the reminder SMS
	studentPhones, err := students.GetPhoneNumbersForDemoReminder()
	if err != nil {
		log.Println("[ERROR] SendReminderSMSForDemoBooking: Failed to retrieve phone numbers for reminder SMS:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Construct the demo booking link
	demoBookingURL := fmt.Sprintf("%s/student/dashboard", utility.GetHostURL())

	// Prepare and send the SMS reminder to each student
	for _, studentPhone := range studentPhones {
		reminderMessage := fmt.Sprintf("Coding for kids Ages 6+. Get your FREE demo scheduled today and build your first game.\n\nBook Now: %s", demoBookingURL)
		err := services.SendMessage(studentPhone.Phone, reminderMessage, studentPhone.DialingCode)
		if err != nil {
			log.Println("[ERROR] SendReminderSMSForDemoBooking: Failed to send reminder SMS to", studentPhone.Phone, ":", err)
			continue
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully send reminder SMS to book free demo class after signUp.",
	})
}

func CreateUserOnStripe(user models.User) {

	stripeId, err := services.CreateCustomer(user.Phone)
	if err != nil {
		log.Printf("[STRIPE-ERROR] CreateUserOnStripe: Error creating Stripe customer for userID %d : %v", user.ID, err)
		return
	}

	err = models.StoreStripeCustomerId(int(user.ID), stripeId)
	if err != nil {
		log.Printf("Create: Error storing Stripe customer ID for userId %d: %v", user.ID, err)
		return
	}
}

// AddChildForUser godoc
// @Summary This controller function handles the adding a child for user.
// @description This controller will add a chiled for user.
// @description This api is taking student name, parent's email, grade as postform.
// @Tags User
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param student-name  formData  string true "Student Name"
// @Param grade  formData  string true "Student Grade"
// @Param parent-email  formData  string true "parent's Email`"
// @Produce json
// @Success 200
// @Router /user/child [POST]
func AddChildForUser(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("[UNAUTHORIZED] AddChildForUser: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	studentID, updateProfileErr := UpdateStudentProfile(c, userID.(int))
	if updateProfileErr != nil {
		log.Println("[ERROR] AddChildForUser: Failed to update student profile with error: ", updateProfileErr)
		return
	}

	err := students.UpdateStudentSessionPreference(userID.(int))
	if err != nil {
		log.Println("[ERROR] :AddChildForUser -->> UpdateStudentSessionPreference failed to update the studen_id on student session preferences")
	}
	// update the leads on ZOHO CRM lead board

	err = controllers.UpdateLeadsOnZohoCRM(userID.(int))
	if err != nil {
		log.Println("AddChildForUser: Failed to update student information in zoho crm lead board with error: ", err)
		//	controllers.HandleJSONErrorResponse(apierror.FailedToCreateLeadOnZohoCRM, err, c)
		//return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":    "Student successfully added.",
		"status":     "Success",
		"student_id": studentID,
	})
}

// GetChildOfUser godoc
// @Summary This controller function handles the listing of children for user.
// @description This controller will fetch all child for user.
// @Tags User
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Produce json
// @Success 200
// @Router /user/child [GET]
func GetChildOfUser(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("[UNAUTHORIZED] GetChildOfUser: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	childList, err := students.GetChildOfUser(userID.(int))
	log.Println(childList)
	if err != nil {
		log.Println("[ERROR] GetChildOfUser : Failed while fetching list of chiled from database with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":    "Student successfully added.",
		"status":     "Success",
		"child_list": childList,
	})
}

// UpdateChildInformation is a handler function that updates a child’s information based on input from a request context.
//
// Parameters:
// - c: *gin.Context, provides access to request and response handling for the endpoint.
//
// Process:
// 1. Retrieves the userID from the session; if not found, responds with an unauthorized error.
// 2. Collects and validates data for the child, including childOldName, childName, parentEmail, and gradeID.
// 3. Converts gradeID to an integer and checks that it is positive.
// 4. Calls UpdateChildDetails to update the child’s information in the database.
// 5. Responds with a success message if the update is successful, otherwise handles and logs errors appropriately.

// UpdateChildInformation godoc
// @Summary This controller function handles the updatation of child details.
// @description This controller will update given child details.
// @Tags User
// @Accept application/x-www-form-urlencoded
// @Param Authorization header string true "Authorization token (Bearer token)"
// @Param grade_id  formData  number false "Grade ID"
// @Param parent-email formData  string false "Parent's Email"
// @Param student-name formData  string false "Child Name"
// @Param child_id formData  number true "Child ID"
// @Produce json
// @Success 200
// @Router /user/child [PUT]
func UpdateChildInformation(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		log.Println("[UNAUTHORIZED] UpdateChildInformation: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	var requestDataOfChild students.RequestDataToUpdateChildDetails
	childIDStr := c.PostForm("child_id")
	if len(childIDStr) == 0 {
		log.Println("[NOT FOUND] UpdateChildInformation : child id not provided.")
		controllers.HandleJSONErrorResponse(apierror.ErrorDataNotProvided, nil, c)
		return
	}

	childID, err := strconv.Atoi(childIDStr)
	if err != nil {
		log.Println("[ERROR] UpdateChildInformation: Failed to convert child id string to integer with error :", err)
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, err, c)
		return
	}
	if childID <= 0 {
		log.Println("[ERROR] UpdateChildInformation: Failed due to invalid child id, must be positive integer.")
		controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
		return
	}

	requestDataOfChild.ID = childID
	childName := c.PostForm("student-name")
	if len(childName) == 0 {
		log.Println("UpdateChildInformation : Child name not provided.")
	}
	requestDataOfChild.Name = childName
	parentEmail := c.PostForm("parent-email")
	if len(parentEmail) == 0 {
		log.Println("UpdateChildInformation : Parent email not provided.")
	}
	requestDataOfChild.ParentEmail = parentEmail
	var (
		gradeID int
	)
	gardeIDStr := c.PostForm("grade_id")
	if len(gardeIDStr) == 0 {
		log.Println("UpdateChildInformation : Grade id not provided.")
	}
	if len(gardeIDStr) != 0 {
		gradeID, err = strconv.Atoi(gardeIDStr)
		if err != nil {
			log.Println("[ERROR] UpdateChildInformation: Failed to convert child's grade id string to integer with error :", err)
			controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, err, c)
			return
		}
		if gradeID <= 0 {
			log.Println("[ERROR] UpdateChildInformation: Failed due to invalid child's grade id, must be positive integer.")
			controllers.HandleJSONErrorResponse(apierror.FieldMustBePositiveInteger, nil, c)
			return
		}

	}
	requestDataOfChild.GradeID = gradeID
	stripeCustomerID, err := students.UpdateChildDetails(requestDataOfChild)
	if err != nil {
		log.Println("[ERROR] UpdateChildInformation : Failed while update details of child in database with error :", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	err = services.UpdateUserInformationOnStripe(stripeCustomerID, requestDataOfChild.ParentEmail, requestDataOfChild.Name)
	if err != nil {
		log.Println("[STRIPE-ERROR] UpdateChildInformation --> UpdateUserInformationOnStripe: Failed to update the user information on stripe with:", err)
	} else {
		err := stripe.UpdateUser2StripeUpdatedAt(stripeCustomerID)
		if err != nil {
			log.Println("[ERROR] UpdateStudentProfile --> UpdateUser2StripeUpdatedA: Failed to update the updated_at with :", err)

		}
	}
	// update the leads on ZOHO CRM lead board
	err = controllers.UpdateLeadsOnZohoCRM(userID.(int))
	if err != nil {
		log.Println("AddChildForUser: Failed to update student information in zoho crm lead board with error: ", err)
		//	controllers.HandleJSONErrorResponse(apierror.FailedToCreateLeadOnZohoCRM, err, c)
		//return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "child details successfully updated.",
		"status":  "Success",
	})
}
