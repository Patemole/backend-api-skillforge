package supabase

import (
	"context"
	"encoding/json"
	"time"
)

// ProfileStripeFields represents the subset of the Supabase "profiles" table that
// is relevant for billing. Having a dedicated struct keeps Stripe persistence
// logic isolated from the rest of the profile model.
type ProfileStripeFields struct {
	UserID                string     `json:"user_id"`
	StripeCustomerID      string     `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID  string     `json:"stripe_subscription_id,omitempty"`
	StripePlan            string     `json:"stripe_plan,omitempty"`
	StripeStatus          string     `json:"stripe_status,omitempty"`
	StripeTrialEnd        *time.Time `json:"stripe_trial_end,omitempty"`
	StripeCurrentPeriodEnd *time.Time `json:"stripe_current_period_end,omitempty"`
	StripeCancelAtPeriodEnd bool     `json:"stripe_cancel_at_period_end"`
	StripeLastInvoiceStatus string   `json:"stripe_last_invoice_status,omitempty"`
	StripeLastEventID     string     `json:"stripe_last_event_id,omitempty"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// UpsertProfileStripeFields writes the latest billing snapshot for a user into Supabase.
// It keeps the record keyed by user_id, ensuring webhook retries remain idempotent.
func UpsertProfileStripeFields(ctx context.Context, p ProfileStripeFields) error {
	p.UpdatedAt = time.Now().UTC()

	_, _, err := Client.
		From("profiles").
		Upsert(p, false, "", "", "").
		Eq("user_id", p.UserID).
		Execute()
	return err
}

// GetProfileStripeFields fetches the existing Stripe metadata for a user.
// Callers can use this to avoid creating duplicate Stripe customers or to
// inspect the latest subscription status.
func GetProfileStripeFields(ctx context.Context, userID string) (*ProfileStripeFields, error) {
	data, _, err := Client.
		From("profiles").
		Select("user_id,stripe_customer_id,stripe_subscription_id,stripe_plan,stripe_status,stripe_trial_end,stripe_current_period_end,stripe_cancel_at_period_end,stripe_last_invoice_status,stripe_last_event_id", "exact", false).
		Eq("user_id", userID).
		Single().
		Execute()
	if err != nil {
		return nil, err
	}

	var rec ProfileStripeFields
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}