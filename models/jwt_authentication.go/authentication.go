package jwtauthenticationgo

import (
	"log"
	"time"
	"tutree/student-apis/config"

	"github.com/dgrijalva/jwt-go"
)

// JWTAuthClaims Structure to store JWT
type JWTAuthClaims struct {
	Phone string `json:"email"`
	jwt.StandardClaims
}

// CreateNewSession function to insert user session into table
// Parameter -
// user id : id of user type int.
// validUntil : Time of validation for OTP type time.Time.
// token  : Token is
// userAgent: The user agent contains information about the client making the request to the web server.
// userIP: ip address of user
// location : location of user
// Return -
// success or error
func CreateNewSession(userID int64, validUntil time.Time, token, userAgent, userIP, location string) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("CreateNewSession: Failed while connecting with the database :", err)
		return err
	}
	defer db.Close()

	sqlInsert := `
	INSERT INTO 
		user_auth (
			user_id,
			jwt_token,
			valid_until,
			browser,
			ip,
			location,
			created_at
		)
	VALUES 
		($1, $2,$3,$4,$5,$6,NOW())
	RETURNING id`

	_, err = db.Exec(sqlInsert, userID, token, validUntil, userAgent, userIP, location)
	if err != nil {
		log.Println("CreateNewSession: failed while executing query with error:", err)
		return err
	}
	return nil
}
