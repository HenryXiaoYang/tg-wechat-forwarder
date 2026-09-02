package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type cheveretoSettings struct {
	BaseURL       string `json:"baseUrl"`
	KeyConfigured bool   `json:"keyConfigured"`
}

type cheveretoUpdate struct {
	BaseURL  string `json:"baseUrl"`
	APIKey   string `json:"apiKey"`
	ClearKey bool   `json:"clearKey"`
}

type cheveretoClient struct {
	store  *store
	client *http.Client
}

func newCheveretoClient(store *store) *cheveretoClient {
	return &cheveretoClient{
		store: store,
		client: &http.Client{
			Timeout:       45 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (c *cheveretoClient) settings(ctx context.Context) (cheveretoSettings, error) {
	baseURL, _, err := c.store.get(ctx, "chevereto_base_url", false)
	if err != nil {
		return cheveretoSettings{}, err
	}
	_, configured, err := c.store.get(ctx, "chevereto_api_key", true)
	return cheveretoSettings{BaseURL: baseURL, KeyConfigured: configured}, err
}

func (c *cheveretoClient) update(ctx context.Context, input cheveretoUpdate) error {
	baseURL := strings.TrimSpace(input.BaseURL)
	if baseURL != "" {
		if _, err := cheveretoEndpoint(baseURL); err != nil {
			return err
		}
	}
	if err := c.store.set(ctx, "chevereto_base_url", baseURL, false); err != nil {
		return err
	}
	if input.ClearKey {
		return c.store.deleteSetting(ctx, "chevereto_api_key")
	}
	if strings.TrimSpace(input.APIKey) != "" {
		return c.store.set(ctx, "chevereto_api_key", strings.TrimSpace(input.APIKey), true)
	}
	return nil
}

func cheveretoEndpoint(baseURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("Chevereto API Base URL 必须是完整的 http(s) 地址")
	}
	u.RawQuery, u.Fragment = "", ""
	clean := strings.TrimRight(u.Path, "/")
	switch {
	case strings.HasSuffix(clean, "/api/1/upload"):
	case strings.HasSuffix(clean, "/api/1"):
		clean += "/upload"
	default:
		clean = path.Join(clean, "/api/1/upload")
	}
	u.Path = clean
	return u.String(), nil
}

func (c *cheveretoClient) upload(ctx context.Context, filename, mime string, data []byte, expiration string) (string, error) {
	filename = filepath.Base(strings.NewReplacer("\r", "", "\n", "").Replace(filename))
	if filename == "." || filename == "" {
		filename = "telegram-image"
	}
	baseURL, ok, err := c.store.get(ctx, "chevereto_base_url", false)
	if err != nil {
		return "", err
	}
	if !ok || strings.TrimSpace(baseURL) == "" {
		return "", errors.New("Chevereto API Base URL 未配置")
	}
	apiKey, ok, err := c.store.get(ctx, "chevereto_api_key", true)
	if err != nil {
		return "", err
	}
	if !ok || apiKey == "" {
		return "", errors.New("Chevereto API Key 未配置")
	}
	endpoint, err := cheveretoEndpoint(baseURL)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("source", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	_ = writer.WriteField("key", apiKey)
	_ = writer.WriteField("format", "json")
	if expiration != "" {
		_ = writer.WriteField("expiration", expiration)
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload to Chevereto: %w", err)
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Chevereto returned %s: %s", resp.Status, strings.TrimSpace(string(response)))
	}
	var result struct {
		StatusCode int    `json:"status_code"`
		StatusText string `json:"status_txt"`
		Image      struct {
			URL        string `json:"url"`
			DisplayURL string `json:"display_url"`
		} `json:"image"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("decode Chevereto response: %w", err)
	}
	imageURL := result.Image.URL
	if imageURL == "" {
		imageURL = result.Image.DisplayURL
	}
	parsed, err := url.Parse(imageURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("Chevereto response has no valid image URL (%s)", result.StatusText)
	}
	_ = mime // Kept for callers and future compatible APIs; multipart detects by filename.
	return imageURL, nil
}

func (c *cheveretoClient) test(ctx context.Context) (string, error) {
	// 1x1 transparent PNG; PT5M asks compatible Chevereto servers to clean it up.
	data, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M/wHwAF/gL+X8W+WQAAAABJRU5ErkJggg==")
	return c.upload(ctx, "forwarder-test.png", "image/png", data, "PT5M")
}
