package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MockAuthenticator struct {
	secret string
}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{
		secret: "secret",
	}
}

func (a *MockAuthenticator) CreateJwt(sub, aud string, exp int64) (string, error) {
	testClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
		"iss": "test-issuer",
		"aud": aud,
		"exp": time.Now().Add(1 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := testClaims.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *MockAuthenticator) ValidateJwt(tokenString string) (*jwt.Token, error) {

	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(a.secret), nil
	})
}

