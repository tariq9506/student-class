package students

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"tutree/student-apis/config"
	"tutree/student-apis/constants"
	"tutree/student-apis/models"
	"tutree/student-apis/models/tutor"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"
)

// Student represents a student record in the database.
type Student struct {
	ID                int64       `json:"id,omitempty"`
	StudentName       string      `json:"student_name,omitempty"`
	ParentName        string      `json:"parent_name,omitempty"`
	ParentEmail       string      `json:"parent_email,omitempty"`
	SchoolID          int         `json:"school_id,omitempty"`
	StudentPictureURL string      `json:"picture_url,omitempty"`
	Grade             string      `json:"grade,omitempty"`
	IsPhoneVerified   bool        `json:"phone_verified,omitempty"`
	IsEmailVerified   bool        `json:"email_verified,omitempty"`
	RefferalCode      string      `json:"refferal_code,omitempty"`
	Role              string      `json:"role,omitempty"`
	SessionStartTime  time.Time   `json:"session_start_time,omitempty"`
	User              models.User `json:"user,omitempty"`
	MeetingLink       string      `json:"meeting_link,omitempty"`
	SessionTimezone   string      `json:"session_timezone,omitempty"`
	SchoolName        string      `json:"school_name,omitempty"`
	SessionType       string      `json:"session_type,omitempty"`
	SessionID         int         `json:"session_id,omitempty"`
	Tutor             tutor.Tutor
	GradeID           int `json:"grade_id,omitempty"`
}
type Child struct {
	ID                     int64  `json:"id,omitempty"`
	StudentName            string `json:"student_name,omitempty"`
	ParentEmail            string `json:"parent_email,omitempty"`
	SchoolID               int    `json:"school_id,omitempty"`
	StudentPictureURL      string `json:"picture_url,omitempty"`
	Grade                  string `json:"grade,omitempty"`
	IsPhoneVerified        bool   `json:"phone_verified,omitempty"`
	IsEmailVerified        bool   `json:"email_verified,omitempty"`
	RefferalCode           string `json:"refferal_code,omitempty"`
	Role                   string `json:"role,omitempty"`
	SchoolName             string `json:"school_name,omitempty"`
	GradeID                int    `json:"grade_id,omitempty"`
	ParentPhone            string `json:"parent_phone,omitempty"`
	ParentPhoneDialingCode string `json:"parent_phone_dialing_code,omitempty"`
}

type StudentToScheduleReminder struct {
	ID       int64
	Name     string
	Phone    string
	Email    string
	Timezone string
}

// GetStudentProfile retrieves the profile details of a student from the database based on the provided user ID.
// It returns a Student struct containing the student's information or an error if the operation fails.
//
// Parameters:
//   - userID (int): The unique identifier for the student whose profile is to be retrieved.
//
// Returns:
//   - Student: A struct containing the student's profile details such as name, grade, parent email, picture URL, and school ID.
//   - error: An error object if there is any failure during database connection or query execution.
func GetStudentProfile(userID int) (Student, error) {
	db, err := config.GetDB2()
	var studentDetails Student
	if err != nil {
		log.Println("GetStudentProfile: Failed when try to connect with database with error: ", err)
		return studentDetails, err
	}
	defer db.Close()
	var (
		role            sql.NullString
		studentName     sql.NullString
		gradeID         sql.NullInt64
		gardeName       sql.NullString
		parentEmail     sql.NullString
		parentName      sql.NullString
		studentPicture  sql.NullString
		schoolID        sql.NullInt64
		isPhoneVerified sql.NullBool
		isEmailVerified sql.NullBool
		referralCode    sql.NullString
		phoneNumber     sql.NullString
		timeZone        sql.NullString
		dialingCode     sql.NullString
		location        sql.NullString
		queryParam      sql.NullString
		schoolName      sql.NullString
	)
	query := `
		SELECT                                     
    			r.role,
    			s.name AS student_name,
    			s.parent_name,
    			s.parent_email,
    			s.grade_id AS grade_id,
    			s.picture_url,
    			s.school_id AS school_id,
    			u.phone_verified,
    			u.email_verified,
    			u.referral_code,
    			u.phone_number,
    			u.timezone,
    			u.dialing_code,
    			u.location,
				u.query_param,
				sc.name AS school_name,
				gr.name AS grade_name
		FROM 
		    public.user AS u 
		LEFT JOIN
		    student AS s ON u.id = s.user_id
		LEFT JOIN 
		    user2roles AS ur ON u.id = ur.user_id
		LEFT JOIN 
		    roles AS r ON ur.role_id = r.id
		LEFT JOIN
			schools AS sc ON s.school_id = sc.id
		LEFT JOIN
			grade AS gr ON gr.id =s.grade_id	
		WHERE 
   			 u.id = $1
`

	err = db.QueryRow(query, userID).Scan(
		&role,
		&studentName,
		&parentName,
		&parentEmail,
		&gradeID,
		&studentPicture,
		&schoolID,
		&isPhoneVerified,
		&isEmailVerified,
		&referralCode,
		&phoneNumber,
		&timeZone,
		&dialingCode,
		&location,
		&queryParam,
		&schoolName,
		&gardeName,
	)
	if err == sql.ErrNoRows {
		return studentDetails, nil
	}
	if err != nil {
		log.Println("GetStudentProfile: Failed while execute the query with error: ", err)
		return studentDetails, err
	}
	var locStruct services.Location
	loc := strings.Split(location.String, ", ")
	if len(loc) > 2 {
		locStruct.City = loc[0]
		locStruct.State = loc[1]
		locStruct.CountryCode = loc[2]

	}
	// Create a map to store the session details.
	studentDetails = Student{
		StudentName:       studentName.String,
		GradeID:           int(gradeID.Int64),
		Grade:             gardeName.String,
		ParentEmail:       parentEmail.String,
		ParentName:        parentName.String,
		StudentPictureURL: studentPicture.String,
		SchoolID:          int(schoolID.Int64),
		Role:              role.String,
		IsPhoneVerified:   isPhoneVerified.Bool,
		IsEmailVerified:   isEmailVerified.Bool,
		RefferalCode:      referralCode.String,
		SchoolName:        schoolName.String,
		User: models.User{
			Phone:       phoneNumber.String,
			Timezone:    timeZone.String,
			DialingCode: dialingCode.String,
			Location:    locStruct,
			QueryParams: queryParam.String,
		},
	}
	return studentDetails, nil
}

