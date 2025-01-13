package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

func GetWebhookEvents(c *gin.Context) (stripe.Event, error) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("GetWebhookEvents: failed to get the data with", err)
		return stripe.Event{}, fmt.Errorf("request body too large")
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		log.Println("GetWebhookEvents: failed to get secret", err)
		return stripe.Event{}, fmt.Errorf("webhook secret not configured")
	}

	signatureHeader := c.Request.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signatureHeader, endpointSecret)
	if err != nil {
		log.Println("GetWebhookEvents: failed to get Stripe signature", err)
		return stripe.Event{}, fmt.Errorf("failed to verify signature: %v", err)
	}

	return event, nil
}
