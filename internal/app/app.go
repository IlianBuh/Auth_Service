package app

import (
	grpcapp "Service/internal/app/grpc"
	"Service/internal/services/auth"
	"Service/internal/storage/sqlite"
	"log/slog"
	"time"
)

type App struct {
	GRPCApp *grpcapp.App
}

func New(
	log *slog.Logger,
	storagePath string,
	secret string,
	tokenTTL time.Duration,
	port int,
	timeout time.Duration,
) *App {

	st := sqlite.New(storagePath)

	authsrvc := auth.New(log, st, st, secret, tokenTTL)

	gRPCApp := grpcapp.New(log, port, timeout, authsrvc)

	return &App{
		GRPCApp: gRPCApp,
	}
}
