package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"webanalyzer/internal/log"
)

type Config struct {
	BasicAuthUser string `mapstructure:"BASIC_AUTH_USER"`
	BasicAuthPass string `mapstructure:"BASIC_AUTH_PASS"`
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

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Logger.Fatal("Failed to unmarshal config", zap.Error(err))
	}

	AppConfig = &cfg

	if AppConfig.BasicAuthUser == "" || AppConfig.BasicAuthPass == "" {
		log.Logger.Fatal("BASIC_AUTH_USER and BASIC_AUTH_PASS must be set")
	}

}
