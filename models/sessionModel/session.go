package sessionmodel

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"tutree/student-apis/config"
	"tutree/student-apis/constants"
	"tutree/student-apis/models"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/models/tutor"
	"tutree/student-apis/utility"
)

type Session struct {
	SessionID       int              `json:"session_id"`
	SessionType     string           `json:"session_type,omitempty"`
	StartSession    time.Time        `json:"start_session,omitempty"`
	EndSession      time.Time        `json:"end_session,omitempty"`
	Tutor           tutor.Tutor      `json:"tutor,omitempty"`
	Student         students.Student `json:"students,omitempty"`
	Status          string           `json:"status,omitempty"`
	SessionDuration int              `json:"session_duration,omitempty"`
	RecordingURL    string           `json:"recording_url,omitempty"`
	SessionLiked    bool             `json:"session_liked,omitempty"`
	Review          string           `json:"review,omitempty"`
	TimeZone        string           `json:"timezone,omitempty"`
	SessionDayTime  []SessionDayTime `json:"slots,omitempty"`
	TutorID         int              `json:"tutor_id,omitempty"`
	StartFrom       string           `json:"start_from,omitempty"`
	GameURL         string           `json:"game_url,omitempty"`
	StudentName     string           `json:"student_name,omitempty"`
	StudentID       int              `json:"student_id,omitempty"`
	SubjectName     string           `json:"subject_name,omitempty"`
	SubjectID       int              `json:"subject_id,omitempty"`
	Grade           string           `json:"grade,omitempty"`
	GradeID         int              `json:"grade_id,omitempty"`
	Credited        bool             `json:"credited,omitempty"`
}

type SessionBookRequest struct {
	SessionType    string           `json:"session_type,omitempty"`
	TutorID        int              `json:"tutor_id,omitempty"`
	StartFrom      string           `json:"start_from,omitempty"`
	SessionDayTime []SessionDayTime `json:"day_time,omitempty"`
	TimeZone       string           `json:"timezone,omitempty"`
	Slots          []SessionSlots   `json:"slots,omitempty"`
	SubjectID      int              `json:"subject_id,omitempty"`
	StudentID      int              `json:"student_id,omitempty"`
}

type SessionDayTime struct {
	UserID    int    `json:"user_id,omitempty"`
	TutorID   int    `json:"tutor_id,omitempty"`
	Day       string `json:"day,omitempty"`
	Time      string `json:"time,omitempty"`
	TimeZone  string `json:"timezone,omitempty"`
	SubjectID int    `json:"subject_id,omitempty"`
	StudentID int    `json:"student_id,omitempty"`
	StartDate string `json:"start_date,omitempty"`
}
type SessionSlots struct {
	SessionStart time.Time `json:"start,omitempty"`
	SessionEnd   time.Time `json:"end,omitempty"`
}
type RequestDataForDemoSessionBooking struct {
	SessionDateTime time.Time
	TutorID         int
	TimeZone        string
	SessionType     string
	SubjectID       int
	StudentID       int
	StudentName     string
	Email           string
}

// SaveDemoSessionInfoOfStudent inserts a new demo session record into the 'session' table for a specific student and returns the newly created session ID.
// It sets the session type to 'demo', the status to 'active', and records the current timestamp for the creation of the session.
//
// Input Parameters:
// - StudentID: int - The ID of the student who is booking the demo session.
// - Sessiondetails: students.DemoSession - A struct containing the details of the demo session, including the start time and the ID of the tutor.
//
// Output:
// - (int, error) - The function returns the ID of the newly created session and an error if there are issues connecting to the database,
// executing the query, or other database operations; otherwise, it returns the session ID and nil.
//
// On success, the function inserts a new record into the 'session' table and returns the ID of the new session.
//
// On failure, it logs the error and returns 0 with the error message.

func SaveDemoSessionInfoOfStudent(userID int, requestData RequestDataForDemoSessionBooking) (error, int) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("SaveDemoSessionInfoOfStudent: failed while trying to connect with database with error: ", err)
		return err, 0
	}
	defer db.Close()
	// Ensure the database connection is closed when the function returns.

	// Define the SQL query to insert a new record into the 'session' table.
	// This query inserts the student ID, session start time, type of session ('demo'),
	// tutor ID, status ('active'), and the current timestamp for the creation of the session.
	endSession := requestData.SessionDateTime.Add(40 * time.Minute)
	query :=
		`INSERT INTO
    	session
    (
        user_id,
        session_start,
        session_end,
        tutor_id,
        duration_mins,
        type,
        status,
        created_at,
		timezone,
		student_id,
		subject_id,
		updated_at
    )
	VALUES
    	($1, $2, $3, $4, $5, $6, 'active', NOW(),$7,$8,$9,NOW())
	RETURNING id`
	var sessionID int
	// Execute the query with the provided student ID, session start time, and tutor ID.
	// Retrieve the ID of the newly created session using the RETURNING clause.
	err = db.QueryRow(query,
		userID,
		requestData.SessionDateTime,
		endSession,
		requestData.TutorID,
		constants.SessionDuration,
		requestData.SessionType,
		requestData.TimeZone,
		requestData.StudentID,
		requestData.SubjectID).Scan(&sessionID)
	// If there is an error executing the query, log the error and return 0 with the error.
	if err != nil {
		log.Println("SaveDemoSessionInfoOfStudent: failed while execute the query with error: ", err)
		return err, 0
	}
	// Return the ID of the newly created session and nil indicating no errors.sss
	return nil, sessionID
}

type SessionListFilters struct {
	DateFrom    string
	DateTo      string
	Status      string
	SessionType string
	SubjectName string
	StudentName string
	Grade       string
	TutorID     int64
	UserID      int
}

type SessionListResponse struct {
	Sessions []Session `json:"sessions"`
}

