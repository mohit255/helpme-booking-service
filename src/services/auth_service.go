package services

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"go-helpme-booking/src/config"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ValidateJWT parses and validates a JWT token string, returning the claims.
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.App.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New(config.MsgTokenInvalid)
	}
	return claims, nil
}
