package middleware

import (
	"net/http"
	"time"
)

var lastRequest = make(map[string]time.Time)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		now := time.Now()

		if t, ok := lastRequest[ip]; ok && now.Sub(t) < time.Second {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		lastRequest[ip] = now
		next.ServeHTTP(w, r)
	})
}