func GetSessionlist(filters SessionListFilters) ([]Session, error) {
	var sessions []Session

	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetSessionlist: Failed while connecting with the database:", err)
		return nil, err
	}
	defer db.Close()

	query := `SELECT
	            s.id, 
				s.session_start, 
				s.session_end, 
				s.status, 
				s.type,
				s.recording_url,
				s.timezone,
				s.game_url,
				s.credited,
				sub.name,
				sub.id,
				st.name,
				gr.name,
				gr.id,
				t.id, 
				t.name, 
				t.picture_url, 
				t.phone_number, 
				t.email_id,
				t.meeting_link 
			  FROM public.session s 
			  JOIN public.tutor t ON s.tutor_id = t.id
			  LEFT JOIN public.subject sub ON s.subject_id = sub.id
			  LEFT JOIN public.student st  ON s.student_id = st.id
			  LEFT JOIN public.grade gr  ON gr.id = st.grade_id
			  WHERE s.user_id = $1 AND s.credited = true`

	var args []interface{}
	args = append(args, filters.UserID)

	argIndex := 2 // Starting index for optional filters (after UserID)
	// NOTE :-
	// [argIndex] is dynamically created according to filters it is creating and appending
	// $1 , $2 values dynamically which resolves the hardcoding of variables.

	// Add date range filter (DateFrom and DateTo)
	if filters.DateFrom != "" && filters.DateTo != "" {
		query += ` AND DATE(s.session_start) BETWEEN DATE($` + strconv.Itoa(argIndex+1) + `) AND DATE($` + strconv.Itoa(argIndex) + `)`
		args = append(args, filters.DateFrom, filters.DateTo)
		argIndex += 2
	}

	// Add status filter
	if len(filters.Status) != 0 && filters.Status != "" {
		query += ` AND s.status = $` + strconv.Itoa(argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	// Add tutor ID filter
	if filters.TutorID != 0 {
		query += ` AND s.tutor_id = $` + strconv.Itoa(argIndex)
		args = append(args, filters.TutorID)
		argIndex++
	}

	// Add session type filter
	if len(filters.SessionType) != 0 && filters.SessionType != "" {
		query += ` AND s.type = $` + strconv.Itoa(argIndex)
		args = append(args, filters.SessionType)
		argIndex++
	}

	// Add session Grade filter
	if len(filters.Grade) != 0 && filters.Grade != "" {
		query += ` AND TRIM(LOWER(gr.name)) = $` + strconv.Itoa(argIndex)
		args = append(args, filters.Grade)
		argIndex++
	}

	// Add session Subject filter
	if len(filters.SubjectName) != 0 && filters.SubjectName != "" {
		query += ` AND TRIM(LOWER(sub.name)) = $` + strconv.Itoa(argIndex)
		args = append(args, filters.SubjectName)
		argIndex++
	}

	// Add session studentName filter
	if len(filters.StudentName) != 0 && filters.StudentName != "" {
		query += ` AND TRIM(LOWER(st.name)) = $` + strconv.Itoa(argIndex)
		args = append(args, filters.StudentName)
		argIndex++
	}
	query += ` ORDER BY s.session_start DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("GetSessionlist: Failed while executing the query:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			sessionStart, sessionEnd                                                      sql.NullTime
			tutorName, tutorPictureURL, tutorPhoneNumber, tutorEmail, status, sessionType sql.NullString
			tutorId, sessionId, subjectId, gradeId                                        sql.NullInt64
			meetingLink, recordingUrl, timeZone, gameUrl, grade, studentName, subjectName sql.NullString
			credited                                                                      sql.NullBool
		)
		err := rows.Scan(&sessionId, &sessionStart, &sessionEnd, &status, &sessionType, &recordingUrl, &timeZone, &gameUrl, &credited, &subjectName, &subjectId, &studentName, &grade, &gradeId, &tutorId, &tutorName, &tutorPictureURL, &tutorPhoneNumber, &tutorEmail, &meetingLink)
		if err != nil {
			log.Println("GetSessionlist: Failed while scanning row:", err)
			return nil, err
		}

		session := Session{
			SessionID:    int(sessionId.Int64),
			StartSession: sessionStart.Time,
			EndSession:   sessionEnd.Time,
			Status:       status.String,
			SessionType:  sessionType.String,
			RecordingURL: recordingUrl.String,
			TimeZone:     timeZone.String,
			GameURL:      gameUrl.String,
			Credited:     credited.Bool,
			SubjectID:    int(subjectId.Int64),
			SubjectName:  subjectName.String,
			StudentName:  studentName.String,
			GradeID:      int(gradeId.Int64),
			Grade:        grade.String,
			Tutor: tutor.Tutor{
				ID:         int(tutorId.Int64),
				TutorName:  tutorName.String,
				PictureURL: tutorPictureURL.String,
				Phone:      tutorPhoneNumber.String,
				TutorEmail: tutorEmail.String,
				ZoomLink:   meetingLink.String,
			},
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		log.Println("GetSessionlist: Error after iterating rows:", err)
		return nil, err
	}

	return sessions, nil
}

// GetRecentDemoSesssion retrieves demo session information for a given student from the database.
// Parameter -
//
//	userID - the ID of the user whose demo sessions are to be fetched.
//
// Return -
//
//	a slice of Session structs containing session details and an error if any issues occur
//
// during database operations.
func GetRecentDemoSessions(userID int) ([]map[string]interface{}, error) {
	var demoSessions []map[string]interface{}
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] GetRecentDemoSessions: Failed while trying to connect with the database with error: ", err)
		return demoSessions, err
	}
	// Ensure the database connection is closed when the function returns.
	defer db.Close()

	// Define the SQL query to retrieve session details and feedback.
	query := `
		WITH RankedSessions AS (
    		SELECT 
    		    s.id,
    		    s.session_start,
    		    s.status,
    		    session_liked,
    		    sf.comment,
    		    s.timezone,
    		    s.subject_id AS subject_id,
    		    st.grade_id AS grade_id,
    		    st.id AS student_id,
    		    sb.name AS subject_name,
    		    ROW_NUMBER() OVER (
    		        PARTITION BY st.grade_id, st.id 
    		        ORDER BY 
    		            CASE 
    		                WHEN s.status = 'active' THEN 1
    		                ELSE 2 
    		            END, 
    		            s.session_start DESC,
    		            sf.created_at DESC
    		    ) AS rank
    		FROM
    		    session AS s
    		LEFT JOIN 
    		    session_feedback AS sf ON s.id = sf.session_id
    		JOIN 
    		    student AS st ON s.student_id = st.id
    		JOIN 
    		    subject AS sb ON s.subject_id = sb.id
    		WHERE 
    		    s.user_id = $1
    		    AND s.type = 'demo'
		)
		SELECT *
		FROM RankedSessions
		WHERE rank = 1;`

	// Execute the query with the provided user ID.
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Println("[ERROR] GetRecentDemoSessions : Failed while executing the query with error:", err)
		return demoSessions, err
	}
	defer rows.Close()

	// Loop through all rows.
	for rows.Next() {
		var (
			sessionID, subjectID, gradeID, studentID, rank sql.NullInt64
			startDate                                      sql.NullTime
			status, subject                                sql.NullString
			sessionLiked                                   sql.NullBool
			review                                         sql.NullString
			timezone                                       sql.NullString
		)
		// Scan the row into variables.
		err := rows.Scan(&sessionID, &startDate, &status, &sessionLiked, &review, &timezone, &subjectID, &gradeID, &studentID, &subject, &rank)
		if err != nil {
			log.Println("[ERROR] GetRecentDemoSessions : Failed while scanning the rows from database with error:", err)
			return demoSessions, err
		}

		// Create a map to store the session details.
		demoSession := make(map[string]interface{})
		demoSession["date_time"] = startDate.Time
		demoSession["session_id"] = int(sessionID.Int64)
		demoSession["status"] = status.String
		demoSession["timezone"] = timezone.String
		demoSession["student_id"] = studentID.Int64
		demoSession["subject_id"] = subjectID.Int64
		demoSession["subject_name"] = subject.String
		demoSession["grade_id"] = gradeID.Int64

		// Set session_liked to nil if it's not valid.
		if sessionLiked.Valid {
			demoSession["session_liked"] = sessionLiked.Bool
		} else {
			demoSession["session_liked"] = nil
		}

		demoSession["review"] = review.String

		// Append the map to the slice.
		demoSessions = append(demoSessions, demoSession)
	}

	// Return the slice of demo sessions.
	return demoSessions, nil
}

func GetSession(sessionID, userID int64) (Session, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetSession: Failed while connecting with the database:", err)
		return Session{}, err
	}
	defer db.Close()
	var (
		sessionStart, sessionEnd                                                                    sql.NullTime
		tutorName, tutorPictureURL, tutorPhoneNumber, tutorEmail, status, sessionType, recordingUrl sql.NullString
		timeZone, gameURL, timezone, grade, studentName, subjectName, parentEmail                   sql.NullString
		tutorId, sessionDuration, subjectID, StudentID                                              sql.NullInt64
		credited                                                                                    sql.NullBool
		session                                                                                     Session
	)
	var where string
	if userID != 0 {
		where = fmt.Sprintf(" AND s.user_id = %d", userID)
	}
	query := fmt.Sprintf(`
    			SELECT 
    			    s.session_start, 
    			    s.session_end, 
    			    s.status, 
    			    s.type,
    			    s.recording_url,
    			    s.timezone,
    			    s.game_url,
    			    s.credited,
					sub.name,
				    st.name,
				    gr.name,
    			    t.id, 
    			    t.name, 
    			    t.picture_url, 
    			    t.phone_number, 
    			    t.email_id,
    			    s.timezone,
					s.duration_mins,
					s.subject_id,
					s.student_id,
					st.parent_email
    			FROM public.session s
    			JOIN public.tutor t ON s.tutor_id = t.id
				LEFT JOIN public.subject sub ON s.subject_id = sub.id
			    LEFT JOIN public.student st  ON s.student_id = st.id
			    LEFT JOIN public.grade gr  ON gr.id = st.grade_id
    			WHERE s.id = $1 AND s.credited = true %s`, where)

	row := db.QueryRow(query, sessionID)

	err = row.Scan(
		&sessionStart,
		&sessionEnd,
		&status,
		&sessionType,
		&recordingUrl,
		&timeZone,
		&gameURL,
		&credited,
		&subjectName,
		&studentName,
		&grade,
		&tutorId,
		&tutorName,
		&tutorPictureURL,
		&tutorPhoneNumber,
		&tutorEmail,
		&timezone,
		&sessionDuration,
		&subjectID,
		&StudentID,
		&parentEmail,
	)
	if err != nil {
		log.Println("GetSessionlist: Failed while scanning row:", err)
		return session, err
	}

	session = Session{
		StartSession:    sessionStart.Time,
		EndSession:      sessionEnd.Time,
		Status:          status.String,
		SessionType:     sessionType.String,
		RecordingURL:    recordingUrl.String,
		TimeZone:        timeZone.String,
		GameURL:         gameURL.String,
		Credited:        credited.Bool,
		StudentName:     studentName.String,
		SubjectName:     subjectName.String,
		Grade:           grade.String,
		SessionDuration: int(sessionDuration.Int64),
		SubjectID:       int(subjectID.Int64),
		StudentID:       int(StudentID.Int64),
		Student:         students.Student{ParentEmail: parentEmail.String},
		Tutor: tutor.Tutor{
			ID:         int(tutorId.Int64),
			TutorName:  tutorName.String,
			PictureURL: tutorPictureURL.String,
			Phone:      tutorPhoneNumber.String,
			TutorEmail: tutorEmail.String,
		},
	}
	return session, nil

}

func CancelSession(sessionID int64, cancelMinutes int) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("CancelSession: Failed while connecting with the database:", err)
		return err
	}
	defer db.Close()

	var returnedID int64
	query := `DELETE FROM public.session 
	                WHERE id = $1
	                  AND session_start > NOW() + ($2 || ' minutes')::interval 
	                RETURNING id`
	err = db.QueryRow(query, sessionID, cancelMinutes).Scan(&returnedID)
	if err != nil {
		log.Println("CancelSession: Failed to delete session:", err)
		return err
	}

	return nil
}

type SessionFeedback struct {
	SessionID int     `json:"session_id"`
	Liked     bool    `json:"liked"`
	Rating    float64 `json:"rating,omitempty"`
	Comment   string  `json:"comment,omitempty"`
}

func AddSessionFeedback(sf SessionFeedback) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("AddSessionFeedback: Failed while connecting with the database:", err)
		return err
	}
	defer db.Close()

	query := `INSERT INTO public.session_feedback (session_liked, comment, rating, session_id)
	          VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(query, sf.Liked, sf.Comment, sf.Rating, sf.SessionID)
	if err != nil {
		log.Println("AddSessionFeedback: Failed to insert session feedback:", err)
		return err
	}

	return nil
}

// CheckSessionReschedulingStatus determines if a session with the given ID can be rescheduled.
// Rescheduling is allowed if the current time is at least 60 minutes before the session's start time.
// It returns 'allowed' if rescheduling is permitted, otherwise 'not allowed', along with any error encountered.
func CheckSessionReschedulingStatus(rescheduleTime, sessioID int64) (string, error) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("CheckSessionReschedulingStatus: failed while trying to connect with database with error: ", err)
		return "", err
	}
	defer db.Close()
	// Ensure the database connection is closed when the function returns.
	query := `SELECT                                                                
    		CASE
        	WHEN 
				(NOW() AT TIME ZONE 'UTC') < session_start - INTERVAL '` + fmt.Sprintf("%d", rescheduleTime) + ` minutes'
			THEN 
				'allowed'
        	ELSE 
				'not allowed'
   			END AS rescheduling_status
			FROM 
				session
			WHERE 
				id = $1`
	var reschedulingStatus string
	err = db.QueryRow(query, sessioID).Scan(&reschedulingStatus)
	if err != nil {
		log.Println("CheckSessionReschedulingStatus: Failed while execute the query with error: ", err)
		return "", err
	}
	return reschedulingStatus, nil
}

