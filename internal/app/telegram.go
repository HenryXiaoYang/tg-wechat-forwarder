package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/query"
	querydialogs "github.com/gotd/td/telegram/query/dialogs"
	querymessages "github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"rsc.io/qr"
)

type telegramState struct {
	Status    string `json:"status"`
	Account   string `json:"account,omitempty"`
	QRCode    string `json:"qrCode,omitempty"`
	QRExpires int64  `json:"qrExpires,omitempty"`
	Error     string `json:"error,omitempty"`
}

type preparedMessage struct {
	peerKey  string
	dedupe   string
	title    string
	text     string
	imageURL string
	created  time.Time
	groupID  int64
}

type pendingAlbum struct {
	messages []preparedMessage
	timer    *time.Timer
}

type telegramJob struct {
	entities tg.Entities
	message  *tg.Message
}

type telegramService struct {
	cfg       config
	store     *store
	push      *pushService
	chevereto *cheveretoClient
	ctx       context.Context

	mu        sync.RWMutex
	state     telegramState
	client    *telegram.Client
	api       *tg.Client
	runCancel context.CancelFunc
	liveSince time.Time
	passwords chan string
	rules     []*regexp.Regexp
	jobs      chan telegramJob

	albumMu sync.Mutex
	albums  map[string]*pendingAlbum
}

func newTelegramService(cfg config, store *store, push *pushService, chevereto *cheveretoClient) *telegramService {
	return &telegramService{
		cfg: cfg, store: store, push: push, chevereto: chevereto,
		state:     telegramState{Status: "starting"},
		passwords: make(chan string, 1), albums: make(map[string]*pendingAlbum),
		jobs: make(chan telegramJob, 256),
	}
}

func (t *telegramService) start(ctx context.Context) {
	t.ctx = ctx
	if err := t.reloadFilters(ctx); err != nil {
		slog.Error("load ad filters", "error", err)
	}
	for range 4 {
		go t.processLoop(ctx)
	}
	go t.loop(ctx)
}

func (t *telegramService) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-t.jobs:
			t.processMessage(ctx, job.entities, job.message)
		}
	}
}

func (t *telegramService) loop(ctx context.Context) {
	for ctx.Err() == nil {
		runCtx, cancel := context.WithCancel(ctx)
		t.mu.Lock()
		t.runCancel = cancel
		t.mu.Unlock()
		err := t.run(runCtx)
		cancel()
		t.mu.Lock()
		t.client, t.api, t.runCancel = nil, nil, nil
		if ctx.Err() == nil && !errors.Is(err, context.Canceled) {
			t.state = telegramState{Status: "error", Error: cleanError(err)}
		}
		t.mu.Unlock()
		if ctx.Err() != nil {
			return
		}
		delay := 5 * time.Second
		if errors.Is(err, context.Canceled) {
			delay = time.Second
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func (t *telegramService) run(ctx context.Context) error {
	dispatcher := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(&dispatcher)
	dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		if message, ok := update.Message.(*tg.Message); ok {
			t.handleMessage(entities, message)
		}
		return nil
	})
	dispatcher.OnNewChannelMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
		if message, ok := update.Message.(*tg.Message); ok {
			t.handleMessage(entities, message)
		}
		return nil
	})

	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, updates tg.UpdatesClass) error {
		if short, ok := updates.(*tg.UpdateShortChatMessage); ok {
			t.handleMessage(tg.Entities{}, &tg.Message{
				Out: short.Out, ID: short.ID, FromID: &tg.PeerUser{UserID: short.FromID},
				PeerID: &tg.PeerChat{ChatID: short.ChatID}, Message: short.Message,
				Date: short.Date, Entities: short.Entities,
			})
			return nil
		}
		// UpdatesTooLong and private short updates are deliberately ignored: this app
		// forwards only live group/channel traffic and never replays a gap.
		return dispatcher.Handle(ctx, updates)
	})

	client := telegram.NewClient(t.cfg.TelegramAppID, t.cfg.TelegramAppHash, telegram.Options{
		SessionStorage: t.store,
		UpdateHandler:  handler,
		AllowCDN:       true,
		Device: telegram.DeviceConfig{
			DeviceModel: "tg-wechat-forwarder", SystemVersion: runtime.GOOS,
			AppVersion: version, SystemLangCode: "zh-Hans", LangCode: "zh-hans",
		},
		OnConnectionState: func(state telegram.ConnectionState) {
			t.mu.Lock()
			if state == telegram.ConnectionStateReady {
				t.liveSince = time.Now()
				if t.state.Account != "" {
					t.state.Status = "connected"
				}
			}
			if t.state.Status == "connected" && state != telegram.ConnectionStateReady {
				t.state.Status = state.String()
			}
			t.mu.Unlock()
		},
	})
	t.mu.Lock()
	t.client, t.api = client, client.API()
	t.state = telegramState{Status: "connecting"}
	t.mu.Unlock()
	connectWatchdog := time.AfterFunc(15*time.Second, t.markConnectTimeout)
	defer connectWatchdog.Stop()

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("Telegram auth status: %w", err)
		}
		if !status.Authorized {
			t.setState(telegramState{Status: "qr"})
			show := func(_ context.Context, token qrlogin.Token) error {
				// Not token.Image(): rsc.io/qr's image.Image adapter reports a scaled
				// canvas but draws one pixel per module in its corner. Its own PNG
				// encoder honours Scale and the quiet zone.
				code, err := qr.Encode(token.URL(), qr.M)
				if err != nil {
					return err
				}
				t.setState(telegramState{
					Status: "qr", QRCode: "data:image/png;base64," + base64.StdEncoding.EncodeToString(code.PNG()),
					QRExpires: token.Expires().Unix(),
				})
				return nil
			}
			if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
				if !tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
					return fmt.Errorf("Telegram QR login: %w", err)
				}
				if err := t.waitPassword(ctx, client.Auth()); err != nil {
					return err
				}
			}
		}

		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("load Telegram account: %w", err)
		}
		name := strings.TrimSpace(self.FirstName + " " + self.LastName)
		if self.Username != "" {
			name += " (@" + self.Username + ")"
		}
		t.mu.Lock()
		t.liveSince = time.Now()
		t.state = telegramState{Status: "connected", Account: name}
		t.mu.Unlock()
		if err := t.refreshDialogs(ctx); err != nil {
			slog.Warn("refresh Telegram dialogs", "error", err)
		}

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if err := t.refreshDialogs(ctx); err != nil {
					slog.Warn("refresh Telegram dialogs", "error", err)
				}
			}
		}
	})
}

