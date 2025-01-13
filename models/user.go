package models

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"
	"tutree/student-apis/config"
	"tutree/student-apis/services"
)

// User roles constant
const (
	StudentRole    = "student"
	AmbassadorRole = "ambassador"
)

// User represents a user(student/ambassador) record in the database.
type User struct {
	ID              int64             `json:"id,omitempty"`
	Email           string            `json:"email,omitempty"`
	Phone           string            `json:"phone,omitempty"`
	DialingCode     string            `json:"dialing_code,omitempty"`
	OTP             string            `json:"otp,omitempty"`
	OTPValidUntil   time.Time         `json:"otp_valid_until,omitempty"`
	UserIP          string            `json:"user_ip,omitempty"`
	Location        services.Location `json:"location,omitempty"`
	IsPhoneVerified bool              `json:"IsPhoneVerified,omitempty"`
	IsEmailVerified bool              `json:"is_email_verified,omitempty"`
	ReferralCode    string            `json:"referral_code,omitempty"`
	CreatedAt       time.Time         `json:"created_at,omitempty"`
	Timezone        string            `json:"timezone,omitempty"`
	QueryParams     string            `json:"query_params,omitempty"`
}

func GetUserByReferralCode(referralCode string) (int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetUserByReferralCode: Failed while connecting with the database :", err)
		return 0, err
	}
	defer db.Close()
	query := `select id from public.user where referral_code = $1`
	var userId sql.NullInt64
	err = db.QueryRow(query, referralCode).Scan(&userId)
	if err != nil {
		log.Println("GetUserByReferralCode: Failed to make query :", err)
		return 0, err
	}

	return int(userId.Int64), nil
}

func SaveInviterDetails(invitee, inviter int) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SaveInviterDetails: Failed while connecting with the database :", err)
		return err
	}
	defer db.Close()

	query := `
			INSERT INTO
				 user_invitee(invitee,inviter,created_at)
			VALUES($1,$2,NOW()) 
			ON CONFLICT (invitee, inviter) DO NOTHING
			RETURNING id`

	var userId sql.NullInt64
	err = db.QueryRow(query, invitee, inviter).Scan(&userId)
	if err != nil {
		log.Println("SaveInviterDetails: Failed to make query :", err)
		return err
	}

	return nil
}
func GetUserDetailsWithPhone(phone string) (User, error) {
	query := `
	SELECT
		id, 
		email,
		phone_number,
		phone_verified,
		email_verified,
		otp,
		query_param,
		location   
	FROM 
		public.user 
	WHERE 
		phone_number = '` + phone + "' "

	return getUser(query)

}

// GetUserByID retrieves a student record from the database by user ID.
// Parameters:
// - userID: The ID of the user to retrieve.
// Returns:
// - Student: The user record retrieved from the database.
// - error: Any error encountered during the process.
func GetUserByID(userID int) (User, error) {

	query := `
		SELECT
			id, 
			email,
			phone_number,
			phone_verified,
			email_verified,
			otp,
			query_param,
			location
		FROM 
			public.user 
		WHERE 
			id=` + strconv.Itoa(userID)

	return getUser(query)
}

func getUser(query string) (User, error) {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetUserByID: Failed while connecting with the database :", err)
		return User{}, err
	}
	defer db.Close()
	var (
		ID              sql.NullInt64
		email           sql.NullString
		phone           sql.NullString
		isPhoneVerified sql.NullBool
		isEmailVerified sql.NullBool
		otp             sql.NullString
		queryParam      sql.NullString
		location        sql.NullString
		user            User
	)

	err = db.QueryRow(query).Scan(
		&ID,
		&email,
		&phone,
		&isPhoneVerified,
		&isEmailVerified,
		&otp,
		&queryParam,
		&location,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("GetUserByID : user doesnt exist error: ", err)
		}
	}

	var locStruct services.Location
	loc := strings.Split(location.String, ", ")
	if len(loc) > 2 {
		locStruct.City = loc[0]
		locStruct.State = loc[1]
		locStruct.CountryCode = loc[2]

	}

	user = User{
		ID:              ID.Int64,
		Email:           email.String,
		Phone:           phone.String,
		IsPhoneVerified: isPhoneVerified.Bool,
		IsEmailVerified: isEmailVerified.Bool,
		OTP:             otp.String,
		QueryParams:     queryParam.String,
		Location:        locStruct,
	}

	return user, err
}

