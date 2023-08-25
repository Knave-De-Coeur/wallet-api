package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"wallet-api/internal/api"
	"wallet-api/internal/pkg"
)

func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		authSplit := strings.Split(auth, "Bearer ")
		if len(authSplit) < 2 {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				api.GenerateMessageResponse("no token", nil, errors.New("missing token in request")),
			)
			return
		}
		tokenString := authSplit[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				api.GenerateMessageResponse("no token", nil, errors.New("missing token in request")),
			)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(pkg.UnexpectedMethod, token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				api.GenerateMessageResponse("something went wrong with teh token", nil, err),
			)
			return
		}

		var userID int
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, _ = strconv.Atoi(fmt.Sprint(claims["sub"]))
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				c.AbortWithStatusJSON(
					http.StatusForbidden,
					api.GenerateMessageResponse("expired token", nil, errors.New("token is no longer valid")),
				)
				return
			}
		} else {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				api.GenerateMessageResponse("bad token", nil, errors.New("token is not valid")),
			)
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}
