package app

import (
	grpcapp "AuthService/internal/app/grpc"
	"AuthService/internal/services/auth"
	"AuthService/internal/storage/postgres"
	"AuthService/internal/storage/redis"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort, redisStorage, redisPassword string, redisDbNumber int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	redisDB, err := redis.InitRedis(redisStorage, redisPassword, redisDbNumber)
	if err != nil {
		panic(err)
	}

	authSerivce := auth.New(log, redisDB, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authSerivce, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
