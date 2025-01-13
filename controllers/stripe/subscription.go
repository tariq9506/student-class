package stripe

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	apierror "tutree/student-apis/apiError"
	"tutree/student-apis/controllers"
	"tutree/student-apis/controllers/session"
	"tutree/student-apis/models"
	"tutree/student-apis/models/stripe"
	students "tutree/student-apis/models/student"
	"tutree/student-apis/services"

	"github.com/gin-gonic/gin"
)

// This struct is here only in purpose of generating a Swagger documentation for an error responses from the API.
// I believe we should put it to apierror package
//
//	@Description	Error response.
type ErrorResponse struct {
	// API operation status.
	Status string `json:"status" enums:"failed"`

	// Error message.
	Message string `json:"message" example:"Something went wrong"`

	// Errors code.
	Code int `json:"code" example:"1000"`

	// An optional field containing an additional error information. May be omitted.
	DetailedMsg string `json:"detailed_msg,omitempty" example:"pq: database \"happy_tree_friends\" does not exist"`
}

func BuySubscription(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		log.Println("BuySubscription: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	//SUBSSCRIPTION PRICE ID
	planId, err := strconv.Atoi(c.PostForm("plan_id"))
	if err != nil || planId <= 0 {
		log.Println("BuySubscription: failed to convert subscription plan id into a valid number.")
		controllers.HandleJSONErrorResponse(apierror.InvalidPlanID, err, c)
		return
	}

	clientSecret, err := GenerateStripePaymentIntent(userID.(int))
	if err != nil {
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}
	// Get plan based on price id
	plan, err := stripe.FetchStripeSubscriptionPlanId(planId)
	if err != nil {
		log.Printf("BuySubscription: Error fetching Stripe subscription plan ID for planId %d: %v", planId, err)
		return
	}

	CreateInactiveSubscriptionEntry(plan.PlanID, plan.PriceID, userID.(int))

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"secret":         clientSecret,
		"total_classes":  plan.MaxSession,
		"plan_name":      plan.PlanName,
		"amount":         plan.ListPrice,
		"total_amount":   plan.ActualPrice,
		"per_week_limit": plan.PerWeekLimit,
	})

}

func SetPaymentMethod(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		log.Println("BuySubscription: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	clientSecret, err := GenerateStripePaymentIntent(userID.(int))
	if err != nil {
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"secret": clientSecret,
	})

}

func GenerateStripePaymentIntent(userID int) (string, error) {
	// Fetch customer information based on userID.
	customer, err := stripe.FetchCustomerInfo(userID)
	if err != nil {
		log.Printf("GenerateStripePaymentIntent: Error fetching customer info for userId %d: %v", userID, err)
		return "", err
	}

	// Create Stripe customer if it doesn't exist.
	if customer.StripeId == "" {
		stripeId, err := services.CreateCustomer(customer.Phone)
		if err != nil {
			log.Printf("GenerateStripePaymentIntent: Error creating Stripe customer for userId %d: %v", userID, err)
			return "", err
		}
		err = models.StoreStripeCustomerId(userID, stripeId)
		if err != nil {
			log.Printf("GenerateStripePaymentIntent: Error storing Stripe customer ID for userId %d: %v", userID, err)
			return "", err
		}
		customer.StripeId = stripeId
	}

	// Check if the customer object is valid and has a Stripe ID.
	if customer.StripeId == "" {
		errMsg := fmt.Sprintf("GenerateStripePaymentIntent: No valid Stripe ID found for userId %d", userID)
		log.Println(errMsg)
		return "", errors.New(errMsg)
	}

	// Proceed to generate the setup intent only if Stripe ID exists.
	secret, err := services.GetClientSecretWithSetupPaymentIntent(customer.StripeId)
	if err != nil {
		log.Printf("GenerateStripePaymentIntent: Error generating setup intent for Stripe ID %s: %v", customer.StripeId, err)
		return "", err
	}

	// Return the client secret.
	return secret, nil
}

func CreateInactiveSubscriptionEntry(planID, priceID, userId int) {

	if planID != 0 && userId != 0 {
		err := stripe.InsertInactiveSubscription(planID, priceID, userId)
		if err != nil {
			log.Println("[ERROR-IN-CREATE] CreateInactiveSubscriptionEntry: Failed to InsertInactiveSubscription with :", err)
		}
	} else {
		log.Println("[MISSING-INFO] CreateInactiveSubscriptionEntry: Failed to InsertInactiveSubscription because either planID or userId missing")
	}

}

