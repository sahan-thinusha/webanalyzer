package util

import (
	"net/http"
	"net/url"
	"regexp"
)

func GetClientIPAddress(r *http.Request) string {
	if forwardedIP := r.Header.Get("X-Forwarded-For"); forwardedIP != "" {
		return forwardedIP
	}
	ip := r.RemoteAddr
	return ip
}

var urlPattern = regexp.MustCompile(`^(https?://)?([a-zA-Z0-9.-]+)(:[0-9]+)?(/.*)?$`)

func IsValidURL(input string) bool {
	if input == "" {
		return false
	}

	if !urlPattern.MatchString(input) {
		return false
	}

	u, err := url.Parse(input)
	if err != nil {
		return false
	}

	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}
