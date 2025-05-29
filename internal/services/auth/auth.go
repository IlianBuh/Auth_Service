package auth

import (
	"Service/internal/domain/models"
	e "Service/internal/lib/errors"
	"Service/internal/lib/jwt"
	"Service/internal/lib/logger/sl"
	"Service/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserProvider interface {
	User(ctx context.Context, key interface{}) (models.User, error)
}

type UserSaver interface {
	Save(ctx context.Context, login, email string, passHash []byte) (uint64, error)
}

type TokenProvider interface {
	StoreToken(ctx context.Context, refreshToken string, accessToken string) error
	DeleteToken(ctx context.Context, refreshToken string) error
	Token(ctx context.Context, token string) (string, error)
}
type Auth struct {
	log        *slog.Logger
	usrPrv     UserProvider
	usrSv      UserSaver
	tknPrv     TokenProvider
	secret     string
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

// New creates auth service instance
func New(
	log *slog.Logger,
	usrPrv UserProvider,
	usrSv UserSaver,
	tknPrv TokenProvider,
	secret string,
	tokenTTL time.Duration,
	refreshTTL time.Duration,
) *Auth {
	return &Auth{
		log:        log,
		usrPrv:     usrPrv,
		usrSv:      usrSv,
		tknPrv:     tknPrv,
		secret:     secret,
		tokenTTL:   tokenTTL,
		refreshTTL: refreshTTL,
	}
}

// Login implements login business logic. It returns JWT token with uuid and login, or error
func (a *Auth) Login(
	ctx context.Context,
	login, password string,
) (models.TokensPair, error) {
	const op = "auth.Login"
	log := a.log.With(slog.String("op", op))
	log.Info("starting to login user")

	user, err := a.usrPrv.User(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user is not found", slog.String("login", login))
			return models.TokensPair{}, fmt.Errorf("%s: %w", op, ErrInvalidArgument)
		}

		log.Error("failed to get user", sl.Err(err))
		return models.TokensPair{}, fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("password mismatched", sl.Err(err))
		return models.TokensPair{}, fmt.Errorf("%s: %w", op, ErrInvalidArgument)
	}

	token, err := jwt.NewTokensPair(user.UUID, user.Login, a.secret, a.tokenTTL, a.refreshTTL)
	if err != nil {
		log.Error("failed to generate JWT", sl.Err(err))
		return models.TokensPair{}, fmt.Errorf("%s: %w", op, err)
	}

	err = a.tknPrv.StoreToken(ctx, token.RefreshToken.Val, token.AccessToken.Val)
	if err != nil {
		log.Error(
			"failed to save token",
			sl.Err(err),
			slog.String("refresh-token", token.RefreshToken.Val),
			slog.String("access-token", token.AccessToken.Val),
		)
		return models.TokensPair{}, e.Fail(op, err)
	}

	log.Info("successfully logged in")
	return token, nil
}

// SignUp implements sign up business logic. It returns JWT token with uuid and login, or error
func (a *Auth) SignUp(
	ctx context.Context,
	login, email, password string,
) (models.TokensPair, error) {
	const op = "grpcapp.SignUp"
	fail := func(err error) (models.TokensPair, error) {
		return models.TokensPair{}, fmt.Errorf("%s: %w", op, err)
	}
	log := a.log.With(slog.String("op", op))
	log.Info("starting to sign up user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to compute hash", sl.Err(err))
		return fail(err)
	}

	uuid, err := a.usrSv.Save(ctx, login, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("failed to save user", sl.Err(err))
			return fail(err)
		}

		log.Error("failed to save user", sl.Err(err))
		return fail(err)
	}

	token, err := jwt.NewTokensPair(uuid, login, a.secret, a.tokenTTL, a.refreshTTL)
	if err != nil {
		log.Error("failed to generate JWT token", sl.Err(err))
		return fail(err)
	}

	err = a.tknPrv.StoreToken(ctx, token.RefreshToken.Val, token.AccessToken.Val)
	if err != nil {
		log.Error(
			"failed to save tokens",
			sl.Err(err),
			slog.String("refresh-token", token.RefreshToken.Val),
			slog.String("access-token", token.AccessToken.Val),
		)
		return fail(err)
	}

	log.Info("successfully signed up")
	return token, nil
}

func (a *Auth) UpdateTokens(
	ctx context.Context,
	refreshToken string,
) (models.TokensPair, error) {
	const op = "auth.UpdateTokens"
	fail := func(err error) (models.TokensPair, error) {
		return models.TokensPair{}, e.Fail(op, err)
	}
	log := a.log.With(slog.String("op", op))
	log.Info("starting to update tokens")

	err := jwt.ValidateToken(refreshToken, a.secret)
	if err != nil {
		if errors.Is(err, jwt.ErrExpired) {
			log.Warn("trying to update expired token")
			return fail(ErrExpired)
		}

		log.Error("failed to validate token", sl.Err(err))
		return fail(err)
	}

	accessToken, err := a.tknPrv.Token(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("trying to update pair with token, which is not tracking")
			return fail(ErrNoToken)
		}

		log.Error("failed to get access-token by refresh", sl.Err(err))
		return fail(err)
	}

	payload, err := jwt.ParseToken(accessToken, a.secret)
	if err != nil {
		log.Error("failed to parse access token")
		return fail(err)
	}

	tokens, err := jwt.NewTokensPair(
		uint64(payload.Id),
		payload.Login,
		a.secret,
		a.tokenTTL,
		a.refreshTTL,
	)
	if err != nil {
		log.Error("failed to generate new tokens pair", sl.Err(err))
		return fail(err)
	}

	err = a.tknPrv.StoreToken(ctx, tokens.RefreshToken.Val, tokens.AccessToken.Val)
	if err != nil {
		log.Error(
			"failed to save token",
			sl.Err(err),
			slog.String("refresh-token", tokens.RefreshToken.Val),
			slog.String("access-token", tokens.AccessToken.Val),
		)
		return models.TokensPair{}, e.Fail(op, err)
	}

	err = a.tknPrv.DeleteToken(ctx, refreshToken)
	if err != nil {
		log.Error(
			"failed to delete old token",
			sl.Err(err),
			slog.String("refresh-token", tokens.RefreshToken.Val),
			slog.String("access-token", tokens.AccessToken.Val),
		)
		return models.TokensPair{}, e.Fail(op, err)
	}

	log.Info("tokens are updated")
	return tokens, nil
}