type RequestDataForReschedule struct {
	SessionID        int
	TutorID          int
	SessionStartTime time.Time
	SessionEndTime   time.Time
	Type             string
	TimeZone         string
	SubjectName      string
}

// ResheduleDemoSession updates the tutor and session time for a given demo session.
//
// This function updates the session details in the database to reschedule a student's demo session.
// It updates the tutor ID and the session start and end times based on the provided parameters.
//
// Parameters:
// - userID (int): The ID of the user (student) who is rescheduling the session.
// - tutorID (int): The ID of the available tutor for the new session time.
// - sessioID (int): The ID of the session to be rescheduled.
// - startSession (time.Time): The new start time of the session.
//
// Returns:
//   - error: An error object if there is any failure during database connection or query execution,
//     otherwise returns nil indicating a successful update.
func ResheduleDemoSession(userID int, sessionReschedulingData RequestDataForReschedule) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("ResheduleDemoSession: failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()
	// Ensure the database connection is closed when the function returns.

	// Define the SQL query to update a record of given session ID into the 'session' table.
	query :=
		`UPDATE 
		    session
		SET
		    tutor_id = $1,
		    session_start = $2::timestamp,
		    session_end = $3::timestamp,
			timezone = $4,
			status = $5,
			updated_at = NOW()
		WHERE 
		    id = $6`
	_, err = db.Exec(query,
		sessionReschedulingData.TutorID,
		sessionReschedulingData.SessionStartTime,
		sessionReschedulingData.SessionEndTime,
		sessionReschedulingData.TimeZone,
		constants.SessionStatusActive,
		sessionReschedulingData.SessionID,
	)
	if err != nil {
		log.Println("ResheduleDemoSession: Failed while execute the query with error: ", err)
		return err
	}
	return nil
}
func SessionSaveAndCredit(userID, tutorID, subjectID, studentID int, sessionDateTime []SessionSlots, timezone string) error {

	return SaveRegularSessionInfoOfStudent(userID, tutorID, subjectID, studentID, sessionDateTime, timezone, true)

}
func SaveSessions(userID, tutorID, subjectID, studentID int, sessionDateTime []SessionSlots, timezone string) error {
	fmt.Println("INTO THE SAVE SESSION FOR GRACE PERIOD")
	return SaveRegularSessionInfoOfStudent(userID, tutorID, subjectID, studentID, sessionDateTime, timezone, false)

}

// session booking using `SaveRegularSessionInfoOfStudent`. If saving any session fails, it logs the error and
// sends a generic error response.
func SaveRegularSessionInfoOfStudent(userID, tutorID, subjectID, studentID int, sessionDateTime []SessionSlots, timezone string, credit bool) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return the error.
		log.Println("SaveRegularSessionInfoOfStudent: failed while trying to connect with database: ", err)
		return err
	}
	defer db.Close()

	var value []interface{}
	var placeHolder []string
	sessionDuration := constants.SessionDuration

	// Dynamically construct placeholders and append values
	for i, sessionSlot := range sessionDateTime {
		sessionStart := sessionSlot.SessionStart
		sessionEnd := sessionStart.Add(time.Duration(sessionDuration) * time.Minute)

		// Create placeholders for each set of values
		placeHolder = append(placeHolder, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, '%s', '%s', NOW(), $%d, $%d,$%d,$%d,NOW())",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, constants.SessionStatusActive, constants.SessionTypePaid, i*9+6, i*9+7, i*9+8, i*9+9))

		// Append the actual values
		value = append(value, userID, tutorID, sessionStart, sessionEnd, sessionDuration, timezone, credit, subjectID, studentID)
	}

	// Construct the final query by joining the placeholders
	query := fmt.Sprintf(`INSERT INTO
								session
								(
									user_id,
									tutor_id,
									session_start,
									session_end,
									duration_mins,
									status,
									type,
									created_at,
									timezone,
									credited,
									subject_id,
									student_id,
									updated_at
								)
								VALUES %s
								RETURNING id`, strings.Join(placeHolder, ", "))

	// Execute the query
	rows, err := db.Query(query, value...)
	log.Println("QUERY-------->>>", query, "<<<<--------")
	if err != nil {
		// Check for duplicate key errors
		if strings.Contains(err.Error(), "duplicate key") {
			log.Println("SaveRegularSessionInfoOfStudent: conflict error: ", err)
			return errors.New("sorry, this time is already taken. Please try a different time as the tutor is unavailable at the selected time")
		}
		// Log other errors
		log.Println("SaveRegularSessionInfoOfStudent: query execution error: ", err)
		return err
	}
	defer rows.Close()

	// Optionally retrieve the inserted IDs
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		fmt.Println("Inserted session ID:", id)
	}

	return nil
}

