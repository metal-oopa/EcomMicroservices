package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/metal-oopa/EcomMicroservices/services/user-service/auth"
)

func GenerateJWT(userID int, secretKey string, duration time.Duration) (string, error) {
	claims := &auth.AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
