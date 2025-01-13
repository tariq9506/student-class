package metadata

import (
	"log"
	"net/http"
	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	"tutree/student-apis/models/metadata"

	"github.com/gin-gonic/gin"
)

// GetSubjects godoc
// @Summary This API will use to fetch subjects.
// @description This API is designed to fetch all the subjects offered by tutree to students.
// @Tags Metadata
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200
// @Router /metadata/subjects [GET]
func GetSubjects(c *gin.Context) {
	// Fetch the list of subjects from the database.
	subjectList, err := metadata.GetSubjects()
	if err != nil {
		log.Println("[ERROR] GetSubjects: Failed to fetch subjects from the database with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Check if the subject list is empty.
	if len(subjectList) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "Success",
		"subjects": subjectList,
	})
}

// GetGrades godoc
// @Summary This API will use to fetch grades.
// @description This API is designed to fetch all the grades of students.
// @Tags Metadata
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Success 200
// @Router /metadata/grades [GET]
func GetGrades(c *gin.Context) {
	// Fetch the list of grades from the database.
	gradeList, err := metadata.GetGrades()
	if err != nil {
		log.Println("[ERROR] GetGrades: Failed to fetch grades from the database with error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Check if the grade list is empty.
	if len(gradeList) == 0 {
		c.JSON(http.StatusNoContent, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Success",
		"grades": gradeList,
	})
}
