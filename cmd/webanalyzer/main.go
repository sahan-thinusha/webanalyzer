package main

import (
	"go.uber.org/zap"
	"net/http"
	"webanalyzer/internal/api/v1/router"
	"webanalyzer/internal/log"
)

func init() {
	log.InitLogger()
}

func main() {
	defer log.Sync()

	r := router.New()

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Logger.Info("Server started..")
	if err := server.ListenAndServe(); err != nil {
		log.Logger.Fatal("Server failed", zap.Error(err))
	}
}
