package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"tutree/student-apis/utility"

	brevo "github.com/getbrevo/brevo-go/lib"

	"github.com/google/uuid"
)

type DynamicEmailDetailsForBrevo struct {
	To          []string
	DynamicData map[string]interface{}
	TemplateID  int64
}

type ICSDynamicDetails struct {
	EventUID       string
	EventDTStamp   string
	EventStart     string
	OrganizerName  string
	OrganizerEmail string
	EventEnd       string
	Summary        string
}

func (p DynamicEmailDetailsForBrevo) CreateICSDynamicDetails() ICSDynamicDetails {

	eventStart := ""
	eventEnd := ""
	if startTime, ok := p.DynamicData["event_start_time"].(time.Time); ok {
		eventStart = startTime.Format("20060102T150405Z")
		eventEnd = startTime.Add(40 * time.Minute).Format("20060102T150405Z")
	}
	// Set the summary, default to "Tutree Session" if not provided
	summary := "Tutree Session"
	if eventSummary, ok := p.DynamicData["event_summary"].(string); ok {
		summary = eventSummary
	}
	return ICSDynamicDetails{
		EventUID:       uuid.New().String(),
		EventDTStamp:   time.Now().UTC().Format("20060102T150405Z"),
		EventStart:     eventStart,
		OrganizerName:  "Tutree",
		OrganizerEmail: "support@tutree.com",
		EventEnd:       eventEnd,
		Summary:        summary,
	}
}

// input:personalizations
// output: Sending email.
// ComposeDynamicTemplateEmailsForBrevoService composes dynamic template emails for the Brevo service based on provided email
// details. It constructs email objects for each recipient specified in the personalizations data and sends them using
// another function sendDynamicTemplateEmailForBrevoService.
func ComposeDynamicTemplateEmailsForBrevoService(personalizations DynamicEmailDetailsForBrevo) {

	emails := []brevo.SendSmtpEmailTo{}
	// This line of code would iterate over each personalizations.To( or basically would iterate over each user) and then it would
	// append each emailIDs in 'emails'.
	for _, to := range personalizations.To {
		emails = append(emails, brevo.SendSmtpEmailTo{
			Email: to,
		})
	}

	smtpEmail := brevo.SendSmtpEmail{
		TemplateId: personalizations.TemplateID,
		Sender: &brevo.SendSmtpEmailSender{
			Name:  utility.GetBrevoSenderName(),
			Email: utility.GetBrevoSenderEmail(),
		},
		To:     emails,
		Params: personalizations.DynamicData,
	}
	sendDynamicTemplateEmailForBrevoService(smtpEmail)
}

// createICSContent generates the content for the .ics file
func createICSContent(personalizations DynamicEmailDetailsForBrevo) string {

	icsDyamicDetails := personalizations.CreateICSDynamicDetails()

	icsTemplate := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Tutree//tutree//EN
BEGIN:VEVENT
UID:%s
DTSTAMP:%s
DTSTART:%s
DTEND:%s
ORGANIZER;CN=%s:MAILTO:%s
SUMMARY:%s
END:VEVENT
END:VCALENDAR`

	return fmt.Sprintf(strings.TrimSpace(icsTemplate),
		icsDyamicDetails.EventUID,
		icsDyamicDetails.EventDTStamp,
		icsDyamicDetails.EventStart,
		icsDyamicDetails.EventEnd,
		icsDyamicDetails.OrganizerName,
		icsDyamicDetails.OrganizerEmail,
		icsDyamicDetails.Summary,
	)
}

func ComposeDynamicTemplateEmailsForBrevoServiceWithICS(personalizations DynamicEmailDetailsForBrevo) {

	emails := []brevo.SendSmtpEmailTo{}
	// This line of code would iterate over each personalizations.To( or basically would iterate over each user) and then it would
	// append each emailIDs in 'emails'.
	for _, to := range personalizations.To {
		emails = append(emails, brevo.SendSmtpEmailTo{
			Email: to,
		})
	}

	icsContent := createICSContent(personalizations)
	log.Println(icsContent)
	icsBase64 := base64.StdEncoding.EncodeToString([]byte(icsContent))

	// Prepare the attachment
	attachments := []brevo.SendSmtpEmailAttachment{
		{
			Content: icsBase64,
			Name:    "event.ics",
		},
	}

	smtpEmail := brevo.SendSmtpEmail{
		TemplateId: personalizations.TemplateID,
		Sender: &brevo.SendSmtpEmailSender{
			Name:  utility.GetBrevoSenderName(),
			Email: utility.GetBrevoSenderEmail(),
		},
		To:         emails,
		Params:     personalizations.DynamicData,
		Attachment: attachments,
	}

	sendDynamicTemplateEmailForBrevoService(smtpEmail)
}

// input: smtpEmail
// output: Call Brevo API
// sendDynamicTemplateEmailForBrevoService sends a dynamic template email via the Brevo service API using the provided SMTP
// email object. It configures API key authorization, creates a new API client, and sends the email. Any errors
// encountered during the process are logged.
func sendDynamicTemplateEmailForBrevoService(smtpEmail brevo.SendSmtpEmail) {
	var ctx context.Context
	cfg := brevo.NewConfiguration()
	// Configure API key authorization: api-key
	cfg.AddDefaultHeader("api-key", utility.GetBrevoAPIKey())
	// Configure API key authorization: partner-key
	cfg.AddDefaultHeader("partner-key", utility.GetBrevoAPIKey())
	br := brevo.NewAPIClient(cfg)
	_, resp, err := br.TransactionalEmailsApi.SendTransacEmail(ctx, smtpEmail)
	if err != nil {
		// Log detailed error information
		if resp != nil {
			log.Printf("sendDynamicTemplateEmailForBrevoService: Error when calling SendTransacEmail: %s\nResponse status: %s\nResponse body: %s", err.Error(), resp.Status, resp.Body)
		} else {
			log.Printf("sendDynamicTemplateEmailForBrevoService: Error when calling SendTransacEmail: %s", err.Error())
		}
		return
	}
	log.Println("sendDynamicTemplateEmailForBrevoService Response:", resp)
}
