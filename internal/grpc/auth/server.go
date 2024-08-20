package auth

import (
	"AuthService/internal/services/auth"
	"context"
	"errors"
	ssov1 "github.com/ryzhy1/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (token string, err error)
	Logout(ctx context.Context, token string) (boolean bool, err error)
	Register(
		ctx context.Context,
		login string,
		email string,
		password string,
	) (userID string, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServiceServer
	auth Auth
}

const (
	ErrInvalidCredentials = "invalid credentials"
	ErrUserNotFound       = "user not found"
	ErrUserAlreadyExists  = "user already exists"
	ErrNoActiveSession    = "user already logged out"
)

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServiceServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	boolean, err := s.auth.Logout(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.LogoutResponse{
		Success: boolean,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.GetLogin(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}

	return nil
}
