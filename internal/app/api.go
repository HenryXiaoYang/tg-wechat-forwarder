package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type apiServer struct {
	auth      *authenticator
	store     *store
	telegram  *telegramService
	push      *pushService
	chevereto *cheveretoClient
	static    fs.FS
}

type dashboardResponse struct {
	Version     string            `json:"version"`
	Telegram    telegramState     `json:"telegram"`
	Dialogs     []dialog          `json:"dialogs"`
	PushPlus    pushSettings      `json:"pushplus"`
	Chevereto   cheveretoSettings `json:"chevereto"`
	FilterRules []string          `json:"filterRules"`
	Deliveries  []delivery        `json:"deliveries"`
	PushStatus  string            `json:"pushStatus"`
}

func (a *apiServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": version})
	})
	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("POST /api/logout", a.logout)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/dashboard", a.dashboard)
	protected.HandleFunc("PATCH /api/dialogs/{peerKey}", a.updateDialog)
	protected.HandleFunc("POST /api/telegram/password", a.telegramPassword)
	protected.HandleFunc("POST /api/telegram/logout", a.telegramLogout)
	protected.HandleFunc("POST /api/telegram/refresh", a.telegramRefresh)
	protected.HandleFunc("PUT /api/settings/pushplus", a.savePushPlus)
	protected.HandleFunc("POST /api/settings/pushplus/test", a.testPushPlus)
	protected.HandleFunc("POST /api/push/message", a.sendMessage)
	protected.HandleFunc("PUT /api/settings/chevereto", a.saveChevereto)
	protected.HandleFunc("POST /api/settings/chevereto/test", a.testChevereto)
	protected.HandleFunc("PUT /api/settings/filters", a.saveFilters)
	mux.Handle("/api/", a.auth.require(protected))
	mux.Handle("/", spaHandler(a.static))
	return securityHeaders(mux)
}

func (a *apiServer) login(w http.ResponseWriter, r *http.Request) {
	if a.auth.limited(r.RemoteAddr) {
		writeError(w, http.StatusTooManyRequests, "登录失败次数过多，请稍后再试")
		return
	}
	var input struct{ Username, Password string }
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !a.auth.validCredentials(input.Username, input.Password) {
		a.auth.failed(r.RemoteAddr)
		time.Sleep(250 * time.Millisecond)
		writeError(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	a.auth.clearFailures(r.RemoteAddr)
	a.auth.setSession(w, r)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *apiServer) logout(w http.ResponseWriter, _ *http.Request) {
	a.auth.clearSession(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) dashboard(w http.ResponseWriter, r *http.Request) {
	dialogs, err := a.store.dialogs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if dialogs == nil {
		dialogs = []dialog{}
	}
	pushSettings, err := a.push.settings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	imageSettings, err := a.chevereto.settings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rules, err := a.telegram.filterRules(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rules == nil {
		rules = []string{}
	}
	deliveries := a.push.deliveries(r.Context())
	if deliveries == nil {
		deliveries = []delivery{}
	}
	writeJSON(w, http.StatusOK, dashboardResponse{
		Version: version, Telegram: a.telegram.snapshot(), Dialogs: dialogs,
		PushPlus: pushSettings, Chevereto: imageSettings, FilterRules: rules,
		Deliveries: deliveries, PushStatus: a.push.status(),
	})
}

func (a *apiServer) updateDialog(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Selected *bool `json:"selected"`
		AdFilter *bool `json:"adFilter"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Selected == nil && input.AdFilter == nil {
		writeError(w, http.StatusBadRequest, "没有可更新的字段")
		return
	}
	if err := a.store.updateDialog(r.Context(), r.PathValue("peerKey"), input.Selected, input.AdFilter); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, context.Canceled) {
			status = http.StatusRequestTimeout
		}
		writeError(w, status, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) telegramPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password string `json:"password"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.telegram.submitPassword(r.Context(), input.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) telegramLogout(w http.ResponseWriter, r *http.Request) {
	if err := a.telegram.logout(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) telegramRefresh(w http.ResponseWriter, r *http.Request) {
	if err := a.telegram.refresh(r.Context()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) savePushPlus(w http.ResponseWriter, r *http.Request) {
	var input pushSettingsUpdate
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.push.updateSettings(r.Context(), input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) testPushPlus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
	defer cancel()
	shortCode, err := a.push.test(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"shortCode": shortCode})
}

func (a *apiServer) sendMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content string `json:"content"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.push.sendManual(r.Context(), input.Content); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) saveChevereto(w http.ResponseWriter, r *http.Request) {
	var input cheveretoUpdate
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.chevereto.update(r.Context(), input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *apiServer) testChevereto(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	imageURL, err := a.chevereto.test(ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"imageUrl": imageURL})
}

func (a *apiServer) saveFilters(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Rules []string `json:"rules"`
	}
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.telegram.saveFilters(r.Context(), input.Rules); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func readJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return errors.New("请求内容不是有效 JSON")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("请求只能包含一个 JSON 对象")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		if r.URL.Path != "/api/login" && r.Method != http.MethodGet && r.Method != http.MethodHead {
			if origin := r.Header.Get("Origin"); origin != "" {
				if parsed, err := url.Parse(origin); err != nil || !strings.EqualFold(parsed.Host, r.Host) {
					writeError(w, http.StatusForbidden, "请求来源无效")
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func spaHandler(static fs.FS) http.Handler {
	files := http.FileServer(http.FS(static))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(static, name); err != nil {
			r.URL.Path = "/"
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}
