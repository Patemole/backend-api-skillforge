package stripe

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	stripeapi "github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"

	"backend-api-skillforge/internal/supabase"
)

func StripeWebhook(c *gin.Context) {
	const maxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		log.Println("STRIPE_WEBHOOK_SECRET is missing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration error"})
		return
	}

	event, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		log.Printf("stripe webhook signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	ctx := c.Request.Context()

	switch event.Type {

	case "checkout.session.completed":
		var sess stripeapi.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			log.Printf("unmarshal session: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
			return
		}

		userID := strings.TrimSpace(sess.ClientReferenceID)
		if userID == "" && sess.Metadata != nil {
			userID = strings.TrimSpace(sess.Metadata["user_id"])
		}
		if strings.TrimSpace(userID) == "" {
			log.Printf("checkout.session.completed missing user reference (session %s)", sess.ID)
			c.Status(http.StatusOK)
			return
		}

		plan := ""
		if sess.Metadata != nil {
			plan = sess.Metadata["plan"]
		}

		rec := supabase.ProfileStripeFields{
			UserID:               userID,
			StripeCustomerID:     stringFromCheckoutCustomer(sess.Customer),
			StripeSubscriptionID: stringFromCheckoutSubscription(sess.Subscription),
			StripePlan:           plan,
			StripeLastEventID:    event.ID,
		}

		if err := supabase.UpsertProfileStripeFields(ctx, rec); err != nil {
			log.Printf("supabase upsert (checkout.session.completed) failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist billing state"})
			return
		}

		// TODO: handle session, update customer/subscription in Supabase
	case "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripeapi.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("unmarshal subscription: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription"})
			return
		}
		// TODO: update status in Supabase
	case "invoice.paid", "invoice.payment_failed":
		var inv stripeapi.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			log.Printf("unmarshal invoice: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice"})
			return
		}
		// TODO: mark payment success/failure
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	c.Status(http.StatusOK)
}

func stringFromCheckoutCustomer(c *stripeapi.Customer) string {
	if c == nil {
		return ""
	}
	return c.ID
}

func stringFromCheckoutSubscription(s *stripeapi.Subscription) string {
	if s == nil {
		return ""
	}
	return s.ID
}
