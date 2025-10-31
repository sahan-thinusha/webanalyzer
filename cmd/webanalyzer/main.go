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
	"webanalyzer/internal/cache"
	"webanalyzer/internal/config"
	"webanalyzer/internal/debug"
	"webanalyzer/internal/log"
)

func init() {
	log.InitLogger()
	config.LoadEnv()
}

func main() {
	defer log.Sync()
	cache.Init()

	r := router.New()

	server := &http.Server{
		Addr:    ":" + config.AppConfig.PublicWebServerPort,
		Handler: r,
	}

	metricsServer := &http.Server{
		Addr:              ":" + config.AppConfig.MetricsWebServerPort,
		Handler:           router.NewMetricsRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Channel to listen for interrupt or terminate signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	//Web Analyzer Server
	go func() {
		log.Logger.Info("Server started on :" + config.AppConfig.PublicWebServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Pprof only enabled in dev env
	if config.AppConfig.IsDev == "true" {
		go func() {
			debug.StartPprof(":" + config.AppConfig.PprofWebServerPort)
		}()
	}

	//Prometheus server
	go func() {
		log.Logger.Info("Metrics server started on :8081")
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("Metrics server failed", zap.Error(err))
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
