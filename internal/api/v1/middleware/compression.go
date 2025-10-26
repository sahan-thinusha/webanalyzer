package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				return
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		next.ServeHTTP(gzw, r)
	})
}

func (gzw gzipResponseWriter) Write(b []byte) (int, error) {
	return gzw.Writer.Write(b)
}
