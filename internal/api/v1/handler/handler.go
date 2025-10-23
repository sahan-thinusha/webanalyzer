package handler

import (
	"net/http"
	"webanalyzer/pkg/response"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	response.Success(w, resp, "")
}
