package app

import (
	grpcapp "AuthService/internal/app/grpc"
	account_manager "AuthService/internal/services/account-manager"
	"AuthService/internal/services/auth"
	"AuthService/internal/storage/postgres"
	"AuthService/internal/storage/redis"
	"log/slog"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort, accountPort, redisStorage, redisPassword string, redisDbNumber int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	redisDB, err := redis.InitRedis(redisStorage, redisPassword, redisDbNumber)
	if err != nil {
		panic(err)
	}

	AuthService := auth.New(log, redisDB, storage, storage, tokenTTL)

	AccountService := account_manager.New(log, storage, tokenTTL)

	grpcApp := grpcapp.New(log, AuthService, AccountService, grpcPort, accountPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
