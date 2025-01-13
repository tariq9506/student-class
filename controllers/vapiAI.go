package controllers

import (
	"net/http"
	"tutree/student-apis/services"

	"github.com/gin-gonic/gin"
)

func GetCallDetails(c *gin.Context) {
	callerID := c.Query("caller-id")
	if len(callerID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone_number is required"})
		return
	}

	status, resone := services.GetCallDetails(callerID)
	c.JSON(http.StatusOK, gin.H{"status": status, "resone": resone})
}
