package user_model

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type JWTClaims struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

const JwtExpiry = time.Hour * 24

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable not set!")
	}
	return []byte(secret)
}
