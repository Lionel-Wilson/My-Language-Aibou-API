package middlewares

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Lionel-Wilson/My-Language-Aibou-API/internal/utils"
)

var secretKey []byte

func init() {
	keystring, err := generateSecureKey()
	if err != nil {
		fmt.Println("Error generating key:", err)
		return
	}

	secretKey = []byte(keystring)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		/* Retrieve the token from the cookie
		tokenString, err := c.Cookie("jwtToken")
		if err != nil {
			utils.NewErrorResponse(c, http.StatusSeeOther, "Authorisation Failed", []string{"Token missing in cookie"})
			c.Abort()
			return
		}*/
		//Using header to look for token
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Authorisation Failed", []string{"Missing authorization header"})
			c.Abort()

			return
		}

		tokenString = tokenString[len("Bearer "):]

		token, err := VerifyToken(tokenString)
		if err != nil {
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Authorisation Failed", []string{"Invalid token"})
			c.Abort()

			return
		}

		userID, err := token.Claims.GetSubject()
		if err != nil {
			utils.NewErrorResponse(c, http.StatusUnauthorized, "Authorisation Failed", []string{"Unable to extract subject from token"})
			c.Abort()

			return
		}

		c.Request.AddCookie(&http.Cookie{
			Name:     "userID",
			Value:    userID,
			MaxAge:   60,
			Domain:   "localhost",
			Secure:   true,
			HttpOnly: true,
		})

		// Continue with the next middleware or route handler
		c.Next()
	}
}

func CreateToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": strconv.FormatInt(int64(userId), 10),  // Subject (user identifier)
			"iss": "my-language-aibou-api",               // Issuer
			"exp": time.Now().Add(time.Hour * 24).Unix(), // Expiration time
			"iat": time.Now().Unix(),                     // Issued at
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func generateSecureKey() (string, error) {
	key := make([]byte, 32) // 32 bytes = 256 bits

	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(key), nil
}
