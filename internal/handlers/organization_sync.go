package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"backend-api-skillforge/internal/supabase"

	"github.com/gin-gonic/gin"
)

// SyncOrganizationSubscriptionToUsers manually triggers the trickle-down
// of subscription details from an organization to its users.
// This is useful for retroactively updating existing users when an organization
// already has an active subscription.
func SyncOrganizationSubscriptionToUsers(c *gin.Context) {
	log.Printf("🔄 Manual sync requested for organization subscription trickle-down")

	organizationID := c.Param("organization_id")
	if organizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "organization_id is required",
		})
		return
	}

	ctx := c.Request.Context()

	// Fetch organization's current Stripe fields
	orgFields, err := getOrganizationStripeFields(ctx, organizationID)
	if err != nil {
		log.Printf("❌ Failed to fetch organization Stripe fields: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch organization subscription details",
		})
		return
	}

	// Check if organization has an active subscription
	if orgFields.StripeStatus != "active" || (orgFields.StripePlan != "pro" && orgFields.StripePlan != "business") {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Organization does not have an active pro/business subscription, no sync needed",
			"plan":    orgFields.StripePlan,
			"status":  orgFields.StripeStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Organization subscription sync is temporarily disabled",
		"organization_id": organizationID,
		"plan":            orgFields.StripePlan,
		"status":          orgFields.StripeStatus,
	})
}

// getOrganizationStripeFields fetches the Stripe subscription details for an organization
func getOrganizationStripeFields(ctx context.Context, organizationID string) (*supabase.OrganizationStripeFields, error) {
	data, _, err := supabase.Client.
		From("organizations").
		Select("id,stripe_customer_id,stripe_subscription_id,stripe_plan,stripe_status,stripe_trial_end,stripe_current_period_end,stripe_cancel_at_period_end,stripe_last_invoice_status,stripe_last_event_id", "exact", false).
		Eq("id", organizationID).
		Single().
		Execute()
	if err != nil {
		return nil, err
	}

	var orgFields supabase.OrganizationStripeFields
	if err := json.Unmarshal(data, &orgFields); err != nil {
		return nil, err
	}

	return &orgFields, nil
}
