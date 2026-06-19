// Package service provides authentication helpers for the URL shortener,
// issuing and validating the JWT stored in the user's auth cookie.
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the set of JWT claims carried by an auth token. It embeds the
// standard registered claims and adds the application's UserID.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// TokenExp is the lifetime of an issued JWT.
const TokenExp = time.Hour * 3

// SecretKey is the HMAC secret used to sign and verify JWTs.
//
// TODO: move to the environment variables.
const SecretKey = "supersecretkey"

// BuildJWTString issues a signed JWT that embeds userID and expires after
// TokenExp.
func BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID parses and validates tokenString and returns the UserID it carries.
// It returns an error if the token is malformed, signed with an unexpected
// method, or otherwise invalid.
func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token is not valid")
	}

	return claims.UserID, err
}
