package utils

import (
	"signalone/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
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

	subscriptions := subscription.List(&stripe.SubscriptionListParams{Customer: stripe.String(user.UserCustomerId)})
	for _, sub := range subscriptions.SubscriptionList().Data {
		stripeSubscription, _ := subscription.Get(sub.ID, nil)
		if stripeSubscription.Status == "active" &&
			stripeSubscription.Items.Data[0].Price.Product.ID == user.ProTierProductId {
			proConfirmed = true
			break
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
