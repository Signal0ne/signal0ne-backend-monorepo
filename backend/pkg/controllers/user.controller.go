package controllers

import (
	"net/http"
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
