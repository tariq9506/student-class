package common

// THIS FOLDER is created to handle the issue "import cycle not allowed"
// code which is using on multiple function then that code should come under this group.

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"tutree/student-apis/config"
	"tutree/student-apis/services"

	"github.com/lib/pq"
)

// UpdateSubscriptionTrialEndToDate updates the trial end date for a user in Stripe
// based on their first paid session end date and time.
func UpdateSubscriptionTrialEndToDate(userID int, firstPaidSessionEndDateAndTime time.Time, loc *time.Location) error {
	// Get the buffer time from the environment or use the default value
	timeAfterPriceDeducted, err := strconv.Atoi(os.Getenv("GRACE_TIME_TO_CHARGE_CUSTOMER"))
	if err != nil || timeAfterPriceDeducted <= 0 {
		timeAfterPriceDeducted = 20 // Default buffer time
	}

	// Calculate the new trial end date and time
	convertedDateAndTime := firstPaidSessionEndDateAndTime.Add(
		time.Duration(timeAfterPriceDeducted) * time.Minute,
	).In(loc)
	fmt.Println("CONVERTED DATE AND TIME", convertedDateAndTime)

	// Update the trial end date in Stripe
	userIDs := []int{userID}
	err = UpdateTrialEndDate(userIDs, convertedDateAndTime)
	if err != nil {
		log.Printf("Failed to update trial end date for user %d: %v", userID, err)
		return err
	}

	log.Printf("Successfully updated trial end date for user %d to %v", userID, convertedDateAndTime)
	return nil
}

func UpdateTrialEndDate(userIDs []int, updatedTrialEndDate time.Time) error {
	// Retrieve the subscription IDs for the given user IDs.
	subscriptionMap, err := GetSubscriptionAndPaymentDetails(userIDs)
	if err != nil {
		log.Println("UpdateTrialEndDate: Failed to retrieve subscription details:", err)
		return err
	}

	log.Println("UpdateTrialEndDate: Retrieved subscription details:", subscriptionMap)

	// Iterate over the subscriptionMap and update the trial end date for each subscription ID.
	for _, subscriptionID := range subscriptionMap {
		if subscriptionID == "" {
			log.Println("UpdateTrialEndDate: Skipping empty subscription ID")
			continue
		}

		err := services.UpdateSubscriptionTrialEndDateOnStripe(subscriptionID, updatedTrialEndDate)
		if err != nil {
			log.Printf("UpdateTrialEndDate: Failed to update trial end date for subscription ID %s: %v", subscriptionID, err)
			// Optionally, continue updating other subscriptions even if one fails.
			continue
		}

		log.Printf("UpdateTrialEndDate: Successfully updated trial end date for subscription ID %s", subscriptionID)
	}

	return nil
}

// GetSubscriptionAndPaymentDetails is a helper function to fetch miscellaneous
// information about payments and subscription details. It retrieves the subscription
// IDs for all given user IDs where the payment was successful and the subscription
// status is active.

func GetSubscriptionAndPaymentDetails(userIDs []int) (map[string]string, error) {
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetSubscriptionAndPaymentDetails: Error connecting to the database:", err)
		return nil, err
	}
	defer db.Close()

	// Construct the query to fetch subscription IDs for all given user IDs.
	query := `
		SELECT 
			swe.user_id,
			swe.data->'object'->>'subscription' AS subscription_id
		FROM 
			stripe_webhook_event swe
		JOIN 
			user2subscription us ON swe.user_id = us.user_id
		WHERE 
			swe.event_type = 'invoice.payment_succeeded'
			AND swe.user_id = ANY($1)
			AND us.status = 'active'
		ORDER BY 
			us.updated_at DESC`

	// Execute the query with pq.Array to pass the userIDs slice.
	rows, err := db.Query(query, pq.Array(userIDs))
	if err != nil {
		log.Println("GetSubscriptionAndPaymentDetails: Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	// Map to store userID -> subscriptionID.
	subscriptionMap := make(map[string]string)

	// Iterate over the rows and populate the map.
	for rows.Next() {
		var userID, subscriptionID string
		if err := rows.Scan(&userID, &subscriptionID); err != nil {
			log.Println("GetSubscriptionAndPaymentDetails: Error scanning row:", err)
			return nil, err
		}
		subscriptionMap[userID] = subscriptionID
	}

	// Check for any row iteration errors.
	if err := rows.Err(); err != nil {
		log.Println("GetSubscriptionAndPaymentDetails: Row iteration error:", err)
		return nil, err
	}

	return subscriptionMap, nil
}