func (t *telegramService) waitPassword(ctx context.Context, client *auth.Client) error {
	t.setState(telegramState{Status: "password"})
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case password := <-t.passwords:
			if _, err := client.Password(ctx, password); err != nil {
				if errors.Is(err, auth.ErrPasswordInvalid) {
					t.setState(telegramState{Status: "password", Error: "两步验证密码错误"})
					continue
				}
				return fmt.Errorf("Telegram 2FA: %w", err)
			}
			return nil
		}
	}
}

func (t *telegramService) submitPassword(ctx context.Context, password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("密码不能为空")
	}
	if t.snapshot().Status != "password" {
		return errors.New("当前不需要两步验证密码")
	}
	select {
	case t.passwords <- password:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *telegramService) logout(ctx context.Context) error {
	t.mu.RLock()
	api, cancel := t.api, t.runCancel
	t.mu.RUnlock()
	if api != nil {
		logoutCtx, stop := context.WithTimeout(ctx, 10*time.Second)
		_, _ = api.AuthLogOut(logoutCtx)
		stop()
	}
	if err := t.store.clearTelegramSession(ctx); err != nil {
		return err
	}
	if cancel != nil {
		cancel()
	}
	t.setState(telegramState{Status: "connecting"})
	return nil
}

func (t *telegramService) snapshot() telegramState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

func (t *telegramService) setState(state telegramState) {
	t.mu.Lock()
	t.state = state
	t.mu.Unlock()
}

func (t *telegramService) markConnectTimeout() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state.Status == "connecting" {
		t.state = telegramState{
			Status: "error",
			Error:  "连接 Telegram 超时，请检查 TELEGRAM_API_ID、TELEGRAM_API_HASH 和服务器网络；后台仍会自动重试",
		}
	}
}

func (t *telegramService) refresh(ctx context.Context) error {
	t.mu.RLock()
	connected := t.state.Status == "connected"
	t.mu.RUnlock()
	if !connected {
		return errors.New("Telegram 尚未登录")
	}
	return t.refreshDialogs(ctx)
}

