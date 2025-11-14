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
	Plan       string            `json:"plan" binding:"required"`
	CustomerID string            `json:"customer_id"`
	UserID     string            `json:"user_id"`
	Metadata   map[string]string `json:"metadata"`
}

func priceIDFromPlan(plan string) string {
	var envVar, priceID string
	switch plan {
	case "pro":
		envVar = "STRIPE_PRO_PRICE"
	case "business":
		envVar = "STRIPE_BUSINESS_PRICE"
	default:
		return ""
	}
	priceID = strings.TrimSpace(os.Getenv(envVar))
	if priceID == "" {
		log.Printf("⚠️ [Billing] Missing price ID for plan=%s (env %s)", plan, envVar)
	}
	return priceID
}

func CreateCheckoutSessionHandler(c *gin.Context) {
	var req createCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	userID := strings.TrimSpace(c.GetString("user_id"))
	if userID == "" {
		userID = strings.TrimSpace(req.UserID)
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	priceID := priceIDFromPlan(req.Plan)
	if priceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown plan"})
		return
	}

	ctx := c.Request.Context()

	customerID := strings.TrimSpace(req.CustomerID)
	if customerID == "" {
		// Try to fetch existing Stripe customer ID from profile, but don't fail if it doesn't exist
		// PGRST116 means "not found" which is fine for new users
		if profileStripe, err := supabase.GetProfileStripeFields(ctx, userID); err == nil {
			customerID = profileStripe.StripeCustomerID
		} else if err != nil && !strings.Contains(err.Error(), "PGRST116") {
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

	cfg := CheckoutConfig{
		PriceID:         priceID,
		CustomerID:      customerID,
		UserID:          userID,
		SuccessURL:      os.Getenv("STRIPE_CHECKOUT_SUCCESS_URL"),
		CancelURL:       os.Getenv("STRIPE_CHECKOUT_CANCEL_URL"),
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
