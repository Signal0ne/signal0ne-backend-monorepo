package controllers

import (
	"fmt"
	"net/http"
	"signalone/cmd/config"
	"signalone/pkg/models"
	"signalone/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	e "github.com/jordan-wright/email"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserAuthController struct {
	usersCollection *mongo.Collection
	emailClientData EmailClientConfig
}

func NewUserAuthController(usersCollection *mongo.Collection,
	emailClientData EmailClientConfig) *UserAuthController {
	return &UserAuthController{
		usersCollection: usersCollection,
		emailClientData: emailClientData,
	}
}

// Auth Handlers
func (c *UserAuthController) LoginWithGithubHandler(ctx *gin.Context) {
	var requestData models.GithubTokenRequest
	var user models.User

	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userData, err = utils.GetGithubData(requestData.Code)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = utils.GetUser(ctx, c.usersCollection, bson.M{"userId": strconv.Itoa(userData.Id)}, &user)

	if user.IsPro {
		utils.VerifyProTierSubscription(ctx, user.UserCustomerId, user.UserId, c.usersCollection)
	}

	if err != nil && err.Error() != mongo.ErrNoDocuments.Error() {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err != nil && err.Error() == mongo.ErrNoDocuments.Error() {
		user = models.User{
			UserId:           strconv.Itoa(userData.Id),
			UserName:         userData.Login,
			Email:            userData.Email,
			IsPro:            false,
			AgentBearerToken: "",
			Counter:          0,
			Type:             "github",
		}

		_, err = c.usersCollection.InsertOne(ctx, user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if user.Email != userData.Email {
		_, err = c.usersCollection.UpdateOne(ctx,
			bson.M{"userId": strconv.Itoa(userData.Id)},
			bson.M{"$set": bson.M{
				"email": userData.Email,
			},
			},
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	accessTokenString, err := utils.CreateToken(user.UserId, user.UserName, "access")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't make authentication token"})
		return
	}

	refreshTokenString, err := utils.CreateToken(user.UserId, user.UserName, "refresh")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't make authentication token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"accessToken":  accessTokenString,
		"expiresIn":    int64(utils.ACCESS_TOKEN_EXPIRATION_TIME) / int64(time.Second),
		"refreshToken": refreshTokenString,
	})
}

func (c *UserAuthController) LoginWithGoogleHandler(ctx *gin.Context) {
	var requestData models.GoogleTokenRequest
	var user models.User

	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ValidateGoogleJWT(requestData.IdToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = utils.GetUser(ctx, c.usersCollection, bson.M{"userId": claims.Subject}, &user)
	if err != nil && err.Error() != mongo.ErrNoDocuments.Error() {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err != nil && err.Error() == mongo.ErrNoDocuments.Error() {
		user = models.User{
			UserId:           claims.Subject,
			UserName:         claims.Email,
			Email:            claims.Email,
			IsPro:            false,
			AgentBearerToken: "",
			Counter:          0,
			Type:             "google",
		}

		_, err = c.usersCollection.InsertOne(ctx, user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if user.IsPro {
		utils.VerifyProTierSubscription(ctx, user.UserCustomerId, user.UserId, c.usersCollection)
	}

	if user.Email != claims.Email {
		_, err = c.usersCollection.UpdateOne(ctx,
			bson.M{"userId": claims.Subject},
			bson.M{"$set": bson.M{
				"email":    claims.Email,
				"userName": claims.Email,
			},
			},
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	accessTokenString, err := utils.CreateToken(user.UserId, user.UserName, "access")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't make authentication token"})
		return
	}

	refreshTokenString, err := utils.CreateToken(user.UserId, user.UserName, "refresh")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't make authentication token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"accessToken":  accessTokenString,
		"expiresIn":    int64(utils.ACCESS_TOKEN_EXPIRATION_TIME) / int64(time.Second),
		"refreshToken": refreshTokenString,
	})
}

func (c *UserAuthController) LoginHandler(ctx *gin.Context) {
	var loginData models.SignalAccountRequest
	var user models.User

	if err := ctx.ShouldBindJSON(&loginData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	loginData.Email = strings.ToLower(loginData.Email)

	err := utils.GetUser(ctx, c.usersCollection, bson.M{"userName": loginData.Email, "type": "signalone"}, &user)
	if err == mongo.ErrNoDocuments {
		ctx.JSON(http.StatusUnauthorized, gin.H{"descriptionKey": "INVALID_CREDENTIALS"})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	if user.IsPro {
		utils.VerifyProTierSubscription(ctx, user.UserCustomerId, user.UserId, c.usersCollection)
	}

	if !utils.ComparePasswordHashes(user.PasswordHash, loginData.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"descriptionKey": "INVALID_CREDENTIALS"})
		return
	}

	if !user.EmailConfirmed {
		ctx.JSON(http.StatusUnauthorized, gin.H{"descriptionKey": "ACCOUNT_NOT_ACTIVE"})
		return
	}

	accessTokenString, err := utils.CreateToken(user.UserId, user.UserName, "access")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	refreshTokenString, err := utils.CreateToken(user.UserId, user.UserName, "refresh")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	if user.Email == "" {
		c.usersCollection.UpdateOne(ctx, bson.M{"userId": user.UserId},
			bson.M{
				"$set": bson.M{
					"email": loginData.Email},
			},
		)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"accessToken":  accessTokenString,
		"expiresIn":    int64(utils.ACCESS_TOKEN_EXPIRATION_TIME) / int64(time.Second),
		"refreshToken": refreshTokenString,
	})

}

func (c *UserAuthController) RegisterHandler(ctx *gin.Context) {
	var loginData models.SignalAccountRequest
	var user models.User

	if err := ctx.ShouldBindJSON(&loginData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	loginData.Email = strings.ToLower(loginData.Email)

	if !utils.PasswordValidation(loginData.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "INVALID_PASSWORD"})
		return
	}

	err := utils.GetUser(ctx, c.usersCollection, bson.M{"userName": loginData.Email, "type": "signalone"}, &user)
	if err != nil && err != mongo.ErrNoDocuments {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}
	if err == nil {
		if !user.EmailConfirmed {
			ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "DUPLICATE_NOT_CONFIRMED_USER_EMAIL"})
			return
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "EMAIL_ALREADY_IN_USE"})
			return
		}
	}

	hashedPassword, err := utils.HashPassword(loginData.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	userId := uuid.New().String()
	confirmationToken, err := utils.CreateToken(userId, loginData.Email, "refresh")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	confirmationLink := fmt.Sprintf("https://signaloneai.com/email-verification?email=%s&verificationCode=%s", loginData.Email, confirmationToken)

	emailObj := e.NewEmail()
	emailObj.From = c.emailClientData.From
	emailObj.To = []string{loginData.Email}
	emailObj.Subject = "Confirm your email address"
	emailObj.HTML = []byte(fmt.Sprintf(`<img alt="Signal0ne" title="Signal0ne Logo" width="196px" height="57px" src="https://signaloneai.com/online-assets/Signal0ne.jpg"
	style="margin-top: 40px;">
	<h1 style="color: black">Hello,</h1> 
	<p style="color: black">Welcome to <span style="font-family: consolas;">Signal0ne!</span> We're excited you're joining us.</p>
	<p style="color: black">Ready to get started? First, verify your email address by clicking the button below: </p>
	<a href="%s" target="_blank"><button style="background-color: #3f51b5; border: none; border-radius: 6px; color: #ffffff; cursor: pointer; padding-bottom: 8px; padding-top: 8px; padding-left: 16px; padding-right: 16px;">Confirm</button></a><br>
	<p>Or you can click the following link: <a href="%s" target="_blank" style="color: #3f51b5">%s</a></p>
	<p style="color: black; margin-bottom: 0; margin-top: 4px;">Best regards,</p>
	<p style="color: black; font-family: consolas; font-size: 15px; font-weight: bold; margin-top: 6px;";>Signal0ne Team</p>`, confirmationLink, confirmationLink, confirmationLink))

	err = emailObj.SendWithStartTLS(c.emailClientData.HostAddress, c.emailClientData.AuthData, c.emailClientData.TlsConfig)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "INVALID_EMAIL"})
	}

	user = models.User{
		UserId:                userId,
		UserName:              loginData.Email,
		Email:                 loginData.Email,
		PasswordHash:          hashedPassword,
		IsPro:                 false,
		AgentBearerToken:      "",
		Counter:               0,
		Type:                  "signalone",
		EmailConfirmed:        false,
		EmailConfirmationCode: confirmationToken,
	}
	_, err = c.usersCollection.InsertOne(ctx, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})

}

