package main

import (
	"log/slog"
	"os"

	"github.com/hxy/tg-wechat-forwarder/internal/app"
)

var version = "dev"

func main() {
	if err := app.Run(version); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
