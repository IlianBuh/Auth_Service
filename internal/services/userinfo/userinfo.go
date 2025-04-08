package userinfo

import (
	"Service/internal/domain/models"
	"Service/internal/lib/logger/sl"
	"Service/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type UserProvider interface {
	User(ctx context.Context, key interface{}) (models.User, error)
	Users(ctx context.Context, uuid []int) ([]models.User, error)
}

type UserInfo struct {
	log    *slog.Logger
	usrPrv UserProvider
}

func New(
	log *slog.Logger,
	usrPrv UserProvider,
) *UserInfo {
	return &UserInfo{
		log:    log,
		usrPrv: usrPrv,
	}
}

func (u *UserInfo) User(ctx context.Context, uuid int) (models.User, error) {
	const op = "userinfo.User"
	log := u.log.With(slog.String("op", op))
	log.Info("Starting to get user")

	user, err := u.usrPrv.User(ctx, uuid)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Warn("user is not found", slog.Int("uuid", uuid))
			return models.User{}, ErrNotFound
		}

		log.Error("failed to get user", sl.Err(err))
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully got user")
	return user, nil
}

func (u *UserInfo) Users(ctx context.Context, uuids []int) ([]models.User, error) {
	const op = "userinfo.Users"
	log := u.log.With(slog.String("op", op))
	log.Info("Starting to get users list")

	users, err := u.usrPrv.Users(ctx, uuids)
	if err != nil {
		log.Error("failed to get user", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully got user")
	return users, nil
}

func (u *UserInfo) UsersExist(ctx context.Context, uuids []int) (bool, error) {
	const op = "userinfo.UsersExist"
	log := u.log.With(slog.String("op", op))
	log.Info(
		"starting to check if users exist",
		slog.Any("uuids", uuids),
	)
	var res bool

	users, err := u.usrPrv.Users(ctx, uuids)
	if err != nil {
		log.Error("failed to get users", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	res = len(users) == len(uuids)
	if !res {
		log.Warn("some users don't exist")
	} else {
		log.Info("all users exist")
	}

	return res, nil
}