func (t *telegramService) refreshDialogs(ctx context.Context) error {
	t.mu.RLock()
	api := t.api
	t.mu.RUnlock()
	if api == nil {
		return errors.New("Telegram 未连接")
	}
	byKey := make(map[string]dialog)
	collect := func(archived bool) error {
		builder := query.GetDialogs(api).BatchSize(100)
		if archived {
			builder = builder.FolderID(1)
		}
		return builder.ForEach(ctx, func(_ context.Context, elem querydialogs.Elem) error {
			d, ok := dialogFromTelegram(elem, archived)
			if ok {
				byKey[d.PeerKey] = d
			}
			return nil
		})
	}
	if err := collect(false); err != nil {
		return err
	}
	if err := collect(true); err != nil {
		return err
	}
	dialogs := make([]dialog, 0, len(byKey))
	for _, d := range byKey {
		dialogs = append(dialogs, d)
	}
	sort.Slice(dialogs, func(i, j int) bool {
		if dialogs[i].Archived != dialogs[j].Archived {
			return !dialogs[i].Archived
		}
		return strings.ToLower(dialogs[i].Title) < strings.ToLower(dialogs[j].Title)
	})
	return t.store.replaceDialogs(ctx, dialogs)
}

func dialogFromTelegram(elem querydialogs.Elem, archived bool) (dialog, bool) {
	if elem.Deleted() {
		return dialog{}, false
	}
	switch peer := elem.Peer.(type) {
	case *tg.InputPeerUser:
		user, ok := elem.Entities.User(peer.UserID)
		if !ok {
			return dialog{}, false
		}
		title := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if title == "" {
			title = user.Username
		}
		subtitle := "私聊"
		if user.Self {
			subtitle = "收藏夹"
		} else if user.Bot {
			subtitle = "机器人"
		}
		if user.Username != "" && !user.Self {
			subtitle += " · @" + user.Username
		}
		return dialog{PeerKey: peerKey("user", peer.UserID), Kind: "user", Title: title, Subtitle: subtitle, Archived: archived}, true
	case *tg.InputPeerChat:
		chat, ok := elem.Entities.Chat(peer.ChatID)
		if !ok {
			return dialog{}, false
		}
		subtitle, selectable := "群组", !chat.Noforwards
		if chat.Noforwards {
			subtitle += " · 内容受保护"
		}
		return dialog{PeerKey: peerKey("chat", peer.ChatID), Kind: "group", Title: chat.Title, Subtitle: subtitle, Selectable: selectable, Archived: archived}, true
	case *tg.InputPeerChannel:
		channel, ok := elem.Entities.Channel(peer.ChannelID)
		if !ok {
			return dialog{}, false
		}
		kind, subtitle := "group", "超级群组"
		if channel.Broadcast {
			kind, subtitle = "channel", "频道"
		}
		if channel.Forum {
			subtitle = "论坛"
		}
		if channel.Username != "" {
			subtitle += " · @" + channel.Username
		}
		selectable := !channel.Noforwards
		if channel.Noforwards {
			subtitle += " · 内容受保护"
		}
		return dialog{PeerKey: peerKey("channel", peer.ChannelID), Kind: kind, Title: channel.Title, Subtitle: subtitle, Selectable: selectable, Archived: archived}, true
	default:
		return dialog{}, false
	}
}

func (t *telegramService) handleMessage(entities tg.Entities, message *tg.Message) {
	if message == nil || message.Out || message.Noforwards {
		return
	}
	t.mu.RLock()
	liveSince := t.liveSince
	t.mu.RUnlock()
	if int64(message.Date) < liveSince.Unix() {
		return
	}
	select {
	case t.jobs <- telegramJob{entities: entities, message: message}:
	default:
		// ponytail: bounded live queue; persist raw MTProto updates if burst loss becomes measurable.
		slog.Warn("Telegram processing queue is full; dropping live message")
	}
}

func (t *telegramService) processMessage(ctx context.Context, entities tg.Entities, message *tg.Message) {
	key := peerKeyFromTG(message.PeerID)
	if key == "" {
		return
	}
	d, ok, err := t.store.dialog(ctx, key)
	if err != nil || !ok || !d.Selected || message.Date < int(d.SelectedAt) {
		return
	}
	sender := senderName(entities, message)
	text, image, skip := messageContent(message)
	if skip {
		return
	}
	if d.AdFilter && t.isAd(sender+"\n"+text) {
		return
	}

	prepared := preparedMessage{
		peerKey: key, dedupe: fmt.Sprintf("%s:%d", key, message.ID),
		title: truncate(d.Title+" · "+sender, 100), text: text,
		created: time.Unix(int64(message.Date), 0),
	}
	if groupID, ok := message.GetGroupedID(); ok {
		prepared.groupID = groupID
	}
	if image {
		file, ok := (querymessages.Elem{Msg: message}).File()
		if ok {
			t.mu.RLock()
			client, api := t.client, t.api
			t.mu.RUnlock()
			if client != nil && api != nil {
				var output limitedBuffer
				output.max = 20 << 20
				downloadCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
				_, downloadErr := client.Downloader().Download(api, file.Location).Stream(downloadCtx, &output)
				cancel()
				if downloadErr == nil {
					uploadCtx, stop := context.WithTimeout(ctx, time.Minute)
					prepared.imageURL, err = t.chevereto.upload(uploadCtx, file.Name, file.MIMEType, output.Bytes(), "")
					stop()
				}
				if downloadErr != nil || err != nil {
					prepared.text = strings.TrimSpace(prepared.text + "\n[图片上传失败]")
				}
			}
		}
	}
	if prepared.text == "" && prepared.imageURL == "" {
		return
	}
	if prepared.groupID != 0 {
		t.addAlbum(prepared)
		return
	}
	t.enqueuePrepared(ctx, []preparedMessage{prepared})
}

