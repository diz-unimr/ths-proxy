package main

import (
	"log/slog"

	"github.com/diz-unimr/ths-proxy/pkg/config"
	"github.com/diz-unimr/ths-proxy/pkg/web"
)

func main() {
	appConfig := config.LoadConfig()
	config.ConfigureLogger(appConfig)

	server := web.NewServer(appConfig)
	slog.Error("Server failed to run", "error", server.Run())
}
