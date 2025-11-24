package stripe

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-api-skillforge/internal/supabase"
)

type createCheckoutRequest struct {
	Plan          string            `json:"plan" binding:"required"`
	UserID        string            `json:"user_id" binding:"required"`
	CustomerID    string            `json:"customer_id"`
	BillingPeriod string            `json:"billing_period"` // "monthly" or "annual"
	Metadata      map[string]string `json:"metadata"`
}

var planPriceMap = map[string]map[string]string{
	"pro": {
		"monthly": os.Getenv("STRIPE_PRO_PRICE_MONTHLY"),
		"annual":  os.Getenv("STRIPE_PRO_PRICE_ANNUAL"),
	},
	"business": {
		"monthly": os.Getenv("STRIPE_BUSINESS_PRICE_MONTHLY"),
		"annual":  os.Getenv("STRIPE_BUSINESS_PRICE_ANNUAL"),
	},
}

// Legacy support: fallback to old env vars if new ones are not set
func getPriceID(plan, billingPeriod string) string {
	// Default to monthly if not specified
	if billingPeriod == "" {
		billingPeriod = "monthly"
	}

	// Normalize billing period
	if billingPeriod != "monthly" && billingPeriod != "annual" {
		billingPeriod = "monthly"
	}

	// Try new format first
	if priceMap, ok := planPriceMap[plan]; ok {
		if priceID := priceMap[billingPeriod]; priceID != "" {
			return priceID
		}
	}

	// Fallback to legacy env vars for backward compatibility
	if billingPeriod == "monthly" {
		switch plan {
		case "pro":
			return os.Getenv("STRIPE_PRO_PRICE")
		case "business":
			return os.Getenv("STRIPE_BUSINESS_PRICE")
		}
	}

	return ""
}

func CreateCheckoutSessionHandler(c *gin.Context) {
	var req createCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	// Get user_id from request body (required field)
	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id"})
		return
	}

	billingPeriod := strings.TrimSpace(req.BillingPeriod)
	if billingPeriod == "" {
		billingPeriod = "monthly" // Default to monthly
	}

	priceID := strings.TrimSpace(getPriceID(req.Plan, billingPeriod))
	if priceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown plan or billing period"})
		return
	}

	ctx := c.Request.Context()

	customerID := strings.TrimSpace(req.CustomerID)
	if customerID == "" {
		// Try to fetch existing Stripe customer ID from profile, but don't fail if it doesn't exist
		// PGRST116 means "not found" which is fine for new users
		if profileStripe, err := supabase.GetProfileStripeFields(ctx, userID); err == nil {
			customerID = profileStripe.StripeCustomerID
		} else if !strings.Contains(err.Error(), "PGRST116") {
			// Log non-critical errors but continue - Stripe can create a new customer
			log.Printf("⚠️ [Billing] Failed to fetch billing profile for user %s: %v (continuing anyway)", userID, err)
		}
		// If error is PGRST116 (not found), that's fine - Stripe will create a new customer
		// If it's any other error, we log it but continue - Stripe can still create a customer
	}

	metadata := make(map[string]string, len(req.Metadata)+1)
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	metadata["plan"] = req.Plan

	successURL := os.Getenv("STRIPE_CHECKOUT_SUCCESS_URL")
	cancelURL := os.Getenv("STRIPE_CHECKOUT_CANCEL_URL")

	log.Printf("🔗 [Billing] Environment variables - SuccessURL: %s, CancelURL: %s", successURL, cancelURL)

	if successURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "STRIPE_CHECKOUT_SUCCESS_URL not configured"})
		return
	}
	if cancelURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "STRIPE_CHECKOUT_CANCEL_URL not configured"})
		return
	}

	cfg := CheckoutConfig{
		PriceID:         priceID,
		CustomerID:      customerID,
		UserID:          userID,
		SuccessURL:      successURL,
		CancelURL:       cancelURL,
		AllowPromoCodes: true,
		TrialFromPlan:   true,
		Metadata:        metadata,
	}

	session, err := CreateCheckoutSession(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":  session.ID,
		"url": session.URL,
	})
}

type initializeTrialRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// InitializeTrialHandler initializes a 14-day free trial for a user if they don't already have one.
func InitializeTrialHandler(c *gin.Context) {
	var req initializeTrialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id"})
		return
	}

	ctx := c.Request.Context()

	initialized, err := supabase.InitializeTrial(ctx, userID)
	if err != nil {
		log.Printf("❌ [Billing] Failed to initialize trial for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize trial"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"initialized": initialized,
		"message":     map[bool]string{true: "Trial initialized", false: "Trial already exists"}[initialized],
	})
}

type getBillingStatusRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// GetBillingStatusHandler returns the combined billing status for a user,
// taking into account both their organization's subscription and their individual profile.
// Organization subscription takes priority over individual profile.
func GetBillingStatusHandler(c *gin.Context) {
	var req getBillingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user_id"})
		return
	}

	ctx := c.Request.Context()

	billingStatus, err := supabase.GetCombinedBillingStatus(ctx, userID)
	if err != nil {
		log.Printf("❌ [Billing] Failed to get billing status for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get billing status"})
		return
	}

	// Format time fields for JSON response
	response := gin.H{
		"stripe_status": billingStatus.StripeStatus,
		"stripe_plan":   billingStatus.StripePlan,
		"source":        billingStatus.Source,
	}

	if billingStatus.StripeTrialEnd != nil {
		response["stripe_trial_end"] = billingStatus.StripeTrialEnd.Format("2006-01-02T15:04:05Z07:00")
	} else {
		response["stripe_trial_end"] = nil
	}

	if billingStatus.StripeCurrentPeriodEnd != nil {
		response["stripe_current_period_end"] = billingStatus.StripeCurrentPeriodEnd.Format("2006-01-02T15:04:05Z07:00")
	} else {
		response["stripe_current_period_end"] = nil
	}

	if billingStatus.OrganizationID != "" {
		response["organization_id"] = billingStatus.OrganizationID
	}

	c.JSON(http.StatusOK, response)
}