// CreateSubscription godoc
//
//	@Summary		Creates a Stripe subscription prepared for a first payment.
//	@Description	Creates a Stripe subscription prepared for a first payment.
//	@Description	Returns a client secret that is used to complete the payment on client's side.
//	@Tags			Subscription
//	@Accept			application/x-www-form-urlencoded
//	@Produce		json
//	@Param			Authorization	header		string	true	"Authorization token (Bearer token)"
//	@Param			plan-id			formData	int		true	"ID of the subscription plan to use"
//	@Success		200				{object}	subscription_controller.CreateSubscription.response
//	@Failure		400
//	@Failure		401
//	@Failure		500	{object}	ErrorResponse
//	@Router			/subscription [POST]
// func CreateSubscription(c *gin.Context) {

// 	// userID, exists := c.Get("userID")
// 	// if !exists {
// 	// 	log.Println("CreateSubscription: UserID not found in context.")
// 	// 	controllers. HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
// 	// 	return
// 	// }

// 	plan_id, err := strconv.Atoi(c.PostForm("plan_id"))
// 	if err != nil || plan_id <= 0 {
// 		log.Println("CreateSubscription: failed to convert subscription plan id into a valid number.")
// 		controllers.HandleJSONErrorResponse(apierror.InvalidPlanID, err, c)
// 		return
// 	}

// 	// client_secret, planName, maxSession, amount, price, err := stripe.Create(16, plan_id)
// 	// fmt.Println("client Secret <<<<<<----->>>>", client_secret)
// 	// if err != nil {
// 	// 	controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
// 	// 	return
// 	// }

// 	// // Done.
// 	// // Return success response with the subscription details using gin.H
// 	// c.JSON(http.StatusOK, gin.H{
// 	// 	"status":        "success",
// 	// 	"secret":        client_secret,
// 	// 	"total_classes": maxSession,
// 	// 	"plan_name":     planName,
// 	// 	"amount":        price,
// 	// 	"total_amount":  amount,
// 	// })
// }

func GetSubscription(c *gin.Context) {
	// Get auth token.
	token := c.GetHeader("Authorization")
	if len(token) == 0 {
		log.Println("GetSubscription: missing \"Authorization\" header.")
		controllers.HandleJSONErrorResponse(apierror.ErrorOnCheckingSession, nil, c)
		return
	}

	// Get user ID by auth token.
	user_id, err := models.GetUserIdByToken(token)
	if err != nil {
		log.Println("GetSubscription: failed to get user id by token with error:", err)
		controllers.HandleJSONErrorResponse(apierror.ErrorOnCheckingSession, err, c)
		return
	}

	// FIXME: incomplete
	log.Println("GET", user_id)
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   nil,
	})
}

func ListSubscriptionPlans(ctx *gin.Context) {
	type response struct {
		Status string                    `json:"status" enums:"success"` // API operation status.
		Plans  []stripe.SubscriptionPlan // The list of available subscription plans.
	}
	priceID := ctx.Query("price_id")
	if len(priceID) == 0 || priceID == "" {
		log.Println("ListSubscriptionPlans: price id not provided")
	}
	subscription_plans, err := stripe.ListPlans(priceID)
	if err != nil {
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, ctx)
	}
	if len(subscription_plans) == 1 {
		planDetails := subscription_plans[0]
		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"Plans":  planDetails})
	} else {
		ctx.JSON(http.StatusOK, response{"success", subscription_plans})
	}

}

func GetActivePlanDetails(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Println("[ERROR] GetActivePlan: UserID not found in context.")
		controllers.HandleJSONErrorResponse(apierror.ErrorUserNotFound, nil, c)
		return
	}

	plan, err := stripe.GetActivePlanDetails(userID.(int))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("[NO-ACTIVE-PLAN] GetActivePlan: No active plan found for the user")
			controllers.HandleJSONErrorResponse(apierror.ErrorNoDataFound, err, c)
			return
		}
		log.Println("[ERROR] GetActivePlan: Error fetching active plan for user:", 2, "Error:", err)
		controllers.HandleJSONErrorResponse(apierror.SomethingWentWrong, err, c)
		return
	}

	// Success case
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"plan":   plan,
	})
}

