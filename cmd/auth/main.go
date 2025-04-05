package main

import (
	"Service/internal/app"
	"Service/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.New()

	log := setUpLogger(cfg.Env)

	log.Info("logger set up", slog.Any("cfg", cfg))

	application := app.New(
		log,
		cfg.StoragePath,
		cfg.Secret,
		cfg.TokenTTL,
		cfg.GRPC.Port,
		cfg.GRPC.Timeout,
	)

	go application.GRPCApp.MustRun()

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop
	log.Info("receive signal", slog.Any("signal", sign))
	application.GRPCApp.Stop()
}

// setUpLogger returns set logger according to current environment
func setUpLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
