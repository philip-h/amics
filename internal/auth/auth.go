package auth

import "github.com/golang-jwt/jwt/v5"

type Authenticator interface {
	CreateJwt(sub, aud string, exp int64) (string, error)
	ValidateJwt(tokenString string) (*jwt.Token, error)
}