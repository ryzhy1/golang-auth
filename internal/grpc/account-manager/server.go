package account_manager

import (
	"AuthService/internal/domain/models"
	account_manager "AuthService/internal/services/account-manager"
	"AuthService/middlewares"
	"context"
	"errors"
	ssov1 "github.com/ryzhy1/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Account interface {
	CreatePurchase(ctx context.Context, userID string, amount float64) error
	UpdateBalance(ctx context.Context, userID string, amount float64) error
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	UpdateEmail(ctx context.Context, username, oldEmail, newEmail string) error
	UpdatePassword(ctx context.Context, username, oldPassword, newPassword string) error
}

type serverAPI struct {
	ssov1.UnimplementedAccountServiceServer
	account Account
}

const (
	ErrInvalidCredentials = "invalid credentials"
	ErrUserNotFound       = "user not found"
	ErrUserAlreadyExists  = "user already exists"
	ErrNoActiveSession    = "user already logged out"
)

func Register(gRPC *grpc.Server, account Account) {
	ssov1.RegisterAccountServiceServer(gRPC, &serverAPI{account: account})
}

func ServerInterceptor() grpc.UnaryServerInterceptor {
	return middlewares.ServerInterceptor()
}

func (s *serverAPI) CreatePurchase(context.Context, *ssov1.CreatePurchaseRequest) (*ssov1.CreatePurchaseResponse, error) {

	return &ssov1.CreatePurchaseResponse{}, nil
}

func (s *serverAPI) UpdateBalance(ctx context.Context, req *ssov1.UpdateBalanceRequest) (*ssov1.UpdateBalanceResponse, error) {
	if req.GetUsername() == "" || req.GetAmount() == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty field")
	}

	err := s.account.UpdateBalance(ctx, req.GetUsername(), float64(req.GetAmount()))
	if err != nil {
		if errors.Is(err, account_manager.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.UpdateBalanceResponse{}, nil
}

func (s *serverAPI) GetUserByID(ctx context.Context, req *ssov1.GetUserRequest) (*ssov1.GetUserResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is empty")
	}

	var user *models.User

	user, err := s.account.GetUserByID(ctx, req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.GetUserResponse{
		Username: user.Login,
		Email:    user.Email,
		Balance:  user.Balance,
		Discount: user.Discount,
	}, nil
}
func (s *serverAPI) UpdateEmail(ctx context.Context, req *ssov1.UpdateEmailRequest) (*ssov1.UpdateEmailResponse, error) {
	if req.GetNewEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}

	if req.GetNewEmail() == req.GetOldEmail() {
		return nil, status.Error(codes.InvalidArgument, "new email is equal to old email")
	}

	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is empty")
	}

	err := s.account.UpdateEmail(ctx, req.GetUsername(), req.GetOldEmail(), req.GetNewEmail())
	if err != nil {
		if errors.Is(err, account_manager.ErrEmailAlreadyTaken) {
			return nil, status.Error(codes.AlreadyExists, "email already taken")
		}

		if errors.Is(err, account_manager.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		if errors.Is(err, account_manager.ErrWrongEmail) {
			return nil, status.Error(codes.InvalidArgument, "wrong email")
		}

		if errors.Is(err, account_manager.ErrFailedToUpdate) {
			return nil, status.Error(codes.Internal, "failed to update email")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.UpdateEmailResponse{
		Status:  "true",
		Message: "email updated successfully",
	}, nil
}
func (s *serverAPI) UpdatePassword(ctx context.Context, req *ssov1.UpdatePasswordRequest) (*ssov1.UpdatePasswordResponse, error) {
	if req.GetUsername() == "" || req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty field")
	}

	if req.GetNewPassword() == req.GetOldPassword() {
		return nil, status.Error(codes.InvalidArgument, "new password is equal to old password")
	}

	err := s.account.UpdatePassword(ctx, req.GetUsername(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		if errors.Is(err, account_manager.ErrWrongPassword) {
			return nil, status.Error(codes.InvalidArgument, "wrong password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.UpdatePasswordResponse{
		Status:  "true",
		Message: "password updated successfully",
	}, nil
}
