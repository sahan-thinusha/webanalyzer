package router

import (
	"net/http"
	"webanalyzer/internal/api/v1/handler"
)

func New() http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthCheckHandler)

	return mux
}
