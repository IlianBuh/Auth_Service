package grpcauth

import (
	"Service/internal/domain/models"
	"Service/internal/services/auth"
	"context"
	"errors"
	"fmt"
	"net/mail"

	authv1 "github.com/IlianBuh/SSO_Protobuf/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		login, password string,
	) (models.TokensPair, error)
	SignUp(
		ctx context.Context,
		login, email, password string,
	) (models.TokensPair, error)
	UpdateTokens(
		ctx context.Context,
		refreshToken string,
	) (models.TokensPair, error)
}
type serverAPI struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func Register(grpcsrv *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(grpcsrv, &serverAPI{auth: auth})
}

// Login handlers Login-API request
func (s *serverAPI) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {

	if err := validateLogin(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.auth.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, "invalid arguments")
		}

		return nil, status.Error(codes.Internal, "internal error occurred")
	}

	return &authv1.LoginResponse{
		RefreshToken: token.RefreshToken.Val,
		AccessToken:  token.AccessToken.Val,
	}, nil

}

// SignUp handlers SignUp-API request
func (s *serverAPI) SignUp(
	ctx context.Context,
	req *authv1.SignUpRequest,
) (*authv1.SignUpResponse, error) {

	if err := validateSignUp(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.auth.SignUp(ctx, req.GetLogin(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, "invalid arguments")
		}

		return nil, status.Error(codes.Internal, "internal error occurred")
	}

	return &authv1.SignUpResponse{
		RefreshToken: token.RefreshToken.Val,
		AccessToken:  token.AccessToken.Val,
	}, nil
}

func (s *serverAPI) UpdateTokens(
	ctx context.Context,
	req *authv1.UpdateRequest,
) (*authv1.UpdateResponse, error) {
	tokens, err := s.auth.UpdateTokens(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, auth.ErrExpired) {
			return nil, status.Error(codes.Unauthenticated, "token is expiret")
		}
		if errors.Is(err, auth.ErrNoToken) {
			return nil, status.Error(codes.Unauthenticated, "token does not exist")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.UpdateResponse{
		AccessToken:  tokens.AccessToken.Val,
		RefreshToken: tokens.RefreshToken.Val,
	}, nil
}

// validateLogin validates user's request to log in
func validateLogin(req *authv1.LoginRequest) error {

	if req.GetLogin() == "" {
		return fmt.Errorf("login is required")
	}

	if len(req.GetPassword()) < 8 {
		return fmt.Errorf("password must be at least 8 chars")
	}

	return nil
}

// validateSignUp validates user's request to sign up
func validateSignUp(req *authv1.SignUpRequest) error {

	if req.GetLogin() == "" {
		return fmt.Errorf("login is required")
	}

	if _, err := mail.ParseAddress(req.GetEmail()); err != nil {
		return fmt.Errorf("invalid email")
	}

	if len(req.GetPassword()) < 8 {
		return fmt.Errorf("password must be at least 8 chars")
	}

	return nil
}
