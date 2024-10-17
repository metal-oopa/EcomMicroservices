package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type AuthClaims struct {
	jwt.RegisteredClaims
	UserID int
}

func ExtractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	tokenValues := md.Get("authorization")
	if len(tokenValues) == 0 {
		return "", errors.New("missing authorization token")
	}

	// Format: Bearer <token>
	var tokenString string
	fmt.Sscanf(tokenValues[0], "Bearer %s", &tokenString)

	if tokenString == "" {
		return "", errors.New("invalid authorization token format")
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, secretKey string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
