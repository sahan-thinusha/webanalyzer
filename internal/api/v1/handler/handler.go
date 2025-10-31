package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
	"webanalyzer/internal/service"
	"webanalyzer/internal/util"
	"webanalyzer/pkg/response"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	response.Success(w, resp, "")
}

func AnalyzePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		response.Error(w, http.StatusBadRequest, "missing 'url' query parameter")
		return
	}

	if !util.IsValidURL(url) {
		response.Error(w, http.StatusBadRequest, "invalid 'url' format")
		return
	}

	result, err := service.AnalyzePage(url)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			statusCode = http.StatusGatewayTimeout
		case strings.Contains(err.Error(), "connection refused"),
			strings.Contains(err.Error(), "no such host"),
			strings.Contains(err.Error(), "unexpected status code: 403"),
			strings.Contains(err.Error(), "timeout"):
			statusCode = http.StatusBadGateway
		case strings.Contains(err.Error(), "service unavailable"):
			statusCode = http.StatusServiceUnavailable
		default:
			statusCode = http.StatusInternalServerError
		}
		response.Error(w, statusCode, fmt.Sprintf("failed to analyze page: %v", err))
		return
	}

	if result == nil {
		response.Error(w, http.StatusInternalServerError, "failed to analyze page")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response.Success(w, result, "")
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