// GetUserProfilByToken retrieves the student ID associated with the provided JWT token from the 'student_auth' table.
// This function queries the database to find the student ID linked to the given token, which is used for identifying the student during authentication.
//
// Input Parameters:
// - token: string - The JWT token that was used for the student’s login session.
//
// Output:
// - (int, error) - The function returns the student ID associated with the provided token and an error if there are issues with connecting to the database or executing the query; otherwise, it returns the student ID and nil.
//
// On success, the function returns the student ID of the student associated with the JWT token.
//
// On failure, it logs the error and returns 0 with the error message.
func GetUserProfilByToken(token string) (int, error) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("GetUserProfilByToken: failed while trying to connect with database with error: ", err)
		return 0, err
	}
	defer db.Close()
	// Ensure the database connection is closed when the function returns.

	// Define the SQL query to fetch the student ID from the 'student_auth' table using the provided JWT token.
	query := `
				SELECT student_id FROM student_auth WHERE jwt_token=$1`
	var studentID int
	// Execute the query with the provided token and retrieve the student ID.

	err = db.QueryRow(query, token).Scan(&studentID)
	if err != nil {
		// If there is an error executing the query or scanning the result, log the error and return 0 with the error.
		log.Println("GetUserProfilByToken: failed while execute the query with error: ", err)
		return 0, err
	}
	// Return the student ID associated with the JWT token and nil indicating no errors.
	return studentID, nil
}