func (t *telegramService) addAlbum(message preparedMessage) {
	key := fmt.Sprintf("%s:%d", message.peerKey, message.groupID)
	t.albumMu.Lock()
	album := t.albums[key]
	if album == nil {
		album = &pendingAlbum{}
		t.albums[key] = album
		album.timer = time.AfterFunc(1200*time.Millisecond, func() { t.flushAlbum(key) })
	} else {
		album.timer.Reset(1200 * time.Millisecond)
	}
	album.messages = append(album.messages, message)
	t.albumMu.Unlock()
}

func (t *telegramService) flushAlbum(key string) {
	t.albumMu.Lock()
	album := t.albums[key]
	delete(t.albums, key)
	t.albumMu.Unlock()
	if album != nil {
		t.enqueuePrepared(t.ctx, album.messages)
	}
}

func (t *telegramService) enqueuePrepared(ctx context.Context, messages []preparedMessage) {
	if len(messages) == 0 {
		return
	}
	sort.SliceStable(messages, func(i, j int) bool {
		if messages[i].created.Equal(messages[j].created) {
			return messages[i].dedupe < messages[j].dedupe
		}
		return messages[i].created.Before(messages[j].created)
	})
	first := messages[0]
	var texts, images []string
	for _, message := range messages {
		if strings.TrimSpace(message.text) != "" {
			texts = append(texts, strings.TrimSpace(message.text))
		}
		if message.imageURL != "" {
			images = append(images, message.imageURL)
		}
	}
	text := strings.Join(unique(texts), "\n\n")
	template, content := "txt", first.created.Format("2006-01-02 15:04")
	if text != "" {
		content += "\n" + text
	}
	if len(images) > 0 {
		template = "html"
		content = html.EscapeString(first.created.Format("2006-01-02 15:04"))
		if text != "" {
			content += "<br>" + strings.ReplaceAll(html.EscapeString(text), "\n", "<br>")
		}
		for _, imageURL := range images {
			content += `<br><img src="` + html.EscapeString(imageURL) + `" alt="Telegram image">`
		}
	}
	dedupe := first.dedupe
	if first.groupID != 0 {
		dedupe = fmt.Sprintf("album:%s:%d", first.peerKey, first.groupID)
	}
	if err := t.push.enqueue(ctx, queuedMessage{DedupeKey: dedupe, Title: first.title, Content: truncate(content, 19000), Template: template, CreatedAt: first.created}); err != nil {
		slog.Error("enqueue Telegram message", "error", err)
	}
}

func messageContent(message *tg.Message) (text string, image, skip bool) {
	text = strings.TrimSpace(message.Message)
	switch media := message.Media.(type) {
	case nil, *tg.MessageMediaEmpty, *tg.MessageMediaWebPage:
		return text, false, false
	case *tg.MessageMediaPhoto:
		return text, true, false
	case *tg.MessageMediaDocument:
		document, ok := media.Document.AsNotEmpty()
		if !ok {
			return "", false, true
		}
		sticker, animated := false, false
		for _, attribute := range document.Attributes {
			switch attribute.(type) {
			case *tg.DocumentAttributeSticker:
				sticker = true
			case *tg.DocumentAttributeAnimated, *tg.DocumentAttributeVideo:
				animated = true
			}
		}
		if sticker && !animated && strings.HasPrefix(document.MimeType, "image/") {
			return text, true, false
		}
		return "", false, true
	case *tg.MessageMediaPoll:
		lines := []string{"投票：" + media.Poll.Question.Text}
		for _, answer := range media.Poll.Answers {
			if value, ok := answer.(*tg.PollAnswer); ok {
				lines = append(lines, "• "+value.Text.Text)
			}
		}
		return strings.TrimSpace(strings.Join(appendNonEmpty([]string{text}, lines...), "\n")), false, false
	case *tg.MessageMediaContact:
		contact := strings.TrimSpace(media.FirstName + " " + media.LastName)
		return strings.TrimSpace(strings.Join(appendNonEmpty([]string{text}, "联系人："+contact+" "+media.PhoneNumber), "\n")), false, false
	case *tg.MessageMediaVenue:
		return strings.TrimSpace(strings.Join(appendNonEmpty([]string{text}, "位置："+media.Title+" "+media.Address, geoText(media.Geo)), "\n")), false, false
	case *tg.MessageMediaGeo:
		return strings.TrimSpace(strings.Join(appendNonEmpty([]string{text}, "位置："+geoText(media.Geo)), "\n")), false, false
	case *tg.MessageMediaGeoLive:
		return strings.TrimSpace(strings.Join(appendNonEmpty([]string{text}, "实时位置："+geoText(media.Geo)), "\n")), false, false
	default:
		return "", false, true
	}
}

