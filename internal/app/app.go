package app

import (
	grpcapp "Service/internal/app/grpc"
	"time"
)

type App struct {
	GRPCApp *grpcapp.App
}

func New(
	storagePath string,
	tokenTTL time.Duration,
	port int,
	timeout time.Duration,
) *App {

	// TODO : init storage

	// TODO : init auth service

	// TODO : init grpc-application

	return &App{}
}
