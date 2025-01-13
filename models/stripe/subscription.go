package stripe

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"tutree/student-apis/common"
	"tutree/student-apis/config"
	"tutree/student-apis/constants"
	messageservices "tutree/student-apis/controllers/message_service"
	"tutree/student-apis/controllers/slack"
	sessionmodel "tutree/student-apis/models/sessionModel"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/services"
	"tutree/student-apis/utility"

	"github.com/stripe/stripe-go"
)

// type SubscriptionPlans struct {
// 	Id   int    `json:"id" example:"1"`
// 	Name string `json:"name" example:"Basic"`

// 	// A brief description of the subscription plan.
// 	Description string `json:"description" example:"(4 Classes)"`

// 	// A list of marketing features, each line as a separate string. May be empty.
// 	MarketingFeatures []string `json:"marketing_features" example:"Free Demo Class,1 Student : 1 Tutor"`

// 	// Subscription price as a string formatted to display, like $119.99
// 	Price string `json:"price" example:"$119"`

// 	// Same as above; but this field represent a price without a discount.
// 	// (It serves only marketing purpose, not actually used anywhere.)
// 	// Also, unlike Price, this field's value is a pointer, it can be nil.
// 	ListPrice  *string `json:"list_price" example:"$199" extensions:"x-nullable"`
// 	ProductID  string  `json:"product_id"`
// 	MaxSession int     `json:"total_classes"`
// 	Frequency  string  `json:"frequency"`
// }

type Customer struct {
	Name     string
	Email    string
	Phone    string
	StripeId string
}

type SubscriptionPlan struct {
	Id                int      `json:"id" example:"1"`
	Name              string   `json:"name" example:"Basic"`
	Description       string   `json:"description" example:"(4 Classes)"`
	MarketingFeatures []string `json:"marketing_features" example:"Free Demo Class,1 Student : 1 Tutor"`
	Price             int      `json:"price,omitempty" example:"$119"`
	ListPrice         *int     `json:"list_price,omitempty" example:"$199" extensions:"x-nullable"`
	Currency          string
	ProductId         string                   `json:"product_id"`
	MaxSession        int                      `json:"total_classes"`
	PerWeekLimit      int                      `json:"per_week_limit"`
	Frequency         string                   `json:"frequency"`
	Prices            []SubscriptionPlanPrices `json:"prices"`
}
type SubscriptionPlanPrices struct {
	Id            int    `json:"price_id"`
	Price         string `json:"price"`
	ListPrice     string `json:"list_price" example:"$199" extensions:"x-nullable"`
	StripePriceID string `json:"stripe_price_id" example:"" extensions:"x-nullable"`
}

func FetchCustomerInfo(userId int) (*Customer, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("FetchCustomerInfo: Error connecting to the database  :", err)
		return nil, err
	}
	defer db.Close()

	const query = `
		SELECT
			s.name, u.email, u.phone_number, u2s.stripe_id
		FROM
			public.user AS u
		LEFT JOIN
			student AS s ON u.id = s.user_id
		LEFT JOIN
			user2stripe AS u2s ON u.id = u2s.user_id
		WHERE
			u.id = $1`

	var (
		cusName, email, phone, stripeId sql.NullString
		info                            Customer
	)

	err = db.QueryRow(query, userId).Scan(&cusName, &email, &phone, &stripeId)
	if err == sql.ErrNoRows {
		errMsg := fmt.Sprintf("FetchCustomerInfo: No customer found with userId: %d", userId)
		log.Println(errMsg)
		return nil, fmt.Errorf(errMsg)
	} else if err != nil {
		log.Println("FetchCustomerInfo: Error querying customer info :", err)
		return nil, err
	}

	info = Customer{
		Name:     cusName.String,
		Email:    email.String,
		Phone:    phone.String,
		StripeId: stripeId.String,
	}

	return &info, nil
}

type SubscriptionPlanDetails struct {
	PlanID        int    `json:"plan_id"`
	MaxSession    int    `json:"total_classes"`
	StripePlanID  string `json:"stripe_plan_id"`
	PlanName      string `json:"plan_name"`
	ListPrice     string `json:"list_price"`
	ActualPrice   string `json:"actual_price"`
	PriceID       int    `json:"price_id"`
	StripePriceID string `json:"stripe_price_id"`
	PerWeekLimit  int    `json:"per_week_limit"`
}

func FetchStripeSubscriptionPlanId(dBplanId int) (*SubscriptionPlanDetails, error) {
	// Get database connection
	db, err := config.GetDB2()
	if err != nil {
		log.Println("FetchStripeSubscriptionPlanId: Error connecting to the database:", err)
		return nil, err
	}
	defer db.Close()

	// SQL query to fetch the subscription plan details
	// const query = `
	// 	SELECT
	// 		stripe_id, name, max_session, list_price, price
	// 	FROM
	// 		subscription_plan
	// 	WHERE
	// 		id = $1`

	const query = `
		SELECT
    		plan.id,plan.stripe_id, plan.name, plan.max_session, price.list_price, price.price, price.id,stripe_price_id,plan.per_week_limit
		FROM
			subscription_plan as plan,
		    subscription_prices as price
		WHERE
		    plan.id=price.subscription_plan_id
			AND price.id = $1`

	// Variables to hold the results from the query
	var (
		planID        sql.NullInt64
		stripePlanId  sql.NullString
		planName      sql.NullString
		maxSession    sql.NullInt32
		listPrice     sql.NullInt64
		price         sql.NullInt64
		priceID       sql.NullInt64
		stripePriceId sql.NullString
		perWeekLimit  sql.NullInt32
	)

	// Execute the query and scan the result into variables
	err = db.QueryRow(query, dBplanId).Scan(&planID, &stripePlanId, &planName, &maxSession, &listPrice, &price, &priceID, &stripePriceId, &perWeekLimit)
	if err != nil {
		log.Println("FetchStripeSubscriptionPlanId: Error querying subscription plan:", err)
		return nil, err
	}

	// Convert the SQL Null types to their respective Go types
	stripePlanIdStr := utility.SQLNullStringToString(stripePlanId)
	planNameStr := utility.SQLNullStringToString(planName)
	maxSessionInt := int(maxSession.Int32)
	listPriceInt := utility.SQLNullIntToInt(listPrice) // Convert to int
	priceInt := utility.SQLNullIntToInt(price)         // Convert to int

	// Format the prices to display
	listPriceWithDollar := formatPriceToDisplay(int(listPriceInt), "USD")
	actualPriceWithDollar := formatPriceToDisplay(int(priceInt), "USD")

	//TODO: CHANGE TO SUBSCRIPTION STRUCT, KEEP
	// Create and return the SubscriptionPlanDetails struct
	subscriptionPlanDetails := &SubscriptionPlanDetails{
		PlanID:        int(planID.Int64),
		StripePlanID:  stripePlanIdStr,
		PlanName:      planNameStr,
		MaxSession:    maxSessionInt,
		ListPrice:     listPriceWithDollar,
		ActualPrice:   actualPriceWithDollar,
		PriceID:       int(priceID.Int64),
		StripePriceID: stripePriceId.String,
		PerWeekLimit:  int(perWeekLimit.Int32),
	}

	return subscriptionPlanDetails, nil
}

