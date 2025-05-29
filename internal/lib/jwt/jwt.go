package jwt

import (
	"Service/internal/domain/models"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	ErrExpired = errors.New("token is expired")
)

var (
	EncodingMethod = jwt.SigningMethodHS256
)

func NewAccess(id uint64, login string, secret string, exp time.Duration) (string, error) {
	token := jwt.New(EncodingMethod)

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

func NewRefresh(secret string, exp time.Duration) (string, error) {
	token := jwt.New(EncodingMethod)

	claim := token.Claims.(jwt.MapClaims)

	claim["exp"] = time.Now().Add(exp).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func NewTokensPair(
	id uint64,
	login string,
	secret string,
	exp time.Duration,
	refreshTTL time.Duration,
) (models.TokensPair, error) {

	accessToken, err := NewAccess(id, login, secret, exp)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshToken, err := NewRefresh(secret, refreshTTL)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		AccessToken: models.Token{
			Type: models.AccessToken,
			Val:  accessToken,
		},
		RefreshToken: models.Token{
			Type: models.RefreshToken,
			Val:  refreshToken,
		},
	}, err
}

func ValidateToken(token string, secret string) error {
	var claims TokenToValidate
	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(secret), nil
			}

			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func ParseToken(token string, secret string) (TokenPayload, error) {
	var claims TokenPayload
	_, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(secret), nil
			}

			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		},
	)
	if err != nil {
		return TokenPayload{}, err
	}

	return claims, nil
}

type TokenToValidate struct {
	Exp int64 `json:"exp"`
}

func (tc *TokenToValidate) Valid() error {
	if tc.Exp < time.Now().Unix() {
		return ErrExpired
	}

	return nil
}

type TokenPayload struct {
	Id    int    `json:"id"`
	Login string `json:"login"`
}

func (tc *TokenPayload) Valid() error {
	return nil
}
