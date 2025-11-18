package supabase

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// ProfileStripeFields represents the subset of the Supabase "profiles" table that
// is relevant for billing. Having a dedicated struct keeps Stripe persistence
// logic isolated from the rest of the profile model.
type ProfileStripeFields struct {
	UserID                  string     `json:"user_id"`
	StripeCustomerID        string     `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID    string     `json:"stripe_subscription_id,omitempty"`
	StripePlan              string     `json:"stripe_plan,omitempty"`
	StripeStatus            string     `json:"stripe_status,omitempty"`
	StripeTrialEnd          *time.Time `json:"stripe_trial_end,omitempty"`
	StripeCurrentPeriodEnd  *time.Time `json:"stripe_current_period_end,omitempty"`
	StripeCancelAtPeriodEnd bool       `json:"stripe_cancel_at_period_end"`
	StripeLastInvoiceStatus string     `json:"stripe_last_invoice_status,omitempty"`
	StripeLastEventID       string     `json:"stripe_last_event_id,omitempty"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// UpsertProfileStripeFields writes the latest billing snapshot for a user into Supabase.
// It keeps the record keyed by user_id, ensuring webhook retries remain idempotent.
func UpsertProfileStripeFields(ctx context.Context, p ProfileStripeFields) error {
	updateData := map[string]interface{}{
		"stripe_customer_id":          p.StripeCustomerID,
		"stripe_subscription_id":      p.StripeSubscriptionID,
		"stripe_plan":                 p.StripePlan,
		"stripe_status":               p.StripeStatus,
		"stripe_cancel_at_period_end": p.StripeCancelAtPeriodEnd,
		"stripe_last_invoice_status":  p.StripeLastInvoiceStatus,
		"stripe_last_event_id":        p.StripeLastEventID,
		"updated_at":                  time.Now().UTC().Format(time.RFC3339),
	}

	// Handle time fields
	if p.StripeTrialEnd != nil {
		updateData["stripe_trial_end"] = p.StripeTrialEnd.Format(time.RFC3339)
	} else {
		updateData["stripe_trial_end"] = nil
	}

	if p.StripeCurrentPeriodEnd != nil {
		updateData["stripe_current_period_end"] = p.StripeCurrentPeriodEnd.Format(time.RFC3339)
	} else {
		updateData["stripe_current_period_end"] = nil
	}

	log.Printf("📤 Updating profile for user %s with data: customer_id=%s, subscription_id=%s, plan=%s, status=%s",
		p.UserID, p.StripeCustomerID, p.StripeSubscriptionID, p.StripePlan, p.StripeStatus)

	_, _, err := Client.
		From("profiles").
		Update(updateData, "", "").
		Eq("user_id", p.UserID).
		Execute()

	if err != nil {
		log.Printf("❌ Error updating profile for user %s: %v", p.UserID, err)
	} else {
		log.Printf("✅ Successfully updated profile for user %s in database", p.UserID)
	}

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

// InitializeTrial sets a 14-day free trial for a user if they don't already have one.
// Returns true if a trial was initialized, false if they already had one.
func InitializeTrial(ctx context.Context, userID string) (bool, error) {
	// First check if user already has a trial_end set or an active subscription
	profile, err := GetProfileStripeFields(ctx, userID)
	if err != nil {
		// If profile doesn't exist, we'll still try to initialize trial
		// The update will fail gracefully if the profile doesn't exist
	} else {
		// User already has a trial_end set, don't overwrite it
		if profile.StripeTrialEnd != nil {
			return false, nil
		}
		// User has an active subscription (Pro or Business), don't initialize trial
		if profile.StripeStatus == "active" && (profile.StripePlan == "pro" || profile.StripePlan == "business") {
			return false, nil
		}
	}

	// Calculate trial end date: 14 days from now
	trialEnd := time.Now().UTC().AddDate(0, 0, 14)

	// Update only the trial_end field
	updateData := map[string]interface{}{
		"stripe_trial_end": trialEnd.Format(time.RFC3339),
	}

	_, _, err = Client.
		From("profiles").
		Update(updateData, "", "").
		Eq("user_id", userID).
		Execute()

	if err != nil {
		return false, err
	}

	return true, nil
}

// OrganizationStripeFields represents the subset of the "organizations" table that
// is relevant for billing at the organization level.
type OrganizationStripeFields struct {
	OrganizationID          string     `json:"id"`
	StripeCustomerID        string     `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID    string     `json:"stripe_subscription_id,omitempty"`
	StripePlan              string     `json:"stripe_plan,omitempty"`
	StripeStatus            string     `json:"stripe_status,omitempty"`
	StripeTrialEnd          *time.Time `json:"stripe_trial_end,omitempty"`
	StripeCurrentPeriodEnd  *time.Time `json:"stripe_current_period_end,omitempty"`
	StripeCancelAtPeriodEnd bool       `json:"stripe_cancel_at_period_end"`
	StripeLastInvoiceStatus string     `json:"stripe_last_invoice_status,omitempty"`
	StripeLastEventID       string     `json:"stripe_last_event_id,omitempty"`
}

// GetOrganizationIDFromUserID fetches the organization_id for a given user_id.
func GetOrganizationIDFromUserID(ctx context.Context, userID string) (string, error) {
	data, _, err := Client.
		From("profiles").
		Select("organization_id", "exact", false).
		Eq("user_id", userID).
		Single().
		Execute()
	if err != nil {
		return "", err
	}

	var result struct {
		OrganizationID string `json:"organization_id"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	return result.OrganizationID, nil
}

// UpsertOrganizationStripeFields writes the latest billing snapshot for an organization into Supabase.
// It keeps the record keyed by organization_id, ensuring webhook retries remain idempotent.
func UpsertOrganizationStripeFields(ctx context.Context, o OrganizationStripeFields) error {
	updateData := map[string]interface{}{
		"stripe_customer_id":          o.StripeCustomerID,
		"stripe_subscription_id":      o.StripeSubscriptionID,
		"stripe_plan":                 o.StripePlan,
		"stripe_status":               o.StripeStatus,
		"stripe_cancel_at_period_end": o.StripeCancelAtPeriodEnd,
		"stripe_last_invoice_status":  o.StripeLastInvoiceStatus,
		"stripe_last_event_id":        o.StripeLastEventID,
	}

	// Handle time fields
	if o.StripeTrialEnd != nil {
		updateData["stripe_trial_end"] = o.StripeTrialEnd.Format(time.RFC3339)
	} else {
		updateData["stripe_trial_end"] = nil
	}

	if o.StripeCurrentPeriodEnd != nil {
		updateData["stripe_current_period_end"] = o.StripeCurrentPeriodEnd.Format(time.RFC3339)
	} else {
		updateData["stripe_current_period_end"] = nil
	}

	log.Printf("📤 Updating organization %s with data: customer_id=%s, subscription_id=%s, plan=%s, status=%s",
		o.OrganizationID, o.StripeCustomerID, o.StripeSubscriptionID, o.StripePlan, o.StripeStatus)

	_, _, err := Client.
		From("organizations").
		Update(updateData, "", "").
		Eq("id", o.OrganizationID).
		Execute()

	if err != nil {
		log.Printf("❌ Error updating organization %s: %v", o.OrganizationID, err)
	} else {
		log.Printf("✅ Successfully updated organization %s in database", o.OrganizationID)
	}

	return err
}
