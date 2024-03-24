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

func CreateToken(id string, userName string, tokenType string) (string, error) {
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
			"exp":      time.Now().Add(expTime).Unix(),
			"id":       id,
			"userName": userName,
		})

	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetGithubData(code string) (models.GithubUserData, error) {
	var cfg = config.GetInstance()
	var githubData = models.GithubUserData{}
	var githubJWTData = models.GithubTokenResponse{}
	var httpClient = &http.Client{}

	ghJWTReqBody := map[string]string{
		"client_id":     cfg.GithubClientId,
		"client_secret": cfg.GithubClientSecret,
		"code":          code,
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
