package middleware

import (
	"Ustasjs/yp-url-shortener/internal/logger"
	"Ustasjs/yp-url-shortener/internal/service"
	"context"
	"net/http"

	"go.uber.org/zap"
)

const AuthCookieName = "auth_token"

type contextKey string

const UserIDContextKey contextKey = "user_id"

type UserRepository interface {
	CreateUser(ctx context.Context) (string, error)
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

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