// Function to check if a plan already exists in the slice
func findPlanIndex(plans []SubscriptionPlan, planID int) int {
	for i, plan := range plans {
		if plan.Id == planID {
			return i
		}
	}
	return -1 // Return -1 if the plan does not exist
}

// Returns the list of all available subscription plans. The plans are sorted by the subscription price.
func FetchSubscriptionPlans(priceID string) ([]SubscriptionPlan, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("FetchSubscriptionPlans: Error connecting to the database :", err)
		return nil, err
	}
	defer db.Close()

	// const query = `
	// 	SELECT
	// 		id, name, description, marketing_features, price, list_price, currency, max_session, frequency, stripe_id
	// 	FROM
	// 		subscription_plan
	// 	WHERE
	// 		status='active'
	// 	ORDER BY
	// 		price`

	query := `SELECT
		plan.id as plan_id, 
        plan.name, 
        plan.description, 
        plan.marketing_features, 
        price.price, 
        price.list_price, 
        plan.currency, 
        plan.max_session, 
		plan.per_week_limit,
        plan.frequency, 
        plan.stripe_id,
        price.id as price_id,
		price.default_price
			FROM
				subscription_plan as plan,
		        subscription_prices as price
			WHERE 
				plan.status='active'
		        AND plan.id=price.subscription_plan_id `
	if len(priceID) == 0 || priceID == "" {
		query += ` AND price.default_price = true `
	}

	// Add condition if priceid is provided
	if len(priceID) != 0 || priceID != "" {
		query += ` AND price.id = ` + priceID
	}

	query += ` ORDER BY price.price`

	rows, err := db.Query(query)
	if err != nil {
		log.Println("FetchSubscriptionPlans: Error querying fetch subscription plan :", err)
		return nil, err
	}
	defer rows.Close()

	// Fetch response data.
	var subscriptionPlans []SubscriptionPlan
	for rows.Next() {
		var (
			rawDescription     sql.NullString
			rawFeatures        sql.NullString
			planListPrice      sql.NullInt32
			planPriceID        sql.NullInt64
			planPrice          sql.NullInt32
			isPlanPriceDefault sql.NullBool
			planId             sql.NullInt64
			planName           sql.NullString
			planCurrency       sql.NullString
			planMaxSession     sql.NullInt16
			planFrequency      sql.NullString
			planProductId      sql.NullString
			perWeekLimit       sql.NullInt64
		)
		if err = rows.Scan(
			&planId,
			&planName,
			&rawDescription,
			&rawFeatures,
			&planPrice,
			&planListPrice,
			&planCurrency,
			&planMaxSession,
			&perWeekLimit,
			&planFrequency,
			&planProductId,
			&planPriceID,
			&isPlanPriceDefault,
		); err != nil {
			log.Println("FetchSubscriptionPlans: Failed while scanning with error: ", err)
			return nil, err
		}
		planIndex := findPlanIndex(subscriptionPlans, int(planId.Int64))
		if planIndex == -1 { // new plan
			var plan SubscriptionPlan
			plan.Id = int(planId.Int64)
			plan.Name = planName.String
			plan.Currency = planCurrency.String
			plan.MaxSession = int(planMaxSession.Int16)
			plan.Frequency = planFrequency.String
			plan.ProductId = planProductId.String
			plan.PerWeekLimit = int(perWeekLimit.Int64)

			// Processing "description" field...
			plan.Description = utility.SQLNullStringToString(rawDescription)
			// "marketing_features" are stored in database as a text. Converting them into an array of strings...

			if !rawFeatures.Valid { // This field can also be NULL; returning an empty array in this case.
				plan.MarketingFeatures = []string{}
			} else {
				for _, feature := range strings.Split(rawFeatures.String, "\n") {
					plan.MarketingFeatures = append(plan.MarketingFeatures, strings.TrimSpace(feature))
				}
			}

			// Processing "list_price"...
			// if !planListPrice.Valid {
			// 	plan.ListPrice = nil
			// } else {
			// 	listPrice := int(planListPrice.Int32)
			// 	plan.ListPrice = &listPrice
			// }
			subscriptionPlansPrice := SubscriptionPlanPrices{
				Id:        int(planPriceID.Int64),
				Price:     formatPriceToDisplay(int(planPrice.Int32), plan.Currency),
				ListPrice: formatPriceToDisplay(int(planListPrice.Int32), plan.Currency),
			}
			plan.Prices = append(plan.Prices, subscriptionPlansPrice)
			subscriptionPlans = append(subscriptionPlans, plan)

		} else {
			plan := subscriptionPlans[planIndex]
			subscriptionPlans[planIndex].Prices = append(subscriptionPlans[planIndex].Prices, SubscriptionPlanPrices{
				Id:        int(planPriceID.Int64),
				Price:     formatPriceToDisplay(int(planPrice.Int32), plan.Currency),
				ListPrice: formatPriceToDisplay(int(planListPrice.Int32), plan.Currency),
			})

		}
	}

	// Check for errors.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subscriptionPlans, nil
}

func StoreRawEvent(event stripe.Event) (int64, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("StoreRawEvent: Error connecting to the database :", err)
		return 0, err
	}
	defer db.Close()

	// Unmarshal the raw event data to get the customer ID
	var eventData map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &eventData); err != nil {
		log.Println(" StoreRawEvent: Error unmarshalling event data DOMNNNNNNNN:", err)
		return 0, err
	}

	// Extract the customer ID if it exists
	customerID := ""
	if customerData, ok := eventData["customer"].(string); ok {
		customerID = customerData
	}

	// Assuming event.Data is a struct and needs to be serialized
	data, err := json.Marshal(event.Data)
	if err != nil {
		log.Println("StoreRawEvent: Error marshalling event data:", err)
		return 0, err
	}

	// Single query to lookup user_id and insert event data
	query := `
        WITH user_lookup AS (
            SELECT user_id 
            FROM user2stripe 
            WHERE stripe_id = $1
        )
        INSERT INTO stripe_webhook_event (event_id, event_type, data, received_at, user_id, created_at) 
        VALUES ($2, $3, $4, NOW(), (SELECT user_id FROM user_lookup), NOW())
		RETURNING user_id
    `

	var userID sql.NullInt64
	err = db.QueryRow(query, customerID, event.ID, event.Type, data).Scan(&userID)
	if err != nil {
		log.Println("StoreRawEvent: Error executing query:", err)
		return 0, err
	}

	return userID.Int64, nil
}

