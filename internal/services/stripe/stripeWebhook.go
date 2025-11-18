package stripe

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	stripeapi "github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/subscription"
	"github.com/stripe/stripe-go/v78/webhook"

	"backend-api-skillforge/internal/supabase"
)

func StripeWebhook(c *gin.Context) {
	log.Printf("🔔 Webhook received - Method: %s, Path: %s", c.Request.Method, c.Request.URL.Path)

	const maxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("❌ Failed to read webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	log.Printf("📝 Webhook signature header present: %v", signature != "")

	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		log.Println("❌ STRIPE_WEBHOOK_SECRET is missing")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration error"})
		return
	}

	event, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		log.Printf("❌ Stripe webhook signature verification failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	log.Printf("✅ Webhook event received - Type: %s, ID: %s", event.Type, event.ID)
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
		// Get organization_id from user_id
		log.Printf("🔍 Processing checkout.session.completed for user %s, plan: %s", userID, plan)
		organizationID, err := supabase.GetOrganizationIDFromUserID(ctx, userID)
		if err != nil {
			log.Printf("❌ failed to get organization_id for user %s: %v", userID, err)
			// Continue with user-level update as fallback
		} else if organizationID == "" {
			log.Printf("⚠️ organization_id is empty for user %s", userID)
		} else {
			log.Printf("✅ Found organization_id %s for user %s", organizationID, userID)
			// Extract customer ID - Customer field is a pointer to Customer object
			customerID := ""
			if sess.Customer != nil {
				customerID = sess.Customer.ID
			}

			// Extract subscription ID - Subscription field is a pointer to Subscription object
			subscriptionID := ""
			if sess.Subscription != nil {
				subscriptionID = sess.Subscription.ID
			}

			log.Printf("📋 Extracted customer_id: %s, subscription_id: %s", customerID, subscriptionID)

			orgStripeFields := supabase.OrganizationStripeFields{
				OrganizationID:       organizationID,
				StripeCustomerID:     customerID,
				StripeSubscriptionID: subscriptionID,
				StripePlan:           plan,
				StripeStatus:         "active", // Default to active for new subscriptions
				StripeLastEventID:    event.ID,
			}
			if subscriptionID != "" {
				// Fetch subscription details from Stripe API to get full details
				sub, err := subscription.Get(subscriptionID, nil)
				if err != nil {
					log.Printf("failed to fetch subscription %s from Stripe: %v (using session data)", subscriptionID, err)
				} else {
					// Extract plan from subscription metadata if not already set
					if plan == "" && sub.Metadata != nil {
						plan = sub.Metadata["plan"]
						orgStripeFields.StripePlan = plan
					}

					// Update with full subscription details
					orgStripeFields.StripeStatus = string(sub.Status)
					orgStripeFields.StripeCancelAtPeriodEnd = sub.CancelAtPeriodEnd

					// Handle time fields
					if sub.TrialEnd > 0 {
						trialEnd := time.Unix(sub.TrialEnd, 0).UTC()
						orgStripeFields.StripeTrialEnd = &trialEnd
					}
					if sub.CurrentPeriodEnd > 0 {
						periodEnd := time.Unix(sub.CurrentPeriodEnd, 0).UTC()
						orgStripeFields.StripeCurrentPeriodEnd = &periodEnd
					}
				}
			}

			// Update organization Stripe fields
			log.Printf("💾 Attempting to update organization %s with Stripe fields...", organizationID)
			if err := supabase.UpsertOrganizationStripeFields(ctx, orgStripeFields); err != nil {
				log.Printf("❌ supabase upsert organization (checkout.session.completed) failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist organization billing state"})
				return
			}
			log.Printf("✅ Successfully updated organization %s with subscription plan %s", organizationID, plan)
		}

		// Also update user profile for backward compatibility
		customerID := ""
		if sess.Customer != nil {
			customerID = sess.Customer.ID
		}

		subscriptionID := ""
		if sess.Subscription != nil {
			subscriptionID = sess.Subscription.ID
		}

		rec := supabase.ProfileStripeFields{
			UserID:               userID,
			StripeCustomerID:     customerID,
			StripeSubscriptionID: subscriptionID,
			StripePlan:           plan,
			StripeLastEventID:    event.ID,
		}

		if err := supabase.UpsertProfileStripeFields(ctx, rec); err != nil {
			log.Printf("supabase upsert profile (checkout.session.completed) failed: %v", err)
			// Don't fail the request if user update fails, org update succeeded
		}

	case "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripeapi.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("unmarshal subscription: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription"})
			return
		}
		// Find organization by subscription_id
		// We need to query organizations table to find which org has this subscription
		// For now, we'll try to get it from the subscription metadata or customer metadata
		userID := ""
		if sub.Metadata != nil {
			userID = sub.Metadata["user_id"]
		}
		if userID == "" && sub.Customer != nil {
			// Try to get user_id from customer metadata
			cust, err := customer.Get(sub.Customer.ID, nil)
			if err == nil && cust.Metadata != nil {
				userID = cust.Metadata["user_id"]
			}
		}
		if userID != "" {
			organizationID, err := supabase.GetOrganizationIDFromUserID(ctx, userID)
			if err == nil && organizationID != "" {
				// Extract plan from subscription metadata
				plan := ""
				if sub.Metadata != nil {
					plan = sub.Metadata["plan"]
				}

				customerID := ""
				if sub.Customer != nil {
					customerID = sub.Customer.ID
				}
				orgStripeFields := supabase.OrganizationStripeFields{
					OrganizationID:          organizationID,
					StripeCustomerID:        customerID,
					StripeSubscriptionID:    sub.ID,
					StripePlan:              plan,
					StripeStatus:            string(sub.Status),
					StripeCancelAtPeriodEnd: sub.CancelAtPeriodEnd,
					StripeLastEventID:       event.ID,
				}
				// Handle time fields
				if sub.TrialEnd > 0 {
					trialEnd := time.Unix(sub.TrialEnd, 0).UTC()
					orgStripeFields.StripeTrialEnd = &trialEnd
				}
				if sub.CurrentPeriodEnd > 0 {
					periodEnd := time.Unix(sub.CurrentPeriodEnd, 0).UTC()
					orgStripeFields.StripeCurrentPeriodEnd = &periodEnd
				}
				if err := supabase.UpsertOrganizationStripeFields(ctx, orgStripeFields); err != nil {
					log.Printf("supabase upsert organization (subscription.updated) failed: %v", err)
				} else {
					log.Printf("✅ Updated organization %s subscription status to %s", organizationID, sub.Status)
				}
			}
		}
	case "invoice.paid", "invoice.payment_failed":
		var inv stripeapi.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			log.Printf("unmarshal invoice: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice"})
			return
		}

		// Update invoice status for organization
		if inv.Subscription != nil {
			sub, err := subscription.Get(inv.Subscription.ID, nil)
			if err == nil {
				userID := ""
				if sub.Metadata != nil {
					userID = sub.Metadata["user_id"]
				}

				if userID != "" {
					organizationID, err := supabase.GetOrganizationIDFromUserID(ctx, userID)
					if err == nil && organizationID != "" {
						orgStripeFields := supabase.OrganizationStripeFields{
							OrganizationID:          organizationID,
							StripeLastInvoiceStatus: string(inv.Status),
							StripeLastEventID:       event.ID,
						}

						// Only update invoice status, preserve other fields
						updateData := map[string]interface{}{
							"stripe_last_invoice_status": orgStripeFields.StripeLastInvoiceStatus,
							"stripe_last_event_id":       orgStripeFields.StripeLastEventID,
						}

						_, _, err = supabase.Client.
							From("organizations").
							Update(updateData, "", "").
							Eq("id", organizationID).
							Execute()

						if err != nil {
							log.Printf("supabase upsert organization (invoice) failed: %v", err)
						} else {
							log.Printf("✅ Updated organization %s invoice status to %s", organizationID, inv.Status)
						}
					}
				}
			}
		}
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
