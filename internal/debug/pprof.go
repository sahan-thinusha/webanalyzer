package debug

import (
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"webanalyzer/internal/log"
)

func StartPprof(host string) {
	go func() {
		log.Logger.Info("pprof listening", zap.String("host", host))
		if err := http.ListenAndServe(host, nil); err != nil {
			log.Logger.Fatal("pprof failed", zap.Error(err))
		}
	}()
}
