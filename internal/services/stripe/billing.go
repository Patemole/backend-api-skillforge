package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	supabaseSvc "backend-api-skillforge/internal/supabase"
	stripeSvc "backend-api-skillforge/internal/services/stripe"
)

type createCheckoutRequest struct {
	Plan       string            `json:"plan" binding:"required"`
	CustomerID string            `json:"customer_id"` // optional, pass if you already created one
	Metadata   map[string]string `json:"metadata"`    // optional extra tags
}

// map plan names to their Stripe price IDs (could live in env or Supabase)
var planPriceMap = map[string]string{
	"pro":      os.Getenv("STRIPE_PRO_PRICE"),
	"business": os.Getenv("STRIPE_BUSINESS_PRICE"),
}

// CreateCheckoutSessionHandler starts a Stripe Checkout session for a subscription.
func CreateCheckoutSessionHandler(c *gin.Context) {
	var req createCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}
	// Get user ID first (before using it)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	ctx := c.Request.Context()
	
	// Try to fetch existing Stripe customer ID from Supabase
	profileStripe, err := supabase.GetProfileStripeFields(ctx, userID)
	customerID := ""
	if err != nil {
		// If error is "not found" (PGRST116), that's okay - user might not have a Stripe customer yet
		// For other errors, log but continue (we'll create a new customer if needed)
		if !strings.Contains(err.Error(), "PGRST116") {
			// Log non-404 errors but don't fail - we can still create checkout
			log.Printf("Error fetching Stripe customer: %v", err)
		}
	} else if profileStripe != nil && profileStripe.StripeCustomerID != "" {
		customerID = profileStripe.StripeCustomerID
	}
	
	priceID, ok := planPriceMap[req.Plan]
	if !ok || priceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown plan"})
		return
	}

	// Build metadata - include plan and merge any additional metadata from request
	metadata := make(map[string]string)
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}
	metadata["plan"] = req.Plan

	cfg := stripeSvc.CheckoutConfig{
		PriceID:         priceID,
		CustomerID:      req.CustomerID,
		UserID:          userID,
		SuccessURL:      os.Getenv("STRIPE_CHECKOUT_SUCCESS_URL"),
		CancelURL:       os.Getenv("STRIPE_CHECKOUT_CANCEL_URL"),
		AllowPromoCodes: true,
		TrialFromPlan:   true,
		Metadata:        req.Metadata,
	}

	session, err := stripeSvc.CreateCheckoutSession(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":  session.ID,
		"url": session.URL,
	})
}