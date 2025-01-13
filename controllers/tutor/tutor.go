package tutor

import (
	"log"
	"net/http"
	"strconv"
	"time"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	"tutree/student-apis/models/stripe"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/models/tutor"
	"tutree/student-apis/utility"

	"github.com/gin-gonic/gin"
)

// This function retrieves the available time slots for a tutor within a specified date range.
// Input:
// Context (c): Contains the HTTP request and response, including:
// userID (from context): The unique identifier for the user making the request.
// start_date (query parameter): The start date for the available slots query.
// end_date (query parameter): The end date for the available slots query.
// Output:
// Status Code: 200 OK
// Body: JSON object containing available_slots with the list of available time slots for the tutor.

// GetTutorAvailableSlots godoc
// @Summary This function retrieves the available time slots for a tutor within a specified date range.
// @description This function retrieves the available time slots for a tutor within a specified date range so that student can
// @description book demo session with available tutor.
// @Tags Tutor
// @Accept application/x-www-form-urlencoded
// @Param start_date  query  string true "Start Date"
// @Param end_date  query  string true "End Date"
// @Produce json
// @Success 200
// @Router /tutor/slots [get]
func GetTutorAvailableDemoSlots(c *gin.Context) {

	userID, exist := c.Get("userID")
	if !exist {
		log.Println("GetTutorAvailableSlots: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		log.Println("GetTutorAvailableSlots: Failed to fetch start and end date time from query params.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	subjectID, err := strconv.Atoi(c.Query("subject_id"))
	if err != nil || subjectID <= 0 {
		log.Println("[INVALID] GetSessionlist: Missing subject Id")
		controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
		return
	}
	gradeID, err := strconv.Atoi(c.Query("grade_id"))
	if err != nil || gradeID <= 0 {
		log.Println("[INVALID] GetSessionlist: Missing grade id")

		//check for student id
		studentID, err := strconv.Atoi(c.Query("student_id"))
		if err != nil || studentID <= 0 {
			log.Println("[INVALID] GetSessionlist: Missing grade id")
			controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
			return
		}
		gradeID, err = students.GetStudentGrade(studentID)
		if err != nil {
			log.Println("[INVALID] GetSessionlist: Missing grade id")
			controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
		}

	}

	// Parse the input dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
		return
	}

	// Get the current date and time
	now := time.Now()

	log.Println("startDate", startDate)
	// If startDate is in the past, set it to the current date and time
	if startDate.Before(now) {
		startDate = now
		startDate = startDate.Add(time.Minute * 60) // add 60 minutes to return slots in future.
	}
	log.Println("now ", now)
	log.Println("startDate 2 ", startDate)
	availableSlots, err := tutor.GetAvailableDemoSlots(startDate, endDate, userID.(int), subjectID, gradeID)
	if err != nil {
		log.Println("GetTutorAvailableSlots: Failed to fetch available tutor from database in given time slot with error :", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available_slots": availableSlots})
}
func GetTutorAvailableSlotsForReschedule(c *gin.Context) {

	userID, exist := c.Get("userID")
	if !exist {
		log.Println("GetTutorAvailableSlots: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	// endDateStr, err := students.GetActiveSubscriptionDetailsRenewalDate(userID.(int))
	// if err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	// 	return
	// }

	plan, err := stripe.GetActivePlanDetails(userID.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active plan found."})
		return
	}

	endDateStr := plan.ValidUntil.String

	log.Println("endDate", endDateStr)
	layout := time.RFC3339
	parsedTime, err := time.Parse(layout, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
		return
	}
	endDate := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 23, 59, 59, 0, parsedTime.Location())

	startDateStr := plan.StartDate.String
	log.Println("startDateStr", startDateStr)
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
		return
	}

	// Get the current date and time
	now := time.Now()
	if startDate.Before(now) {
		startDate = now
		startDate = startDate.Add(time.Minute * 60) // add 60 minutes to return slots in future.
	}

	// start date of plan
	//startDate := time.Now()

	// Add 30 days to the start date
	//endDate := startDate.AddDate(0, 0, 30)

	availableSlots, err := tutor.GetAvailableRescheduleSlots(startDate, endDate, userID.(int))
	if err != nil {
		log.Println("GetTutorAvailableSlots: Failed to fetch available tutor from database in given time slot with error :", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available_slots": availableSlots})
}
func GetTutorAvailableSlots(c *gin.Context) {

	userID, exist := c.Get("userID")
	if !exist {
		log.Println("GetTutorAvailableSlots: Failed to validate user's session.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}
	var (
		consecutiveSessionLimit int
		plan                    stripe.ActivePlan
		endDateStr              string
	)

	// Accept "per_week_limit" and "max_session" as URL query parameters from the PPC page.
	// These parameters represent the maximum sessions allowed per week and the total maximum sessions for the plan.
	planPerWeekLimit, err := strconv.Atoi(c.Query("per_week_limit"))
	planMaxSession, err := strconv.Atoi(c.Query("max_session"))

	// If valid values (greater than zero) for both `planMaxSession` and `planPerWeekLimit` were provided by the PPC page:
	if planMaxSession > 0 && planPerWeekLimit > 0 {
		// Calculate `consecutiveSessionLimit`, which determines the maximum consecutive sessions
		// a user can book within a week by dividing total sessions by weekly limit.
		consecutiveSessionLimit = planMaxSession / planPerWeekLimit
	}

	// If no valid values were provided by the PPC page (i.e., `consecutiveSessionLimit` is zero or negative),
	// we proceed to fetch active plan details as a fallback.
	if consecutiveSessionLimit <= 0 {
		plan, err = stripe.GetActivePlanDetails(userID.(int))
		if err != nil {
			log.Println("BookRegularSession: Failed to fetch active subscription plan slots for given user ID with error:", err)
			controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
			return
		}

		endDateStr, err = students.GetActiveSubscriptionDetailsRenewalDate(userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// Recalculate `consecutiveSessionLimit` based on the active plan's details as a fallback.
		// This ensures that valid values are available, even if the PPC page didn’t provide correct inputs.
		consecutiveSessionLimit = plan.MaxSession / plan.PerWeekLimit

		planPerWeekLimit = plan.PerWeekLimit
	}

	subjectID, err := strconv.Atoi(c.Query("subject_id"))
	if err != nil || subjectID <= 0 {
		log.Println("[INVALID] GetSessionlist: Missing subject Id")
		controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
		return
	}

	//check for student id
	studentID, err := strconv.Atoi(c.Query("student_id"))
	if err != nil || studentID <= 0 {
		log.Println("[INVALID] GetSessionlist: Missing grade id")
		// controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
		// return

	}
	var gradeID int
	if studentID != 0 {
		gradeID, err = students.GetStudentGrade(studentID)
		if err != nil {
			log.Println("[INVALID] GetSessionlist: Missing grade id")
			controllers.HandleJSONErrorResponse(apierror.FailedToParseSubjectIDIntoInteger, nil, c)
		}
	} else {
		gradeID = 5
	}

	// Get the current date and time

	startDateString := c.Query("start_from")
	var startDate time.Time
	if startDateString == "" {
		startDate = time.Now()
	} else {
		// Parse the input dates
		startDate, err = utility.ConvertToUTCTime(startDateString) //time.Parse("2006-01-02 15:04:05", startDateString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_from format"})
			return
		}

		//check if its today's date
		today := time.Now()
		if startDate.Year() == today.Year() && startDate.Month() == today.Month() && startDate.Day() == today.Day() {
			startDate = time.Now()
		}

	}

	now := time.Now()
	if startDate.Before(now) {
		startDate = now
		startDate = startDate.Add(time.Minute * 60) // add 60 minutes to return slots in future.
	}

	log.Println("endDate", endDateStr)
	var endDate time.Time
	if endDateStr == "" || len(endDateStr) == 0 {

		// Add 27 days to startDate to include both start and end dates in a 28-day range
		endDate = startDate.AddDate(0, 0, 27)
	} else {

		layout := time.RFC3339
		endDate, err = time.Parse(layout, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}

	}
	endDateEndOFDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	// Extract only the date portion (Year, Month, Day)
	//endDate := parsedTime.Format("2006-01-02")

	availableSlots, err := tutor.GetAvailableSlots(startDate, endDateEndOFDay, userID.(int), consecutiveSessionLimit, subjectID, gradeID, planPerWeekLimit)
	if err != nil {
		log.Println("GetTutorAvailableSlots: Failed to fetch available tutor from database in given time slot with error :", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available_slots": availableSlots})
}
