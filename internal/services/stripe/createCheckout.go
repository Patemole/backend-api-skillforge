package stripe

import (
	"log"
	"os"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

type CheckoutConfig struct {
	PriceID          string
	CustomerID       string // optional – if omitted, Checkout creates one
	UserID           string // internal reference
	SuccessURL       string
	CancelURL        string
	AllowPromoCodes  bool
	TrialFromPlan    bool // true → use price’s trial; false → skip
	Metadata         map[string]string
}

func CreateCheckoutSession(cfg CheckoutConfig) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(cfg.SuccessURL),
		CancelURL:  stripe.String(cfg.CancelURL),
		ClientReferenceID: stripe.String(cfg.UserID), // or store inside metadata
		AllowPromotionCodes: stripe.Bool(cfg.AllowPromoCodes),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialSettings: &stripe.SubscriptionTrialSettingsParams{
				EndBehavior: &stripe.SubscriptionTrialSettingsEndBehaviorParams{
					MissingPaymentMethod: stripe.String(string(stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodCancel)),
				},
			},
		},
	}

	if cfg.PriceID == "" {
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

	if len(cfg.Metadata) > 0 {
		if params.Metadata == nil {
			params.Metadata = make(map[string]string, len(cfg.Metadata))
		}
		for k, v := range cfg.Metadata {
			params.Metadata[k] = v
		}
	}
	params.Metadata["user_id"] = cfg.UserID

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		return nil, err
	}
	log.Printf("Checkout session created: %s", sess.ID)
	return sess, nil
}