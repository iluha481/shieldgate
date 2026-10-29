package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/iluha481/shieldgate/internal/app"
	"github.com/iluha481/shieldgate/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	cfg := config.MustLoad()
	application, err := app.New(context.Background(), cfg)
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}
	if err := application.Run(); err != nil {
		slog.Error("Application terminated with error", "error", err)
		os.Exit(1)
	}
}
