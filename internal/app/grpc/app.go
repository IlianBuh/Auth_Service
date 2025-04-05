package grpcapp

import (
	"google.golang.org/grpc"
	"log/slog"
	"time"
)

type App struct {
	log     *slog.Logger
	gRPCSrv *grpc.Server
}

func New(
	port int,
	timeout time.Duration,
) *App {
	// TODO : make grpc server

	// TODO : bind auth service

	return &App{}
}