//		For GetAvailableTutorSlotsForRegularSession each entry  to retrieve available time slots for the specified tutor.
//	   If an error occurs during this process, it logs the error and sends a generic error response.
func GetAvailableTutorSlotsForRegularSession(sessionInfo []SessionDayTime, tutorID, maxSession int, endDate time.Time, startFrom string) ([]time.Time, error) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("GetAvailableTutorSlotsForRegularSession : failed while trying to connect with database with error: ", err)
		return nil, err
	}
	defer db.Close()
	var sessionDatetime []time.Time

	whereClause := []string{}
	args := []interface{}{startFrom, endDate, tutorID}
	if endDate.IsZero() {
		startTime, err := time.Parse("2006-01-02", startFrom)
		if err != nil {
			log.Println("GetAvailableTutorSlotsForRegularSession : Failed to parse start from time string to time.Time  with error :", err)
			return nil, err
		}
		endDate = startTime.AddDate(0, 0, 27)
		args = []interface{}{startFrom, endDate, tutorID}
	}

	argsIndex := 4
	for _, slot := range sessionInfo {
		whereClause = append(whereClause, fmt.Sprintf("day_of_week = $%d AND to_char(start_date, 'HH24:MI:SS') = $%d", argsIndex, argsIndex+1))
		args = append(args, slot.Day, slot.Time)
		argsIndex += 2
	}

	query := fmt.Sprintf(`
	WITH cte_slots AS (
			SELECT 
   				start_date,
				end_date
			FROM 
    			tutor_available_slots as ts
			WHERE
				start_date >= $1 AND
	    		end_date <= $2
	    		AND tutor_id = $3 
	    		AND (%s)
			),
		cte_available_slots AS (
			SELECT 
				*
    		FROM 
				cte_slots AS tas
   			 WHERE 
       			NOT EXISTS (
           				SELECT 1
            			FROM 	
							tutor_unavailable_times AS tut
           				WHERE 
                			tut.tutor_id = $3
                			AND (tas.start_date < tut.end_date AND tas.end_date > tut.start_date)
       				) 
					AND		
       				NOT EXISTS (
            			SELECT 1
            			FROM 
							session s
            			WHERE 
               				s.tutor_id = $3
                			AND s.status IN ('active', 'cancelled')
                			AND (tas.start_date < s.session_end AND tas.end_date > s.session_start)
        			)
			)
				SELECT 
					start_date
				FROM 
					cte_available_slots
				ORDER BY 
					start_date ASC
				LIMIT 
					$%d`,
		strings.Join(whereClause, " OR "), argsIndex)

	args = append(args, maxSession)
	// Execute the query
	// this is a sample query
	// 	WITH cte_slots AS (
	// 		SELECT
	// 			start_date,
	// 			end_date
	// 		FROM
	// 			tutor_available_slots as ts
	// 		WHERE
	// 				start_date > NOW() AND
	// 				end_date < $1
	// 				AND tutor_id = $2
	// 				AND (day_of_week = $3 AND to_char(start_date, 'HH24:MI:SS') = $4)),
	// 		cte_available_slots AS (
	// 				  SELECT *
	//					FROM
	//						cte_slots AS tas
	// 					WHERE
	// 				NOT EXISTS (
	// 				SELECT 1
	// 				FROM
	//					tutor_unavailable_times AS tut
	// 				WHERE
	// 					tut.tutor_id = $2
	// 					AND (tas.start_date < tut.end_date AND tas.end_date > tut.start_date)
	// )
	// 				AND
	// 				NOT EXISTS (
	// 					SELECT 1
	// 					FROM
	//						session s
	// 					WHERE
	// 						s.tutor_id = $2
	// 						AND s.status IN ('active', 'cancelled')
	// 						AND (tas.start_date < s.session_end AND tas.end_date > s.session_start)
	// )
	// 				)
	// 				SELECT start_date
	// 		FROM cte_available_slots
	// 		ORDER BY
	// 				start_date ASC
	// 		LIMIT $5;
	log.Println(query, args)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("GetAvailableTutorSlotsForRegularSession : Failed while trying execute the query with error :", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var startDate time.Time
		if err := rows.Scan(&startDate); err != nil {
			log.Println("GetAvailableTutorSlotsForRegularSession: Failed to scan rows from dataase with error : ", err)
			return nil, err
		}
		sessionDatetime = append(sessionDatetime, startDate)
	}
	if len(sessionDatetime) != maxSession {
		log.Println("GetAvailableTutorSlotsForRegularSession: No available slots found for the given conditions")
		return nil, fmt.Errorf("no available slots found for the given conditions")
	}
	return sessionDatetime, nil
}

// GetPreviousSessionDetails retrieves the start time, timezone, type, and subject name of a session
// based on the given sessionID. It fetches the session data by joining the session and subject tables.
//
// Parameters:
// - sessionID (int): The unique identifier for the session to retrieve.
//
// Returns:
// - Session: A struct containing the session start time, timezone, type, and subject name.
// - error: Returns an error if the database connection fails or if the query encounters an issue.
//
// Errors:
// - Logs an error if there's an issue connecting to the database or executing the query.
// - Returns the session struct with the retrieved data, or an error if any step fails.
func GetPreviousSessionDetails(sessionID int) (Session, error) {
	var session Session
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("[ERROR] GetPreviousSessionDetails : failed while trying to connect with database with error: ", err)
		return session, err
	}
	defer db.Close()
	query :=
		`SELECT 
					s.session_start,
					s.timezone,
					s.type,
					sb.name
			FROM
				session AS s
			JOIN subject AS sb ON sb.id = s.subject_id
			WHERE
				s.id = $1`
	err = db.QueryRow(query, sessionID).Scan(
		&session.StartSession,
		&session.TimeZone,
		&session.SessionType,
		&session.SubjectName,
	)
	if err != nil {
		log.Println("[ERROR] GetPreviousSessionDetails : Failed while trying execute the query with error :", err)
		return session, err
	}
	return session, nil
}

// CheckStudentAllowedToBookDemo checks if a student is allowed to book a demo session by querying the database
// for any existing active demo sessions for the given user ID, if there is an active demo session,
// and returns a boolean indicating whether demo booking is allowed.
// If an active demo session is found, demoBookingAllowed is set to false, and an error is returned.
// If no active demo session exists, demoBookingAllowed remains true. In case of database connection errors or
// query execution errors, the error is logged and returned.
func CheckStudentAllowedToBookDemo(userID int) (bool, error) {
	demoBookingAllowed := true
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("CheckStudentAllowedToBookDemo : failed while trying to connect with database with error: ", err)
		return demoBookingAllowed, err
	}
	defer db.Close()
	var isAllowed bool
	query := `
		SELECT EXISTS(
			SELECT 
				session_start
			FROM
				session
			WHERE 
				user_id = $1 AND
				type = 'demo' AND
				status  = 'active'
			ORDER BY 
				created_at DESC
			LIMIT 1
		)`
	err = db.QueryRow(query, userID).Scan(&isAllowed)
	if err == sql.ErrNoRows {
		return demoBookingAllowed, nil
	}
	if err != nil {
		log.Println("CheckStudentAllowedToBookDemo : Failed while trying execute the query with error :", err)
		return demoBookingAllowed, err
	}

	if isAllowed {
		log.Println("CheckStudentAllowedToBookDemo : Failed,You have  already an active demo session.")
		demoBookingAllowed = false
		return demoBookingAllowed, errors.New("You have  already an active demo session.")
	}

	return demoBookingAllowed, err
}

