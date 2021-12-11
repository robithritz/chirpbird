package middleware

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type AuthCustomClaims struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	CreatedAt string `json:"createdAt"`
	jwt.StandardClaims
}

func AuthorizeJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")

		_, payload, err := JWTVerifyToken(auth)
		if err != nil {
			ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Unauthorized",
			})
			ctx.Abort()
			return
		}

		ctx.Set("username", payload.Username)
		ctx.Set("id", payload.Id)
		ctx.Set("name", payload.Name)
		ctx.Next()
	}
}

func CheckToken(ctx *gin.Context) {
	auth := ctx.GetHeader("Authorization")

	_, payload, err := JWTVerifyToken(auth)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Unauthorized",
		})
		ctx.Abort()
		return
	}
	ctx.IndentedJSON(http.StatusOK, gin.H{
		"id":       payload.Id,
		"name":     payload.Name,
		"username": payload.Username,
	})
}

func JWTGenToken(id int, name string, username string, createdAt string) (string, error) {
	sKey := getSKey()

	claims := &AuthCustomClaims{
		Id:        id,
		Name:      name,
		Username:  username,
		CreatedAt: createdAt,
		StandardClaims: jwt.StandardClaims{
			Issuer:   "github.com/robithritz/chirpbird",
			IssuedAt: time.Now().Unix(),
		},
	}

	tokenInstance := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := tokenInstance.SignedString([]byte(sKey))
	if err != nil {
		return "", err
	}

	return token, nil

}

func JWTVerifyToken(encodedToken string) (bool, *AuthCustomClaims, error) {
	tokenInstance, err := jwt.ParseWithClaims(encodedToken, &AuthCustomClaims{}, func(parsedToken *jwt.Token) (interface{}, error) {
		if _, isValid := parsedToken.Method.(*jwt.SigningMethodHMAC); !isValid {
			return nil, errors.New("invalid token")
		}
		sKey := getSKey()
		return sKey, nil
	})
	if err != nil {
		return false, nil, err
	}
	if tokenInstance.Valid {
		payload := tokenInstance.Claims.(*AuthCustomClaims)

		return true, payload, nil
	} else {
		return false, nil, errors.New("invalid token")
	}
}

func getSKey() []byte {
	sKey := os.Getenv("SKEY")
	if sKey == "" {
		sKey = "secret"
	}

	return []byte(sKey)
}
