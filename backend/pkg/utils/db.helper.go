package utils

import (
	"signalone/pkg/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUser(ctx *gin.Context, usersCollection *mongo.Collection, filter bson.M, user *models.User) error {
	result := usersCollection.FindOne(ctx, filter)
	err := result.Decode(&user)
	return err
}
