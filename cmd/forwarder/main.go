package main

import (
	"log/slog"
	"os"

	"github.com/HenryXiaoYang/tg-wechat-forwarder/internal/app"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "hash" {
		if err := app.PrintPasswordHash(); err != nil {
			slog.Error("hash password", "error", err)
			os.Exit(1)
		}
		return
	}
	if err := app.Run(version); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
