package timezone

import (
	"tutree/student-apis/models/timezone"

	"log"
	"tutree/student-apis/controllers"

	"tutree/student-apis/apiError"

	"net/http"

	"github.com/gin-gonic/gin"
)

func GetTimezones(c *gin.Context) {
	timezones, err := timezone.GetTimezones()

	if err != nil {
		log.Println("GetTimezones: Failed to get timezones with error: ", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"timezones": timezones,
	})

}
