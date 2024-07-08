package app

import (
	grpcapp "AuthService/internal/app/grpc"
	"AuthService/internal/services/auth"
	"AuthService/internal/storage/postgres"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort string, storagePath string, tokenTTL time.Duration) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	authSerivce := auth.New(log, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authSerivce, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
