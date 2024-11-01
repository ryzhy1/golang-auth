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
	) (accessToken, refreshToken string, err error)
	Register(
		ctx context.Context,
		login string,
		email string,
		password string,
	) (userID string, err error)

	UpdateUserEmail(
		ctx context.Context,
		userId string,
		oldEmail string,
		newEmail string,
	) (message string, err error)

	UpdateUserPassword(ctx context.Context, userId, oldPassword, newPassword string) (message string, err error)
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
	if req.GetInput() == "" {
		return nil, status.Error(codes.InvalidArgument, "input is empty")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}

	accessToken, refreshToken, err := s.auth.Login(ctx, req.GetInput(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "login is empty")
	}

	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}

	userID, err := s.auth.Register(ctx, req.GetUsername(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) UpdateEmail(ctx context.Context, req *ssov1.EmailRequest) (*ssov1.UpdateResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is empty")
	}

	if req.GetOldEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "old password is empty")
	}

	if req.GetNewEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "new email is empty")
	}

	message, err := s.auth.UpdateUserEmail(ctx, req.GetUserId(), req.GetOldEmail(), req.GetNewEmail())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.UpdateResponse{
		Message: message,
	}, nil

}

func (s *serverAPI) UpdatePassword(ctx context.Context, req *ssov1.PasswordRequest) (*ssov1.UpdateResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is empty")
	}

	if req.GetOldPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "old password is empty")
	}

	if req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "new password is empty")
	}

	message, err := s.auth.UpdateUserPassword(ctx, req.GetUserId(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		} else {
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &ssov1.UpdateResponse{
		Message: message,
	}, nil
}

//func validateRegister(req *ssov1.RegisterRequest) error {
//	if req.GetUsername() == "" {
//		return status.Error(codes.InvalidArgument, "login is empty")
//	}
//
//	if req.GetEmail() == "" {
//		return status.Error(codes.InvalidArgument, "email is empty")
//	}
//
//	if req.GetPassword() == "" {
//		return status.Error(codes.InvalidArgument, "password is empty")
//	}
//
//	return nil
//}
