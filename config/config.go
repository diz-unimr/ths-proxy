package config

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"strings"
)

type AppConfig struct {
	App  App  `mapstructure:"app"`
	Gics Gics `mapstructure:"gics"`
}

type Http struct {
	Port string `mapstructure:"port"`
}

type App struct {
	LogLevel string `mapstructure:"log-level"`
	Http     Http   `mapstructure:"http"`
}

type Gics struct {
	BaseUrl string `mapstructure:"base-url"`
	Host    string
}

func LoadConfig() AppConfig {
	c, err := parseConfig(".")
	if err != nil {
		slog.Error("Unable to load config file", "error", err)
		os.Exit(1)
	}

	return *c
}

func parseConfig(path string) (config *AppConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	return config, err
}

func ConfigureLogger(c AppConfig) {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))
	slog.SetDefault(logger)

	// set configured log level
	err := lvl.UnmarshalText([]byte(c.App.LogLevel))
	if err != nil {
		slog.Error("Unable to set Log level from application properties", "level", c.App.LogLevel, "error", err)
	}
}
