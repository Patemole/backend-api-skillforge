package stripe

import (
	"fmt"
	"log"
	"strings"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

type CheckoutConfig struct {
	PriceID         string
	CustomerID      string // optional – if omitted, Checkout creates one
	UserID          string // internal reference
	SuccessURL      string
	CancelURL       string
	AllowPromoCodes bool
	TrialFromPlan   bool // true → use price’s trial; false → skip
	Metadata        map[string]string
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
