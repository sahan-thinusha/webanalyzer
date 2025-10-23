package log

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	Logger = l
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