func (c *UserAuthController) VerifyEmail(ctx *gin.Context) {
	var user models.User
	var verificationData models.EmailConfirmationRequest

	if err := ctx.ShouldBindJSON(&verificationData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	_, err := utils.VerifyToken(verificationData.ConfirmationToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_VERIFICATION_TOKEN"})
		return
	}

	err = utils.GetUser(ctx, c.usersCollection, bson.M{"emailConfirmationCode": verificationData.ConfirmationToken}, &user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "INVALID_VERIFICATION_CODE"})
		return
	}

	_, err = c.usersCollection.UpdateOne(ctx,
		bson.M{"userId": user.UserId},
		bson.M{"$set": bson.M{
			"emailConfirmed":        true,
			"emailConfirmationCode": "",
		},
		},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (c *UserAuthController) ResendConfirmationEmail(ctx *gin.Context) {
	var user models.User
	var verificationData struct {
		Email string `json:"email"`
	}

	if err := ctx.ShouldBindJSON(&verificationData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	err := utils.GetUser(ctx, c.usersCollection, bson.M{"userName": verificationData.Email}, &user)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "EMAIL_NOT_FOUND"})
		return
	}

	if user.EmailConfirmed {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "EMAIL_ALREADY_IN_USE"})
		return
	}

	confirmationToken, err := utils.CreateToken(user.UserId, user.UserName, "refresh")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	_, err = c.usersCollection.UpdateOne(ctx,
		bson.M{"userId": user.UserId},
		bson.M{"$set": bson.M{
			"emailConfirmationCode": confirmationToken,
		},
		},
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"descriptionKey": "ERROR_OCCURED"})
		return
	}

	confirmationLink := fmt.Sprintf("https://signaloneai.com/email-verification?email=%s&verificationCode=%s", verificationData.Email, confirmationToken)

	emailObj := e.NewEmail()
	emailObj.From = c.emailClientData.From
	emailObj.To = []string{verificationData.Email}
	emailObj.Subject = "Confirm your email address"
	emailObj.HTML = []byte(fmt.Sprintf(`<img alt="Signal0ne" title="Signal0ne Logo" width="196px" height="57px" src="https://signaloneai.com/online-assets/Signal0ne.jpg"
	style="margin-top: 40px;">
	<h1 style="color: black">Hello,</h1> 
	<p style="color: black">Welcome to <span style="font-family: consolas;">Signal0ne!</span> We're excited you're joining us.</p>
	<p style="color: black">Ready to get started? First, verify your email address by clicking the button below: </p>
	<a href="%s" target="_blank"><button style="background-color: #3f51b5; border: none; border-radius: 6px; color: #ffffff; cursor: pointer; padding-bottom: 8px; padding-top: 8px; padding-left: 16px; padding-right: 16px;">Confirm</button></a><br>
	<p>Or you can click the following link: <a href="%s" target="_blank" style="color: #3f51b5">%s</a></p>
	<p style="color: black; margin-bottom: 0; margin-top: 4px;">Best regards,</p>
	<p style="color: black; font-family: consolas; font-size: 15px; font-weight: bold; margin-top: 6px;";>Signal0ne Team</p>`, confirmationLink, confirmationLink, confirmationLink))
	err = emailObj.SendWithStartTLS(c.emailClientData.HostAddress, c.emailClientData.AuthData, c.emailClientData.TlsConfig)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"descriptionKey": "INVALID_EMAIL"})
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (c *UserAuthController) RefreshTokenHandler(ctx *gin.Context) {
	const RefreshTokenExpirationDeltaThreshold = 7200
	var cfg = config.GetInstance()
	var claims = &models.JWTClaimsWithUserData{}
	var data models.RefreshTokenRequest
	var user models.User
	var refreshTokenString string
	var SECRET_KEY = []byte(cfg.SignalOneSecret)

	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := utils.VerifyToken(data.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	expirationDelta := utils.GetTokenExpirationDateInUnixFormat(data.RefreshToken) - time.Now().Unix()
	if expirationDelta >= 0 && expirationDelta < RefreshTokenExpirationDeltaThreshold {
		err = utils.GetUser(ctx, c.usersCollection, bson.M{"userId": userId}, &user)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if user.IsPro {
			utils.VerifyProTierSubscription(ctx, user.UserCustomerId, user.UserId, c.usersCollection)
		}

		refreshTokenString, err = utils.CreateToken(claims.Id, claims.UserName, "refresh")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't make authentication token"})
			return
		}
	} else {
		refreshTokenString = data.RefreshToken
	}

	token, err := jwt.ParseWithClaims(data.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !token.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	accessTokenString, err := utils.CreateToken(claims.Id, claims.UserName, "access")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't make authentication token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Success",
		"accessToken":  accessTokenString,
		"expiresIn":    int64(utils.ACCESS_TOKEN_EXPIRATION_TIME) / int64(time.Second),
		"refreshToken": refreshTokenString,
	})
}
