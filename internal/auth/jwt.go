package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/philip-h/amics/internal/errs"
)

type JwtAuthenticator struct {
	SecretKey []byte
	iss       string
}

func NewJwtAuthenticator(secretKey, iss string) *JwtAuthenticator {
	return &JwtAuthenticator{
		SecretKey: []byte(secretKey),
		iss:       iss,
	}
}

func (a *JwtAuthenticator) CreateJwt(userId, role string, exp int64) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId,
		"aud": role,
		"exp": exp,
		"iss": a.iss,
		"iat": jwt.NewNumericDate(time.Now()),
	})

	tokenString, err := claims.SignedString(a.SecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *JwtAuthenticator) ValidateJwt(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, &errs.JwtError{}
		}

		return []byte(a.SecretKey), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(a.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
