package app

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	pushplus "github.com/pushplus/perk-pushplus-go-sdk"
)

func TestCoreHelpers(t *testing.T) {
	t.Run("Chevereto endpoint", func(t *testing.T) {
		for input, want := range map[string]string{
			"https://img.example.com":              "https://img.example.com/api/1/upload",
			"https://img.example.com/api/1":        "https://img.example.com/api/1/upload",
			"https://img.example.com/api/1/upload": "https://img.example.com/api/1/upload",
		} {
			got, err := cheveretoEndpoint(input)
			if err != nil || got != want {
				t.Fatalf("cheveretoEndpoint(%q) = %q, %v; want %q", input, got, err, want)
			}
		}
	})

	t.Run("delivery states replace instead of duplicate", func(t *testing.T) {
		p := &pushService{}
		p.addRecent(delivery{Key: "channel:1:2", Title: "queued", Content: "body", Status: "queued"})
		p.addRecent(delivery{Key: "channel:1:2", ShortCode: "short", Title: "sent", Status: "accepted"})
		p.history = []delivery{{ShortCode: "short", Title: "remote", Status: "history"}}
		items := p.deliveries(context.Background())
		if len(items) != 1 || items[0].Content != "body" || items[0].Status != "accepted" {
			t.Fatalf("unexpected deliveries: %#v", items)
		}
	})

	t.Run("only transport and server errors are retried", func(t *testing.T) {
		for _, permanent := range []int{600, 900, 903, 905, 999} {
			if retryableSendError(&pushplus.Error{Code: permanent, Msg: "permanent"}) {
				t.Fatalf("code %d was queued for retry", permanent)
			}
		}
		if !retryableSendError(&pushplus.Error{Code: 500, Msg: "server"}) {
			t.Fatal("server error was not retried")
		}
		if !retryableSendError(&pushplus.Error{Code: 999, Cause: errors.New("connection reset")}) {
			t.Fatal("transport failure was not retried")
		}
		if !retryableSendError(errors.New("dial tcp: timeout")) {
			t.Fatal("plain transport error was not retried")
		}
	})

	t.Run("manual push rejects empty text and missing configuration", func(t *testing.T) {
		p := &pushService{}
		if err := p.sendManual(context.Background(), "   "); err == nil {
			t.Fatal("empty manual message was accepted")
		}
		if err := p.sendManual(context.Background(), "hello"); err == nil {
			t.Fatal("manual message was accepted without PushPlus configuration")
		}
	})

	t.Run("Telegram connection timeout is actionable", func(t *testing.T) {
		service := &telegramService{state: telegramState{Status: "connecting"}}
		service.markConnectTimeout()
		if state := service.snapshot(); state.Status != "error" || state.Error == "" {
			t.Fatalf("unexpected timeout state: %#v", state)
		}
		service.state = telegramState{Status: "qr"}
		service.markConnectTimeout()
		if state := service.snapshot(); state.Status != "qr" {
			t.Fatalf("watchdog replaced active state: %#v", state)
		}
	})

	t.Run("filter rules", func(t *testing.T) {
		keyword, err := compileRule("推广")
		if err != nil || !keyword.MatchString("限时推广") {
			t.Fatal("keyword rule did not match")
		}
		regex, err := compileRule("re:(?i)promo\\s+code")
		if err != nil || !regex.MatchString("PROMO CODE") {
			t.Fatal("regex rule did not match")
		}
		if _, err := compileRule("re:["); err == nil {
			t.Fatal("invalid regexp was accepted")
		}
	})

	t.Run("signed session", func(t *testing.T) {
		a := newAuthenticator(config{AdminUsername: "admin", AdminPassword: "long-test-password", AppSecret: "test-secret-that-is-at-least-32-bytes"})
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "http://example.test", nil)
		a.setSession(recorder, request)
		response := recorder.Result()
		cookies := response.Cookies()
		if len(cookies) != 1 {
			t.Fatal("session cookie was not created")
		}
		request.AddCookie(cookies[0])
		if !a.authenticated(request) {
			t.Fatal("valid session cookie was rejected")
		}
		cookies[0].Value += "x"
		tampered := httptest.NewRequest("GET", "http://example.test", nil)
		tampered.AddCookie(cookies[0])
		if a.authenticated(tampered) {
			t.Fatal("tampered session cookie was accepted")
		}
	})

	if got := truncate("12345", 4); got != "123…" {
		t.Fatalf("truncate = %q", got)
	}
}