// UpdateStudentProfile updates the details of a student in the database based on the provided student ID.
// It updates the student's name, grade, and parent's email address.
//
// Input Parameters:
// - id: int - The ID of the student whose details need to be updated.
// - details: DemoSession - A struct containing the new details for the student, including the student’s name, grade, and parent’s email.
//
// Output:
// - error - The function returns an error if there are issues connecting to the database, executing the update query, or other database operations; otherwise, it returns nil.
//
// On success, the function performs the update operation and returns nil.
//
// On failure, it logs the error and returns the error object with a descriptive message.
func UpdateStudentProfile(userID int, details Student) (int, string, error) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return the error.
		log.Println("UpdateStudentProfile: failed while trying to connect with database with error: ", err)
		return 0, "", err
	}
	defer db.Close()
	// Ensure the database connection is closed when the function returns.
	var updateFields []string

	if details.SchoolID != 0 {
		updateFields = append(updateFields, fmt.Sprintf("school_id = %d", details.SchoolID))
	}
	if len(details.StudentPictureURL) > 0 {
		updateFields = append(updateFields, fmt.Sprintf("picture_url = '%s'", details.StudentPictureURL))
	}
	if len(details.ParentName) > 0 {
		updateFields = append(updateFields, fmt.Sprintf("parent_name = '%s'", details.ParentName))
	}

	updateQuery := ""
	if len(updateFields) > 0 {
		updateQuery = ", " + strings.Join(updateFields, ", ")
	}
	// Define the SQL query to update the student’s name, grade, and parent’s email in the 'student' table.
	query :=
		`WITH upserted_student AS (
		INSERT INTO
				student(
					user_id,
					name,
					grade_id,
					parent_email
				)
			VALUES(
				$1,$2,$3,$4
			)ON CONFLICT (user_id,name) DO UPDATE 
				 SET
				name = $2,
				grade_id = $3,
				parent_email = $4` + updateQuery + ` RETURNING user_id, id
			)
			SELECT 
                us.id AS student_id,
                u2s.stripe_id
            FROM 
                upserted_student us
           left JOIN 
                public.user2stripe u2s
            ON 
                us.user_id = u2s.user_id`
	// Execute the query with the provided details and the student ID.
	var (
		studentID        int
		customerStripeID sql.NullString
	)
	err = db.QueryRow(query, userID, details.StudentName, details.GradeID, details.ParentEmail).Scan(&studentID, &customerStripeID)
	// If there is an error executing the query, log the error and return the error.
	if err != nil {
		log.Println("UpdateStudentProfile: failed while execute the query with error: ", err)
		return 0, "", err
	}
	cusStripeID := customerStripeID.String
	// If the update is successful, return nil indicating no errors.
	return studentID, cusStripeID, nil
}

// SendReminderEmailForSession retrieves student information from the database to send reminder emails
// for sessions that are scheduled to start within the next 15 minutes. The function connects to the
// database, executes a query to fetch relevant details including the student's email, phone number,
// dialing code, session start time, student name, meeting link, session type, and timezone. The retrieved
// data is then stored in a slice of `Student` structs and returned. If there is an error during database
// connection, query execution, or row scanning, the error is logged and returned.
func SendReminderEmailForSession() ([]Student, error) {
	student := []Student{}
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SendReminderEmailForSession : Failed while trying to connect with database with error :", err)
		return student, err
	}
	defer db.Close()
	query := `
						SELECT 
                            st.parent_email,
                            u.phone_number,
							u.dialing_code,
							s.timezone,
                            st.name,
                            s.session_start,
							t.meeting_link,
							s.type
                        FROM
                            public.user AS u
                        JOIN 
                            session AS s ON u.id = s.user_id
                        JOIN
                            student as st ON u.id = st.user_id
						JOIN 
							tutor as t ON t.id= s.tutor_id
                        WHERE 
                             s.session_start >(NOW() AT TIME ZONE 'UTC') 
                           AND s.session_start <= (NOW() AT TIME ZONE 'UTC') + INTERVAL '15 minutes'`
	rows, err := db.Query(query)
	if err != nil {
		log.Println("SendReminderEmailForSession : Failed while execut the query with error :", err)
		return student, err
	}
	defer rows.Next()
	for rows.Next() {
		var (
			email        sql.NullString
			studentName  sql.NullString
			startSession sql.NullTime
			phoneNumber  sql.NullString
			dialingCode  sql.NullString
			timeZone     sql.NullString
			meetingLink  sql.NullString
			sessioType   sql.NullString
		)

		err := rows.Scan(&email, &phoneNumber, &dialingCode, &timeZone, &studentName, &startSession, &meetingLink, &sessioType)
		if err != nil {
			log.Println("SendReminderEmailForSession : Failed while scanning the rows with error :", err)
			return student, err
		}
		student = append(student, Student{
			ParentEmail:      email.String,
			SessionStartTime: startSession.Time,
			StudentName:      studentName.String,
			User: models.User{
				Phone:       phoneNumber.String,
				DialingCode: dialingCode.String,
			},
			MeetingLink:     utility.SQLNullStringToString(meetingLink),
			SessionTimezone: timeZone.String,
			SessionType:     sessioType.String,
		})
	}
	return student, nil
}

// IsRegularSessionBooked checks whether there is a booked regular session for a specific user in the database.
// A regular session is defined as a session with type 'paid'.
//
// Input Parameters:
// - userID: int - The ID of the user for whom the session booking status is being checked.
//
// Output:
// - bool - Returns true if there is at least one paid session booked for the given user, otherwise false.
// - error - Returns an error if there are issues connecting to the database or executing the query; otherwise, returns nil.
//
// On success, the function performs a query to check the existence of a paid session for the specified user and returns the result along with nil for error.
//
// On failure, it logs the error with a descriptive message and returns the error object along with the current value of `isSessionBooked` (which will be false in the case of an error).

