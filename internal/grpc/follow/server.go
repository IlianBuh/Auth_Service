package follow

import (
	"context"
	"errors"

	"Service/internal/domain/models"
	"Service/internal/lib/mappers"
	"Service/internal/services/follow"

	followv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/follow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowProvider interface {
	Follow(
		ctx context.Context,
		src, target int,
	) error
	Unfollow(
		ctx context.Context,
		src, target int,
	) error
	Followers(
		ctx context.Context,
		uuid int,
	) ([]models.User, error)
	Followees(
		ctx context.Context,
		uuid int,
	) ([]models.User, error)
}

type serverAPI struct {
	followProvider FollowProvider
	followv1.UnimplementedFollowServer
}

func Register(srvr *grpc.Server, followProvider FollowProvider) {
	followv1.RegisterFollowServer(srvr, &serverAPI{followProvider: followProvider})
}

func (s *serverAPI) Follow(
	ctx context.Context,
	req *followv1.FollowRequest,
) (*followv1.FollowResponse, error) {
	if err := validateIds(int(req.GetSrc()), int(req.GetTarget())); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := s.followProvider.Follow(ctx, int(req.GetSrc()), int(req.GetTarget()))
	if err != nil {
		switch {
		case errors.Is(err, follow.ErrFollowing):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, follow.ErrInvalidUUIDs):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	return &followv1.FollowResponse{}, nil
}

func (s *serverAPI) Unfollow(
	ctx context.Context,
	req *followv1.UnfollowRequest,
) (*followv1.UnfollowResponse, error) {
	if err := validateIds(int(req.GetSrc()), int(req.GetTarget())); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err := s.followProvider.Unfollow(ctx, int(req.GetSrc()), int(req.GetTarget()))
	if err != nil {
		switch {
		case errors.Is(err, follow.ErrNoFollowing):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal")
		}
	}

	return &followv1.UnfollowResponse{}, nil
}

func (s *serverAPI) Followers(
	ctx context.Context,
	req *followv1.FollowersRequest,
) (*followv1.FollowersResponse, error) {
	clientId := mappers.Int32ToInt(req.GetUuid())[0]

	if err := validateIds(clientId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	followers, err := s.followProvider.Followers(ctx, clientId)
	if err != nil {
		// TODO : handle error when users are not found

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &followv1.FollowersResponse{User: mappers.ModelUsersToAPI(followers...)}, nil
}

func (s *serverAPI) Followees(
	ctx context.Context,
	req *followv1.FolloweesRequest,
) (*followv1.FolloweesResponse, error) {
	clientId := mappers.Int32ToInt(req.GetUuid())[0]

	if err := validateIds(clientId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	followees, err := s.followProvider.Followees(ctx, clientId)
	if err != nil {
		// TODO : handle error when users are not found

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &followv1.FolloweesResponse{User: mappers.ModelUsersToAPI(followees...)}, nil
}

func validateIds(ids ...int) error {

	for _, id := range ids {
		if id < 0 {
			return errors.New("id can't be negative")
		}
	}

	return nil
}
