package main

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webanalyzer/internal/api/v1/router"
	"webanalyzer/internal/config"
	"webanalyzer/internal/log"
)

func init() {
	log.InitLogger()
	config.LoadEnv()
}

func main() {
	defer log.Sync()

	r := router.New()

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Logger.Info("Server started on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt
	<-stop
	log.Logger.Info("Shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	log.Logger.Info("Server exited successfully")
}
