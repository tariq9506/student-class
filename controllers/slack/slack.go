package slack

import (
	"fmt"
	"log"
	"os"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/services"
)

// SendSuccessfulPaymentMessageOnSlack sends a Slack notification when a student's demo is successfully converted.
// Retrieves student details, formats a message, and sends it to a specific Slack channel.
func SendSuccessfulPaymentMessageOnSlack(userID int, planName string, maxSession int, price string) {
	// Retrieve the student profile using the user ID
	student, err := students.GetStudentProfile(userID)
	if err != nil {
		log.Println("SendSuccessfulPaymentMessageOnSlack: failed to get student details by user id with error:", err)
		return
	}

	// Get Slack channel ID from environment
	channelID := os.Getenv("DEMO_CONVERTED_CHANNEL")
	if len(channelID) == 0 {
		log.Println("[ERROR] Failed to get DEMO_CONVERTED_CHANNEL")
		return
	}

	// Format student's location details
	location := student.User.Location.City + ", " + student.User.Location.State + ", " + student.User.Location.CountryCode

	// Create the message to be sent to Slack
	message := fmt.Sprintf("*Demo Successfully Converted for (Payment Received):* \nKid's Name: %s \nPhone Number: %s \nEmail: %s \nLocation: %s \nGrade: %s \nPlan-Choosen: %s \nNumber of Classes: %v \nPrice: %s",
		student.StudentName, student.User.Phone, student.ParentEmail, location, student.Grade, planName, maxSession, price)

	// Send the message to the specified Slack channel
	services.SendSlackMessage(channelID, message)
}
func SendPaymentSetupMessageOnSlack(userID int64, planName string, maxSession int, price string) {

	student, err := students.GetStudentProfile(int(userID))
	if err != nil {
		log.Println("SendSuccessfulPaymentMessageOnSlack: failed to get student details by user id with error:", err)
		return
	}
	// Get Slack channel ID from environment
	channelID := os.Getenv("DEMO_CONVERTED_CHANNEL")
	if len(channelID) == 0 {
		log.Println("[ERROR] Failed to get DEMO_CONVERTED_CHANNEL")
		return
	}

	// Format student's location details
	location := student.User.Location.City + ", " + student.User.Location.State + ", " + student.User.Location.CountryCode

	// Create the message to be sent to Slack
	message := fmt.Sprintf("*Demo Successfully Converted for (Payment Setup):* \nKid's Name: %s \nPhone Number: %s \nEmail: %s \nLocation: %s \nGrade: %s \nPlan-Choosen: %s \nNumber of Classes: %v \nPrice: %s",
		student.StudentName, student.User.Phone, student.ParentEmail, location, student.Grade, planName, maxSession, price)

	log.Println("message", message)
	// Send the message to the specified Slack channel
	services.SendSlackMessage(channelID, message)
}