func SaveNewUser(user *User) (string, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SaveOTP: Failed while connecting with the database :", err)
		return "", err
	}
	defer db.Close()

	query := `
			INSERT INTO public.user (phone_number, otp,otp_valid_until,referral_code,location, ip,query_param,timezone,dialing_code)
    VALUES ($1, $2, $3, $4, $5, $6,$7,$8,$9)
			ON CONFLICT (phone_number)
			DO UPDATE SET
			otp = $2,
			otp_valid_until = $3
			RETURNING id,phone_verified,email_verified,referral_code,query_param,
			 CASE WHEN xmax = 0 THEN 'insert' ELSE 'update' END AS action`
	log.Println("===", query)
	var (
		user_id       int
		phoneVerified bool
		emailVerified bool
		referralCode  sql.NullString
		queryParam    sql.NullString
		action        string
	)
	userLocation := user.Location.City + ", " + user.Location.State + ", " + user.Location.Country
	otpValidUntil := time.Now().Add(time.Minute * 5)
	// Set phone_otp_expire time to 24 hrs from date of student generated.
	if len(user.ReferralCode) != 0 {
		err = db.QueryRow(query, user.Phone, user.OTP, otpValidUntil, user.ReferralCode, userLocation, user.UserIP, user.QueryParams, user.Timezone, user.DialingCode).Scan(&user_id, &phoneVerified, &emailVerified, &referralCode, &queryParam, &action)
	} else {
		err = db.QueryRow(query, user.Phone, user.OTP, otpValidUntil, userLocation, user.UserIP, user.QueryParams, user.Timezone, user.DialingCode).Scan(&user_id, &phoneVerified, &emailVerified, &referralCode, &queryParam, &action)
	}
	if err != nil {
		log.Println("SaveOTP: failed while execute the query for saving otp in database with error :", err)
		return "", err
	}

	user.ID = int64(user_id)
	user.IsPhoneVerified = phoneVerified
	user.IsEmailVerified = emailVerified
	user.ReferralCode = referralCode.String
	user.QueryParams = queryParam.String

	return action, nil
}

func GetRoleID(role string) (int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetRoleID: Failed while connecting with the database :", err)
		return 0, err
	}
	defer db.Close()
	query := `SELECT id FROM roles WHERE role = $1`
	var id sql.NullInt64
	err = db.QueryRow(query, role).Scan(&id)
	if err != nil {
		log.Println("GetRoleID: failed while execute the query with error :", err)
		return 0, err
	}
	return int(id.Int64), nil
}

func SaveUserRole(userID, roleID int) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SaveUserRole: Failed while connecting with the database :", err)
		return err
	}
	defer db.Close()

	query := `INSERT INTO
					user2roles(user_id,role_id)
				VALUES($1,$2)
				ON CONFLICT (user_id, role_id) DO NOTHING
				RETURNING id`

	var id sql.NullInt64
	err = db.QueryRow(query, userID, roleID).Scan(&id)
	if err != nil {
		log.Println("SaveUserRole: failed while execute the query with error :", err)
		return err
	}
	return nil
}

// GetUserIdByToken retrieves the student ID associated with the provided JWT token from the 'student_auth' table.
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
func GetUserIdByToken(token string) (int, error) {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the error and return 0 with the error.
		log.Println("GetUserIdByToken: failed while trying to connect with database with error: ", err)
		return 0, err
	}
	defer db.Close()
	// Ensure the database connection is closed when the function returns.

	// Define the SQL query to fetch the student ID from the 'student_auth' table using the provided JWT token.
	query := `
				SELECT 
					user_id 
				FROM 
					user_auth 
				WHERE 
					jwt_token=$1 AND
					is_active = true`
	var userID int
	// Execute the query with the provided token and retrieve the student ID.

	err = db.QueryRow(query, token).Scan(&userID)
	if err != nil {
		// If there is an error executing the query or scanning the result, log the error and return 0 with the error.
		log.Println("GetUserIdByToken: failed while execute the query with error: ", err)
		return 0, err
	}
	// Return the student ID associated with the JWT token and nil indicating no errors.
	return userID, nil
}

