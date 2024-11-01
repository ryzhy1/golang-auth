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

func New(log *slog.Logger, grpcPort, storagePath string, tokenTTL time.Duration) *App {
	storage, err := postgres.NewPostgres(storagePath)
	if err != nil {
		panic(err)
	}

	//redisDB, err := redis.InitRedis(redisStorage, redisPassword, redisDbNumber)
	if err != nil {
		panic(err)
	}

	AuthService := auth.New(log, storage, tokenTTL)

	grpcApp := grpcapp.New(log, AuthService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