func GetUsersWithTodaySession(date string) ([]int, error) {
	// Connect to the database
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return an empty slice with the error.
		log.Println("GetUsersWithTodaySession: failed while trying to connect with database with error:", err)
		return nil, err
	}
	defer db.Close()

	// Define the query to select user IDs with sessions scheduled for today
	query := `SELECT
                    s.user_id
                FROM
                    public.session AS s
                JOIN
                    public.user2subscription AS u2s ON s.user_id = u2s.user_id
                WHERE
                    u2s.status = 'active' AND s.type = 'paid'
                GROUP BY
                    s.user_id
                HAVING
                    DATE(MIN(s.session_start)) = `
	// Append the date condition dynamically
	if len(date) == 0 || date == "" {
		// If no date is provided, use the current date
		query += `DATE(NOW())`
	} else {
		// If a date is provided, use the provided date
		query += `'` + date + `'`
	}
	fmt.Println("GET USERS LIST Who's first class is scheduled today", query)
	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetUsersWithTodaySession: failed to execute query with error:", err)
		return nil, err
	}
	defer rows.Close()

	// Initialize a slice to hold the user IDs
	var userIDs []int

	// Iterate over the rows and scan the user IDs into the slice
	for rows.Next() {
		var userID sql.NullInt64
		if err := rows.Scan(&userID); err != nil {
			log.Println("GetUsersWithTodaySession: failed to scan row with error:", err)
			return nil, err
		}
		userIDs = append(userIDs, int(userID.Int64))
	}

	// Check for any error that might have occurred during iteration
	if err := rows.Err(); err != nil {
		log.Println("GetUsersWithTodaySession: row iteration error:", err)
		return nil, err
	}

	// Return the slice of user IDs
	return userIDs, nil
}

func UpdateSubscriptionStartDateForUser(userID int) (time.Time, error) {
	// Connect to the database
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return an empty slice with the error.
		log.Println("UpdateSubscriptionStartDateForUser: failed while trying to connect with database with error:", err)
		return time.Time{}, nil
	}
	defer db.Close()

	var session_endDate sql.NullTime

	// Define the query to select user IDs with sessions scheduled for today
	query := `WITH cte AS (
         		SELECT 
         			s.user_id, s.id, s.session_start, s.session_end
         		FROM 
         			public.session AS s
         		JOIN 
         			public.user2subscription AS u2s ON s.user_id = u2s.user_id
         		WHERE 
         			u2s.valid_until IS NULL AND s.type = 'paid' AND s.user_id = $1 
         		ORDER BY 
         			s.session_start ASC 
         		LIMIT 1
         	)
         	UPDATE 
         		public.user2subscription 
         	SET 
         		start_date = cte.session_start,
         		valid_until = cte.session_start + INTERVAL '27 days'
         	FROM 
         		cte
         	WHERE 
         		public.user2subscription.user_id = cte.user_id 
			RETURNING cte.session_end`

	// Execute the query
	err = db.QueryRow(query, userID).Scan(&session_endDate)
	if err != nil {
		log.Println("UpdateSubscriptionStartDateForUser: failed to execute query with error:", err)
		return time.Time{}, nil
	}
	return session_endDate.Time, nil

}

func CreditSessionsForUser(userID int, startDate, validUntil string) (int64, error) {
	// Get a connection to the database
	db, err := config.GetDB2()
	if err != nil {
		log.Println("CreditSessionsForUser: failed while trying to connect with database:", err)
		return 0, err
	}
	defer db.Close()

	// Update query without RETURNING
	query := `UPDATE public.session 
	          SET credited = true
	          WHERE user_id = $1 AND session_start BETWEEN $2 AND $3`

	// Execute the update query
	result, err := db.Exec(query, userID, startDate, validUntil)
	if err != nil {
		log.Println("CreditSessionsForUser: error executing update query:", err)
		return 0, err
	}

	// Get the number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rowsAffected = 0
	}

	// Return the number of affected rows
	return rowsAffected, nil
}

func GetStudentSessionPreference(userID int) ([]SessionDayTime, error) {
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return an error
		log.Println("GetStudentSessionPreference: failed while trying to connect with database:", err)
		return nil, err
	}
	defer db.Close()

	var sessionPreferences []SessionDayTime

	query := `SELECT tutor_id, 
	                 day_of_week, 
					 start_time,
					 timezone,
					 subject_id,
					 student_id
			  FROM  public.student_session_preference 
			  WHERE user_id = $1`
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Println("GetStudentSessionPreference: query execution error:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionDayTime SessionDayTime
		var (
			dayOfWeek, timeZone           sql.NullString
			startTime                     sql.NullTime
			tutorID, subjectID, studentID sql.NullInt64
		)

		err = rows.Scan(&tutorID, &dayOfWeek, &startTime, &timeZone, &subjectID, &studentID)
		if err != nil {
			log.Println("GetStudentSessionPreference: error scanning row:", err)
			return nil, err
		}

		sTime := startTime.Time
		startTimeString := sTime.Format(utility.TimeLayout)

		sessionDayTime.TutorID = int(tutorID.Int64)
		sessionDayTime.Day = dayOfWeek.String
		sessionDayTime.Time = startTimeString
		sessionDayTime.TimeZone = timeZone.String
		sessionDayTime.StudentID = int(studentID.Int64)
		sessionDayTime.SubjectID = int(subjectID.Int64)

		sessionPreferences = append(sessionPreferences, sessionDayTime)
	}

	if err = rows.Err(); err != nil {
		log.Println("GetStudentSessionPreference: rows iteration error:", err)
		return nil, err
	}

	return sessionPreferences, nil
}

