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
	p.UpdatedAt = time.Now().UTC()

	_, _, err := Client.
		From("profiles").
		Upsert(p, "", "", "").
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

// GetOrganizationStripeFields fetches the Stripe subscription details for an organization.
func GetOrganizationStripeFields(ctx context.Context, organizationID string) (*OrganizationStripeFields, error) {
	data, _, err := Client.
		From("organizations").
		Select("id,stripe_customer_id,stripe_subscription_id,stripe_plan,stripe_status,stripe_trial_end,stripe_current_period_end,stripe_cancel_at_period_end,stripe_last_invoice_status,stripe_last_event_id", "exact", false).
		Eq("id", organizationID).
		Single().
		Execute()
	if err != nil {
		return nil, err
	}

	var orgFields OrganizationStripeFields
	if err := json.Unmarshal(data, &orgFields); err != nil {
		return nil, err
	}

	return &orgFields, nil
}

// CombinedBillingStatus represents the effective billing status for a user,
// taking into account both their organization's subscription and their individual profile.
// Organization subscription takes priority over individual profile.
type CombinedBillingStatus struct {
	StripeStatus           string     `json:"stripe_status"`
	StripePlan             string     `json:"stripe_plan"`
	StripeTrialEnd         *time.Time `json:"stripe_trial_end,omitempty"`
	StripeCurrentPeriodEnd *time.Time `json:"stripe_current_period_end,omitempty"`
	Source                 string     `json:"source"` // "organization" or "profile"
	OrganizationID         string     `json:"organization_id,omitempty"`
}

