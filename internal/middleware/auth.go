// Package middleware provides HTTP middleware for the URL shortener:
// authentication and user identification via a JWT cookie, and gzip
// request decompression.
package middleware

import (
	"context"
	"net/http"

	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/service"

	"go.uber.org/zap"
)

// AuthCookieName is the name of the cookie that stores the user's JWT.
const AuthCookieName = "auth_token"

type contextKey string

// UserIDContextKey is the request-context key under which the authenticated user
// ID is stored.
const UserIDContextKey contextKey = "user_id"

// UserRepository creates users for callers that do not yet have an identity.
type UserRepository interface {
	CreateUser(ctx context.Context) (string, error)
}

// GetUserIDFromContext returns the user ID stored in ctx by the Auth middleware
// and reports whether it was present.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

// Auth returns middleware that authenticates each request via the auth cookie.
// When the cookie is missing or invalid, a new user is created through userRepo
// and a fresh JWT cookie is issued. The resolved user ID is stored in the
// request context, where it can be read with GetUserIDFromContext.
func Auth(userRepo UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie(AuthCookieName)
			if err == nil && cookie.Value != "" {
				userID, err = service.GetUserID(cookie.Value)
			}

			if err != nil || userID == "" {
				newUserID, createErr := userRepo.CreateUser(r.Context())
				if createErr != nil {
					logger.Log.Error("Failed to create user", zap.Error(createErr))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				userID = newUserID

				tokenString, tokenErr := service.BuildJWTString(userID)
				if tokenErr != nil {
					logger.Log.Error("Failed to build JWT token", zap.Error(tokenErr))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     AuthCookieName,
					Value:    tokenString,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   int(service.TokenExp.Seconds()),
				})
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth returns middleware that rejects requests without a valid
// authenticated user, responding with 401 Unauthorized.
func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(AuthCookieName)
			if err == nil && cookie.Value != "" {
				if _, err := service.GetUserID(cookie.Value); err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			_, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
