package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res := Response{
		Status:  http.StatusText(statusCode),
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(res)
}

func Success(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusOK, data, message)
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, nil, message)
}
