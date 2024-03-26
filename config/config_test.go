package config

import (
	"context"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestConfigureLoggerSetsLogLevel(t *testing.T) {
	setProjectDir()

	expected := "debug"

	ConfigureLogger(AppConfig{App: App{LogLevel: expected}})

	assert.True(t, slog.Default().Enabled(context.Background(), slog.LevelDebug))
}

func TestParseConfigWithEnv(t *testing.T) {
	setProjectDir()

	expected := "test"
	t.Setenv("GICS_BASE_URL", expected)

	config := LoadConfig()

	assert.Equal(t, expected, config.Gics.BaseUrl)
}

func TestParseConfigFileNotFound(t *testing.T) {
	setProjectDir()

	// config file not found
	_, err := parseConfig("./bla")

	assert.ErrorIs(t, err, err.(viper.ConfigFileNotFoundError))
}

func setProjectDir() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	_ = os.Chdir(dir)

	viper.Reset()
}
