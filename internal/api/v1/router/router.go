package router

import (
	"net/http"
	"webanalyzer/internal/api/v1/handler"
	"webanalyzer/internal/api/v1/middleware"
	"webanalyzer/internal/log"
)

func New() http.Handler {
	appName := "webanalyzer"
	apiVersion := "v1"
	basePath := "/" + appName + "/api/" + apiVersion

	mux := http.NewServeMux()

	register := func(path string, h http.HandlerFunc) {
		mux.HandleFunc(basePath+path, h)
	}

	register("/health", handler.HealthCheckHandler)
	register("/analyze", handler.AnalyzePageHandler)

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
