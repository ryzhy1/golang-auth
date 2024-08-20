package middlewares

import (
	"AuthService/internal/lib/jwt"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

// ServerInterceptor Интерсептор для проверки токена
func ServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		// Извлекаем заголовок авторизации
		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		tokenString := authHeader[0]
		if !strings.HasPrefix(tokenString, "Bearer ") {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token format")
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Проверяем токен
		token, err := jwt.VerifyToken(tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Извлекаем userID из токена
		userID, err := jwt.GetUserIDFromToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Добавляем userID в контекст для последующего использования
		ctx = context.WithValue(ctx, "userID", userID)
		return handler(ctx, req)
	}
}
