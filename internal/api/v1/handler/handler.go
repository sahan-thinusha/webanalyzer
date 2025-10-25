package handler

import (
	"encoding/json"
	"net/http"
	"webanalyzer/internal/service"
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

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		response.Error(w, http.StatusBadRequest, "missing 'url' query parameter")
		return
	}

	result := service.AnalyzePage(targetURL)
	if result == nil {
		response.Error(w, http.StatusInternalServerError, "failed to analyze page")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
}