func IsRegularSessionBooked(userID int) (bool, error) {
	var isSessionBooked bool
	db, err := config.GetDB2()
	if err != nil {
		log.Println("IsRegularSessionBooked: Failed when try to connect with database with error: ", err)
		return false, err
	}
	defer db.Close()
	query := `
				SELECT EXISTS(
					SELECT 1
						FROM session
						WHERE 
							type = 'paid' AND
							user_id = $1
							)`
	err = db.QueryRow(query, userID).Scan(&isSessionBooked)
	if err != nil {
		log.Println("IsRegularSessionBooked: Failed while execute the query with error: ", err)
		return isSessionBooked, err
	}
	return isSessionBooked, nil
}

// GetLatestSubscriptionPlan retrieves the latest subscription plan for a specific user from the database.
// It queries the database for the most recent subscription associated with the given user ID and returns details about this subscription.
//
// Input Parameters:
// - userID: int - The ID of the user whose latest subscription plan is to be retrieved.
//
// Output:
// - map[string]interface{} - A map containing the details of the latest subscription plan. The map includes:
//   - "plan_id": int64 - The ID of the subscription plan.
//   - "frequency": string - The billing frequency of the subscription (e.g., monthly, quarterly).
//   - "max_session": int64 - The maximum number of sessions allowed under the subscription.
//   - "per_week_limit": int64 - The limit on the number of sessions per week for the subscription.
//   - "status": string - The current status of the subscription (e.g., active, expired).
//   - "name": string - The name of the subscription plan.
//
// - error - The function returns an error if there are issues connecting to the database, executing the query, or other database operations; otherwise, it returns nil.
//
// On success, the function retrieves the latest subscription plan details and returns them in a map.
//
// On failure, it logs the error and returns the map with no data and an error object. If no subscription is found for the user, the function logs this and returns an empty map and a nil error.

func GetLatestSubscriptionPlan(userID int) (map[string]interface{}, error) {
	var subscription map[string]interface{}
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetLatestSubscriptionPlan: Failed when try to connect with database with error: ", err)
		return subscription, err
	}
	defer db.Close()
	var (
		planID            sql.NullInt64
		frequency         sql.NullString
		maxSession        sql.NullInt64
		classPerWeekLimit sql.NullInt64
		status            sql.NullString
		name              sql.NullString
	)
	query := `
		SELECT  sp.id,                                   
    			sp.frequency,
    			sp.max_session,
    			sp.per_week_limit,
    			usb.status,
				sp.name
		FROM 
		    public.user AS u 
		JOIN 
		    user2subscription AS usb ON usb.user_id = u.id
		JOIN 
		    subscription_plan AS sp ON usb.subscription_plan_id = sp.id
		WHERE 
   			u.id = $1 AND usb.status = 'active' 
		ORDER BY 
			purchased_date DESC
		LIMIT 1`
	err = db.QueryRow(query, userID).Scan(
		&planID,
		&frequency,
		&maxSession,
		&classPerWeekLimit,
		&status,
		&name,
	)
	if err == sql.ErrNoRows {
		log.Println("GetLatestSubscriptionPlan: No subscription found: ", err)
		return subscription, err
	}
	if err != nil {
		log.Println("GetLatestSubscriptionPlan: Failed while execute the query with error: ", err)
		return subscription, err
	}
	subscription = make(map[string]interface{})
	subscription["max_session"] = maxSession.Int64
	subscription["per_week_limit"] = classPerWeekLimit.Int64
	subscription["status"] = status.String
	subscription["frequency"] = frequency.String
	subscription["name"] = name.String
	subscription["plan_id"] = planID.Int64
	return subscription, nil
}

func GetActiveSubscriptionDetailsRenewalDate(userID int) (string, error) {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetActiveSubscriptionDetailsRenewalDate: Failed when try to connect with database with error: ", err)
		return "", err
	}
	defer db.Close()
	var (
		renewalDate sql.NullString
	)
	query := `
		SELECT  usb.valid_until                                  
    			
		FROM 
		    public.user AS u 
		JOIN 
		    user2subscription AS usb ON usb.user_id = u.id
		JOIN 
		    subscription_plan AS sp ON usb.subscription_plan_id = sp.id
		WHERE 
   			u.id = $1
		ORDER BY 
			purchased_date DESC
		LIMIT 1`
	err = db.QueryRow(query, userID).Scan(
		&renewalDate,
	)
	if err == sql.ErrNoRows {
		log.Println("GetActiveSubscriptionDetailsRenewalDate: No subscription found: ", err)
		return "", err
	}
	if err != nil {
		log.Println("GetActiveSubscriptionDetailsRenewalDate: Failed while execute the query with error: ", err)
		return "", err
	}
	return renewalDate.String, nil
}

