package auth

import (
	"Service/internal/domain/models"
	"Service/internal/lib/jwt"
	"Service/internal/lib/logger/sl"
	"Service/internal/storage"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type UserProvider interface {
	User(ctx context.Context, login string) (models.User, error)
}

type UserSaver interface {
	Save(ctx context.Context, login, email string, passHash []byte) (uint64, error)
}

type Auth struct {
	log      *slog.Logger
	usrPrv   UserProvider
	usrSv    UserSaver
	secret   string
	tokenTTL time.Duration
}

// New creates auth service instance
func New(
	log *slog.Logger,
	usrPrv UserProvider,
	usrSv UserSaver,
	secret string,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:      log,
		usrPrv:   usrPrv,
		usrSv:    usrSv,
		secret:   secret,
		tokenTTL: tokenTTL,
	}
}

// Login implements login business logic. It returns JWT token with uuid and login, or error
func (a *Auth) Login(
	ctx context.Context,
	login, password string,
) (string, error) {
	const op = "auth.Login"
	log := a.log.With(slog.String("op", op))
	log.Info("starting to login user")

	user, err := a.usrPrv.User(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user is not found", slog.String("login", login))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidArgument)
		}

		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Warn("password mismatched", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidArgument)
	}

	token, err := jwt.New(user.UUID, user.Login, a.secret, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate JWT", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully logged in")
	return token, nil
}

// SignUp implements sign up business logic. It returns JWT token with uuid and login, or error
func (a *Auth) SignUp(
	ctx context.Context,
	login, email, password string,
) (string, error) {
	const op = "grpcapp.SignUp"
	log := a.log.With(slog.String("op", op))
	log.Info("starting to sign up user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to compute hash", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	uuid, err := a.usrSv.Save(ctx, login, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("failed to save user", sl.Err(err))
			return "", fmt.Errorf("%s:%w", op, ErrInvalidArgument)
		}

		log.Error("failed to save user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.New(uuid, login, a.secret, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate JWT token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully signed up")
	return token, nil
}
