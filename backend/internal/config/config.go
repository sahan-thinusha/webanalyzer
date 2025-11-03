package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"webanalyzer/internal/log"
)

type Config struct {
	BasicAuthUser        string `mapstructure:"BASIC_AUTH_USER"`
	BasicAuthPass        string `mapstructure:"BASIC_AUTH_PASS"`
	IsDev                string `mapstructure:"IS_DEV"`
	PublicWebServerPort  string `mapstructure:"PUBLIC_WEB_SERVER_PORT"`
	MetricsWebServerPort string `mapstructure:"METRICS_WEB_SERVER_PORT"`
	PprofWebServerPort   string `mapstructure:"PPROF_WEB_SERVER_PORT"`
}

var AppConfig *Config

func LoadEnv() {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		log.Logger.Error(".env file not found")
	}

	v.AutomaticEnv()

	v.SetDefault(BASIC_AUTH_USER, "")
	v.SetDefault(BASIC_AUTH_PASS, "")
	v.SetDefault(PUBLIC_WEB_SERVER_PORT, "8080")
	v.SetDefault(PPROF_WEB_SERVER_PORT, "6061")
	v.SetDefault(METRICS_WEB_SERVER_PORT, "8081")
	v.SetDefault(IS_DEV, "false")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Logger.Fatal("Failed to unmarshal config", zap.Error(err))
	}

	AppConfig = &cfg

	if AppConfig.BasicAuthUser == "" || AppConfig.BasicAuthPass == "" {
		log.Logger.Fatal("BASIC_AUTH_USER and BASIC_AUTH_PASS must be set")
	}

}