// SendReminderEmailForBookRegularSession retrieves student information from the database and sends reminder emails to students
// who have purchased a subscription plan but have not yet booked a paid session.
// The function connects to the PostgreSQL database and executes a query to fetch the
// phone number, parent email, dialing code, subscription plan name, and maximum session count for each student who meets the criteria.
// The retrieved data is stored in a slice of maps, where each map contains the relevant information for a student.
// The function returns the slice of maps containing student data, or an error if the database connection or query execution fails.

func SendReminderEmailForBookRegularSession() ([]map[string]interface{}, error) {
	var studentsToRemind []map[string]interface{}
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SendReminderEmailForBookRegularSession: Failed when try to connect with database with error: ", err)
		return studentsToRemind, err
	}
	defer db.Close()
	query := `
			SELECT 
				u.phone_number,
				s.parent_email,
				u.dialing_code,
				sp.name,
				sp.max_session
			FROM
				public.user as u
			JOIN
				student AS s ON s.user_id = u.id
			JOIN
				user2subscription AS u2s ON u2s.user_id  = u.id
			JOIN 
				subscription_plan AS sp ON sp.id   = u2s.subscription_plan_id
			WHERE
				u.id NOT IN
				(
				SELECT
					user_id
				FROM 
					session
				WHERE 
					type = 'paid'
				)`
	rows, err := db.Query(query)
	if err != nil {
		log.Println("SendReminderEmailForBookRegularSession : Failed while execut the query with error :", err)
		return studentsToRemind, err
	}
	defer rows.Next()
	for rows.Next() {
		var (
			email       sql.NullString
			phoneNumber sql.NullString
			planName    sql.NullString
			dialingCode sql.NullString
			maxSession  sql.NullInt64
		)

		err := rows.Scan(&phoneNumber, &email, &dialingCode, &planName, &maxSession)
		if err != nil {
			log.Println("SendReminderEmailForBookRegularSession : Failed while scanning the rows with error :", err)
			return studentsToRemind, err
		}
		record := map[string]interface{}{
			"email":        email.String,
			"phone":        phoneNumber.String,
			"dialing_code": dialingCode.String,
			"plan_name":    planName.String,
			"max_session":  maxSession.Int64,
		}
		studentsToRemind = append(studentsToRemind, record)
	}
	return studentsToRemind, nil
}

// GetPhoneNumbersForDemoReminder retrieves the phone numbers and dialing codes of students who need to receive a reminder SMS for demo booking.
// The function performs the following steps:
// 1. Connects to the database. If the connection fails, logs the error and returns an empty slice and the error.
// 2. Executes a SQL query to fetch phone numbers and dialing codes from the `public.user` table where:
//    - The phone number is verified (`phone_verified = true`).
//    - The dialing code is not null (`dialing_code IS NOT NULL`).
//    - The user has not booked any sessions (`id NOT IN (SELECT user_id FROM session)`).
// 3. Iterates over the query results to populate the `phoneNumberInfo` slice with the retrieved data, which is stored as instances of the `models.User` struct.
//    - For each row, it retrieves the `phone_number` and `dialing_code`, handling null values appropriately.
//    - If an error occurs while scanning the rows, logs the error and returns the partially populated slice along with the error.
// 4. Returns the `phoneNumberInfo` slice containing the phone numbers and dialing codes of eligible students, or an error if any issues occurred during the process.

func GetPhoneNumbersForDemoReminder() ([]models.User, error) {
	var phoneNumberInfo []models.User
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] GetPhoneNumbersForDemoReminder: Failed when try to connect with database with error: ", err)
		return phoneNumberInfo, err
	}
	defer db.Close()
	query := `
		 	SELECT 
                phone_number,
				dialing_code
            FROM 
                public.user
            WHERE
                phone_verified = true AND
				dialing_code IS NOT NULL AND
                id NOT IN(
                         	SELECT
                                user_id
                            FROM
                            	session)`
	rows, err := db.Query(query)
	if err != nil {
		log.Println("[ERROR] GetPhoneNumbersForDemoReminder: Failed when try to execute query with database with error: ", err)
		return phoneNumberInfo, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			phone       sql.NullString
			dialingCode sql.NullString
		)
		err := rows.Scan(&phone, &dialingCode)
		if err != nil {
			log.Println("[ERROR] GetPhoneNumbersForDemoReminder: Failed when try to scanning rows from database with error: ", err)
			return phoneNumberInfo, err
		}
		phoneNumberInfo = append(phoneNumberInfo, models.User{
			Phone:       phone.String,
			DialingCode: dialingCode.String,
		})
	}
	return phoneNumberInfo, nil
}

