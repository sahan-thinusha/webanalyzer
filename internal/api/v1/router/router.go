package router

import (
	"net/http"
	"webanalyzer/internal/api/v1/handler"
	"webanalyzer/internal/api/v1/middleware"
	"webanalyzer/internal/log"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthCheckHandler)

	return middleware.RecoverPanic(
		log.Logger,
		func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		},
		middleware.SecureHeaders(
			middleware.Logging(
				middleware.Compression(
					middleware.CORS(
						middleware.RateLimit(mux),
					),
				),
			),
		),
	)
}
