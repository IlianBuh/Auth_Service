package follow

import (
	"Service/internal/domain/models"
	"Service/internal/lib/logger/sl"
	"Service/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Follower interface {
	Follow(context.Context, int, int) error
}
type Unfollower interface {
	Unfollow(context.Context, int, int) error
}
type FollowingsProvider interface {
	Followers(context.Context, int) ([]models.User, error)
	Followees(context.Context, int) ([]models.User, error)
}

type Follow struct {
	log    *slog.Logger
	flw    Follower
	unflw  Unfollower
	flwPrv FollowingsProvider
}

// New returns new instance of service layer
func New(
	log *slog.Logger,
	flw Follower,
	unflw Unfollower,
	flwPrv FollowingsProvider,
) *Follow {
	return &Follow{
		log:    log,
		flw:    flw,
		unflw:  unflw,
		flwPrv: flwPrv,
	}
}

// Follow follows user src on target
func (f *Follow) Follow(
	ctx context.Context,
	src, target int,
) error {
	const op = "follow.Follow"
	log := f.log.With(slog.String("op", op))
	log.Info(
		"starting to follow",
		slog.Int("src", src),
		slog.Int("target", target),
	)

	err := f.flw.Follow(ctx, src, target)
	if err != nil {
		if errors.Is(err, storage.ErrFollowing) {
			log.Warn("user already following")
			return fmt.Errorf("%s: %w", op, ErrFollowing)
		}

		log.Error("failed to follow user", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully followed user")
	return nil
}

// Unfollow unfollows user src on target
func (f *Follow) Unfollow(
	ctx context.Context,
	src, target int,
) error {
	const op = "follow.Unfollow"
	log := f.log.With(slog.String("op", op))
	log.Info(
		"starting to unfollow",
		slog.Int("src", src),
		slog.Int("target", target),
	)

	err := f.unflw.Unfollow(ctx, src, target)
	if err != nil {
		if errors.Is(err, storage.ErrNoFollowing) {
			log.Warn("user has not followed")
			return fmt.Errorf("%s: %w", op, ErrNoFollowing)
		}

		log.Error("failed to unfollow user", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully unfollowed user")
	return nil
}

// ListFollowers returns all followers of the user with the uuid
func (f *Follow) Followers(
	ctx context.Context,
	uuid int,
) ([]models.User, error) {
	const op = "follow.ListFollowers"
	log := f.log.With(slog.String("op", op))
	log.Info("starting to list followers", slog.Int("uuid", uuid))

	followers, err := f.flwPrv.Followers(ctx, uuid)
	if err != nil {
		log.Error("failed to list followers", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully listed followers")
	return followers, nil
}

// ListFollowees returns all followees of the user with the uuid
func (f *Follow) Followees(
	ctx context.Context,
	uuid int,
) ([]models.User, error) {
	const op = "follow.ListFollowees"
	log := f.log.With(slog.String("op", op))
	log.Info("starting to list followees", slog.Int("uuid", uuid))

	followees, err := f.flwPrv.Followees(ctx, uuid)
	if err != nil {
		log.Error("failed to list followees", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully listed followees")
	return followees, nil
}
