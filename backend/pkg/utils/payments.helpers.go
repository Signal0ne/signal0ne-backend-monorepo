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
