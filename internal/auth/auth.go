package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Authenticator interface {
	CreateJwt(sub, aud string, exp time.Time) (string, error)
	ValidateJwt(tokenString string) (*jwt.Token, error)
}

