package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func New(id uint64, login string, secret string, exp time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claim := token.Claims.(jwt.MapClaims)

	claim["uuid"] = id
	claim["login"] = login
	claim["exp"] = time.Now().Add(exp).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}