func InsertStudentPreferences(preferences SessionDayTime) error {
	db, err := config.GetDB2()
	if err != nil {
		// Log the error if there's a problem connecting to the database
		log.Println("InsertStudentPreferences: failed to connect to the database:", err)
		return err
	}
	defer db.Close()

	// Prepare the query for inserting session preferences
	query := `INSERT INTO public.student_session_preference 
	          (user_id, tutor_id, day_of_week, start_time, timezone,student_id,subject_id, start_date) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	// Insert each session preference using db.Exec

	_, err = db.Exec(query, preferences.UserID, preferences.TutorID, preferences.Day, preferences.Time, preferences.TimeZone, preferences.StudentID, preferences.SubjectID, preferences.StartDate)
	if err != nil {
		log.Println("InsertStudentPreferences: failed to execute insert query for user:", preferences.UserID, "error:", err)
		return err
	}

	return nil
}

// SaveExceptionSessionOfStudent saves a session to the session_exceptions table in the database.
// It captures information about a student's session exception, including session start and end times,
// tutor and user IDs, session duration, status, session type, and timezone. The session duration is set
// to 40 minutes by default, and the current timestamp is used for the record's creation time.
//
// Parameters:
// - session (Session): The session object containing information about the session exception to be saved.
//
// Returns:
//   - error: Returns nil if the session exception is saved successfully. If an error occurs, the function logs
//     the error and returns the error object.
func SaveExceptionSessionOfStudent(session Session) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("SaveExceptionSessionOfStudent: failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()
	query :=
		`INSERT INTO
						session_exceptions
						(
							user_id,
							tutor_id,
							session_start,
							session_end,
							duration_mins,
							status,
							type,
							timezone,
							subject_id,
							student_id,
							created_at
						)
						VALUES
							($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())`

	_, err = db.Exec(query, session.Student.User.ID, session.Tutor.ID, session.StartSession, session.EndSession,
		session.SessionDuration, session.Status, session.SessionType, session.TimeZone, session.SubjectID, session.StudentID)
	if err != nil {
		log.Println("SaveExceptionSessionOfStudent: failed while trying to execute the query with database with error: ", err)
		return err
	}

	return nil
}

// DeleteSession removes a session from the database based on the provided sessionID.
//
// Parameters:
// - sessionID (int): The ID of the session to be deleted.
//
// Returns:
//   - error: Returns nil if the session is deleted successfully. If an error occurs,
//     the function logs the error and returns the error object.
func DeleteSession(sessionID int) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("[ERROR] DeleteSession: failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()
	query := `
			DELETE FROM 
				session 
			where id = $1`
	_, err = db.Exec(query, sessionID)
	if err != nil {
		log.Println("[ERROR] DeleteSession : Failed to delete session from database with error : ", err)
		return err
	}
	return nil
}

// GetExpectedSession retrieves a list of sessions for a specific user from the database.
// It connects to the database, executes a query to fetch session details for the given user ID,
// including session status and start time, and returns these details as a slice of `Session` structs.
//
// Parameters:
// - userID (int): The ID of the user for whom to fetch the sessions.
//
// Returns:
// - ([]Session, error): A slice of `Session` structs containing the session details, or an empty slice and an error if something goes wrong.
//
// Errors:
// - If there is an issue connecting to the database, executing the query, or scanning the results, an error is logged and returned.
func GetExpectedSession(userID int) ([]Session, error) {
	var SessionList []Session
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("[ERROR] GetExpectedSession: Failed while trying to connect with database with error: ", err)
		return SessionList, err
	}
	defer db.Close()
	query := `
		SELECT
			s.id,
		    s.status,
		    s.session_start,
			s.timezone
		FROM
		    public.user AS u
		JOIN 
		    session_exceptions AS s ON s.user_id = u.id
		WHERE 
		    u.id = $1
		ORDER BY
		    s.session_start DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Println("[ERROR] GetExpectedSession: Failed while trying to execute the query with database with error: ", err)
		return SessionList, err
	}
	for rows.Next() {
		var (
			sessionID        sql.NullInt64
			status, timezone sql.NullString
			startSession     sql.NullTime
		)
		err := rows.Scan(&sessionID, &status, &startSession, &timezone)
		if err != nil {
			log.Println("[ERROR] GetExpectedSession: Failed while trying to scannig the rows with database with error: ", err)
			return SessionList, err
		}
		SessionList = append(SessionList, Session{
			SessionID:    int(sessionID.Int64),
			Status:       status.String,
			StartSession: startSession.Time,
			TimeZone:     timezone.String,
		})
	}
	return SessionList, nil
}
func VarifyRequestedSessionDateTime(sessionStart []SessionSlots, tutorID int) ([]string, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] VarifyRequestedSessionDateTime: failed to connect to the database:", err)
		return nil, err
	}
	defer db.Close()

	// Prepare WHERE clause and arguments
	argsIndex := 1
	whereClause := []string{}
	args := []interface{}{}

	// Construct the WHERE clause dynamically for each session slot
	for _, slot := range sessionStart {
		whereClause = append(whereClause, fmt.Sprintf("SELECT to_timestamp($%d, 'YYYY-MM-DD\"T\"HH24:MI:SS') AS start_date, to_timestamp($%d, 'YYYY-MM-DD\"T\"HH24:MI:SS') AS end_date", argsIndex, argsIndex+1))
		args = append(args, slot.SessionStart, slot.SessionEnd)
		argsIndex += 2
	}
	// Define your date ranges
	query := fmt.Sprintf(`
		WITH date_ranges AS (%s)
		SELECT start_date, end_date
		FROM date_ranges
		EXCEPT
		SELECT start_date, end_date
		FROM tutor_available_slots AS tas
		WHERE tutor_id = $%d
		AND NOT EXISTS (
            			SELECT 1
            			FROM 
							session s
            			WHERE 
               				s.tutor_id = $%d
                			AND s.status IN ('active', 'cancelled')
                			AND (tas.start_date < s.session_end AND tas.end_date > s.session_start)
        			)ORDER BY start_date ASC`, strings.Join(whereClause, " UNION ALL "), argsIndex, argsIndex)

	// Append the tutorID to the arguments list

	// Execute the query
	args = append(args, tutorID)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("[ERROR] VarifyRequestedSessionDateTime: query execution failed:", err)
		return nil, err
	}
	defer rows.Close()

	// Collect missing date ranges
	var missingRanges []string
	for rows.Next() {
		var startDate, endDate string
		if err := rows.Scan(&startDate, &endDate); err != nil {
			log.Println("[ERROR] VarifyRequestedSessionDateTime: row scan failed:", err)
			return nil, err
		}
		missingRanges = append(missingRanges, fmt.Sprintf("Missing: %s to %s", startDate, endDate))
	}
	// Check for any errors encountered while iterating over rows
	if err := rows.Err(); err != nil {
		log.Println("[ERROR] VarifyRequestedSessionDateTime: error during rows iteration:", err)
		return nil, err
	}

	return missingRanges, nil
}
func DeletePendingSession(userID int) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] DeletePendingSession: failed to connect to the database:", err)
		return err
	}
	defer db.Close()
	query := `
			 WITH targeted_rows AS(
			 	SELECT
					id
				FROM
					session_exceptions
				WHERE
					user_id = $1
				AND status IN ('absent','active','cancelled')
				ORDER BY
					CASE status
					WHEN 'absent' THEN 1
					WHEN 'active' THEN 2
					WHEN 'cancelled' THEN 3
					END
				LIMIT 1
			)
			DELETE 
			FROM
				session_exceptions
			WHERE id = (SELECT id FROM targeted_rows)
			RETURNING id`
	var deletedSessionID int
	err = db.QueryRow(query, userID).Scan(&deletedSessionID)
	if err != nil {
		log.Println("[ERROR] DeletePendingSession : Failed to delete pending session with error :", err)
		return err
	} else if deletedSessionID == 0 {
		log.Println("[ERROR] DeletePendingSession : Failed to delete pending session from database with error : No rows deleted")
		return errors.New("no pending session available")
	}
	return nil
}
func DeleteCancelledSession(userID int) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("[ERROR] DeleteCancelledSession: failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()
	query := `
				DELETE FROM 
					session_exceptions 
				WHERE id IN(
					SELECT 
						id
					FROM
						session_exceptions
					WHERE
						user_id = $1 AND
						status = 'cancelled'
					ORDER BY 
						created_at
					LIMIT 1
				)
					RETURNING id`
	var deletedSessionID int
	err = db.QueryRow(query, userID).Scan(&deletedSessionID)
	if deletedSessionID == 0 {
		log.Println("[ERROR] DeleteCancelledSession : Failed to delete pending session from database with error : No rows deleted")
		return errors.New("subscription plan expire, only book session which cancelled by tutor")
	} else if err != nil {
		log.Println("[ERROR] DeleteCancelledSession : Failed to delete pending session from database with error : ", err)
		return err
	}
	return nil
}