// DeactivateUserSession marks the user's session as inactive in the database.
//
// Parameters:
// - userID int: The unique identifier of the user whose session needs to be marked as inactive.
//
// Returns:
// - error: An error object if there was an issue with the database connection or query execution; otherwise, `nil`.
func DeactivateUserSession(userID int, token string) error {
	// Get a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		// If there is an error connecting to the database, log the with the error.
		log.Println("DeactivateUserSession: Failed while trying to connect with database with error: ", err)
		return err
	}
	defer db.Close()
	query := `
			UPDATE 
					user_auth
			SET
					is_active = false
			WHERE
					user_id = $1 and jwt_token=$2`
	_, err = db.Exec(query, userID, token)
	if err != nil {
		// If there is an error executing the query or scanning the result, log with the error.
		log.Println("DeactivateUserSession: Failed while execute the query with error: ", err)
		return err
	}
	return nil
}

// Stores Stripe Customer ID in database.
func StoreStripeCustomerId(userId int, stripeId string) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("StoreStripeCustomerId: Error connecting to the database :", err)
		return err
	}
	defer db.Close()
	const query = `
		INSERT INTO
			user2stripe (stripe_id, user_id) VALUES ($1, $2)`

	_, err = db.Exec(query, stripeId, userId)
	if err != nil {
		log.Println("StoreStripeCustomerId: Error executing query:", err)
		return err
	}

	return nil
}

// SaveCRMLeadsID saves the relationship between a user and their corresponding Zoho CRM lead ID.
//
// This function inserts a mapping of user ID and Zoho CRM lead ID into the 'user2zoho_leads' table.
//
// Parameters:
// - userID: The ID of the user to be mapped.
// - leadID: The Zoho CRM lead ID associated with the user.
//
// Steps:
// 1. Connect to the database using the provided configuration.
// 2. Execute an INSERT query to store the user-lead mapping.
// 3. Handle errors during database connection or query execution.
//
// Returns:
// - An error if the database connection or query execution fails.
func SaveCRMLeadsID(userID, leadID int) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("SaveCRMLeadsID: Error connecting to the database :", err)
		return err
	}
	defer db.Close()
	query := `INSERT INTO user2zoho_leads (user_id,lead_id) VALUES ($1,$2)`
	_, err = db.Exec(query, userID, leadID)
	if err != nil {
		log.Println("SaveCRMLeadsID: Failed while execute query with error :", err)
		return err
	}
	return nil
}

// GetZohoCRMLeadIDByUserID retrieves the Zoho CRM lead ID associated with a given user ID.
//
// This function queries the 'user2zoho_leads' table to find the lead ID linked to the specified user ID.
//
// Parameters:
// - userID: The ID of the user for which the lead ID is being fetched.
//
// Steps:
// 1. Connect to the database using the provided configuration.
// 2. Execute a SELECT query to retrieve the lead ID from the 'user2zoho_leads' table.
// 3. If no lead is found, return 0 and nil to indicate no association.
// 4. Handle errors during database connection or query execution.
//
// Returns:
// - The Zoho CRM lead ID associated with the user, or 0 if no lead is found.
// - An error if the database connection or query execution fails.
func GetZohoCRMLeadIDByUserID(userID int) (int, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetZohoCRMLeadByUserID: Error connecting to the database:", err)
		return 0, err
	}
	defer db.Close()

	var zohoLeadID sql.NullInt64
	query := `SELECT lead_id FROM user2zoho_leads WHERE user_id = $1`
	err = db.QueryRow(query, userID).Scan(&zohoLeadID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("GetZohoCRMLeadByUserID: No lead found for the given user ID")
			return 0, nil
		}
		log.Println("GetZohoCRMLeadByUserID: Failed to execute query:", err)
		return 0, err
	}

	return (int(zohoLeadID.Int64)), nil
}
