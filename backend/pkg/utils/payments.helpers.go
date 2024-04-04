package utils

import (
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
)

func HandleStripeCustomer(customerId string) (*stripe.Customer, error) {
	stripeCustomer, err := customer.Get(customerId, &stripe.CustomerParams{})
	if err != nil {
		return nil, err
	}
	return stripeCustomer, nil
}

func VerifyProTierSubscription(customerId string) bool {
	var proConfirmed bool = false
	stripeCustomer, err := HandleStripeCustomer(customerId)
	if err != nil {
		proConfirmed = false
	} else {
		for _, subscription := range stripeCustomer.Subscriptions.Data {
			if subscription.Status == "active" {
				proConfirmed = true
				break
			}
		}
	}
	return proConfirmed
}
