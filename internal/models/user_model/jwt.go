package user_model

import (
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type JwtClaims struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const JwtExpiry = time.Hour * 24

func GetJwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable not set!")
	}
	return []byte(secret)
}

func ValidateToken(tokenString string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, custom_errors.ErrSigningMethod
		}
		return GetJwtSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, custom_errors.ErrInvalidToken
}
