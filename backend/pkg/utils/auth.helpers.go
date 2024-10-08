package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"signalone/cmd/config"
	"signalone/pkg/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const ACCESS_TOKEN_EXPIRATION_TIME = time.Minute * 10
const REFRESH_TOKEN_EXPIRATION_TIME = time.Hour * 24

func GetUserIdFromToken(ctx *gin.Context) (string, error) {
	bearerToken := ctx.GetHeader("Authorization")

	jwtToken := strings.TrimPrefix(bearerToken, "Bearer ")

	userId, err := VerifyToken(jwtToken)
	if err != nil {
		return "", err
	}
	return userId, nil
}

func CreateToken(user models.User, tokenType string) (string, error) {
	var cfg = config.GetInstance()
	var expTime time.Duration
	var SECRET_KEY = []byte(cfg.SignalOneSecret)

	if tokenType == "refresh" {
		expTime = REFRESH_TOKEN_EXPIRATION_TIME
	} else if tokenType == "access" {
		expTime = ACCESS_TOKEN_EXPIRATION_TIME
	} else {
		expTime = time.Second * 0
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp":                time.Now().Add(expTime).Unix(),
			"id":                 user.UserId,
			"userName":           user.UserName,
			"email":              user.Email,
			"isPro":              user.IsPro,
			"canRateApplication": user.CanRateApplication,
		})

	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetGithubData(githubCodePayload models.GithubTokenRequest) (models.GithubUserData, error) {
	var cfg = config.GetInstance()
	var githubData = models.GithubUserData{}
	var githubJWTData = models.GithubTokenResponse{}
	var httpClient = &http.Client{}
	var client_id string
	var client_secret string

	switch githubCodePayload.Source {
	case "docker":
		client_id = cfg.GithubClientId
		client_secret = cfg.GithubClientSecret
	case "customIdp":
		client_id = cfg.CustomIDPGithubClientId
		client_secret = cfg.CustomIDPGithubSecret
	default:
		client_id = cfg.GithubClientId
		client_secret = cfg.GithubClientSecret
	}

	ghJWTReqBody := map[string]string{
		"client_id":     client_id,
		"client_secret": client_secret,
		"code":          githubCodePayload.Code,
	}

	jsonData, _ := json.Marshal(ghJWTReqBody)

	ghJWTReq, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(jsonData))
	if err != nil {
		return models.GithubUserData{}, err
	}

	ghJWTReq.Header.Set("Accept", "application/json")
	ghJWTReq.Header.Set("Content-Type", "application/json")

	ghJWTResp, err := httpClient.Do(ghJWTReq)
	if err != nil {
		return models.GithubUserData{}, err
	}

	ghJWTRespBody, err := io.ReadAll(ghJWTResp.Body)
	if err != nil {
		return models.GithubUserData{}, err
	}

	err = json.Unmarshal(ghJWTRespBody, &githubJWTData)
	if err != nil {
		return models.GithubUserData{}, err
	}

	ghUserDataReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return models.GithubUserData{}, err
	}

	ghUserDataReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubJWTData.AccessToken))

	ghUserDataResp, err := httpClient.Do(ghUserDataReq)
	if err != nil {
		return models.GithubUserData{}, err
	}

	ghUserDataRespBody, err := io.ReadAll(ghUserDataResp.Body)
	if err != nil {
		return models.GithubUserData{}, err
	}

	err = json.Unmarshal(ghUserDataRespBody, &githubData)
	if err != nil {
		return models.GithubUserData{}, err
	}

	return githubData, nil
}

func GetGooglePublicKey(keyId string) (string, error) {
	var googleData = map[string]string{}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(data, &googleData)
	if err != nil {
		return "", err
	}

	key, ok := googleData[keyId]
	if !ok {
		return "", errors.New("key not found")
	}

	return key, nil
}

func ValidateGoogleJWT(tokenString string) (models.GoogleClaims, error) {
	var cfg = config.GetInstance()
	var claimsStruct = models.GoogleClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
			pem, err := GetGooglePublicKey(fmt.Sprintf("%s", token.Header["kid"]))
			if err != nil {
				return nil, err
			}

			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
			if err != nil {
				return nil, err
			}

			return key, nil
		},
	)

	if err != nil {
		return models.GoogleClaims{}, err
	}

	claims, ok := token.Claims.(*models.GoogleClaims)
	if !ok {
		return models.GoogleClaims{}, errors.New("invalid claims")
	}

	if claims.Issuer != "accounts.google.com" && claims.Issuer != "https://accounts.google.com" {
		return models.GoogleClaims{}, errors.New("iss is invalid")
	}

	audienceToCheck := cfg.GoogleClientId
	found := false

	for _, audience := range claims.Audience {
		if audience == audienceToCheck {
			found = true
			break
		}
	}

	if !found {
		return models.GoogleClaims{}, errors.New("aud is invalid")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Unix() < time.Now().UTC().Unix() {
		return models.GoogleClaims{}, errors.New("jwt is expired")
	}

	return *claims, nil
}

func VerifyToken(tokenString string) (string, error) {
	var cfg = config.GetInstance()
	var claims = &models.JWTClaimsWithUserData{}
	var SECRET_KEY = []byte(cfg.SignalOneSecret)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	return claims.Id, nil
}

func GetTokenExpirationDateInUnixFormat(tokenString string) int64 {
	var cfg = config.GetInstance()
	var claims = &models.JWTClaimsWithUserData{}
	var SECRET_KEY = []byte(cfg.SignalOneSecret)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})
	if err != nil {
		return -1
	}

	expirationTimestamp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return -1
	}

	return expirationTimestamp.Unix()
}

func VerifyRatingAbility(ctx *gin.Context, user models.User, issuesCollection *mongo.Collection, usersCollection *mongo.Collection) {
	const AnalysisNoThreshold = 6
	if !user.CanRateApplication && user.Metrics.OverallScore == 0 {
		filter := bson.M{"userId": user.UserId}
		count, err := issuesCollection.CountDocuments(ctx, filter)
		if err != nil {
			return
		}

		if count >= AnalysisNoThreshold {
			usersCollection.UpdateOne(ctx, bson.M{"userId": user.UserId}, bson.M{"$set": bson.M{"canRateApplication": true}})
		}
	}
}
