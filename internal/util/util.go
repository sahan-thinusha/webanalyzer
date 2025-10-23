package util

import "net/http"

func GetClientIPAddress(r *http.Request) string {
	if forwardedIP := r.Header.Get("X-Forwarded-For"); forwardedIP != "" {
		return forwardedIP
	}
	ip := r.RemoteAddr
	return ip
}
