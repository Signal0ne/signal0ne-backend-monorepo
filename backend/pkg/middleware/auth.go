package middlewares

import (
	"net/http"
	"signalone/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func CheckAuthorization(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	var jwtToken = strings.TrimPrefix(authHeader, "Bearer ")

	_, err := utils.VerifyToken(jwtToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.Next()
}
