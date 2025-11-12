package stripe

import(
	"github.com/stripe/stripe-go/v78"
    "github.com/stripe/stripe-go/v78/customer"
    "os"
	"log"
)

func init() {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		log.Fatal("STRIPE_SECRET_KEY is not set")
	}
	stripe.Key = key
	log.Println("Stripe initialized.")
}

func CreateStripeCustomer(email, userID string, metadata map[string]string) (*stripe.Customer, error) {
	if metadata == nil {
		metadata = make(map[string]string, 1)
	}
	metadata["user_id"] = userID

	params := &stripe.CustomerParams{
		Email:    stripe.String(email),
		Metadata: metadata,
	}

	cus, err := customer.New(params)
	if err != nil {
		log.Printf("Error creating Stripe customer: %v", err)
		return nil, err
	}
	log.Printf("Stripe customer created: %s", cus.ID)
	return cus, nil
}
