package debug

import (
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
	"webanalyzer/internal/api/v1/middleware"
	"webanalyzer/internal/log"
)

func StartPprof(host string) {
	go func() {
		log.Logger.Info("pprof listening", zap.String("host", host))

		mux := http.NewServeMux()

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		authHandler := middleware.BasicAuth()(mux)

		server := &http.Server{
			Addr:    host,
			Handler: authHandler,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Logger.Fatal("pprof failed", zap.Error(err))
		}
	}()
}
