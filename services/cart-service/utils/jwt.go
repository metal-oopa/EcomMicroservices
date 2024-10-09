package utils

import (
	"errors"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractUserIDFromToken(tokenString, secretKey string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			userID, err := strconv.Atoi(sub)
			if err != nil {
				return 0, errors.New("invalid user ID in token")
			}
			return userID, nil
		}
	}

	return 0, errors.New("user ID not found in token")
}
