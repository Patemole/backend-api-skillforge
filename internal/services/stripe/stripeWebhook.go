package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
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

	event, err := stripe.WebhookConstructEvent(payload, signature, secret)
	if err != nil {
		log.Printf("stripe webhook signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("unmarshal session: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
			return
		}
		// TODO: handle session, update customer/subscription in Supabase
	case "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("unmarshal subscription: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription"})
			return
		}
		// TODO: update status in Supabase
	case "invoice.paid", "invoice.payment_failed":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
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