package utils

import (
	"signalone/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleStripeCustomer(customerId string) (*stripe.Customer, error) {
	stripeCustomer, err := customer.Get(customerId, &stripe.CustomerParams{})
	if err != nil {
		return nil, err
	}
	return stripeCustomer, nil
}

func VerifyProTierSubscription(ctx *gin.Context, user models.User, usersCollection *mongo.Collection) {
	var proConfirmed bool = false
	stripeCustomer, err := HandleStripeCustomer(user.UserCustomerId)
	if err != nil {
		proConfirmed = false
	} else {
		for _, subscription := range stripeCustomer.Subscriptions.Data {
			if subscription.Status == "active" &&
				subscription.Items.Data[0].Price.Product.ID == user.ProTierProductId {
				proConfirmed = true
				break
			}
		}
	}
	if !proConfirmed {
		usersCollection.UpdateOne(ctx, bson.M{"userId": user.UserId},
			bson.M{
				"$set": bson.M{
					"isPro":            false,
					"proTierProductId": "",
				},
			})
	}
}
