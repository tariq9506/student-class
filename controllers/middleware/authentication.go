package middleware

import (
	"log"
	"net/http"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	"tutree/student-apis/models"

	"github.com/gin-gonic/gin"
)

// TokenAuthMiddleware is a middleware function for authenticating requests using a token.
// It checks the Authorization header for a token, validates it, and extracts the user ID.
// If the token is valid, the user ID is added to the request context for further use by handlers.
func TokenAuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	log.Println("Token ", token)
	log.Println("header", c.Request.Header)

	if len(token) == 0 {
		log.Println("[UN-AUTHORIZED] TokenAuthMiddleware: No token provided.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUnAuthorizedAccess, nil, c)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	// using token fetch student's information.
	userID, err := models.GetUserIdByToken(token)
	if err != nil {
		log.Println("[UN-AUTHORIZED] TokenAuthMiddleware: Failed to fetch student details by token with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorUnAuthorizedAccess, err, c)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Add userID to the context for further use in handlers
	c.Set("userID", userID)

	// Continue to the next middleware/handler
	c.Next()
}
