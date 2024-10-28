package grpcapp

import (
	authgrpc "AuthService/internal/grpc/auth"
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	ssov1 "github.com/ryzhy1/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"net/http"
	"sync"
)

type App struct {
	log           *slog.Logger
	authServer    *grpc.Server
	accountServer *grpc.Server
	authPort      string
	accountPort   string
}

func New(log *slog.Logger, authService authgrpc.Auth, authPort string) *App {
	authServer := grpc.NewServer()
	authgrpc.Register(authServer, authService)
	reflection.Register(authServer)

	return &App{
		log:        log,
		authServer: authServer,
		authPort:   ":" + authPort,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	ctx := context.Background()

	log := a.log.With(slog.String("op", op))

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		l, err := net.Listen("tcp", a.authPort)
		if err != nil {
			log.Error("failed to listen for auth server", "error", err)
			return
		}
		log.Info("Auth gRPC server started", "port", a.authPort)
		if err := a.authServer.Serve(l); err != nil {
			log.Error("failed to serve auth server", "error", err)
		}
	}()

	go func() {
		defer wg.Done()

		mux := runtime.NewServeMux()

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		// Register AuthService endpoint
		if err := ssov1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, a.authPort, opts); err != nil {
			log.Error("failed to register AuthService handler", "error", err)
			return
		}

		log.Info("Http server listening at", "port", ":8081")

		handler := allowCORS(mux) // Добавлено CORS middleware
		if err := http.ListenAndServe(`localhost:8081`, handler); err != nil {
			log.Error("failed to serve grpc auth server", "error", err)
			return
		}
	}()

	wg.Wait()
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("Stopping gRPC servers")

	a.authServer.GracefulStop()
	a.accountServer.GracefulStop()
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PATCH, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}
