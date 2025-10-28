package handler

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
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

	result := service.AnalyzePage(url)
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
