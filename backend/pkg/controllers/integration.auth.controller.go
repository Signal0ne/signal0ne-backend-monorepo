package controllers

import (
	"errors"
	"net/http"
	"signalone/cmd/config"
	"signalone/pkg/models"
	"signalone/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IntegrationAuthController struct {
	issuesCollection        *mongo.Collection
	usersCollection         *mongo.Collection
	analysisStoreCollection *mongo.Collection
}

func NewIntegrationAuthController(issuesCollection *mongo.Collection,
	usersCollection *mongo.Collection,
	analysisStoreCollection *mongo.Collection) *IntegrationAuthController {
	return &IntegrationAuthController{
		issuesCollection:        issuesCollection,
		usersCollection:         usersCollection,
		analysisStoreCollection: analysisStoreCollection,
	}
}

func (c *IntegrationAuthController) AuthenticateAgent(ctx *gin.Context) {
	var user models.User

	userId, err := utils.GetUserIdFromToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = utils.GetUser(ctx, c.usersCollection, bson.M{"userId": userId}, &user)
	if err != nil {
		return
	}

	if user.AgentBearerToken != "" {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"token":   user.AgentBearerToken,
		})
		return
	}

	token, err := utils.CreateToken(userId, user.UserName, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = c.usersCollection.UpdateOne(ctx,
		bson.M{"userId": userId},
		bson.M{"$set": bson.M{
			"agentBearerToken": token,
		},
		},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"token":   token,
	})
}

func (c *IntegrationAuthController) CheckAgentAuthorization(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	var token = strings.TrimPrefix(authHeader, "Bearer ")

	err := c.VerifyAgentToken(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.Next()
}

func (c *IntegrationAuthController) VerifyAgentToken(ctx *gin.Context, token string) (err error) {
	var user models.User
	var cfg = config.GetInstance()
	var claims = &models.JWTClaimsWithUserData{}
	var SECRET_KEY = []byte(cfg.SignalOneSecret)

	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
	)

	_, err = parser.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})
	if err != nil {
		return
	}

	err = utils.GetUser(ctx, c.usersCollection, bson.M{"userId": claims.Id}, &user)
	if err != nil {
		return
	}

	if user.AgentBearerToken == "" || user.AgentBearerToken != token {
		err = errors.New("unauthorized")
		return
	}

	return
}