func AddSuccessPaymentStatus(userIDFromRawEvent int64) (int, error) {
	if userIDFromRawEvent == 0 {
		log.Println("[NO_USERID_FOUND_FROM_RAW_EVENT]: AddSuccessPaymentStatus not inserting any data from")
		return 0, nil

	}
	db, err := config.GetDB2()
	if err != nil {
		log.Println("AddSuccessPaymentStatus: Error connecting to the database:", err)
		return 0, err
	}
	defer db.Close()

	var (
		userID                  sql.NullInt64
		subscriptionPlanID      sql.NullInt64
		subscriptionPlanPriceID sql.NullInt64
		stripeIntentID          sql.NullString
		planStripeID            sql.NullString
		planStripePriceID       sql.NullString
		purchasedAt             sql.NullTime
		// validUntil         sql.NullTime
		invoicePDF     sql.NullString
		customerID     sql.NullString
		amount         sql.NullString
		currency       sql.NullString
		paymentType    sql.NullString
		cardBrand      sql.NullString
		last4          sql.NullString
		expYear        sql.NullString
		expMonth       sql.NullString
		billingAddress sql.NullString
		customerEmail  sql.NullString
	)

	// Query to select all required values
	selectQuery := `
      SELECT
          swe.user_id,
          sp.subscription_plan_id as subscription_plan_id,
		  sp.id as subscription_plan_price_id,
          swe.data->'object'->>'payment_intent' AS stripe_intent_id,
          swe.data->'object'->'lines'->'data'->0->'plan'->>'product' AS plan_stripe_id,
		  swe.data->'object'->'lines'->'data'->0->'plan'->>'id' AS plan_stripe_price_id,
          to_timestamp((swe.data->'object'->'status_transitions'->>'paid_at')::double precision) AS purchased_at,
          swe.data->'object'->>'invoice_pdf' AS invoice_pdf,
		  swe2.data->'object'->>'receipt_email' AS customer_email,
          swe.data->'object'->>'customer' AS customer_id,
          swe.data->'object'->>'amount_paid' AS amount,
          swe.data->'object'->>'currency' AS currency,
		  swe2.data->'object'->'charges'->'data'->0->'payment_method_details'->>'type' AS payment_type,
		  swe2.data->'object'->'charges'->'data'->0->'payment_method_details'->'card'->>'brand' AS card_brand,
		  swe2.data->'object'->'charges'->'data'->0->'payment_method_details'->'card'->>'last4' AS last4,
		  swe2.data->'object'->'charges'->'data'->0->'payment_method_details'->'card'->>'exp_year' AS exp_year,
		  swe2.data->'object'->'charges'->'data'->0->'payment_method_details'->'card'->>'exp_month' AS exp_month,
		  CONCAT('(',
			COALESCE(swe2.data->'object'->'charges'->'data'->0->'billing_details'->'address'->>'city', 'NULL'), ', ',
			COALESCE(swe2.data->'object'->'charges'->'data'->0->'billing_details'->'address'->>'line1', 'NULL'), ', ',
			COALESCE(swe2.data->'object'->'charges'->'data'->0->'billing_details'->'address'->>'line2', 'NULL'), ', ',
			COALESCE(swe2.data->'object'->'charges'->'data'->0->'billing_details'->'address'->>'state', 'NULL'), ', ',
			COALESCE(swe2.data->'object'->'charges'->'data'->0->'billing_details'->'address'->>'country', 'NULL'), ', ',
			COALESCE(swe2.data->'object'->'charges'->'data'->0->'billing_details'->'address'->>'postal_code', 'NULL'),
			')') AS billing_address
      FROM
          stripe_webhook_event AS swe
	  JOIN stripe_webhook_event AS swe2
	  ON swe.data->'object'->>'payment_intent' = swe2.data->'object'->>'id'
      JOIN subscription_prices AS sp
      ON swe.data->'object'->'lines'->'data'->0->'plan'->>'id' = sp.stripe_price_id
      WHERE
          swe.data->'object'->>'status' = 'paid' 
		   AND swe2.event_type = 'payment_intent.succeeded' AND swe.user_id  = $1
		   ORDER BY swe.received_at DESC LIMIT 1`

	// Execute the select query and scan the values into variables
	err = db.QueryRow(selectQuery, userIDFromRawEvent).Scan(&userID, &subscriptionPlanID, &subscriptionPlanPriceID, &planStripePriceID, &stripeIntentID, &planStripeID, &purchasedAt, &invoicePDF, &customerEmail, &customerID, &amount,
		&currency, &paymentType, &cardBrand, &last4, &expYear, &expMonth, &billingAddress)
	if err != nil {
		log.Println("AddSuccessPaymentStatus: Error executing select query and scanning values:", err)
		return 0, err
	}

	//TODO: UPDATE INVOICE PDF WHEN INVOICE EVENT IS RECEIVED BASED ON  stripe_intent_id
	// Insert into payment_history using the scanned values
	insertPaymentHistoryQuery := `
      INSERT INTO payment_history (
          user_id,
          subscription_plan_id,
          purchased_date,
          stripe_intent_id,
          invoice_pdf_link,
          status,
          currency,
          amount,
		  payment_type,
		  brand,
		  card_last4,
		  exp_year,
		  exp_month,
		  billing_address,
          created_at
      )
      VALUES ($1, $2, $3, $4, $5, 'succeeded', $6, $7, $8, $9, $10, $11, $12, $13, NOW())`

	_, err = db.Exec(insertPaymentHistoryQuery, userID, subscriptionPlanID, purchasedAt, stripeIntentID, invoicePDF, currency, amount,
		paymentType, cardBrand, last4, expYear, expMonth, billingAddress)
	if err != nil {
		log.Println("AddSuccessPaymentStatus: Error inserting into payment_history:", err)
		return 0, err
	}

	subscription_plan_price_id := int(subscriptionPlanPriceID.Int64)
	subscription_plan_id := int(subscriptionPlanID.Int64)

	amountInInteger, _ := strconv.Atoi(amount.String) // Convert string to integer
	// Pass the integer value to formatPriceToDisplay
	formattedAmount := formatPriceToDisplay(amountInInteger, "USD")

	plan, err := FetchStripeSubscriptionPlanId(int(subscription_plan_price_id))
	if err != nil {
		log.Println("SendSuccessfulPaymentMessageOnSlack: failed to get subscription plan details with error: ", err)

	} else {

		studentDetails, err := students.GetStudentProfile(int(userID.Int64))
		if err != nil || studentDetails.ParentEmail != "" || len(studentDetails.ParentEmail) != 0 {
			go func() {
				// SendSuccessfulPaymentEmail(paymentDetails)
				paymentData := messageservices.PaymentDataAndPlanDetail{
					CardBrand:      cardBrand.String,
					CardExpYear:    expYear.String,
					CardExpMonth:   expMonth.String,
					CardLast4:      last4.String,
					InvoiceLink:    invoicePDF.String,
					PurchasedAt:    purchasedAt.Time.Format("2006-01-02"),
					AmountPaid:     formattedAmount,
					PlanName:       plan.PlanName,
					PlanMaxSession: plan.MaxSession,
					PerWeekLimit:   plan.PerWeekLimit,
					PaymentMethod:  paymentType.String,
				}
				studentData := messageservices.StudentData{
					ID:    userID.Int64,
					Email: studentDetails.ParentEmail,
				}
				messageservices.ScheduleMsgSuccessfullPaymentEmail(studentData, paymentData)
			}()
		}

		// Sends a Slack notification when a student puchase the plan successfully.
		go func() {
			//price, _ := strconv.Atoi(plan.ActualPrice)
			slack.SendSuccessfulPaymentMessageOnSlack(int(userID.Int64), plan.PlanName, plan.MaxSession, plan.ActualPrice)

		}()
	}
	// Insert into user2subscription using the scanned values
	// insertUserSubscriptionQuery := `
	//   INSERT INTO user2subscription (
	//       user_id,
	//       subscription_plan_id,
	//       purchased_date,
	//       status,
	//       created_at
	//   )
	//   VALUES ($1, $2, $3, 'active', NOW())`

	// _, err = db.Exec(insertUserSubscriptionQuery, userID, subscriptionPlanID, purchasedAt)
	// if err != nil {
	// 	log.Println("AddSuccessPaymentStatus: Error inserting into user2subscription:", err)
	// 	return err
	// }

	user_id := int(userID.Int64)
	log.Println("Planid ", subscription_plan_id)
	log.Println("User ", user_id)

	err = InsertInactiveSubscription(subscription_plan_id, subscription_plan_price_id, user_id)
	if err != nil {
		log.Println("AddSuccessPaymentStatus: Failed to insert the InsertInactiveSubscription with :", err)

	}

	err = RenewSubscription(purchasedAt.Time, subscription_plan_id, user_id)
	if err != nil {
		log.Println("AddSuccessPaymentStatus: Failed to update the UpdateSubscriptionStatus with :", err)
		return 0, err
	}

	// session.CreditSessionsForUser(user_id)

	return user_id, nil
}