// SessionReminderNotificationBeforeHour retrieves a list of students who have a session starting
// within the next given hours and compiles the necessary details for sending reminder notifications.
//
// The function connects to the database, executes a query to find sessions scheduled
// within the next given hours, and retrieves associated student and session details.
// It returns a slice of Student structs containing the relevant information or an error if one occurs.
//
// Returns:
// - []Student: A slice containing details of students with upcoming sessions.
// - error: Any error encountered during database connection, query execution, or row scanning.
func SessionReminderNotificationBeforeHour(reminderHour int, sessionType, reminderDate string) ([]Student, error) {
	var reminderDetails []Student

	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] SessionReminderNotificationBeforeHour: Failed to connect to the database: ", err)
		return reminderDetails, err
	}
	defer db.Close()
	query := " where "
	if len(sessionType) != 0 {
		// Prepare the WHERE clause conditionally if session type is "paid"
		if sessionType == constants.SessionTypePaid {
			query += fmt.Sprintf("s.type = '%s' AND ", constants.SessionTypePaid)

		}
	}
	if len(reminderDate) == 0 {
		query += fmt.Sprintf("s.session_start > (NOW() AT TIME ZONE 'UTC') AND s.session_start <= (NOW() AT TIME ZONE 'UTC') + make_interval(hours => %d)", reminderHour)
	} else {
		query += fmt.Sprintf("DATE(session_start) = '%s'", reminderDate)
	}

	// Define the SQL query to retrieve session reminder information.
	query = `
		SELECT 
			s.id,
			st.parent_email,
			u.phone_number,
			u.dialing_code,
			s.timezone,
			s.session_start,
			s.type,
			st.name,
			t.id,
			u.id,
			t.email_id,
			t.timezone,
			t.name,
			t.meeting_link
		FROM
			public.user AS u
		JOIN 
			session AS s ON u.id = s.user_id
		LEFT JOIN
			student AS st ON u.id = st.user_id
		JOIN
			tutor AS t ON t.id = s.tutor_id ` + query

	// Execute the query
	rows, err := db.Query(query)

	if err != nil {
		log.Println("[ERROR] SessionReminderNotificationBeforeHour: Failed to execute the query: ", err)
		return reminderDetails, err
	}
	defer rows.Close()

	// Iterate over the result set and scan each row into the corresponding variables.
	for rows.Next() {
		var (
			sessionID, tutorId, StudentId                                              sql.NullInt64
			parentEmail, tutorEmail, tutorTimezone, tutorName, tutorMeetingLink        sql.NullString
			sessionStartTime                                                           sql.NullTime
			studentPhoneNumber, dialingCode, sessionTimeZone, sessionType, studentName sql.NullString
		)

		// Scan the current row into variables.
		err := rows.Scan(&sessionID, &parentEmail, &studentPhoneNumber, &dialingCode, &sessionTimeZone, &sessionStartTime, &sessionType,
			&studentName, &tutorId, &StudentId, &tutorEmail, &tutorTimezone, &tutorName, &tutorMeetingLink)
		if err != nil {
			log.Println("[ERROR] SessionReminderNotificationBeforeHour: Failed to scan the row: ", err)
			return reminderDetails, err
		}

		// Append the scanned data to the reminderDetails slice.
		reminderDetails = append(reminderDetails, Student{
			User: models.User{
				ID:          StudentId.Int64,
				Phone:       studentPhoneNumber.String,
				DialingCode: dialingCode.String,
			},
			StudentName:      studentName.String,
			SessionStartTime: sessionStartTime.Time,
			SessionType:      sessionType.String,
			SessionTimezone:  sessionTimeZone.String,
			SessionID:        int(sessionID.Int64),
			ParentEmail:      parentEmail.String,
			Tutor: tutor.Tutor{
				ID:         int(tutorId.Int64),
				TutorEmail: tutorEmail.String,
				Timezone:   tutorTimezone.String,
				TutorName:  tutorName.String,
				ZoomLink:   tutorMeetingLink.String,
			},
		})
	}

	// Return the collected reminder details and any potential error.
	return reminderDetails, nil
}
func GetChildOfUser(userID int) ([]Child, error) {
	var chiledList []Child

	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] GetChildOfUser: Failed to connect to the database: ", err)
		return chiledList, err
	}
	defer db.Close()

	query := `
		SELECT   
				s.id,                                  
    			r.role,
    			s.name,
    			s.parent_name,
    			s.parent_email,
    			s.grade_id,
    			s.picture_url,
    			s.school_id,
    			u.phone_verified,
    			u.email_verified,
    			u.referral_code,
    			u.phone_number,
    			u.timezone,
    			u.dialing_code,
    			u.location,
				u.query_param,
				sc.name
		FROM 
		    public.user AS u 
		LEFT JOIN
		    student AS s ON u.id = s.user_id
		LEFT JOIN 
		    user2roles AS ur ON u.id = ur.user_id
		LEFT JOIN 
		    roles AS r ON ur.role_id = r.id
		LEFT JOIN
			schools AS sc ON s.school_id = sc.id	
		WHERE 
   			 u.id = $1
`

	rows, err := db.Query(query, userID)
	if err != nil {
		log.Println("[ERROR] : GetChildOfUser : Failed while trying to execute the query with error :", err)
		return chiledList, err
	}
	for rows.Next() {
		var (
			schoolID, gradeID, studentID                                                          sql.NullInt64
			isPhoneVerified, isEmailVerified                                                      sql.NullBool
			timeZone, dialingCode, location, queryParam, schoolName                               sql.NullString
			role, studentName, parentEmail, parentName, studentPicture, referralCode, phoneNumber sql.NullString
		)
		err := rows.Scan(
			&studentID,
			&role,
			&studentName,
			&parentName,
			&parentEmail,
			&gradeID,
			&studentPicture,
			&schoolID,
			&isPhoneVerified,
			&isEmailVerified,
			&referralCode,
			&phoneNumber,
			&timeZone,
			&dialingCode,
			&location,
			&queryParam,
			&schoolName,
		)
		if err != nil {
			log.Println("[ERROR] GetChildOfUser : Failed while scaning rows with error :", err)
			return chiledList, err
		}
		if studentID.Valid {
			chiledList = append(chiledList, Child{
				ID:                     studentID.Int64,
				StudentName:            studentName.String,
				ParentEmail:            parentEmail.String,
				GradeID:                int(gradeID.Int64),
				StudentPictureURL:      studentPicture.String,
				SchoolID:               int(schoolID.Int64),
				SchoolName:             schoolName.String,
				ParentPhone:            phoneNumber.String,
				ParentPhoneDialingCode: dialingCode.String,
			})
		}

	}
	return chiledList, nil

}

