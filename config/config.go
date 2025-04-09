package config

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"strings"
)

type AppConfig struct {
	App          App          `mapstructure:"app"`
	Gics         Gics         `mapstructure:"gics"`
	Notification Notification `mapstructure:"notification"`
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
}

type Recipients struct {
	Info  string `mapstructure:"info"`
	Debug string `mapstructure:"debug"`
}

type Email struct {
	Recipients Recipients `mapstructure:"recipients"`
	Sender     string     `mapstructure:"sender"`
	Smtp       Smtp       `mapstructure:"smtp"`
}

type Smtp struct {
	Server   string `mapstructure:"server"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Helo     string `mapstructure:"helo"`
}

type Notification struct {
	Email        Email  `mapstructure:"email"`
	MatchService string `mapstructure:"match-service"`
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
