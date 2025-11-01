package middleware

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
	"webanalyzer/internal/log"
	"webanalyzer/internal/util"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		clientIP := util.GetClientIPAddress(r)

		requestID := uuid.New().String()
		w.Header().Set("X-Request-ID", requestID)

		log.Logger.Info("HTTP Request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("ip", clientIP),
			zap.Int("status", lrw.statusCode),
			zap.Duration("duration", duration),
		)
	})
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