func GetStudentGrade(studnetID int) (int, error) {
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] GetStudentGrade: Failed to connect to the database: ", err)
		return 0, err
	}
	defer db.Close()

	query := `select grade_id from student where id=$1`
	var gradeID sql.NullInt64
	err = db.QueryRow(query, studnetID).Scan(&gradeID)
	if err != nil {
		log.Println("[ERROR] GetStudentGrade: Failed to fetch from database: ", err)
		return 0, err
	}
	return int(gradeID.Int64), nil

}

type RequestDataToUpdateChildDetails struct {
	ID           int
	Name         string
	GradeID      int
	ParentEmail  string
	ChildOldName string
}

// UpdateChildDetails updates specific fields for a child's record in the database based on provided data.
//
// Parameters:
// - childData: A struct containing updated details of the child (GradeID, ParentEmail, and StudentName).
// - childOldName: The current name of the child in the database, used to locate the correct record.
//
// Returns:
// - An error if there is an issue connecting to the database or executing the update query.
func UpdateChildDetails(childData RequestDataToUpdateChildDetails) (string, error) {
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] UpdateChildDetails: Failed to connect to the database: ", err)
		return "", err
	}
	defer db.Close()
	var updateFields []string

	if childData.GradeID != 0 {
		updateFields = append(updateFields, fmt.Sprintf(" grade_id = %d ", childData.GradeID))
	}
	if len(childData.ParentEmail) > 0 {
		updateFields = append(updateFields, fmt.Sprintf(" parent_email = '%s' ", childData.ParentEmail))
	}
	if len(childData.Name) > 0 {
		updateFields = append(updateFields, fmt.Sprintf(" name = '%s' ", childData.Name))
	}

	updateQuery := ""
	if len(updateFields) > 0 {
		updateQuery = strings.Join(updateFields, ", ")
	}

	query := `
            UPDATE student
            SET ` + updateQuery + `
            WHERE id = $1
            RETURNING (
                SELECT u2s.stripe_id
                FROM public.user2stripe AS u2s
                WHERE u2s.user_id = student.user_id
            ) AS stripe_id `

	row := db.QueryRow(query, childData.ID)
	var stripeID sql.NullString
	if err := row.Scan(&stripeID); err != nil {
		log.Println("[ERROR] UpdateChildDetails: Failed to execute the combined query with error: ", err)
		return "", err
	}

	return stripeID.String, nil
}

