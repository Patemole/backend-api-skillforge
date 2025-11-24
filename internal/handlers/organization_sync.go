package handlers

import (
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
	orgFields, err := supabase.GetOrganizationStripeFields(ctx, organizationID)
	if err != nil {
		log.Printf("❌ Failed to fetch organization Stripe fields: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch organization subscription details",
		})
		return
	}

	// Sync organization billing status to all members and admins
	if err := supabase.SyncOrganizationBillingToMembers(ctx, organizationID); err != nil {
		log.Printf("❌ Failed to sync organization billing to members: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to sync billing status to members",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Successfully synced organization billing status to all members and admins",
		"organization_id": organizationID,
		"plan":            orgFields.StripePlan,
		"status":          orgFields.StripeStatus,
	})
}
