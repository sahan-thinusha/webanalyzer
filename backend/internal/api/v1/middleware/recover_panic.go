package middleware

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

func RecoverPanic(logger *zap.Logger, errorHandler func(http.ResponseWriter, *http.Request, error), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.ByteString("stack", debug.Stack()),
					zap.String("method", r.Method),
					zap.String("url", r.URL.String()),
					zap.String("remote_addr", r.RemoteAddr),
				)

				errorHandler(w, r, fmt.Errorf("%v", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
