package stripe

import (
	"fmt"
	"log"
	"strings"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

type BillingAddress struct {
	Line1      string
	City       string
	PostalCode string
	State      string
	Country    string
}

type BillingDetails struct {
	Name    string
	Email   string
	Phone   string
	Address *BillingAddress
}

type CheckoutConfig struct {
	PriceID         string
	CustomerID      string // optional – if omitted, Checkout creates one
	UserID          string // internal reference
	SuccessURL      string
	CancelURL       string
	AllowPromoCodes bool
	TrialFromPlan   bool // true → use price's trial; false → skip
	Metadata        map[string]string
	CustomerEmail   string
	BillingDetails  *BillingDetails
	CompanyName     string
	VATNumber       string
}

func CreateCheckoutSession(cfg CheckoutConfig) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:                stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:          stripe.String(cfg.SuccessURL),
		CancelURL:           stripe.String(cfg.CancelURL),
		ClientReferenceID:   stripe.String(cfg.UserID), // or store inside metadata
		AllowPromotionCodes: stripe.Bool(cfg.AllowPromoCodes),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialSettings: &stripe.CheckoutSessionSubscriptionDataTrialSettingsParams{
				EndBehavior: &stripe.CheckoutSessionSubscriptionDataTrialSettingsEndBehaviorParams{
					MissingPaymentMethod: stripe.String(string(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodCancel)),
				},
			},
		},
	}

	if strings.TrimSpace(cfg.PriceID) == "" {
		return nil, fmt.Errorf("PriceID is required")
	}
	params.LineItems = []*stripe.CheckoutSessionLineItemParams{
		{
			Price:    stripe.String(cfg.PriceID),
			Quantity: stripe.Int64(1),
		},
	}

	if cfg.CustomerID != "" {
		params.Customer = stripe.String(cfg.CustomerID)
	}

	// Pre-fill customer email if provided
	if cfg.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(cfg.CustomerEmail)
	}

	// Note: Billing details (address, name, phone) cannot be pre-filled in Stripe Checkout v78 SDK
	// They will be collected by Stripe during the checkout process.
	// If pre-filling is required, you would need to create/update the customer first with billing details,
	// then use that customer ID in the checkout session.

	if !cfg.TrialFromPlan {
		params.SubscriptionData.TrialSettings = nil
		params.SubscriptionData.TrialPeriodDays = stripe.Int64(0)
	}

	// Copy metadata to both the session and the resulting subscription.
	if params.Metadata == nil {
		params.Metadata = make(map[string]string)
	}
	if params.SubscriptionData.Metadata == nil {
		params.SubscriptionData.Metadata = make(map[string]string)
	}

	for k, v := range cfg.Metadata {
		params.Metadata[k] = v
		params.SubscriptionData.Metadata[k] = v
	}
	params.Metadata["user_id"] = cfg.UserID
	params.SubscriptionData.Metadata["user_id"] = cfg.UserID

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		return nil, err
	}
	log.Printf("Checkout session created: %s", sess.ID)
	return sess, nil
}
