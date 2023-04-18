package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"wallet-api/internal/api"
)

func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		authSplit := strings.Split(auth, "Bearer ")
		tokenString := authSplit[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				api.GenerateMessageResponse("no token", nil, fmt.Errorf("missing token in request")),
			)
			return
		}

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(jwtSecret), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				api.GenerateMessageResponse("bad token", nil, fmt.Errorf("cannot extract token")),
			)
			return
		}

		var userID int
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, _ = strconv.Atoi(fmt.Sprint(claims["sub"]))
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				c.AbortWithStatusJSON(
					http.StatusForbidden,
					api.GenerateMessageResponse("bad token", nil, fmt.Errorf("token is expired")),
				)
				return
			}
		} else {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				api.GenerateMessageResponse("bad token", nil, fmt.Errorf("token is not valid")),
			)
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}