func RenewSubscription(purchasedAt time.Time, planId, userID int) error {
	log.Println("RenewSubscription")
	db, err := config.GetDB2()
	if err != nil {
		log.Println("RenewSubscription: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	startOfDay := time.Date(purchasedAt.Year(), purchasedAt.Month(), purchasedAt.Day(), 0, 0, 0, 0, purchasedAt.Location())

	validUntil := purchasedAt.AddDate(0, 0, 27)
	validUntilEndOFDay := time.Date(validUntil.Year(), validUntil.Month(), validUntil.Day(), 23, 59, 59, 0, validUntil.Location())

	query := `UPDATE 
	                public.user2subscription
	            SET status = 'active',
	            	updated_at = NOW(),
					start_date = $1,
					valid_until = $2
	            WHERE user_id = $3 
	              AND subscription_plan_id = $4`

	_, err = db.Exec(query, startOfDay, validUntilEndOFDay, userID, planId)
	if err != nil {
		log.Println("RenewSubscription: Failed to update user2subscription with error :", err)
		return err
	}
	return nil
}

func UpdateSubscriptionStatus(status string, planId, userID int) error {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("UpdateSubscriptionStatus: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	query := `UPDATE 
	                user2subscription
	            SET status = $1,
	            	updated_at = NOW()
	            WHERE user_id = $2 
	              AND subscription_plan_id = $3`

	_, err = db.Exec(query, status, userID, planId)
	if err != nil {
		log.Println("UpdateSubscriptionStatus: Error inserting into user2subscription:", err)
		return err
	}
	return nil
}

func InsertInactiveSubscription(planId, priceID, userID int) error {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("InsertInactiveSubscription: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	insertUserSubscriptionQuery := `
      INSERT INTO user2subscription (
          user_id,
          subscription_plan_id,
		  price_id,
          purchased_date,
          status,
          created_at
      )
      VALUES ($1, $2,$3, NOW(), 'inactive', NOW())`

	_, err = db.Exec(insertUserSubscriptionQuery, userID, planId, priceID)
	if err != nil {
		log.Println("InsertInactiveSubscription: Error inserting into user2subscription:", err)
		return err
	}

	return nil
}

func UpdateUserSusbcriptionStatus(userID string, validUntil time.Time) error {

	db, err := config.GetDB2()
	if err != nil {
		log.Println("UpdateUserSusbcriptionStatus: Error connecting to the database:", err)
		return err
	}
	defer db.Close()
	var id sql.NullInt64
	query := `
			UPDATE 	public.user2subscription
	        SET 
			 valid_until = $1 ,
			 start_date =  NOW(),
			 status = 'active'
			WHERE
			   user_id  = $2
			RETURNING id`

	err = db.QueryRow(query, validUntil, userID).Scan(&id)
	if err != nil {
		log.Println("UpdateUserSusbcriptionStatus: Failed to execute the query with the database:", err)
		return err
	}
	return nil
}

func AddFailedPaymentStatus(event stripe.Event) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("AddFailedPaymentStatus: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	var eventData map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &eventData); err != nil {
		log.Println("AddFailedPaymentStatus: Error unmarshalling event data:", err)
		return err
	}

	fmt.Println("******TRTTRRRTRTRTRRTRTRTRTTR", eventData)

	customerID := eventData["customer"].(string)
	paymentIntentID := eventData["id"].(string)
	errorMessage := ""
	if lastPaymentError, ok := eventData["last_payment_error"].(map[string]interface{}); ok {
		errorMessage = lastPaymentError["message"].(string)
	}

	query := `
        WITH user_info AS (
            SELECT user_id
            FROM user2stripe
            WHERE stripe_id = $1
        )
        INSERT INTO user2subscription (
            user_id,
            stripe_intent_id,
            payment_status,
            error_message,
            updated_at
        )
        SELECT
            user_id,
            $2,
            'failed',
            $3,
            NOW()
        FROM user_info`

	_, err = db.Exec(query, customerID, paymentIntentID, errorMessage)
	if err != nil {
		log.Println("AddFailedPaymentStatus: Error executing query:", err)
		return err
	}

	return nil
}
func IsTrialPeriodInvoice(event stripe.Event, userID int64) bool {
	// Extract the amount paid from the invoice event data.
	paidPrice := event.Data.Object["amount_paid"]
	log.Println("invoice", event.Data.Object["amount_paid"])
	log.Println("customer", event.Data.Object["customer"])

	// Check if the invoice indicates a trial period (i.e., amount paid is 0.0).
	if paidPrice.(float64) == 0.0 {
		//<<<<<Updating subscription trial end date is only for user's who singed up using ppc FLOW..>>>>>
		// Retrieve the student's session preferences (day/time).
		sessionDayTime, err := sessionmodel.GetAllSessionPreferences(int(userID))
		if err == nil && len(sessionDayTime) != 0 {
			// If session preferences are available, proceed with further operations.
			log.Println("BookSessionAfterPaymentSetup: Failed to get student session preferences with error:", err)

			var firstPaidSessionEndDateAndTime time.Time

			// Update the subscription start date for the user.
			firstPaidSessionEndDateAndTime, err = sessionmodel.UpdateSubscriptionStartDateForUser(int(userID))
			if err != nil {
				log.Println("[ERROR] BookRegularSession: Failed to UpdateSubscriptionStartDateForUser with err ", err)
			}

			// Load the timezone from the session preferences.
			loc, _ := time.LoadLocation(sessionDayTime[0].Timezone)

			// Update the trial end date of the subscription.
			err = common.UpdateSubscriptionTrialEndToDate(int(userID), firstPaidSessionEndDateAndTime, loc)
			if err != nil {
				log.Println("[ERROR]: IsTrialPeriodInvoice -> UpdateSubscriptionTrialEndToDate: Failed to update trial end by date")
			}
		}

		// Perform a background task using a goroutine.
		go func() {
			// Fetch the active plan details for the user.
			plan, err := GetActivePlanDetails(int(userID))
			if err != nil {
				log.Println("no active plan")
			} else {
				// Send a notification to Slack about the payment setup.
				slack.SendPaymentSetupMessageOnSlack(userID, plan.Name, plan.MaxSession, plan.Price)
			}
		}()

		// Log that the invoice corresponds to a trial period.
		log.Println("Trial Invoice")
		return true
	}
	return false
}

type SubscriptionCancel struct {
	StripeCustomerID string
	ProductID        string
	CancelledAt      int64
}

func ChangeSubscription(event stripe.Event) error {
	log.Println("ChangeSubscription")
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		log.Printf("Error parsing subscription updated event: %v", err)
		return err
	}
	customerID := subscription.Customer.ID

	sCancelRequest := SubscriptionCancel{
		StripeCustomerID: customerID,
		ProductID:        subscription.Plan.Product.ID,
		CancelledAt:      subscription.CancelAt,
	}

	// Check if the subscription is canceled by verifying that CancelAt is set (non-zero) and CancelAtPeriodEnd is true.
	if subscription.CancelAt != 0 && subscription.CancelAtPeriodEnd {
		log.Println("Subscription is scheduled for cancellation. Marking it as canceled in the database.")
		if err := MarkSubscriptionCancelled(sCancelRequest); err != nil {
			log.Printf("[ERROR] ChangeSubscription --> MarkSubscriptionCancelled: Failed to mark subscription as canceled in the database: %v", err)
			return err
		}
		return nil // Return early as no further actions are needed if the subscription is canceled.
	}

	previousAttributes := event.Data.PreviousAttributes
	// Check if plan or price has changed
	if previousAttributes != nil {
		if previousAttributes["items"] == nil {
			return errors.New("No plan found to change")
		}
		previousItems := previousAttributes["items"].(map[string]interface{})["data"].([]interface{})
		currentItems := subscription.Items.Data

		// Compare old and new plans/products
		for i, previousItem := range previousItems {
			previousPlan := previousItem.(map[string]interface{})["plan"].(map[string]interface{})
			currentPlan := currentItems[i].Plan
			previousPlanProductID := previousPlan["product"]

			log.Println("Previous Plan id", previousPlanProductID)
			log.Println("Previous Plan id", currentPlan.Product.ID)
			if previousPlanProductID != currentPlan.Product.ID {
				log.Printf("Product or plan has changed from %v to %v", previousPlanProductID, currentPlan.ID)
				err := ChangeSubscriptionPlanProduct(customerID, previousPlanProductID.(string), currentPlan.ID)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
	return errors.New("No plan found to change")
}

// MarkSubscriptionCancelled handles the cancellation process for a user subscription in the database.
// It performs the following steps:
// 1. Connects to the database and retrieves necessary subscription and user information based on the provided Stripe IDs.
// 2. Marks the subscription as "cancelled" in the `user2subscription` table, linking the cancellation to the subscription and user IDs.
// 3. Converts the UNIX cancellation timestamp into a human-readable date and time format.
// 4. Updates the `session` table to set sessions as "not credited" if their start date is after the cancellation date.
// 5. Prepares cancellation information and schedules an email to notify the user of the subscription cancellation details.
func MarkSubscriptionCancelled(sCancel SubscriptionCancel) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("MarkSubscriptionCancel: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	var (
		planName, parentEmail, parentName sql.NullString
		userID                            sql.NullInt64
	)

	// First query to mark the subscription as cancelled
	query1 := `
        WITH user_info AS (
            SELECT u2s.user_id, s.parent_email, s.name
            FROM user2stripe u2s
            JOIN public.student s ON u2s.user_id = s.user_id
            WHERE u2s.stripe_id = $1
            LIMIT 1
        ),
        subscription_info AS (
            SELECT id, name
            FROM public.subscription_plan
            WHERE stripe_id = $2
        )
        UPDATE public.user2subscription
        SET status = 'cancelled'
        WHERE subscription_plan_id = (SELECT id FROM subscription_info)
        AND user_id = (SELECT user_id FROM user_info)
        RETURNING 
            (SELECT user_id FROM user_info), 
            (SELECT parent_email FROM user_info), 
            (SELECT name FROM user_info), 
            (SELECT name FROM subscription_info) `

	err = db.QueryRow(query1, sCancel.StripeCustomerID, sCancel.ProductID).Scan(&userID, &parentEmail, &parentName, &planName)
	if err != nil {
		log.Println("MarkSubscriptionCancel: Error executing query1(Mark subscription Cancel):", err)
		return err
	}

	//CONVERT the UNIX time to Human Readable form.
	cancelTime := time.Unix(sCancel.CancelledAt, 0)
	humanReadableTime := cancelTime.Format("15:04:05")   // Time in HH:MM:SS format
	humanReadableDate := cancelTime.Format("2006-01-02") // Date in YYYY-MM-DD format

	subscriptionCancelData := messageservices.CancelSubscriptionData{
		CancellationDate: humanReadableDate,
		CancellationTime: humanReadableTime,
		PlanName:         planName.String,
		UserName:         parentName.String,
	}
	studentData := messageservices.StudentData{
		ID:    userID.Int64,
		Email: parentEmail.String,
	}

	messageservices.ScheduleSubscriptionCancellationEmail(studentData, subscriptionCancelData)

	return nil
}

func ChangeSubscriptionPlanProduct(customerID string, previousProduct string, newProductPrice string) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("ChangeSubscriptionPlanProduct: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	query := `WITH user_info AS (
				    SELECT user_id
				    FROM user2stripe
				    WHERE stripe_id = $1
				),
				previous_plan_info AS (
				    SELECT id
				    FROM subscription_plan 
				    WHERE stripe_id = $2
				),
				new_plan_info AS (
				    SELECT subscription_plan_id,id
				    FROM subscription_prices 
				    WHERE stripe_price_id = $3
				)
				UPDATE user2subscription as u2s 
				SET subscription_plan_id=(SELECT subscription_plan_id FROM new_plan_info),price_id=(SELECT id FROM new_plan_info), status = 'active', updated_at = NOW() 
				WHERE u2s.user_id = (SELECT user_id FROM user_info);`

	_, err = db.Exec(query, customerID, previousProduct, newProductPrice)
	if err != nil {
		log.Println("ChangeSubscriptionPlanProduct: Error executing query:", err)
		return err
	}

	return nil
}

// Formats a price to display.
//
// Parameters:
//   - amount: The price amount, in smallest units possible. For USD it's in cents.
//   - currency: Three-letter ISO currency code, in lowercase.
//     Must be a supported currency (the only currently supported currency is USD).
//
// Returns a price formatted to display, like $119.99
func formatPriceToDisplay(amount int, currency string) string {
	// TODO: Implement formatting for other currencies too.
	if currency != "USD" {
		// Unknown currency; output it like: 1234 XXX
		log.Printf("formatPriceToDisplay: unsupported currency: \"%s\"", currency)
		return fmt.Sprintf("%d %s", amount, currency)
	}

	// Format USD price.
	dollars, cents := amount/100, amount%100
	if cents == 0 {
		return fmt.Sprintf("$%d", dollars)
	} else {
		return fmt.Sprintf("$%d.%02d", dollars, cents)
	}
}

// CreateSubscriptionWithTrial will create subsciption with trial period(28 days).
func CreateSubscriptionWithTrial(userId, planId, priceID int) error {
	customer, err := FetchCustomerInfo(userId)
	if err != nil {
		log.Printf("CreateSubscriptionWithTrial: Error fetching customer info for userId %d: %v", userId, err)
		return err
	}
	fmt.Println("CUSTOMER ID:", customer)

	// Create Stripe customer if it doesn't exist.
	// if customer.StripeId == "" {
	// 	stripeId, err := services.CreateCustomer(customer.Phone)
	// 	if err != nil {
	// 		log.Printf("CreateSubscriptionWithTrial: Error creating Stripe customer for userId %d: %v", userId, err)
	// 		return err
	// 	}
	// 	err = StoreStripeCustomerId(userId, stripeId)
	// 	if err != nil {
	// 		log.Printf("CreateSubscriptionWithTrial: Error storing Stripe customer ID for userId %d: %v", userId, err)
	// 		return err
	// 	}
	// 	customer.StripeId = stripeId
	// }

	// Get subscrition plan basd on price
	plan, err := FetchStripeSubscriptionPlanId(priceID)
	if err != nil {
		log.Printf("CreateSubscriptionWithTrial: Error fetching Stripe subscription plan ID for planId %s: %v", plan.StripePlanID, err)
		return err
	}
	fmt.Println("Fetched Stripe subscription plan ID:", plan.StripePlanID)

	// stripePlan, err := services.GetSubscriptionPlan(plan.StripePlanID)
	// if err != nil {
	// 	log.Printf("CreateSubscriptionWithTrial: Error getting subscription plan details for stripePlanId %s: %v", plan.StripePlanID, err)
	// 	return err
	// }
	// fmt.Println("Fetched Stripe subscription plan details:", stripePlan)

	subscriptionId, err := services.CreateSubscription(customer.StripeId, plan.StripePriceID)
	if err != nil {
		log.Printf("CreateSubscriptionWithTrial: Error creating subscription for Stripe customer ID %s: %v", customer.StripeId, err)
		return err
	}

	status := utility.SubscriptionStatusActive
	err = UpdateSubscriptionStatus(status, planId, userId)
	if err != nil {
		log.Println("CreateSubscriptionWithTrial: UpdateSubscriptionStatus :Failed to update the UpdateSubscriptionStatus with :", err)
		return err
	}
	fmt.Println("Created subscription ID:", subscriptionId)

	// Fetch new subscription's client secret.
	// clientSecret, err := services.GetClientSecretFromSubscription(subscriptionId)
	// if err != nil {
	// 	log.Printf("Create: Error fetching client secret for subscription ID %s: %v", subscriptionId, err)
	// 	return "", "", 0, "", "", err
	// }
	// listPrice := formatPriceToDisplay(amount, "USD")
	// actualPrice := formatPriceToDisplay(price, "USD")

	return nil
}

// Returns the list of all available subscription plans. The plans are sorted by the subscription price.
func ListPlans(priceID string) ([]SubscriptionPlan, error) {
	// Get subscription plans from our database.
	subscriptionPlans, err := FetchSubscriptionPlans(priceID)
	if err != nil {
		return nil, err
	}

	// // Convert the plans to output format for JSON serialization.
	// var subscriptionPlans []SubscriptionPlans
	// for _, plan := range plansFromDb {
	// 	var listPricePtr *string = nil
	// 	if plan.ListPrice != nil {
	// 		listPrice := formatPriceToDisplay(*plan.ListPrice, plan.Currency)
	// 		listPricePtr = &listPrice
	// 	}
	// 	subscriptionPlans = append(subscriptionPlans, SubscriptionPlans{
	// 		plan.Id,
	// 		plan.Name,
	// 		plan.Description,
	// 		plan.MarketingFeatures,
	// 		formatPriceToDisplay(plan.Price, plan.Currency),
	// 		listPricePtr,
	// 		plan.ProductId,
	// 		plan.MaxSession,
	// 		plan.Frequency,
	// 	})
	// }

	return subscriptionPlans, nil
}

func SendSuccessfulPaymentEmail(paymentDetail ActivePlan) {

	pData := map[string]interface{}{
		"card_brand":   paymentDetail.PaymentDetails.CardBrand,
		"card_last4":   paymentDetail.PaymentDetails.CardLast4,
		"exp_year":     paymentDetail.PaymentDetails.CardExpireYear,
		"exp_month":    paymentDetail.PaymentDetails.CardExpireMonth,
		"invoice_link": paymentDetail.InvoicePdfLink,
		"purchased_at": paymentDetail.PurchasedDate,
	}

	// Prepare the request data for sending the email
	requestData := services.DynamicEmailDetailsForBrevo{
		To:          []string{paymentDetail.CustomerEmail},
		DynamicData: pData,
		TemplateID:  constants.SuccessfulPaymentEmail,
	}

	// Send the reminder email using the dynamic template
	services.ComposeDynamicTemplateEmailsForBrevoService(requestData)
}

type ActivePlan struct {
	MaxSession     int            `json:"max_session"`
	Name           string         `json:"name"`
	PurchasedDate  string         `json:"purchased_date"`
	ValidUntil     sql.NullString `json:"valid_until"`
	Price          string         `json:"price"`
	InvoicePdfLink string         `json:"invoice_pdf_link"`
	CustomerEmail  string         `json:"cutomer_email,omitempty"`
	PaymentDetails PaymentDetails `json:"payment_details"`
	PerWeekLimit   int            `json:"per_week_limit"`
	StartDate      sql.NullString `json:"start_date"`
}

type PaymentDetails struct {
	CardExpireYear  string `json:"exp_year"`
	CardExpireMonth string `json:"exp_month"`
	CardLast4       string `json:"last4digit"`
	CardBrand       string `json:"card_brand"`
	BillingAddress  string `json:"billing_address"`
}

func GetActivePlanDetails(userID int) (ActivePlan, error) {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetActivePlan: Error connecting to the database:", err)
		return ActivePlan{}, err
	}
	defer db.Close()

	var (
		plan     ActivePlan
		rawPrice int
	)

	query := `SELECT 
	            sp.name,
			    sp.max_session,
				sprice.price,
			    u2s.purchased_date,
				u2s.start_date,
			    u2s.valid_until,
				sp.per_week_limit

			FROM
			    subscription_plan AS sp 
			JOIN 
			    user2subscription AS u2s 
			ON 
			    u2s.subscription_plan_id = sp.id
			JOIN 
			    subscription_prices AS sprice 
			ON 
			    u2s.price_id = sprice.id
			WHERE 
			    u2s.user_id = $1 AND u2s.status = 'active' 
			ORDER BY 
			    u2s.updated_at DESC 
			LIMIT 1`

	err = db.QueryRow(query, userID).Scan(
		&plan.Name,
		&plan.MaxSession,
		&rawPrice,
		&plan.PurchasedDate,
		&plan.StartDate,
		&plan.ValidUntil,
		&plan.PerWeekLimit,
	)
	if err != nil {
		log.Println("GetActivePlan: Error executing query:", err)
		return ActivePlan{}, err
	}
	plan.Price = formatPriceToDisplay(rawPrice, "USD")

	return plan, nil
}

// MarkTrialEnd adjusts the billing cycle for the specified user IDs by fetching
// their subscription details and updating their billing cycles accordingly. It returns
// a map containing the user ID, subscription ID, current period start, and current period end
// for each updated subscription.

func MarkTrialEnd(userIDs []int) (map[string]interface{}, error) {
	// Retrieve the subscription IDs for the given user IDs.
	subscriptionMap, err := common.GetSubscriptionAndPaymentDetails(userIDs)
	if err != nil {
		log.Println("MarkTrialEnd: Failed to retrieve subscription details:", err)
		return nil, err
	}
	log.Println("MarkTrialEnd: Retrieved subscription details:", subscriptionMap)

	// Initialize a map to store the updated subscription details for each user.
	result := make(map[string]interface{})

	// Iterate through the subscription map to update the billing cycle for each subscription.
	for userID, subscriptionID := range subscriptionMap {
		if subscriptionID == "" {
			log.Println("MarkTrialEnd: No subscription found for user ID:", userID)
			continue
		}

		// Unpause the subscription before updating the billing cycle (commented out for now).
		// err := services.UnPauseTheSubscription(subscriptionID)
		// if err != nil {
		// 	log.Println("MarkTrialEnd: Error unpausing subscription for ID:", subscriptionID, "error:", err)
		// 	continue
		// }

		// Call the service to mark the trial end and update the billing cycle on Stripe.
		subscription, err := services.MarkTrialEndOnStripe(subscriptionID)
		if err != nil {
			log.Println("MarkTrialEnd: Failed to MarkTrialEndOnStripe for subscription ID:", subscriptionID, "error:", err)
			continue // Move to the next subscription if there's an error.
		}

		// Store the updated subscription details in the result map.
		result = map[string]interface{}{
			"user_id":              userID,
			"subscription_id":      subscriptionID,
			"current_period_start": subscription.CurrentPeriodStart,
			"current_period_end":   subscription.CurrentPeriodEnd,
		}

		// Log the success for the updated subscription.
		log.Printf("MarkTrialEnd: Successfully MarkTrialEndOnStripe for user ID: %s\nSubscription ID: %s\nCurrent Period Start: %v\nCurrent Period End: %v\n",
			userID, subscriptionID, subscription.CurrentPeriodStart, subscription.CurrentPeriodEnd)
	}
	return result, nil
}

func GetRecentCustomerIDandPaymentMethodID() (customerID, paymentMethodID string, err error) {
	// Establish a connection to the database.
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetRecentCustomerIDandPaymentMethodID: Error connecting to the database:", err)
		return "", "", err
	}
	defer db.Close()

	// Construct the query to fetch the most recent customer ID and payment method ID.
	query := `
		SELECT 
			swe.data->'object'->>'customer' AS customer_id,
			swe.data->'object'->>'payment_method' AS payment_method_id
		FROM 
			stripe_webhook_event swe
		JOIN 
			user2stripe AS u2s ON swe.data->'object'->>'customer' = u2s.stripe_id
		WHERE 
			swe.event_type = 'setup_intent.succeeded'
		ORDER BY 
			swe.received_at DESC 
		LIMIT 1`

	// Execute the query and scan the result into customerID and paymentMethodID.
	err = db.QueryRow(query).Scan(&customerID, &paymentMethodID)
	if err != nil {
		log.Println("GetRecentCustomerIDandPaymentMethodID: Error executing query:", err)
		return "", "", fmt.Errorf("error retrieving customer and payment method: %v", err)
	}

	return customerID, paymentMethodID, nil
}

func GetPlanIdAndUserID(customerID string) (userID, planID, priceID int, err error) {
	// Get the database connection
	db, err := config.GetDB2()
	if err != nil {
		log.Println("GetPlanIdAndUserID: Error connecting to the database:", err)
		return 0, 0, 0, err
	}
	defer db.Close()

	// Define the query
	query := `SELECT 
                    u2s.user_id, 
                    u2s.subscription_plan_id,
					u2s.price_id
                FROM 
                    user2subscription AS u2s 
                JOIN 
                    user2stripe AS us 
                ON 
                    u2s.user_id = us.user_id 
                WHERE 
                    us.stripe_id = $1`

	// Execute the query and scan the result into userID and planID
	err = db.QueryRow(query, customerID).Scan(&userID, &planID, &priceID)
	if err != nil {
		log.Println("GetPlanIdAndUserID: Error executing query:", err)
		return 0, 0, 0, err
	}

	// Return the userID and planID
	return userID, planID, priceID, nil
}

type ProductPriceRequest struct {
	PriceID        string
	ProductID      string
	DefaultPriceID string
	UnitAmount     float64
}

// parseProductPriceEvents function is designed to extract relevant fields from Stripe events related to pricing and products.
// It handles the following event types:
// - "price.created": When a new price is created in Stripe.
// - "price.updated": When a price is updated in Stripe.
// - "price.deleted": When a price is deleted in Stripe.
// - "product.updated": When a product is updated in Stripe.
//
// The function extracts the following fields:
// - `priceID`: The Stripe price ID (`id` field from the event).
// - `productID`: The Stripe product ID (only applicable for "price.created" and "price.deleted").
// - `unitAmount`: The price amount (only applicable for "price.created" and "price.deleted").
//
// NOTE :- - `defaultPriceID`: The default price ID (only applicable for "product.updated").

func parseProductPriceEvents(event stripe.Event) (ProductPriceRequest, error) {
	// Declare variables
	var (
		defaultPriceID string
		priceID        string
		productID      string
		unitAmount     float64
	)

	// Safely get the priceID from the event data
	priceID = event.Data.Object["id"].(string)

	// Handle product.updated case
	if event.Type == "product.updated" {
		if id, ok := event.Data.Object["default_price"]; ok {
			defaultPriceID = id.(string) // Safe type assertion
		}
	} else {
		// Handle other event types (like price.created, price.deleted)
		productID = event.Data.Object["product"].(string)
		unitAmount = event.Data.Object["unit_amount"].(float64)
	}
	details := ProductPriceRequest{
		PriceID:        priceID,
		ProductID:      productID,
		UnitAmount:     unitAmount,
		DefaultPriceID: defaultPriceID,
	}
	// Return the values, using an empty string for defaultPriceID if not set
	return details, nil
}

// Function to handle price events and database operations
func HandlePriceEvents(event stripe.Event, dbOperation func(ProductPriceRequest) error) {
	// Parse event data
	detailsToBEUpdated, err := parseProductPriceEvents(event)
	if err != nil {
		log.Println("HandlePriceEvents :Failed to parse event data:", err)
		return
	}
	if len(detailsToBEUpdated.PriceID) == 0 && len(detailsToBEUpdated.ProductID) == 0 && len(detailsToBEUpdated.DefaultPriceID) == 0 && detailsToBEUpdated.UnitAmount == 0.0 {
		log.Println("HandlePriceEvents: DO NOT HAVE ANY DATA in Feilds to update product price")
		return
	}

	// Perform the database operation
	err = dbOperation(detailsToBEUpdated)
	if err != nil {
		log.Println("HandlePriceEvents: Failed to execute DB operation FOR PRODUCT PRICES:", err)
		return
	}
}

// DB functions to handle product prices
func InsertNewProductPrice(details ProductPriceRequest) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("InsertNewProductPrice: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	// Define the SQL query using a CTE to fetch id and list_price directly
	query := `
		WITH plan_info AS (
			SELECT id, list_price
			FROM subscription_plan 
			WHERE stripe_id = $1
		)
		INSERT INTO subscription_prices(
			stripe_product_id, 
			subscription_plan_id, 
			stripe_price_id, 
			price, 
			list_price, 
			default_price
		)
		VALUES (
			$1, 
			(SELECT id FROM plan_info), 
			$2, 
			$3, 
			(SELECT list_price FROM plan_info), 
			FALSE)
	`

	// Execute the query with the provided parameters
	_, err = db.Exec(query, details.ProductID, details.PriceID, details.UnitAmount)
	if err != nil {
		log.Println("InsertNewProductPrice: Error executing query:", err)
		return err
	}

	return nil
}

func UpdateProductPrice(details ProductPriceRequest) error {
	// Connect to the database
	db, err := config.GetDB2()
	if err != nil {
		log.Println("UpdateProductPrice: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	// Define the SQL query for updating the price
	query := `
		UPDATE subscription_prices 
		SET price = $3 , updated_at = NOW()
		WHERE stripe_price_id = $1 
		AND stripe_product_id = $2 
	`

	// Execute the update query with the provided parameters
	_, err = db.Exec(query, details.PriceID, details.ProductID, details.UnitAmount)
	if err != nil {
		log.Println("UpdateProductPrice: Error executing query:", err)
		return err
	}

	return nil
}

func DeleteProductPrice(details ProductPriceRequest) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("DeleteProductPrice: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	// Define the SQL query for deleting a product price
	query := `
		DELETE FROM subscription_prices 
		WHERE stripe_price_id = $1 
		AND stripe_product_id = $2
		AND price = $3
	`

	// Execute the delete query with the provided parameters
	_, err = db.Exec(query, details.PriceID, details.ProductID, details.UnitAmount)
	if err != nil {
		log.Println("DeleteProductPrice: Error executing query:", err)
		return err
	}

	return nil
}

// MarkPriceIdDefault is responsible for marking a given price as the default price for a product.
//
// Purpose:
// - This function unmarks any existing default prices for a specific product and then
//   marks the new price as the default.
//
// Inputs:
// - details (ProductPriceRequest struct): Contains the `ProductID` and `PriceID` for the price being marked as default.
//
// Outputs:
// - error: If successful, the function returns nil; otherwise, it returns an error describing what went wrong.
//
// Steps:
// 1. Connect to the database.
// 2. Unmark all current default prices for the specified product.
// 3. Mark the new price as the default.
// 4. Return success or handle any errors.

func MarkPriceIdDefault(details ProductPriceRequest) error {
	// Connect to the database
	db, err := config.GetDB2()
	if err != nil {
		log.Println(" MarkPriceIdDefault: Error connecting to the database:", err)
		return err
	}
	defer db.Close()

	// Explanation:
	// - In the event "product.updated", the `id` field refers to the product ID, while the `default_price` field refers to the price ID.
	// - Therefore, we need to interchange the variables when updating the default price for the product.
	// - The `productID` should come from `details.PriceID` (which refers to the Stripe product ID).
	// - The `priceID` should come from `details.DefaultPriceID` (which refers to the Stripe price ID).
	productID := details.PriceID
	priceID := details.DefaultPriceID

	// Step 1: Unmark all other prices as default for this product
	unmarkQuery := `
		UPDATE subscription_prices 
		SET default_price = FALSE, updated_at = NOW() 
		WHERE stripe_product_id = $1 AND default_price = TRUE
	`
	_, err = db.Exec(unmarkQuery, productID)
	if err != nil {
		log.Println("MarkPriceIdDefault: Error executing unmark query:", err)
		return err
	}

	// Step 2: Mark the specified price as default
	markQuery := `
		UPDATE subscription_prices 
		SET default_price = TRUE, updated_at = NOW() 
		WHERE stripe_price_id = $1 AND stripe_product_id = $2
	`
	_, err = db.Exec(markQuery, priceID, productID)
	if err != nil {
		log.Println("MarkPriceIdDefault: Error executing mark query:", err)
		return err
	}

	return nil

}

// UpdateUser2StripeUpdatedAt updates the `updated_at` field for a specific customer ID
func UpdateUser2StripeUpdatedAt(customerID string) error {
	db, err := config.GetDB2()
	if err != nil {
		log.Println("UpdateUser2StripeUpdatedAt: Error connecting to the database:", err)
		return err
	}
	defer db.Close()
	query := `
		UPDATE user2stripe
		SET updated_at = NOW()
		WHERE stripe_id = $1
	`

	// Execute the query with the current time and customerID
	_, err = db.Exec(query, customerID)
	if err != nil {
		return err
	}

	return nil
}