type InfoToUpdate struct {
	StripeCustomerId string
	StudentName      string
	StudentEmail     string
}

func GetAllCustomerInfoToUpdateOnstripe() ([]InfoToUpdate, error) {
	var info2updateList []InfoToUpdate

	// Establish a connection to the database.
	db, err := config.GetDB2() // Replace this with your DB initialization function
	if err != nil {
		log.Println("[ERROR] GetAllCustomerInfoToUpdateOnstripe: Failed to connect to the database:", err)
		return nil, err
	}
	defer db.Close()

	// SQL query to fetch the required data.
	query := `
        	SELECT 
            u2s.stripe_id,
            st.name,
            st.parent_email
        FROM 
            public.user2stripe AS u2s
        JOIN 
            public.student AS st 
        ON 
            u2s.user_id = st.user_id
        WHERE 
            u2s.updated_at IS NULL`

	// Execute the query.
	rows, err := db.Query(query)
	if err != nil {
		log.Println("[ERROR] GetAllCustomerInfoToUpdateOnstripe: Failed to execute the query:", err)
		return nil, err
	}
	defer rows.Close()

	// Iterate through the results and build the list.
	for rows.Next() {
		var stripeID, studentName, studentEmail sql.NullString

		err := rows.Scan(&stripeID, &studentName, &studentEmail)
		if err != nil {
			log.Println("[ERROR] GetAllCustomerInfoToUpdateOnstripe: Failed to scan row:", err)
			return nil, err
		}

		// Append to the struct after handling null values.
		info2updateList = append(info2updateList, InfoToUpdate{
			StripeCustomerId: utility.SQLNullStringToString(stripeID),
			StudentName:      utility.SQLNullStringToString(studentName),
			StudentEmail:     utility.SQLNullStringToString(studentEmail),
		})
	}

	// Check for any errors during iteration.
	if err := rows.Err(); err != nil {
		log.Println("[ERROR] Error during row iteration:", err)
		return nil, err
	}

	return info2updateList, nil
}

// GetStudentToSendReminderToSchedule fetch the students list who is signed up from ppc page
// but did not take any further action like booking-demo or buying plans.
func GetStudentToSendReminderToSchedule() ([]StudentToScheduleReminder, error) {
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] GetStudentToSendReminderToSchedule: Failed to connect to the database: ", err)
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT
			u.id,
			u.phone_number,
			u.timezone,
			st.name,
			st.parent_email
		FROM public.user AS u
		LEFT JOIN public.student AS st ON st.user_id = u.id
		LEFT JOIN public.session AS s ON s.user_id = u.id AND s.student_id = st.id
		WHERE
			s.id IS NULL
			AND (
      			u.query_param LIKE '%gclid%'
				OR u.query_param LIKE '%gbraid%'
				OR u.query_param LIKE '%wbraid%'
				OR u.query_param LIKE '%msclkid%'
				OR u.query_param LIKE '%flyer%'
   			)`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("[ERROR] GetStudentToSendReminderToSchedule: Failed to execute the query with errro: ", err)
		return nil, err
	}
	var studentList []StudentToScheduleReminder
	for rows.Next() {
		var uid sql.NullInt64
		var name, timezone, phone, email sql.NullString
		err := rows.Scan(&uid, &phone, &timezone, &name, &email)
		if err != nil {
			log.Println("[ERROR] GetStudentToSendReminderToSchedule: Failed to scan the results from database: ", err)
			return nil, err
		}
		studentList = append(studentList, StudentToScheduleReminder{
			ID:       uid.Int64,
			Name:     name.String,
			Phone:    phone.String,
			Email:    email.String,
			Timezone: timezone.String,
		})
	}
	return studentList, nil
}
func UpdateStudentSessionPreference(userID int) error {
	// Connect to the database
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] UpdateStudentSessionPreference: Failed to connect to the database: ", err)
		return err
	}
	defer db.Close()

	// SQL query to update student session preference
	query := `
        WITH existing_entry AS (
            SELECT id 
            FROM public.student
            WHERE user_id = $1
        )
        UPDATE public.student_session_preference
        SET student_id = (SELECT id FROM existing_entry)
        WHERE user_id = $1 AND student_id IS NULL
    `

	// Execute the query
	_, err = db.Exec(query, userID)
	if err != nil {
		log.Println("[ERROR] UpdateStudentSessionPreference: Failed to execute query: ", err)
		return err
	}

	log.Println("[INFO] UpdateStudentSessionPreference: Successfully updated session preference for userID:", userID)
	return nil
}
