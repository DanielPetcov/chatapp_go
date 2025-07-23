package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
}

const (
	secretKey = "daniel123"
)

func NewJWTHandler() *JWTHandler {
	return &JWTHandler{}
}

func (j *JWTHandler) NewJWT(userID string) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "auth-server",
			"sub": userID,
			"exp": jwt.NewNumericDate(now.Add(time.Hour * 24)),
		})

	tokenString, err := t.SignedString([]byte(secretKey))
	if err != nil {
		return "", errors.New("error signing the token")
	}

	return tokenString, nil
}

func (j *JWTHandler) CheckJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", errors.New("parse didn't work")
	}

	if token.Valid {
		return token.Claims.GetSubject()
	} else {
		return "", errors.New("invalid token")
	}
}
