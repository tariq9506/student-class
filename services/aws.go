package services

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"tutree/student-apis/utility"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadFileToS3AtLocation(file bytes.Buffer, fileHeader *multipart.FileHeader, fileInfo FileInfo) string {

	s := AwsSession()

	filename := fileHeader.Filename
	filename = strings.Replace(filename, " ", "", -1)
	filename = strings.Replace(filename, "/", "", -1)
	filename = strings.Replace(filename, "+", "", -1)

	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	fileLocation := filename
	//check filepath is empty or not
	if len(fileInfo.FilePath) != 0 || fileInfo.FilePath != "" {
		fileInfo.FilePath = strings.Replace(fileInfo.FilePath, "/", "", -1)
		fileLocation = fileInfo.FilePath + "/" + filename
	}

	bucketName := os.Getenv("BUCKETNAME")
	if len(fileInfo.Bucket) != 0 || fileInfo.Bucket != "" {
		bucketName = fileInfo.Bucket
	}
	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(fileLocation),
		ACL:                  aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})

	if err != nil {
		log.Println("UploadFileToS3AtLocation : Could not upload file with error: ", err)
	}
	url := fmt.Sprintf("%s/%s", utility.GetAWSURL(bucketName), fileLocation)
	return url

}
func AwsSession() *session.Session {
	// Session
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("SECRETID"),  // id
			os.Getenv("SECRETKEY"), // secret
			""),                    // token can be left blank for now
	})

	if err != nil {
		log.Println(err.Error())
		log.Println("AwsSession : Could not upload file with error: ", err)
	}

	return s
}
