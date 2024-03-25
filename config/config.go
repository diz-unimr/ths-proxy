package config

import (
	"github.com/spf13/viper"
	"log/slog"
	"net/url"
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

func (g *Gics) parse() *url.URL {
	gicsUrl, _ := url.Parse(g.BaseUrl)
	return gicsUrl
}

func (g *Gics) validate() bool {
	_, err := url.Parse(g.BaseUrl)
	return err == nil
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

func ConfigureLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

}
