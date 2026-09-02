package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	webui "github.com/hxy/tg-wechat-forwarder/web"
)

var version = "dev"

func Run(buildVersion string) error {
	if buildVersion != "" {
		version = buildVersion
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	store, err := openStore(cfg.AppSecret)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	push := newPushService(store)
	chevereto := newCheveretoClient(store)
	telegram := newTelegramService(cfg, store, push, chevereto)
	push.start(ctx)
	telegram.start(ctx)

	static, err := webui.Dist()
	if err != nil {
		return err
	}
	api := &apiServer{
		auth: newAuthenticator(cfg), store: store, telegram: telegram,
		push: push, chevereto: chevereto, static: static,
	}
	server := &http.Server{
		Addr: cfg.ListenAddr, Handler: api.handler(),
		ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 70 * time.Second, IdleTimeout: 2 * time.Minute,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = server.Shutdown(shutdownCtx)
	}()
	slog.Info("listening", "address", cfg.ListenAddr, "data", dataFile, "version", version)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
