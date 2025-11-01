package response

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"webanalyzer/internal/log"
)

type Response struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code,omitempty"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res := Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Logger.Error("failed to encode JSON response", zap.Error(err))
		return
	}
}

func Success(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusOK, data, message)
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, nil, message)
}
