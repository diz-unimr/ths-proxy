package main

import (
	"github.com/diz-unimr/ths-proxy/config"
	"github.com/diz-unimr/ths-proxy/web"
	"log/slog"
)

func main() {
	appConfig := config.LoadConfig()
	config.ConfigureLogger()

	server := web.NewServer(appConfig)
	slog.Error("Server failed to run", "error", server.Run())
}
