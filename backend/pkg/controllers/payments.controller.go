package controllers

import (
	"fmt"
	"net/http"
	"signalone/cmd/config"
	"signalone/pkg/models"
	"signalone/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	billingPortal "github.com/stripe/stripe-go/v76/billingportal/session"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/subscription"
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

	stripePriceId := cfg.StripePriceIdProd
	if cfg.Mode == "local" {
		stripePriceId = cfg.StripePriceIdTest
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

	stripeCustomer, _ := utils.HandleStripeCustomer(user.UserCustomerId)

	checkoutSessionParams := &stripe.CheckoutSessionParams{
		CancelURL: stripe.String(requestData.CancelUrl),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePriceId),
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(30),
		},
		SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", requestData.SuccessUrl)),
	}
	if stripeCustomer != nil {
		checkoutSessionParams.Customer = stripe.String(user.UserCustomerId)
	} else if user.Email != "" {
		checkoutSessionParams.CustomerEmail = stripe.String(user.Email)
	}

	checkoutSessionParams.AddExpand("customer")

	checkoutSession, err := session.New(checkoutSessionParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating checkout session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": checkoutSession.ID})
}

func (pc *PaymentController) StripeCheckoutCompleteHandler(ctx *gin.Context) {
	const CheckoutExpirationTimeDelta = 172800
	var user models.User

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

	successfulCheckoutSessionId := ctx.Query("session_id")

	successfulCheckoutSession, err := session.Get(successfulCheckoutSessionId, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "VERSION_UPGRADE_ERROR"})
		return
	}
	if ((successfulCheckoutSession.Created - time.Now().Unix()) * -1) > CheckoutExpirationTimeDelta {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "VERSION_UPGRADE_ERROR"})
		return
	}
	if successfulCheckoutSession.Status != "complete" {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "VERSION_UPGRADE_ERROR"})
		return
	}

	stripeSubscription, err := subscription.Get(successfulCheckoutSession.Subscription.ID, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "VERSION_UPGRADE_ERROR"})
		return
	}

	pc.usersCollection.UpdateOne(ctx, bson.M{"userId": user.UserId},
		bson.M{"$set": bson.M{
			"isPro":            true,
			"userCustomerId":   successfulCheckoutSession.Customer.ID,
			"proTierProductId": stripeSubscription.Items.Data[0].Price.Product.ID,
		},
		})

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (pc *PaymentController) StripeCreateBillingPortalHandler(ctx *gin.Context) {
	const BILLING_PORTAL_RETURN_URL = "http://localhost:37001"
	var user models.User

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

	stripeCustomer, err := utils.HandleStripeCustomer(user.UserCustomerId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error getting stripe customer"})
	}

	if stripeCustomer == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "no stripe customer found"})
		return
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(user.UserCustomerId),
		ReturnURL: stripe.String(BILLING_PORTAL_RETURN_URL),
	}

	session, err := billingPortal.New(params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creating billing portal session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": session.URL})
}