func markDefaultPaymentMethod() {
	// Retrieve the most recent customer ID and payment method ID.
	customerId, paymentMethod, err := stripe.GetRecentCustomerIDandPaymentMethodID()
	if err != nil {
		log.Println("markDefaultPaymentMethod: Error fetching customer and payment method:", err)
		return
	}

	// Check if both customerId and paymentMethod are not empty.
	if customerId != "" && paymentMethod != "" {
		err = services.MakeDefaultPaymentMethod(customerId, paymentMethod)
		if err != nil {
			log.Println("markDefaultPaymentMethod: Error making payment method default:", err)
			return
		}
		log.Println("markDefaultPaymentMethod: Successfully set payment method as default for customer", customerId)
	}

	// Fetch userID and planID from user2subscription table.
	userID, planID, priceID, err := stripe.GetPlanIdAndUserID(customerId)
	if err != nil {
		log.Println("[ERROR] markDefaultPaymentMethod: Error fetching userID and planID:", err)
		return
	}
	if userID == 0 || planID == 0 || priceID == 0 {
		log.Println("markDefaultPaymentMethod: userID or planID is empty, skipping subscription creation")
		return
	}

	// Create a subscription with a trial period.
	err = stripe.CreateSubscriptionWithTrial(userID, planID, priceID)
	if err != nil {
		log.Println("markDefaultPaymentMethod: Failed to create subscription with trial for user", userID, "and plan", planID, ":", err)
		return
	}
	// call the function to update
	err = controllers.UpdateLeadsOnZohoCRM(userID)
	if err != nil {
		log.Println("[ERROR] markDefaultPaymentMethod --->>UpdateLeadsOnZohoCRM : Failed to update the plan details on zoho CRM leadboard with error :", err)
	}
	// This function will book the user's regular session just after the successfully setting the PAYMENT method.
	err = session.BookSessionAfterPaymentSetup(userID)
	if err != nil {
		log.Println("[ERROR] markDefaultPaymentMethod --->>BookSessionAfterPaymentSetup : Failed to book the session", err)
	}

	log.Println("markDefaultPaymentMethod: Successfully created subscription with trial for user", userID)
}

func StripeWebhook(c *gin.Context) {

	event, err := services.GetWebhookEvents(c)
	if err != nil {
		log.Println("StripeWebhook:-->>>> readAndVerifyWebhook: error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("EVENT TYPE-------->>>>>>>", event.Type)
	// Store the raw event
	userID, err := stripe.StoreRawEvent(event)
	log.Println("userId", userID)
	if err != nil {
		log.Println("[ERROR] StripeWebhook --->>>>>readAndVerifyWebhook: Failed to store raw evant data", err)
		return
	}

	switch event.Type {
	case "setup_intent.succeeded":
		markDefaultPaymentMethod()

	case "invoice.payment_succeeded":
		if !stripe.IsTrialPeriodInvoice(event, userID) {
			user_id, err := stripe.AddSuccessPaymentStatus(userID)
			if err != nil {
				log.Println("StripeWebhook:-->>>> AddSuccessPaymentStatus: error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if user_id != 0 {
				session.CreditSessionsForUser(user_id)
			}
		}
	case "payment_intent.payment_failed":
		err := stripe.AddFailedPaymentStatus(event)
		if err != nil {
			log.Println("StripeWebhook:-->>>> AddFailedPaymentStatus: error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "price.created":
		stripe.HandlePriceEvents(event, stripe.InsertNewProductPrice)
	case "price.updated":
		stripe.HandlePriceEvents(event, stripe.UpdateProductPrice)
	case "price.deleted":
		stripe.HandlePriceEvents(event, stripe.DeleteProductPrice)
	case "product.updated":
		stripe.HandlePriceEvents(event, stripe.MarkPriceIdDefault)

	case "customer.subscription.updated":
		log.Println("event", event.Type)
		//user's stripe prodcut is changed
		err := stripe.ChangeSubscription(event)
		if err != nil {
			log.Println("StripeWebhook:-->>>> ChangeSubscription: error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
func RenewSubscription(c *gin.Context) {
	// var userID int64
	// user_id, err := stripe.AddSuccessPaymentStatus(userID)
	// if err != nil {
	// 	log.Println("StripeWebhook:-->>>> AddSuccessPaymentStatus: error", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	user_id, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "wrong user id",
		})
		return
	}
	if user_id != 0 {
		session.CreditSessionsForUser(user_id)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "done",
	})
}

// GetAllCustomerInfoToUpdateOnstripe retrieves customers and updates them on Stripe
func GetAllCustomerInfoToUpdateOnstripe(c *gin.Context) {
	// Example data retrieval function (replace this with your actual implementation)
	customer2Update, err := students.GetAllCustomerInfoToUpdateOnstripe()
	if err != nil {
		log.Printf("Error fetching customer info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer information"})
		return
	}

	for _, s := range customer2Update {
		err := services.UpdateUserInformationOnStripe(s.StripeCustomerId, s.StudentEmail, s.StudentName)
		if err != nil {
			log.Printf("Failed to update customer %s: %v", s.StripeCustomerId, err)
		} else {
			log.Printf("Successfully updated customer %s", s.StripeCustomerId)

			// Update the updated_at field in the database
			err = stripe.UpdateUser2StripeUpdatedAt(s.StripeCustomerId)
			if err != nil {
				log.Printf("Failed to update updated_at for customer %s: %v", s.StripeCustomerId, err)
			}
		}
	}

	c.JSON(200, gin.H{"message": "Customer information update process completed"})
}
