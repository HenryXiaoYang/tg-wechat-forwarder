package app

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const sessionCookie = "forwarder_session"

type authenticator struct {
	username [32]byte
	password [32]byte
	key      [32]byte
	mu       sync.Mutex
	failures map[string][]time.Time
}

func newAuthenticator(cfg config) *authenticator {
	passwordHash := sha256.Sum256([]byte(cfg.AdminPassword))
	return &authenticator{
		username: sha256.Sum256([]byte(cfg.AdminUsername)),
		password: passwordHash,
		key:      sha256.Sum256([]byte("session\x00" + cfg.AppSecret + "\x00" + base64.RawURLEncoding.EncodeToString(passwordHash[:]))),
		failures: make(map[string][]time.Time),
	}
}

func (a *authenticator) validCredentials(username, password string) bool {
	u := sha256.Sum256([]byte(username))
	p := sha256.Sum256([]byte(password))
	return subtle.ConstantTimeCompare(u[:], a.username[:])&subtle.ConstantTimeCompare(p[:], a.password[:]) == 1
}

func (a *authenticator) limited(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	now := time.Now()
	cutoff := now.Add(-15 * time.Minute)
	a.mu.Lock()
	defer a.mu.Unlock()
	items := a.failures[host][:0]
	for _, item := range a.failures[host] {
		if item.After(cutoff) {
			items = append(items, item)
		}
	}
	a.failures[host] = items
	return len(items) >= 8
}

func (a *authenticator) failed(remote string) {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	a.mu.Lock()
	a.failures[host] = append(a.failures[host], time.Now())
	a.mu.Unlock()
}

func (a *authenticator) clearFailures(remote string) {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	a.mu.Lock()
	delete(a.failures, host)
	a.mu.Unlock()
}

func (a *authenticator) setSession(w http.ResponseWriter, r *http.Request) {
	payload := make([]byte, 8+16)
	binary.BigEndian.PutUint64(payload[:8], uint64(time.Now().Add(24*time.Hour).Unix()))
	_, _ = rand.Read(payload[8:])
	mac := hmac.New(sha256.New, a.key[:])
	_, _ = mac.Write(payload)
	value := append(payload, mac.Sum(nil)...)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    base64.RawURLEncoding.EncodeToString(value),
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"),
		SameSite: http.SameSiteStrictMode,
	})
}

func (a *authenticator) clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
}

func (a *authenticator) authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	value, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil || len(value) != 8+16+sha256.Size {
		return false
	}
	payload, signature := value[:24], value[24:]
	mac := hmac.New(sha256.New, a.key[:])
	_, _ = mac.Write(payload)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return false
	}
	expires := int64(binary.BigEndian.Uint64(payload[:8]))
	return time.Now().Unix() < expires
}

func (a *authenticator) require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authenticated(r) {
			writeError(w, http.StatusUnauthorized, "请先登录")
			return
		}
		next.ServeHTTP(w, r)
	})
}
