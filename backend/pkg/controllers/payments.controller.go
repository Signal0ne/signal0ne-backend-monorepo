package controllers

import (
	"net/http"
	"signalone/cmd/config"
	"signalone/pkg/models"
	"signalone/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentController struct {
	paymentsCollection *mongo.Collection
	usersCollection    *mongo.Collection
}

func NewPaymentController(paymentsCollection *mongo.Collection, usersCollection *mongo.Collection) *PaymentController {
	return &PaymentController{
		paymentsCollection: paymentsCollection,
		usersCollection:    usersCollection,
	}
}

func (pc *PaymentController) UpgradeProHandler(ctx *gin.Context) {
	var requestData models.PaymentRequest
	var user models.User

	cfg := config.GetInstance()
	stripe.Key = cfg.StripeApiKeyProd

	if cfg.Mode == "local" {
		stripe.Key = cfg.StripeApiKeyTest
	}

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = utils.GetUser(ctx, pc.usersCollection, bson.M{"userId": userId}, &user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stripeProductId := cfg.StripeProductIdProd

	if cfg.Mode == "local" {
		stripeProductId = cfg.StripeProductIdTest
	}

	//TODO: Check if user exists in the Stripe customer list
	// params := &stripe.CustomerListParams{
	// 	Email: stripe.String("abc2@gmail.com"),
	// }
	// result := customer.List(params)

	// fmt.Printf("result: %+v\n", result)

	// customerParams := &stripe.CustomerParams{
	// 	Email: stripe.String("abc@gmail.com"),
	// }

	// customer, err := customer.New(customerParams)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating customer"})
	// 	return
	// }

	checkoutSessionParams := &stripe.CheckoutSessionParams{
		CancelURL: stripe.String(requestData.CancelUrl),
		// Customer: stripe.String(customer.ID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripeProductId),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", requestData.SuccessUrl)),
	}

	checkoutSessionParams.AddExpand("customer")

	// result, err := session.Get("{{SESSION_ID}}", checkoutSessionParams)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating checkout session"})
	// 	return
	// }

	checkoutSession, err := session.New(checkoutSessionParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating checkout session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": checkoutSession.ID})
}
