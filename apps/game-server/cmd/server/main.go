package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/depuzhiguang/game-server/internal/logger"
	"github.com/depuzhiguang/game-server/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:3000"
	}

	logg := logger.New("game-server")
	logg.Info("De Pu Zhi Guang - Game Server starting", slog.String("port", port))

	srv := server.NewServer(":"+port, apiBaseURL, logg)
	go func() {
		if err := srv.Start(); err != nil {
			logg.Error("Server error", slog.String("error", err.Error()))
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logg.Info("Shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		logg.Error("Shutdown error", slog.String("error", err.Error()))
	}
}
