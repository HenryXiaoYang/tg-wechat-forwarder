package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	pushplus "github.com/pushplus/perk-pushplus-go-sdk"
)

type pushSettings struct {
	Topic            string `json:"topic"`
	TokenConfigured  bool   `json:"tokenConfigured"`
	SecretConfigured bool   `json:"secretConfigured"`
}

type pushSettingsUpdate struct {
	Token       string `json:"token"`
	SecretKey   string `json:"secretKey"`
	Topic       string `json:"topic"`
	ClearToken  bool   `json:"clearToken"`
	ClearSecret bool   `json:"clearSecret"`
}

type delivery struct {
	Key       string    `json:"-"`
	CreatedAt time.Time `json:"-"`
	ShortCode string    `json:"shortCode,omitempty"`
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	Time      string    `json:"time"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

type pushService struct {
	store          *store
	mu             sync.RWMutex
	client         *pushplus.Client
	token          string
	secret         string
	topic          string
	pace           time.Duration
	lastSend       time.Time
	wake           chan struct{}
	recent         []delivery
	history        []delivery
	historyAt      time.Time
	historyLoading bool
}

func newPushService(store *store) *pushService {
	return &pushService{store: store, pace: 13 * time.Second, wake: make(chan struct{}, 1)}
}

func (p *pushService) start(ctx context.Context) {
	if err := p.reload(ctx); err != nil {
		slog.Error("load PushPlus settings", "error", err)
	}
	go p.worker(ctx)
}

func (p *pushService) settings(ctx context.Context) (pushSettings, error) {
	topic, _, err := p.store.get(ctx, "pushplus_topic", false)
	if err != nil {
		return pushSettings{}, err
	}
	_, tokenOK, err := p.store.get(ctx, "pushplus_token", true)
	if err != nil {
		return pushSettings{}, err
	}
	_, secretOK, err := p.store.get(ctx, "pushplus_secret", true)
	return pushSettings{Topic: topic, TokenConfigured: tokenOK, SecretConfigured: secretOK}, err
}

func (p *pushService) updateSettings(ctx context.Context, input pushSettingsUpdate) error {
	if len([]rune(strings.TrimSpace(input.Topic))) > 100 {
		return errors.New("Topic 不能超过 100 个字符")
	}
	if err := p.store.set(ctx, "pushplus_topic", strings.TrimSpace(input.Topic), false); err != nil {
		return err
	}
	if input.ClearToken {
		if err := p.store.deleteSetting(ctx, "pushplus_token"); err != nil {
			return err
		}
	} else if strings.TrimSpace(input.Token) != "" {
		if err := p.store.set(ctx, "pushplus_token", strings.TrimSpace(input.Token), true); err != nil {
			return err
		}
	}
	if input.ClearSecret {
		if err := p.store.deleteSetting(ctx, "pushplus_secret"); err != nil {
			return err
		}
	} else if strings.TrimSpace(input.SecretKey) != "" {
		if err := p.store.set(ctx, "pushplus_secret", strings.TrimSpace(input.SecretKey), true); err != nil {
			return err
		}
	}
	return p.reload(ctx)
}

func (p *pushService) reload(ctx context.Context) error {
	token, _, err := p.store.get(ctx, "pushplus_token", true)
	if err != nil {
		return err
	}
	secret, _, err := p.store.get(ctx, "pushplus_secret", true)
	if err != nil {
		return err
	}
	topic, _, err := p.store.get(ctx, "pushplus_topic", false)
	if err != nil {
		return err
	}
	var client *pushplus.Client
	if token != "" {
		client = pushplus.NewClient(pushplus.WithToken(token), pushplus.WithSecretKey(secret))
	}
	p.mu.Lock()
	p.client, p.token, p.secret, p.topic = client, token, secret, topic
	p.pace = 13 * time.Second
	p.historyAt = time.Time{}
	p.mu.Unlock()
	if client != nil && secret != "" {
		go p.detectVIP(client)
	}
	select {
	case p.wake <- struct{}{}:
	default:
	}
	return nil
}

func (p *pushService) detectVIP(client *pushplus.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if me, err := client.User().MyInfo(ctx); err == nil && me.VipInfo != nil && me.VipInfo.IsVip == 1 {
		p.mu.Lock()
		if p.client == client {
			p.pace = 2100 * time.Millisecond
		}
		p.mu.Unlock()
	}
}

func (p *pushService) enqueue(ctx context.Context, message queuedMessage) error {
	if err := p.store.enqueue(ctx, message); err != nil {
		return err
	}
	p.addRecent(delivery{Key: message.DedupeKey, CreatedAt: message.CreatedAt, Title: message.Title, Content: message.Content, Time: message.CreatedAt.Format("01-02 15:04"), Status: "queued"})
	select {
	case p.wake <- struct{}{}:
	default:
	}
	return nil
}

// sendManual queues a message typed in the control panel; it rides the same
// queue, pacing, retry and delivery tracking as forwarded Telegram messages.
func (p *pushService) sendManual(ctx context.Context, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("消息不能为空")
	}
	p.mu.RLock()
	configured := p.client != nil
	p.mu.RUnlock()
	if !configured {
		return errors.New("请先配置 PushPlus Token")
	}
	now := time.Now()
	return p.enqueue(ctx, queuedMessage{
		DedupeKey: fmt.Sprintf("manual:%d", now.UnixNano()),
		Title:     "控制台消息",
		Content:   truncate(content, 19000),
		Template:  "txt",
		CreatedAt: now,
	})
}

func (p *pushService) worker(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-p.wake:
		}
		p.sendNext(ctx)
	}
}

func (p *pushService) sendNext(ctx context.Context) {
	cutoff := time.Now().Add(-10 * time.Minute)
	if err := p.store.deleteExpiredQueued(ctx, cutoff); err != nil {
		slog.Error("expire send queue", "error", err)
		return
	}
	p.expireRecent(cutoff)
	p.mu.RLock()
	client, topic, pace, last := p.client, p.topic, p.pace, p.lastSend
	p.mu.RUnlock()
	job, ok, err := p.store.nextQueued(ctx, time.Now())
	if err != nil {
		slog.Error("read send queue", "error", err)
		return
	}
	if !ok {
		return
	}
	if time.Since(job.CreatedAt) > 10*time.Minute {
		_ = p.store.deleteQueued(ctx, job.ID)
		p.addRecent(delivery{Key: job.DedupeKey, Title: job.Title, Time: time.Now().Format("01-02 15:04"), Status: "expired", Error: "等待超过 10 分钟"})
		return
	}
	if client == nil || time.Since(last) < pace {
		return
	}
	p.mu.Lock()
	p.lastSend = time.Now()
	p.mu.Unlock()

	sendCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	shortCode, sendErr := client.Send(sendCtx, &pushplus.SendRequest{
		Title: job.Title, Content: job.Content, Topic: topic,
		Template: pushplus.Template(job.Template), Channel: pushplus.ChannelWechat,
		// Validity of this request, not of the original message: a queue-relative
		// deadline expires exactly as the message becomes due and is rejected as 999.
		Timestamp: time.Now().Add(2 * time.Minute).UnixMilli(),
	})
	cancel()
	if sendErr == nil {
		_ = p.store.deleteQueued(ctx, job.ID)
		p.addRecent(delivery{Key: job.DedupeKey, ShortCode: shortCode, Title: job.Title, Content: job.Content, Time: time.Now().Format("01-02 15:04"), Status: "accepted"})
		if shortCode != "" {
			go p.track(shortCode, job.Title, job.DedupeKey)
		}
		return
	}

	if retryableSendError(sendErr) && job.Attempts < 2 {
		delay := time.Duration(1<<job.Attempts) * 30 * time.Second
		_ = p.store.retryQueued(ctx, job.ID, job.Attempts+1, time.Now().Add(delay))
		return
	}
	_ = p.store.deleteQueued(ctx, job.ID)
	p.addRecent(delivery{Key: job.DedupeKey, Title: job.Title, Time: time.Now().Format("01-02 15:04"), Status: "failed", Error: sendErr.Error()})
}

// retryableSendError reports whether a failed send is worth another attempt.
// Only transport failures and genuine server errors are: every other business
// code (quota, token, not verified, validation…) is permanent, and retrying one
// burns a paced send slot that the queued messages behind it are waiting for.
func retryableSendError(err error) bool {
	if sdkErr, ok := pushplus.AsError(err); ok {
		return sdkErr.Cause != nil || sdkErr.Code == int(pushplus.ErrorCodeServerError)
	}
	return true
}

func (p *pushService) track(shortCode, title, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		p.mu.RLock()
		client, secret := p.client, p.secret
		p.mu.RUnlock()
		if client == nil || secret == "" {
			return
		}
		result, err := client.OpenMessage().QueryResult(ctx, shortCode)
		if err != nil {
			continue
		}
		if result.Status == int(pushplus.SendStatusSuccess) {
			p.addRecent(delivery{Key: key, ShortCode: shortCode, Title: title, Time: time.Now().Format("01-02 15:04"), Status: "sent"})
			return
		}
		if result.Status == int(pushplus.SendStatusFailed) {
			p.addRecent(delivery{Key: key, ShortCode: shortCode, Title: title, Time: time.Now().Format("01-02 15:04"), Status: "failed", Error: result.ErrorMessage})
			return
		}
	}
}

func (p *pushService) test(ctx context.Context) (string, error) {
	p.mu.RLock()
	client, topic := p.client, p.topic
	p.mu.RUnlock()
	if client == nil {
		return "", errors.New("请先配置 PushPlus Token")
	}
	return client.Send(ctx, &pushplus.SendRequest{
		Title:   "Telegram 转发器连接测试",
		Content: "PushPlus 配置正常。",
		Topic:   topic, Template: pushplus.TemplateTxt, Channel: pushplus.ChannelWechat,
	})
}

func (p *pushService) addRecent(item delivery) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if item.ShortCode != "" || item.Key != "" {
		for i := range p.recent {
			if (item.ShortCode != "" && p.recent[i].ShortCode == item.ShortCode) || (item.Key != "" && p.recent[i].Key == item.Key) {
				if item.Content == "" {
					item.Content = p.recent[i].Content
				}
				if item.CreatedAt.IsZero() {
					item.CreatedAt = p.recent[i].CreatedAt
				}
				p.recent[i] = item
				return
			}
		}
	}
	p.recent = append([]delivery{item}, p.recent...)
	if len(p.recent) > 50 {
		p.recent = p.recent[:50]
	}
}

func (p *pushService) expireRecent(cutoff time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.recent {
		if p.recent[i].Status == "queued" && !p.recent[i].CreatedAt.IsZero() && p.recent[i].CreatedAt.Before(cutoff) {
			p.recent[i].Status = "expired"
			p.recent[i].Error = "等待超过 10 分钟"
		}
	}
}

func (p *pushService) deliveries(ctx context.Context) []delivery {
	p.mu.RLock()
	client, secret := p.client, p.secret
	recent := append([]delivery{}, p.recent...)
	history := append([]delivery{}, p.history...)
	stale := time.Since(p.historyAt) >= 30*time.Second
	loading := p.historyLoading
	p.mu.RUnlock()
	if stale && !loading && client != nil && secret != "" {
		p.mu.Lock()
		if !p.historyLoading {
			p.historyLoading = true
			go p.refreshHistory(client)
		}
		p.mu.Unlock()
	}
	combined := append(recent, history...)
	result := make([]delivery, 0, len(combined))
	seen := make(map[string]struct{}, len(combined))
	for _, item := range combined {
		if item.ShortCode != "" {
			if _, ok := seen[item.ShortCode]; ok {
				continue
			}
			seen[item.ShortCode] = struct{}{}
		}
		result = append(result, item)
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func (p *pushService) refreshHistory(client *pushplus.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	page, err := client.OpenMessage().List(ctx, pushplus.NewPageQuery(1, 20))
	history := []delivery{}
	if err == nil {
		for _, item := range page.List {
			history = append(history, delivery{ShortCode: item.ShortCode, Title: item.Title, Time: item.UpdateTime, Status: "history"})
		}
	}
	p.mu.Lock()
	if p.client == client && err == nil {
		p.history, p.historyAt = history, time.Now()
	}
	p.historyLoading = false
	p.mu.Unlock()
}

func (p *pushService) status() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.client == nil {
		return "unconfigured"
	}
	if until := p.client.RateLimitGuard().BlockedUntil(p.token); !until.IsZero() {
		return fmt.Sprintf("limited:%s", until.Format(time.RFC3339))
	}
	return "ready"
}
