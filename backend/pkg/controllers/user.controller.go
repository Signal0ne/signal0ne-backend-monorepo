package controllers

import (
	"net/http"
	"signalone/pkg/models"
	"signalone/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	usersCollection *mongo.Collection
}

func NewUserController(usersCollection *mongo.Collection) *UserController {
	return &UserController{
		usersCollection: usersCollection,
	}
}

func (c *UserController) LastActivityHandler(ctx *gin.Context) {
	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	_, err = c.usersCollection.UpdateOne(ctx,
		bson.M{"userId": userId},
		bson.M{"$set": bson.M{
			"lastActivity": time.Now(),
		}},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func (c *UserController) MetricsOverallScoreHandler(ctx *gin.Context) {
	var overallScoreRequest models.OverallScoreRequest

	if err := ctx.ShouldBindJSON(&overallScoreRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"userId": userId}
	update := bson.M{"$set": bson.M{
		"metrics.overallScore": *overallScoreRequest.OverallScore,
	}}

	updatedUserResult, err := c.usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updatedUserResult.MatchedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User cannot be found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func (c *UserController) MetricsProButtonClickHandler(ctx *gin.Context) {
	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "userId", Value: userId}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "metrics.proButtonClicksCount", Value: 1}}}}

	updatedUserResult, err := c.usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updatedUserResult.MatchedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User cannot be found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func (c *UserController) MetricsProCheckoutClickHandler(ctx *gin.Context) {
	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "userId", Value: userId}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "metrics.proCheckoutClicksCount", Value: 1}}}}

	updatedUserResult, err := c.usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if updatedUserResult.MatchedCount == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User cannot be found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Success"})
}
