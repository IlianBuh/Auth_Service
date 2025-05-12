package grpcusrinfo

import (
	"Service/internal/domain/models"
	"Service/internal/lib/mappers"
	"Service/internal/services/userinfo"
	"context"
	"errors"
	userv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/user"
	userinfov1 "github.com/IlianBuh/SSO_Protobuf/gen/go/userinfo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserInfo interface {
	User(ctx context.Context, uuid int) (models.User, error)
	Users(ctx context.Context, uuid []int) ([]models.User, error)
	UsersExist(ctx context.Context, uuid []int) (bool, error)
}

type serverAPI struct {
	userinfov1.UnimplementedUserInfoServer
	usrInfo UserInfo
}

func Register(grpcsrv *grpc.Server, usrInfo UserInfo) {
	userinfov1.RegisterUserInfoServer(grpcsrv, &serverAPI{usrInfo: usrInfo})
}

func (s *serverAPI) Users(ctx context.Context, u *userinfov1.UsersRequest) (*userinfov1.UsersResponse, error) {
	if err := validateUUIDs(u.GetUuids()...); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	users, err := s.usrInfo.Users(ctx, mappers.Int32ToInt(u.GetUuids()...))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userinfov1.UsersResponse{
		Users: mappers.ModelUsersToAPI(users...),
	}, nil
}

func (s *serverAPI) User(ctx context.Context, u *userinfov1.UserRequest) (*userinfov1.UserResponse, error) {
	if err := validateUUIDs(u.GetUuid()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.usrInfo.User(ctx, int(u.GetUuid()))
	if err != nil {
		if errors.Is(err, userinfo.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user is not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userinfov1.UserResponse{
		User: &userv1.User{
			Uuid:  int32(user.UUID),
			Login: user.Login,
			Email: user.Email,
		},
	}, nil
}

func (s *serverAPI) UsersExist(ctx context.Context, u *userinfov1.UsersExistRequest) (*userinfov1.UsersExistResponse, error) {
	if err := validateUUIDs(u.GetUuid()...); err != nil {
		return nil, status.Error(codes.InvalidArgument, "uuid can't be negative")
	}

	exist, err := s.usrInfo.UsersExist(ctx, mappers.Int32ToInt(u.GetUuid()...))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userinfov1.UsersExistResponse{
		Exist: exist,
	}, nil
}
func validateUUIDs(uuids ...int32) error {
	for _, uuid := range uuids {
		if uuid < 0 {
			return errors.New("uuid can't be negative")
		}
	}
	return nil
}
