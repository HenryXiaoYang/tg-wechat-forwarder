package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	ListenAddr      string
	AdminUsername   string
	AdminPassword   string
	AppSecret       string
	TelegramAppID   int
	TelegramAppHash string
}

func loadConfig() (config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return config{}, fmt.Errorf("load .env: %w", err)
	}

	appID, err := strconv.Atoi(strings.TrimSpace(os.Getenv("TELEGRAM_API_ID")))
	if err != nil || appID <= 0 {
		return config{}, errors.New("TELEGRAM_API_ID must be a positive integer")
	}

	cfg := config{
		ListenAddr:      envOr("LISTEN_ADDR", ":8080"),
		AdminUsername:   strings.TrimSpace(os.Getenv("ADMIN_USERNAME")),
		AdminPassword:   os.Getenv("ADMIN_PASSWORD"),
		AppSecret:       os.Getenv("APP_SECRET"),
		TelegramAppID:   appID,
		TelegramAppHash: strings.TrimSpace(os.Getenv("TELEGRAM_API_HASH")),
	}
	if cfg.AdminUsername == "" {
		return config{}, errors.New("ADMIN_USERNAME is required")
	}
	if len(cfg.AdminPassword) < 12 {
		return config{}, errors.New("ADMIN_PASSWORD must be at least 12 characters")
	}
	if cfg.AdminPassword == "change-this-password" {
		return config{}, errors.New("ADMIN_PASSWORD still has the unsafe example value")
	}
	if len(cfg.AppSecret) < 32 {
		return config{}, errors.New("APP_SECRET must be at least 32 characters")
	}
	if strings.HasPrefix(cfg.AppSecret, "replace-with-") {
		return config{}, errors.New("APP_SECRET still has the unsafe example value")
	}
	if cfg.TelegramAppHash == "" {
		return config{}, errors.New("TELEGRAM_API_HASH is required")
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