func geoText(value tg.GeoPointClass) string {
	if point, ok := value.(*tg.GeoPoint); ok {
		return fmt.Sprintf("https://maps.google.com/?q=%f,%f", point.Lat, point.Long)
	}
	return ""
}

func senderName(entities tg.Entities, message *tg.Message) string {
	if message.PostAuthor != "" {
		return message.PostAuthor
	}
	from, ok := message.GetFromID()
	if !ok {
		return "频道"
	}
	switch value := from.(type) {
	case *tg.PeerUser:
		if user := entities.Users[value.UserID]; user != nil {
			name := strings.TrimSpace(user.FirstName + " " + user.LastName)
			if name != "" {
				return name
			}
			if user.Username != "" {
				return "@" + user.Username
			}
		}
		return fmt.Sprintf("用户 %d", value.UserID)
	case *tg.PeerChannel:
		if channel := entities.Channels[value.ChannelID]; channel != nil {
			return channel.Title
		}
	case *tg.PeerChat:
		if chat := entities.Chats[value.ChatID]; chat != nil {
			return chat.Title
		}
	}
	return "Telegram"
}

func (t *telegramService) saveFilters(ctx context.Context, rules []string) error {
	clean := make([]string, 0, len(rules))
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if _, err := compileRule(rule); err != nil {
			return fmt.Errorf("规则 %q 无效: %w", rule, err)
		}
		clean = append(clean, rule)
	}
	if err := t.store.saveJSON(ctx, "ad_filter_rules", clean); err != nil {
		return err
	}
	return t.reloadFilters(ctx)
}

func (t *telegramService) filterRules(ctx context.Context) ([]string, error) {
	var rules []string
	_, err := t.store.loadJSON(ctx, "ad_filter_rules", &rules)
	return rules, err
}

func (t *telegramService) reloadFilters(ctx context.Context) error {
	rules, err := t.filterRules(ctx)
	if err != nil {
		return err
	}
	compiled := make([]*regexp.Regexp, 0, len(rules))
	for _, rule := range rules {
		item, err := compileRule(rule)
		if err != nil {
			return err
		}
		compiled = append(compiled, item)
	}
	t.mu.Lock()
	t.rules = compiled
	t.mu.Unlock()
	return nil
}

func compileRule(rule string) (*regexp.Regexp, error) {
	if strings.HasPrefix(rule, "re:") {
		return regexp.Compile(strings.TrimSpace(strings.TrimPrefix(rule, "re:")))
	}
	return regexp.Compile("(?i)" + regexp.QuoteMeta(rule))
}

func (t *telegramService) isAd(text string) bool {
	t.mu.RLock()
	rules := append([]*regexp.Regexp{}, t.rules...)
	t.mu.RUnlock()
	for _, rule := range rules {
		if rule.MatchString(text) {
			return true
		}
	}
	return false
}

type limitedBuffer struct {
	bytes.Buffer
	max int
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	if b.Len()+len(value) > b.max {
		return 0, errors.New("Telegram image exceeds 20 MiB")
	}
	return b.Buffer.Write(value)
}

func peerKey(kind string, id int64) string { return fmt.Sprintf("%s:%d", kind, id) }

func peerKeyFromTG(peer tg.PeerClass) string {
	switch value := peer.(type) {
	case *tg.PeerUser:
		return peerKey("user", value.UserID)
	case *tg.PeerChat:
		return peerKey("chat", value.ChatID)
	case *tg.PeerChannel:
		return peerKey("channel", value.ChannelID)
	default:
		return ""
	}
}

func truncate(value string, max int) string {
	if utf8.RuneCountInString(value) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max-1]) + "…"
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := values[:0]
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func appendNonEmpty(values []string, items ...string) []string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			values = append(values, strings.TrimSpace(item))
		}
	}
	return values
}

func cleanError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "连接已关闭"
	}
	return err.Error()
}
