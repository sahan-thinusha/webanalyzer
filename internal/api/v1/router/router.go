package router

import (
	"net/http"
	"webanalyzer/internal/api/v1/handler"
	"webanalyzer/internal/api/v1/middleware"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthCheckHandler)

	return middleware.LoggingMiddleware(middleware.CORS(middleware.RateLimit(mux)))
}
