package services

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/product"
	"github.com/stripe/stripe-go/v79/setupintent"
	"github.com/stripe/stripe-go/v79/subscription"
)

// Returns newly created customer's ID in stripe.
func CreateCustomer(phone string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.CustomerParams{
		Phone: stripe.String(phone),
	}
	customer, err := customer.New(params)
	if err != nil {
		return "", err
	}

	return customer.ID, nil
}

// Returns newly created subscription's ID in Stripe.
func CreateSubscription(customerId, priceId string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	trialEnd := time.Now().AddDate(0, 0, 27).Unix()

	paymentSettings := &stripe.SubscriptionPaymentSettingsParams{
		SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		PaymentMethodTypes:       stripe.StringSlice([]string{"card"}),
	}
	subscriptionParams := &stripe.SubscriptionParams{
		TrialEnd: stripe.Int64(trialEnd),
		Customer: stripe.String(customerId),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceId),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: paymentSettings,
	}
	subscription, err := subscription.New(subscriptionParams)
	if err != nil {
		return "", err
	}
	// MarkUnCollectibleJustAfterFirstPayment(subscription.ID)

	return subscription.ID, nil
}

// Returns active subscription plan's ID for a user.
func GetActiveSubscriptionPlan(customerId string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerId),
		Status:   stripe.String("active"),
	}
	params.Limit = stripe.Int64(1) // TODO: we won't need to limit this if we want to check the error below.
	params.AddExpand("data.plan")
	iter := subscription.List(params)
	var res *string = nil
	for iter.Next() {
		// TODO: check if there are more than one active subscription in Stripe;
		//       in this case res would be not nil; handle that as an error.
		res = &iter.Subscription().Items.Data[0].Plan.Product.ID
	}
	if err := iter.Err(); err != nil {
		return "", err
	}
	return *res, nil
}

func GetClientSecretWithSetupPaymentIntent(customerID string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.SetupIntentParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: []*string{stripe.String("card")},
	}
	result, err := setupintent.New(params)
	if err != nil {
		log.Println("[STRIPE-ERROR] GetClientSecretWithSetupPaymentIntent: failed to get secret with :", err)
		return "", err
	}

	return result.ClientSecret, nil

}

func MakeDefaultPaymentMethod(customerID, paymentMethodID string) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create customer update parameters to set the default payment method
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}

	// Update the customer with the default payment method
	_, err := customer.Update(customerID, params)
	if err != nil {
		log.Println("[STRIPE-ERROR] MakeDefaultPaymentMethod: failed to get secret with :", err)
		return err
	}

	return nil
}

// Returns client secret that is used to make a payment on a client's side.
func GetClientSecretFromSubscription(subscriptionId string) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.SubscriptionParams{}
	params.AddExpand("latest_invoice.payment_intent")
	subscription, err := subscription.Get(subscriptionId, params)
	if err != nil {
		return "", err
	}

	return subscription.LatestInvoice.PaymentIntent.ClientSecret, nil
}

// Returns a Stripe subscription.
func GetSubscription(subscriptionId string) (*stripe.Subscription, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.SubscriptionParams{}
	subscription, err := subscription.Get(subscriptionId, params)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

func GetSubscriptionPlan(planId string) (*stripe.Product, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.ProductParams{}
	plan, err := product.Get(planId, params)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// MarkUnCollectibleAfterFirstPayment marks the subscription as uncollectible
// for all customers by default. This function is used to change the billing cycle
// when the user books their first regular session. During the period before the
// first session, any generated bill will not be collected. Once the user books
// their first class, the billing cycle starts from that booking date, covering a
// 28-day period.

func MarkUnCollectibleJustAfterFirstPayment(subscriptionID string) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Define subscription parameters to mark as uncollectible.
	params := &stripe.SubscriptionParams{
		PauseCollection: &stripe.SubscriptionPauseCollectionParams{
			Behavior: stripe.String(string(stripe.SubscriptionPauseCollectionBehaviorMarkUncollectible)),
		},
	}

	// Update the subscription to mark it as uncollectible.
	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		log.Printf("[STRIPE-ERROR] Failed to mark subscription %s as uncollectible: %v\n", subscriptionID, err)
	}
}

// UnPauseTheSubscription is used to unpause a subscription in Stripe by removing the pause collection settings.
// It updates the subscription to resume normal billing.

// NOTE :- BEFORE Updating the billing Cycle we have to Unpause the subscription.
func UnPauseTheSubscription(subscriptionID string) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.SubscriptionParams{}
	params.AddExtra("pause_collection", "")

	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		log.Printf("[STRIPE-ERROR] UnPauseTheSubscription: Failed while unpausing the subscription with ID %s: %v", subscriptionID, err)
		return err
	}

	return nil
}

// MarkTrialEnd updates the subscription's billing cycle anchor to the current time
// and ends the trial period immediately. This is typically used when the user books
// their first regular session, ensuring that the billing cycle starts from the time
// of the booking. The subscription's trial period is also set to end now, aligning
// the service period with the user's first booking (e.g., starting from the first booking
// and lasting for the subsequent 28-day period or other defined billing cycle duration).

func MarkTrialEndOnStripe(subscriptionID string) (*stripe.Subscription, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Define subscription parameters for updating the billing cycle.
	params := &stripe.SubscriptionParams{
		BillingCycleAnchorNow: stripe.Bool(true),
		TrialEndNow:           stripe.Bool(true),
	}

	// Update the subscription with the new billing cycle anchor.
	subscription, err := subscription.Update(subscriptionID, params)
	if err != nil {
		log.Printf("[STRIPE-ERROR] Failed to update MarkTrialEnd for subscription %s: %v", subscriptionID, err)
		return nil, fmt.Errorf("[STRIPE-ERROR] Failed to update  MarkTrialEnd for subscription %s: %v", subscriptionID, err)
	}

	return subscription, nil
}

// UpdateUserInformationOnStripe updates a customer's name and email on Stripe.
// Both name and email are optional. Only non-empty values will be updated.
func UpdateUserInformationOnStripe(customerID, email, name string) error {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Prepare the parameters for updating the customer
	params := &stripe.CustomerParams{}

	if email != "" {
		params.Email = stripe.String(email)
	}

	if name != "" {
		params.Name = stripe.String(name)
	}

	if email != "" || name != "" {
		_, err := customer.Update(customerID, params)
		if err != nil {
			return err
		}
		log.Println("[STRIPE-SUCCESS] UpdateUserInformationOnStripe: Successfully updated the user infromation on stripe")
	}
	return nil
}

func UpdateSubscriptionTrialEndDateOnStripe(subscriptionID string, trialEndByDate time.Time) error {
	// Set the Stripe secret key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Convert trialEndByDate to a Unix timestamp
	trialWillEndAt := trialEndByDate.Unix()

	// Prepare the subscription parameters
	params := &stripe.SubscriptionParams{
		TrialEnd: stripe.Int64(trialWillEndAt), // Use the Unix timestamp directly
	}

	// Update the subscription with the new trial end date
	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		log.Printf("[STRIPE-ERROR] Failed to TrialEndByDate for subscription %s: %v", subscriptionID, err)
		return err // Return the actual error for proper error handling
	}

	// Return nil if no error occurred
	return nil
}
