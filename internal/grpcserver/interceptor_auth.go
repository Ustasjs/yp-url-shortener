// Package grpcserver serves the ShortenerService over gRPC. The RPC handlers
// are thin: they translate protobuf messages to and from the shared use cases
// in internal/service/urlservice and map its errors to gRPC status codes.
package grpcserver

import (
	"context"
	"strings"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/middleware"
	"Ustasjs/yp-url-shortener/internal/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthMetadataKey is the metadata key that carries the user's JWT. gRPC
// lowercases metadata keys, so the lookup ignores case.
const AuthMetadataKey = "authorization"

// AuthInterceptor is the gRPC version of middleware.Auth. It reads the JWT from
// the authorization metadata header and puts the user ID into the context under
// middleware.UserIDContextKey, where GetUserIDFromContext finds it.
//
// With no header it makes a new user and sends the fresh token back in the
// response header, the way the HTTP middleware sets a cookie. A header with a
// broken token is a client error and gets Unauthenticated, the way
// middleware.RequireAuth answers 401.
func AuthInterceptor(users middleware.UserRepository) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if token, ok := tokenFromMetadata(ctx); ok {
			userID, err := service.GetUserID(token)
			if err != nil || userID == "" {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}
			return handler(withUserID(ctx, userID), req)
		}

		userID, err := users.CreateUser(ctx)
		if err != nil {
			logger.Log.Error("grpc: failed to create user", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}

		tokenString, err := service.BuildJWTString(userID)
		if err != nil {
			logger.Log.Error("grpc: failed to build JWT token", zap.Error(err))
			return nil, status.Error(codes.Internal, "internal error")
		}

		if err := grpc.SetHeader(ctx, metadata.Pairs(AuthMetadataKey, tokenString)); err != nil {
			logger.Log.Error("grpc: failed to send auth token", zap.Error(err))
		}

		return handler(withUserID(ctx, userID), req)
	}
}

// tokenFromMetadata returns the bare JWT from the authorization header.
func tokenFromMetadata(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	values := md.Get(AuthMetadataKey)
	if len(values) == 0 {
		return "", false
	}

	token := strings.TrimSpace(values[0])
	if token == "" {
		return "", false
	}
	return token, true
}

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, middleware.UserIDContextKey, userID)
}