// GetCombinedBillingStatus retrieves the effective billing status for a user.
// It checks the organization's subscription first, and falls back to the user's individual profile.
// Organization subscription always takes priority.
func GetCombinedBillingStatus(ctx context.Context, userID string) (*CombinedBillingStatus, error) {
	// First, get the user's organization ID
	organizationID, err := GetOrganizationIDFromUserID(ctx, userID)
	if err != nil {
		// If user has no organization, fall back to profile only
		log.Printf("⚠️ User %s has no organization, using profile billing status only", userID)
		profileFields, err := GetProfileStripeFields(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &CombinedBillingStatus{
			StripeStatus:           profileFields.StripeStatus,
			StripePlan:             profileFields.StripePlan,
			StripeTrialEnd:         profileFields.StripeTrialEnd,
			StripeCurrentPeriodEnd: profileFields.StripeCurrentPeriodEnd,
			Source:                 "profile",
		}, nil
	}

	// Try to get organization's Stripe fields
	orgFields, err := GetOrganizationStripeFields(ctx, organizationID)
	if err != nil {
		// If organization has no Stripe fields, fall back to profile
		log.Printf("⚠️ Organization %s has no Stripe fields, using profile billing status for user %s", organizationID, userID)
		profileFields, err := GetProfileStripeFields(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &CombinedBillingStatus{
			StripeStatus:           profileFields.StripeStatus,
			StripePlan:             profileFields.StripePlan,
			StripeTrialEnd:         profileFields.StripeTrialEnd,
			StripeCurrentPeriodEnd: profileFields.StripeCurrentPeriodEnd,
			Source:                 "profile",
			OrganizationID:         organizationID,
		}, nil
	}

	// Check if organization has an active subscription or valid trial
	now := time.Now().UTC()
	hasActiveOrgSubscription := orgFields.StripeStatus == "active" && (orgFields.StripePlan == "pro" || orgFields.StripePlan == "business")
	hasValidOrgTrial := orgFields.StripeTrialEnd != nil && orgFields.StripeTrialEnd.After(now)
	hasValidOrgPeriod := orgFields.StripeCurrentPeriodEnd != nil && orgFields.StripeCurrentPeriodEnd.After(now)
	hasOrgBillingInfo := orgFields.StripeStatus != "" || orgFields.StripePlan != ""

	// If organization has active subscription or valid trial/period, use organization's status
	if hasActiveOrgSubscription || hasValidOrgTrial || hasValidOrgPeriod {
		log.Printf("✅ User %s inheriting billing status from organization %s (status=%s, plan=%s)", userID, organizationID, orgFields.StripeStatus, orgFields.StripePlan)

		// Synchronize current user's profile immediately to ensure consistency
		if err := syncUserProfileWithOrganization(ctx, userID, orgFields); err != nil {
			log.Printf("⚠️ Failed to sync user %s profile with organization (non-blocking): %v", userID, err)
		}

		// Also sync all other members/admins in background
		go func(orgID string) {
			bgCtx := context.Background()
			if err := SyncOrganizationBillingToMembers(bgCtx, orgID); err != nil {
				log.Printf("⚠️ Background sync of organization billing to members failed (non-blocking): %v", err)
			}
		}(organizationID)

		return &CombinedBillingStatus{
			StripeStatus:           orgFields.StripeStatus,
			StripePlan:             orgFields.StripePlan,
			StripeTrialEnd:         orgFields.StripeTrialEnd,
			StripeCurrentPeriodEnd: orgFields.StripeCurrentPeriodEnd,
			Source:                 "organization",
			OrganizationID:         organizationID,
		}, nil
	}

	// Even if organization doesn't have active subscription, sync billing info if it exists
	// This ensures profile_stripe_plan and profile_stripe_status always match org values
	if hasOrgBillingInfo {
		log.Printf("🔄 Organization %s has billing info (status=%s, plan=%s), syncing to members", organizationID, orgFields.StripeStatus, orgFields.StripePlan)

		// Synchronize current user's profile immediately
		if err := syncUserProfileWithOrganization(ctx, userID, orgFields); err != nil {
			log.Printf("⚠️ Failed to sync user %s profile with organization (non-blocking): %v", userID, err)
		}

		// Also sync all other members/admins in background
		go func(orgID string) {
			bgCtx := context.Background()
			if err := SyncOrganizationBillingToMembers(bgCtx, orgID); err != nil {
				log.Printf("⚠️ Background sync of organization billing to members failed (non-blocking): %v", err)
			}
		}(organizationID)
	}

	// Organization doesn't have active subscription/trial, fall back to profile
	log.Printf("⚠️ Organization %s has no active subscription/trial, using profile billing status for user %s", organizationID, userID)
	profileFields, err := GetProfileStripeFields(ctx, userID)
	if err != nil {
		// If profile also doesn't exist, return organization status anyway
		return &CombinedBillingStatus{
			StripeStatus:           orgFields.StripeStatus,
			StripePlan:             orgFields.StripePlan,
			StripeTrialEnd:         orgFields.StripeTrialEnd,
			StripeCurrentPeriodEnd: orgFields.StripeCurrentPeriodEnd,
			Source:                 "organization",
			OrganizationID:         organizationID,
		}, nil
	}

	return &CombinedBillingStatus{
		StripeStatus:           profileFields.StripeStatus,
		StripePlan:             profileFields.StripePlan,
		StripeTrialEnd:         profileFields.StripeTrialEnd,
		StripeCurrentPeriodEnd: profileFields.StripeCurrentPeriodEnd,
		Source:                 "profile",
		OrganizationID:         organizationID,
	}, nil
}

// syncUserProfileWithOrganization synchronizes a single user's profile with their organization's billing status
func syncUserProfileWithOrganization(ctx context.Context, userID string, orgFields *OrganizationStripeFields) error {
	updateData := map[string]interface{}{
		"stripe_status":               orgFields.StripeStatus,
		"stripe_plan":                 orgFields.StripePlan,
		"stripe_cancel_at_period_end": orgFields.StripeCancelAtPeriodEnd,
		"stripe_last_invoice_status":  orgFields.StripeLastInvoiceStatus,
	}

	// Handle time fields - always set, even if nil
	if orgFields.StripeTrialEnd != nil {
		updateData["stripe_trial_end"] = orgFields.StripeTrialEnd.Format(time.RFC3339)
	} else {
		updateData["stripe_trial_end"] = nil
	}

	if orgFields.StripeCurrentPeriodEnd != nil {
		updateData["stripe_current_period_end"] = orgFields.StripeCurrentPeriodEnd.Format(time.RFC3339)
	} else {
		updateData["stripe_current_period_end"] = nil
	}

	_, _, err := Client.
		From("profiles").
		Update(updateData, "", "").
		Eq("user_id", userID).
		Execute()

	if err != nil {
		log.Printf("❌ Failed to sync profile for user %s: %v", userID, err)
		return err
	}

	log.Printf("✅ Successfully synced profile for user %s with organization billing status", userID)
	return nil
}

// SyncOrganizationBillingToMembers synchronizes the organization's billing status
// to all members and admins of that organization. This is optional since
// GetCombinedBillingStatus already checks the organization in real-time,
// but can be useful for backward compatibility or performance optimization.
func SyncOrganizationBillingToMembers(ctx context.Context, organizationID string) error {
	// Get organization's Stripe fields
	orgFields, err := GetOrganizationStripeFields(ctx, organizationID)
	if err != nil {
		return err
	}

	// Get all members and admins of this organization
	data, _, err := Client.
		From("profiles").
		Select("user_id", "exact", false).
		Eq("organization_id", organizationID).
		In("role", []string{"admin", "member"}).
		Execute()
	if err != nil {
		return err
	}

	var profiles []struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(data, &profiles); err != nil {
		return err
	}

	log.Printf("🔄 Syncing organization %s billing status to %d members/admins", organizationID, len(profiles))

	// Update each member's profile with organization's billing status
	for _, profile := range profiles {
		updateData := map[string]interface{}{
			"stripe_status":               orgFields.StripeStatus,
			"stripe_plan":                 orgFields.StripePlan,
			"stripe_cancel_at_period_end": orgFields.StripeCancelAtPeriodEnd,
			"stripe_last_invoice_status":  orgFields.StripeLastInvoiceStatus,
		}

		// Handle time fields
		if orgFields.StripeTrialEnd != nil {
			updateData["stripe_trial_end"] = orgFields.StripeTrialEnd.Format(time.RFC3339)
		} else {
			updateData["stripe_trial_end"] = nil
		}

		if orgFields.StripeCurrentPeriodEnd != nil {
			updateData["stripe_current_period_end"] = orgFields.StripeCurrentPeriodEnd.Format(time.RFC3339)
		} else {
			updateData["stripe_current_period_end"] = nil
		}

		_, _, err := Client.
			From("profiles").
			Update(updateData, "", "").
			Eq("user_id", profile.UserID).
			Execute()
		if err != nil {
			log.Printf("⚠️ Failed to sync billing status to user %s: %v", profile.UserID, err)
			// Continue with other users even if one fails
		}
	}

	log.Printf("✅ Successfully synced organization %s billing status to members", organizationID)
	return nil
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