func GetNextScheduledSession(userID int64) (Session, error) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return nil object along with the error.
		log.Println("[ERROR] GetNextScheduledSession: failed while trying to connect with database with error: ", err)
		return Session{}, err
	}
	defer db.Close()

	query := ` SELECT 
			    s.id,
			    s.user_id,
			    s.session_start,
				s.session_end,
			    s.status,
			    s.timezone,
				s.tutor_id,
				s.duration_mins,
				s.type,
				st.name,
				st.parent_email,
				tut.email_id,
				tut.name,
				tut.timezone AS tutor_timezone
				
			FROM
			    session AS s,
				student as st,
				tutor AS tut
			WHERE 
			    s.user_id = $1 
				AND st.user_id = s.user_id
				AND s.tutor_id = tut.id
			    AND s.session_start > NOW() AT TIME ZONE 'UTC'
				AND s.session_start - (NOW() AT TIME ZONE 'UTC') < INTERVAL '24 HOURS'
				AND s.status = 'active'
			ORDER BY 
			    s.session_start ASC
			LIMIT 1
			`
	var (
		sessionID                 int64
		userId                    int64
		sessionStart              time.Time
		sessionEnd                time.Time
		status                    string
		timezone, tutorTimezone   sql.NullString
		tutorID                   int64
		durationMinutes           int64
		sessionType               string
		studentName, studentEmail sql.NullString
		tutorName, tutorEmail     sql.NullString
	)
	err = db.QueryRow(query, userID).Scan(
		&sessionID,
		&userId,
		&sessionStart,
		&sessionEnd,
		&status,
		&timezone,
		&tutorID,
		&durationMinutes,
		&sessionType,
		&studentName,
		&studentEmail,
		&tutorEmail,
		&tutorName,
		&tutorTimezone,
	)

	if err != nil {
		// If there is an error connecting to the database, log the error and return nil object along with the error.
		log.Println("[ERROR] GetNextScheduledSession: failed while executing the query with error: ", err)
		return Session{}, err
	}

	session := Session{
		SessionID: int(sessionID),
		Student: students.Student{
			User: models.User{
				ID: userID,
			},
			StudentName: utility.SQLNullStringToString(studentName),
			ParentEmail: studentEmail.String,
		},
		StartSession: sessionStart,
		EndSession:   sessionEnd,
		Status:       status,
		TimeZone:     timezone.String,
		Tutor: tutor.Tutor{
			ID:         int(tutorID),
			TutorEmail: tutorEmail.String,
			TutorName:  tutorName.String,
			Timezone:   tutorTimezone.String,
		},
		SessionDuration: int(durationMinutes),
		SessionType:     sessionType,
	}
	return session, nil
}

func MarkSessionConfirmationStatus(sessionID int64, status bool) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return error.
		log.Println("[ERROR] MarkSessionConfirmationStatus: failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()

	query := ` UPDATE session SET session_confirmed = $1 WHERE id = $2 `
	_, err = db.Exec(query, status, sessionID)
	if err != nil {
		// If there is an error connecting to the database, log the error and return error.
		log.Println("[ERROR] MarkSessionConfirmationStatus: failed while executing the query with error: ", err)
		return err
	}
	return nil
}

func GetSessionsToCancelIfNotConfirmed(dateTime string) (map[int64]Session, error) {

	sessionCancelMessageInfo := make(map[int64]Session)

	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return error.
		log.Println("[ERROR] GetSessionsToCancelIfNotConfirmed: failed while trying to connect with database with error: ", err)
		return map[int64]Session{}, err
	}
	defer db.Close()

	// Get time duration in MINUTES until when confirmation is allowed.
	timeDuration := utility.GetTimeDurationUntilConfirmationAllowed()
	if len(timeDuration) == 0 {
		timeDuration = "210"
	}

	if len(dateTime) == 0 {
		dateTime = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	query := `SELECT 
				s.id,
				u.id,
				u.dialing_code,
				u.phone_number,
				s.type,
				s.status,
				t.id,
				s.session_start,
				s.timezone,
				st.name
			FROM 
				session as s,
				tutor as t,
				public.user as u,
				student as st
			WHERE
				u.id = s.user_id AND
				s.tutor_id = t.id AND
				st.user_id = u.id AND
				s.session_start - ('` + dateTime + `') < INTERVAL '` + timeDuration + ` minutes' AND
				s.status = 'active' AND
				session_confirmed = false 
			`

	log.Println("GetSessionsToCancelIfNotConfirmed query: ", query)
	rows, err := db.Query(query)
	if err != nil {
		log.Println("[ERROR] GetSessionsToCancelIfNotConfirmed: failed while executing the query with error: ", err)
		return map[int64]Session{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID int64
		var userID int64
		var dialingCode string
		var phoneNumber string
		var sessionType string
		var status string
		var tutorID int64
		var sessionStart time.Time
		var timezone string
		var studentName sql.NullString

		err = rows.Scan(&sessionID, &userID, &dialingCode, &phoneNumber, &sessionType, &status, &tutorID, &sessionStart, &timezone, &studentName)
		if err != nil {
			log.Println("[ERROR] GetSessionsToCancelIfNotConfirmed: failed while scanning the query with error: ", err)
			continue
		}
		sessionCancelMessageInfo[sessionID] = Session{
			SessionID: int(sessionID),
			Student: students.Student{
				User: models.User{
					ID:          userID,
					Phone:       phoneNumber,
					DialingCode: dialingCode,
				},
				StudentName: studentName.String,
			},
			SessionType: sessionType,
			Status:      status,
			Tutor: tutor.Tutor{
				ID: int(tutorID),
			},
			StartSession: sessionStart,
			TimeZone:     timezone,
		}

	}
	return sessionCancelMessageInfo, nil
}

func UpdateSessionStatusIfNotConfirmed(sessionIDs, status string) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return error.
		log.Println("[ERROR] UpdateSessionStatusIfNotConfirmed: failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()

	query := `UPDATE session SET status = $1, updated_at = NOW() AT TIME ZONE 'UTC' WHERE id IN (` + sessionIDs + `)`
	log.Println("UpdateSessionStatusIfNotConfirmed: query", query)

	_, err = db.Exec(query, status)
	if err != nil {
		// If there is an error connecting to the database, log the error and return error.
		log.Println("[ERROR] UpdateSessionStatusIfNotConfirmed: failed while executing the query with error: ", err)
		return err
	}
	return nil
}

// StudentSessionPreference represents the schema of the student_session_preference table.
type StudentSessionPreference struct {
	ID        int
	UserID    int64
	TutorID   int64
	DayOfWeek string
	StartTime string
	StartDate string
	Timezone  string
	StudentID int64
	SubjectID int64
}

// GetSessionPreferenceById retrieves a session preference by its ID.
func GetSessionPreferenceById(sessionPreferenceID, userID int64) ([]StudentSessionPreference, error) {
	var preference []StudentSessionPreference
	// Establish database connection
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetSessionPreferenceById: failed to connect to the database:", err)
		return preference, err
	}
	defer db.Close()
	whereClause := ""
	if userID != 0 {
		whereClause = fmt.Sprintf("user_id = %d", userID)
	}
	if sessionPreferenceID != 0 {
		whereClause = fmt.Sprintf("id = %d", sessionPreferenceID)
	}
	// Query to select a session preference by ID
	query := fmt.Sprintf(`SELECT 
		    id, 
		    user_id, 
		    tutor_id, 
		    day_of_week, 
		    start_time,
			start_date, 
		    timezone, 
		    student_id, 
		    subject_id 
		FROM student_session_preference 
		WHERE %s`, whereClause)

	// Execute query and scan the results into nullable variables
	rows, err := db.Query(query)
	if err != nil {
		log.Println("GetSessionPreferenceById: Failed to execute query with :", err)
		return preference, err
	}
	for rows.Next() {
		// Variables for nullable fields
		var (
			preferenceID, userId, tutorID, studentID, subjectID sql.NullInt64
			dayOfWeek, timezone, startTime, startDate           sql.NullString
		)
		err = rows.Scan(
			&preferenceID, &userId, &tutorID, &dayOfWeek, &startTime, &startDate, &timezone,
			&studentID, &subjectID,
		)
		if err != nil {
			log.Println("[ERROR] GetSessionPreferenceById : Failed while scaninng the rows from database with error :", err)
			return preference, err
		}
		preference = append(preference, StudentSessionPreference{
			// Assign nullable values to the struFct
			ID:        int(preferenceID.Int64),
			UserID:    userId.Int64,
			TutorID:   tutorID.Int64,
			DayOfWeek: dayOfWeek.String,
			StartDate: startDate.String,
			Timezone:  timezone.String,
			StartTime: startTime.String,
			StudentID: studentID.Int64,
			SubjectID: subjectID.Int64,
		})
	}

	return preference, nil
}

