package auth

import (
	"context"

	"google.golang.org/grpc"
)

type contextKey string

const UserIDKey contextKey = "userID"

func UnaryAuthInterceptor(secretKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if isUnauthenticatedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		tokenString, err := ExtractTokenFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := ValidateJWT(tokenString, secretKey)
		if err != nil {
			return nil, err
		}

		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)

		return handler(newCtx, req)
	}
}

func isUnauthenticatedMethod(fullMethod string) bool {
	unauthenticatedMethods := []string{
		"/grpc.health.v1.Health/Check",
	}

	for _, method := range unauthenticatedMethods {
		if fullMethod == method {
			return true
		}
	}
	return false
}
