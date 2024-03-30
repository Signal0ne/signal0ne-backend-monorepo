package controllers

import (
	"net/http"
	"signalone/cmd/config"
	"signalone/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentController struct {
	paymentsCollection *mongo.Collection
}

func NewPaymentController(paymentsCollection *mongo.Collection) *PaymentController {
	return &PaymentController{
		paymentsCollection: paymentsCollection,
	}
}

func (pc *PaymentController) UpgradeProHandler(ctx *gin.Context) {
	cfg := config.GetInstance()
	stripe.Key = cfg.StripeApiKey

	var requestData models.PaymentRequest

	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checkoutSessionParams := &stripe.CheckoutSessionParams{
		CancelURL: stripe.String(requestData.CancelUrl),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price: stripe.String("price_1OzpaA06sWxzpbKv9SaOU2gu"),
				// For metered billing, do not pass quantity
				Quantity: stripe.Int64(1),
			},
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(requestData.SuccessUrl),
	}

	checkoutSession, err := session.New(checkoutSessionParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating checkout session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": checkoutSession.ID})
}
