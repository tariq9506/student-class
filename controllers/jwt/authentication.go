package jwt

import (
	"log"
	"os"
	"time"

	"tutree/student-apis/controllers"
	"tutree/student-apis/models"
	jwtauthenticationgo "tutree/student-apis/models/jwt_authentication.go"
	"tutree/student-apis/utility"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// CreateSession creates a new session for a user by generating a JWT token and setting it as a cookie in the HTTP response.
// Parameters:
// - c: The Gin context object for handling the HTTP request and response.
// - phone: The phone number of the user for whom the session is being created.
// Returns:
// - string: The generated JWT token string.
func CreateUserAuth(c *gin.Context, user models.User) string {

	userAgent := c.GetHeader("User-Agent")
	userIP := utility.GetClientIP(c)
	authLocaiton := controllers.GetLocationFromIP(userIP)

	var token string
	// Define the session duration as 5 years in seconds
	// 365 days/year * 24 hours/day * 60 minutes/hour * 60 seconds/minute * 5 years
	maxAge := 365 * 24 * 60 * 60 * 5
	// Get the domain name from the environment variables for the cookie
	domain := os.Getenv("HOST")
	// Generate a JWT token using the provided phone number
	token, _ = createJWT(user.Phone)
	// Set the JWT token as a cookie in the HTTP response
	// Parameters:
	// - "tokenString": The name of the cookie
	// - SessionID: The value of the cookie (the JWT token)
	// - maxAge: The maximum age of the cookie in seconds (5 years)
	// - "/": The path where the cookie is valid (root path)
	// - domain: The domain for which the cookie is valid
	// - false: Whether the cookie should be sent only over HTTPS (false for HTTP and HTTPS)
	// - false: Whether the cookie should be accessible only through HTTP requests (true for HTTP-only cookies)
	tokenValidity := getTimeForCookies()
	location := authLocaiton.City + ", " + authLocaiton.State + ", " + authLocaiton.Country
	jwtauthenticationgo.CreateNewSession(user.ID, tokenValidity, token, userAgent, userIP, location)

	c.SetCookie("token", token, maxAge, "/", domain, false, false)
	return token
}

// GetTimeForCookies returns the expiration time for cookies, set to 5 years from the current time.
// Returns:
// - time.Time: The expiration time for cookies, which is 5 years from now.
func getTimeForCookies() time.Time {
	return time.Now().Add((365 * 5) * 24 * time.Hour)
}

// CreateJWT generates a JWT token for a given phone number.
// Parameters:
// - phone: The phone number for which the JWT token is generated.
// Returns:
// - string: The generated JWT token as a string.
// - error: Any error encountered during the token generation.
func createJWT(phone string) (string, error) {
	expirationTime := getTimeForCookies()

	claims := &jwtauthenticationgo.JWTAuthClaims{
		Phone: phone + time.Now().GoString(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		log.Println("[ERROR] CreateJWT: failed with an error: ", err)
	}

	return tokenString, err
}
