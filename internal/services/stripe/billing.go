package stripe

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-api-skillforge/internal/supabase"
)

type billingAddress struct {
	Line1      string `json:"line1"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	State      string `json:"state"`
	Country    string `json:"country"`
}

type billingDetails struct {
	Name    string         `json:"name"`
	Email   string         `json:"email"`
	Phone   string         `json:"phone"`
	Address billingAddress `json:"address"`
}

type createCheckoutRequest struct {
	Plan           string            `json:"plan" binding:"required"`
	UserID         string            `json:"user_id" binding:"required"`
	CustomerID     string            `json:"customer_id"`
	BillingPeriod  string            `json:"billing_period"` // "monthly" or "annual"
	CustomerEmail  string            `json:"customer_email"`
	BillingDetails *billingDetails   `json:"billing_details"`
	CompanyName    string            `json:"company_name"`
	VATNumber      string            `json:"vat_number"`
	Metadata       map[string]string `json:"metadata"`
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

	// Add company and VAT to metadata if provided
	if req.CompanyName != "" {
		metadata["company_name"] = req.CompanyName
	}
	if req.VATNumber != "" {
		metadata["vat_number"] = req.VATNumber
	}

	// Convert billing details to the format expected by CheckoutConfig
	// Types BillingDetails and BillingAddress are defined in createCheckout.go (same package)
	var checkoutBillingDetails *BillingDetails
	if req.BillingDetails != nil {
		var addr *BillingAddress
		if req.BillingDetails.Address.Line1 != "" || req.BillingDetails.Address.City != "" {
			addr = &BillingAddress{
				Line1:      req.BillingDetails.Address.Line1,
				City:       req.BillingDetails.Address.City,
				PostalCode: req.BillingDetails.Address.PostalCode,
				State:      req.BillingDetails.Address.State,
				Country:    req.BillingDetails.Address.Country,
			}
		}
		checkoutBillingDetails = &BillingDetails{
			Name:    req.BillingDetails.Name,
			Email:   req.BillingDetails.Email,
			Phone:   req.BillingDetails.Phone,
			Address: addr,
		}
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
		CustomerEmail:   strings.TrimSpace(req.CustomerEmail),
		BillingDetails:  checkoutBillingDetails,
		CompanyName:     strings.TrimSpace(req.CompanyName),
		VATNumber:       strings.TrimSpace(req.VATNumber),
	}

	session, err := CreateCheckoutSession(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create checkout session"})
		return
	}

	// Update user and organization profiles with the selected plan information
	// Status is set to "pending" since payment hasn't been completed yet
	// The webhook will update with actual customer_id, subscription_id, and status later
	log.Printf("📝 [Billing] Updating profiles for user %s with plan %s (pending checkout)", userID, req.Plan)

	// Update user profile
	userProfileFields := supabase.ProfileStripeFields{
		UserID:     userID,
		StripePlan: req.Plan,
		StripeStatus: "pending", // Payment is pending until checkout completes
	}
	if customerID != "" {
		userProfileFields.StripeCustomerID = customerID
	}
	if err := supabase.UpsertProfileStripeFields(ctx, userProfileFields); err != nil {
		log.Printf("⚠️ [Billing] Failed to update user profile for %s: %v (non-critical, continuing)", userID, err)
		// Don't fail the request if profile update fails - checkout session was created successfully
	} else {
		log.Printf("✅ [Billing] Updated user profile for %s with plan %s", userID, req.Plan)
	}

	// Update organization profile
	organizationID, err := supabase.GetOrganizationIDFromUserID(ctx, userID)
	if err != nil {
		log.Printf("⚠️ [Billing] Failed to get organization_id for user %s: %v (non-critical, continuing)", userID, err)
	} else if organizationID != "" {
		orgStripeFields := supabase.OrganizationStripeFields{
			OrganizationID: organizationID,
			StripePlan:     req.Plan,
			StripeStatus:   "pending", // Payment is pending until checkout completes
		}
		if customerID != "" {
			orgStripeFields.StripeCustomerID = customerID
		}
		if err := supabase.UpsertOrganizationStripeFields(ctx, orgStripeFields); err != nil {
			log.Printf("⚠️ [Billing] Failed to update organization profile for %s: %v (non-critical, continuing)", organizationID, err)
			// Don't fail the request if org update fails - checkout session was created successfully
		} else {
			log.Printf("✅ [Billing] Updated organization profile for %s with plan %s", organizationID, req.Plan)
		}
	} else {
		log.Printf("⚠️ [Billing] No organization_id found for user %s (user may not be associated with an organization)", userID)
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
