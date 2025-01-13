package services

import (
	"os"

	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Bucket        string
	FilePath      string
	CloudFrontCDN bool
}

// RemoveCookies used to remove cookies
func RemoveCookies(c *gin.Context, cookiesName string) {
	domain := os.Getenv("HOST")
	c.SetCookie(cookiesName, "", -1, "/", domain, false, true)
}