// DeleteSessionPreferenceById deletes a session preference by its ID and returns the deleted ID.
func DeleteSessionPreferenceById(id int64) (int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("DeleteSessionPreferenceById: failed to connect to the database:", err)
		return 0, err
	}
	defer db.Close()

	query := `DELETE FROM student_session_preference WHERE id = $1 RETURNING id`
	var deletedID int
	err = db.QueryRow(query, id).Scan(&deletedID)
	if err != nil {
		log.Println("DeleteSessionPreferenceById: failed to delete preference:", err)
		return 0, err
	}

	return deletedID, nil
}

// GetAllSessionPreferences retrieves all session preferences.
func GetAllSessionPreferences(uid int) ([]StudentSessionPreference, error) {
	var preferences []StudentSessionPreference

	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetAllSessionPreferences: failed to connect to the database:", err)
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT 
		    id, 
		    user_id, 
		    tutor_id, 
		    day_of_week, 
		    start_time,
			start_date, 
		    timezone,  
		    student_id, 
		    subject_id 
		FROM student_session_preference WHERE user_id = $1`
	rows, err := db.Query(query, uid)
	if err != nil {
		log.Println("GetAllSessionPreferences: failed to execute query:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var preference StudentSessionPreference
		var (
			userID, tutorID, studentID, subjectID, preferenceID sql.NullInt64
			dayOfWeek, timezone, startDate                      sql.NullString
			startTime                                           sql.NullString
		)

		err := rows.Scan(
			&preferenceID, &userID, &tutorID, &dayOfWeek, &startTime, &startDate, &timezone,
			&studentID, &subjectID,
		)
		if err != nil {
			log.Println("GetAllSessionPreferences: failed to scan row:", err)
			return nil, err
		}
		preference.ID = int(preferenceID.Int64)
		preference.UserID = userID.Int64
		preference.TutorID = tutorID.Int64
		preference.DayOfWeek = dayOfWeek.String
		preference.StartTime = startTime.String
		preference.StartDate = startDate.String
		preference.Timezone = timezone.String
		preference.StudentID = studentID.Int64
		preference.SubjectID = subjectID.Int64

		preferences = append(preferences, preference)
	}

	if err := rows.Err(); err != nil {
		log.Println("GetAllSessionPreferences: row iteration error:", err)
		return nil, err
	}

	return preferences, nil
}

// UpdateSessionPreference updates an existing session preference by ID and returns the updated ID.
func UpdateSessionPreference(preference StudentSessionPreference) (int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("UpdateSessionPreference: failed to connect to the database:", err)
		return 0, err
	}
	defer db.Close()

	query := `
		UPDATE student_session_preference
		SET subject_id = $1, tutor_id = $2, day_of_week = $3, start_time = $4, start_date = $5, 
			timezone = $6, updated_at = NOW(), student_id = $7
		WHERE id = $8 RETURNING id`
	var (
		updatedID   int
		studentId   *int64
		start_time  *string
		startDate   *string
		day_of_week *string

		tutorID *int64
	)
	if preference.StudentID != 0 {
		studentId = &preference.StudentID
	} else {
		studentId = nil
	}

	if len(preference.StartTime) != 0 {
		start_time = &preference.StartTime
	} else {
		start_time = nil
	}

	if preference.TutorID != 0 {
		tutorID = &preference.TutorID
	} else {
		tutorID = nil
	}

	if len(preference.StartDate) != 0 {
		startDate = &preference.StartDate
	} else {
		startDate = nil
	}

	if len(preference.DayOfWeek) != 0 {
		day_of_week = &preference.DayOfWeek
	} else {
		day_of_week = nil
	}

	err = db.QueryRow(
		query,
		preference.SubjectID, tutorID, day_of_week,
		start_time, startDate, preference.Timezone, studentId, preference.ID,
	).Scan(&updatedID)
	if err != nil {
		log.Println("UpdateSessionPreference: failed to update preference:", err)
		return 0, err
	}

	return updatedID, nil
}

// AddSessionPreference adds new session preferences and returns the inserted IDs.
func AddSessionPreference(preferences []StudentSessionPreference) ([]int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("AddSessionPreference: failed to connect to the database:", err)
		return nil, err
	}
	defer db.Close()

	var insertedIDs []int

	// Prepare the insert query
	query := `
		INSERT INTO student_session_preference 
		   (user_id, tutor_id, day_of_week, start_time, start_date, timezone, created_at, student_id, subject_id)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7, $8)
		RETURNING id`

	// Iterate over the preferences array and execute the insert query for each preference
	for _, preference := range preferences {
		var (
			insertedID  int
			studentID   *int64
			start_time  *string
			day_of_week *string
			startDate   *string
			tutorID     *int64
		)
		if preference.StudentID != 0 {
			studentID = &preference.StudentID // Use a pointer to the non-zero value
		} else {
			studentID = nil
		}

		if len(preference.StartTime) != 0 {
			start_time = &preference.StartTime
		} else {
			start_time = nil
		}

		if preference.TutorID != 0 {
			tutorID = &preference.TutorID
		} else {
			tutorID = nil
		}

		if len(preference.StartDate) != 0 {
			startDate = &preference.StartDate
		} else {
			startDate = nil
		}

		if len(preference.DayOfWeek) != 0 {
			day_of_week = &preference.DayOfWeek
		} else {
			day_of_week = nil
		}

		err := db.QueryRow(
			query,
			preference.UserID, tutorID, day_of_week,
			start_time, startDate, preference.Timezone, studentID, preference.SubjectID,
		).Scan(&insertedID)

		if err != nil {
			log.Printf("AddSessionPreference: failed to insert preference with UserID %d: %v", preference.UserID, err)
			return nil, err
		}

		// Append the inserted ID to the slice
		insertedIDs = append(insertedIDs, insertedID)
	}

	return insertedIDs, nil
}

// function CountRemainingPendingSessionOfStudent is responsible to fetch the no. of remaining session
// in pending session bucket.
func CountRemainingPendingSessionOfStudent(userID int) (int64, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("[ERROR] CountRemainingPendingSessionOfStudent: failed to connect to the database:", err)
		return 0, err
	}
	defer db.Close()
	var remainingPendingSession sql.NullInt64
	query := `SELECT COUNT(*) AS remaining_rows
			FROM session_exceptions
			WHERE user_id  = $1;`
	err = db.QueryRow(query, userID).Scan(&remainingPendingSession)
	if err != nil {
		log.Printf("[ERROR] CountRemainingPendingSessionOfStudent: failed to fetch count of remaining pending session with error: %v", err)
		return remainingPendingSession.Int64, err
	}
	return remainingPendingSession.Int64, nil
}
func AddOrUpdateSessionPreference(preferences []StudentSessionPreference) ([]int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("AddOrUpdateSessionPreference: failed to connect to the database:", err)
		return nil, err
	}
	defer db.Close()

	deleteQuery := `DELETE FROM student_session_preference WHERE user_id=$1 and student_id=$2 and subject_id=$3`

	var insertedOrUpdatedIDs []int

	// Prepare the upsert query
	query := `
		INSERT INTO student_session_preference 
		   (user_id, tutor_id, day_of_week, start_time, start_date, timezone, created_at, student_id, subject_id)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7, $8)
		RETURNING id`

	//delete previous preference
	preference := preferences[0]
	_, _ = db.Exec(deleteQuery, preference.UserID, preference.StudentID, preference.SubjectID)

	// Iterate over the preferences array and execute the upsert query for each preference
	for _, preference := range preferences {

		var upsertedID int
		err := db.QueryRow(
			query,
			preference.UserID, preference.TutorID, preference.DayOfWeek,
			preference.StartTime, preference.StartDate, preference.Timezone, preference.StudentID, preference.SubjectID,
		).Scan(&upsertedID)

		if err != nil {
			log.Printf("AddOrUpdateSessionPreference: failed to upsert preference with UserID %d: %v", preference.UserID, err)
			return nil, err
		}

		// Append the upserted ID to the slice
		insertedOrUpdatedIDs = append(insertedOrUpdatedIDs, upsertedID)
	}

	return insertedOrUpdatedIDs, nil
}
