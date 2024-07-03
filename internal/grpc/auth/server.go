package auth

import (
	"context"
	ssov1 "github.com/ryzhy1/protos/gen/go/sso"
)

import (
	"google.golang.org/grpc"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServiceServer
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterAuthServiceServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	return &ssov1.LoginResponse{
		Token: req.GetEmail(),
	}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	return &ssov1.LogoutResponse{}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	return &ssov1.RegisterResponse{}, nil
}
