package app

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/session"
	_ "modernc.org/sqlite"
)

const dataFile = "data/data.db"

type store struct {
	db   *sql.DB
	aead cipher.AEAD
}

type dialog struct {
	PeerKey    string `json:"peerKey"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle,omitempty"`   // 会话类型：超级群组、频道、论坛…
	Username   string `json:"username,omitempty"`   // 仅用于前端搜索
	LastSender string `json:"lastSender,omitempty"` // 最后一条消息的发送者
	LastText   string `json:"lastText,omitempty"`   // 最后一条消息的摘要
	Selectable bool   `json:"selectable"`
	Selected   bool   `json:"selected"`
	AdFilter   bool   `json:"adFilter"`
	Archived   bool   `json:"archived"`
	Pinned     bool   `json:"pinned"`
	SelectedAt int64  `json:"-"`
	LastAt     int64  `json:"-"`
}

type queuedMessage struct {
	ID          int64
	DedupeKey   string
	Title       string
	Content     string
	Template    string
	CreatedAt   time.Time
	Attempts    int
	NextAttempt time.Time
}

func openStore(secret string) (*store, error) {
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	// MkdirAll is happy with a directory this process cannot write to, and sqlite
	// then reports only "unable to open database file (14)". Probe first so the
	// error names the uid that needs to own ./data.
	probe, err := os.CreateTemp(dir, ".probe")
	if err != nil {
		return nil, fmt.Errorf("data directory %s is not writable by uid %d, chown it to that uid: %w", dir, os.Getuid(), err)
	}
	probe.Close()
	os.Remove(probe.Name())
	db, err := sql.Open("sqlite", "file:"+dataFile+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	key := sha256.Sum256([]byte("database\x00" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		db.Close()
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		db.Close()
		return nil, err
	}
	s := &store{db: db, aead: aead}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	_ = os.Chmod(dataFile, 0o600)
	return s, nil
}

func (s *store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value BLOB NOT NULL
);
CREATE TABLE IF NOT EXISTS dialogs (
  peer_key TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  subtitle TEXT NOT NULL DEFAULT '',
  selectable INTEGER NOT NULL DEFAULT 0,
  selected INTEGER NOT NULL DEFAULT 0,
  ad_filter INTEGER NOT NULL DEFAULT 0,
  archived INTEGER NOT NULL DEFAULT 0,
  pinned INTEGER NOT NULL DEFAULT 0,
  username TEXT NOT NULL DEFAULT '',
  selected_at INTEGER NOT NULL DEFAULT 0,
  last_at INTEGER NOT NULL DEFAULT 0,
  last_sender BLOB,
  last_text BLOB,
  updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS queue (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  dedupe_key TEXT NOT NULL UNIQUE,
  title BLOB NOT NULL,
  content BLOB NOT NULL,
  template TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  next_attempt INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS queue_due ON queue(next_attempt, id);
`)
	if err != nil {
		return fmt.Errorf("migrate sqlite: %w", err)
	}
	// Columns added after the first release; existing databases keep their user state.
	for _, statement := range []string{
		`ALTER TABLE dialogs ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE dialogs ADD COLUMN last_at INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE dialogs ADD COLUMN username TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE dialogs ADD COLUMN last_sender BLOB`,
		`ALTER TABLE dialogs ADD COLUMN last_text BLOB`,
	} {
		if _, err := s.db.Exec(statement); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("migrate sqlite: %w", err)
		}
	}
	return nil
}

func (s *store) Close() error { return s.db.Close() }

func (s *store) seal(plain []byte) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, plain, nil), nil
}

func (s *store) open(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < s.aead.NonceSize() {
		return nil, errors.New("encrypted value is truncated")
	}
	nonce := ciphertext[:s.aead.NonceSize()]
	return s.aead.Open(nil, nonce, ciphertext[s.aead.NonceSize():], nil)
}

// Message previews are chat content, so they follow the same rule as the send
// queue: encrypted at rest, and unreadable rather than fatal after a key change.
func (s *store) sealText(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	return s.seal([]byte(value))
}

func (s *store) openText(ciphertext []byte) string {
	if len(ciphertext) == 0 {
		return ""
	}
	plain, err := s.open(ciphertext)
	if err != nil {
		return ""
	}
	return string(plain)
}

func (s *store) set(ctx context.Context, key, value string, secret bool) error {
	data := []byte(value)
	var err error
	if secret {
		data, err = s.seal(data)
		if err != nil {
			return err
		}
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, data)
	return err
}

func (s *store) get(ctx context.Context, key string, secret bool) (string, bool, error) {
	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key=?`, key).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if secret {
		data, err = s.open(data)
		if err != nil {
			return "", false, fmt.Errorf("decrypt %s: %w", key, err)
		}
	}
	return string(data), true, nil
}

func (s *store) deleteSetting(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM settings WHERE key=?`, key)
	return err
}

func (s *store) LoadSession(ctx context.Context) ([]byte, error) {
	value, ok, err := s.get(ctx, "telegram_session", true)
	if err != nil {
		return nil, err
	}
	if !ok || value == "" {
		return nil, session.ErrNotFound
	}
	return []byte(value), nil
}

func (s *store) StoreSession(ctx context.Context, data []byte) error {
	return s.set(ctx, "telegram_session", string(data), true)
}

func (s *store) clearTelegramSession(ctx context.Context) error {
	return s.deleteSetting(ctx, "telegram_session")
}

func (s *store) replaceDialogs(ctx context.Context, dialogs []dialog) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UnixNano()
	for _, d := range dialogs {
		sender, err := s.sealText(d.LastSender)
		if err != nil {
			return err
		}
		text, err := s.sealText(d.LastText)
		if err != nil {
			return err
		}
		// A live update may already hold a newer preview than this refresh.
		_, err = tx.ExecContext(ctx, `
INSERT INTO dialogs(peer_key,kind,title,subtitle,username,selectable,archived,pinned,last_at,last_sender,last_text,updated_at)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(peer_key) DO UPDATE SET
  kind=excluded.kind,
  title=excluded.title,
  subtitle=excluded.subtitle,
  username=excluded.username,
  selectable=excluded.selectable,
  selected=CASE WHEN excluded.selectable=1 THEN dialogs.selected ELSE 0 END,
  archived=excluded.archived,
  pinned=excluded.pinned,
  last_sender=CASE WHEN excluded.last_at>=dialogs.last_at THEN excluded.last_sender ELSE dialogs.last_sender END,
  last_text=CASE WHEN excluded.last_at>=dialogs.last_at THEN excluded.last_text ELSE dialogs.last_text END,
  last_at=MAX(excluded.last_at,dialogs.last_at),
  updated_at=excluded.updated_at`,
			d.PeerKey, d.Kind, d.Title, d.Subtitle, d.Username, d.Selectable, d.Archived, d.Pinned, d.LastAt, sender, text, now)
		if err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM dialogs WHERE updated_at < ?`, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *store) dialogs(ctx context.Context) ([]dialog, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT peer_key,kind,title,subtitle,username,selectable,selected,ad_filter,archived,pinned,selected_at,last_at,last_sender,last_text FROM dialogs ORDER BY archived,pinned DESC,last_at DESC,title COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []dialog
	for rows.Next() {
		var d dialog
		var sender, text []byte
		if err := rows.Scan(&d.PeerKey, &d.Kind, &d.Title, &d.Subtitle, &d.Username, &d.Selectable, &d.Selected, &d.AdFilter, &d.Archived, &d.Pinned, &d.SelectedAt, &d.LastAt, &sender, &text); err != nil {
			return nil, err
		}
		d.LastSender, d.LastText = s.openText(sender), s.openText(text)
		result = append(result, d)
	}
	return result, rows.Err()
}

func (s *store) dialog(ctx context.Context, peerKey string) (dialog, bool, error) {
	var d dialog
	err := s.db.QueryRowContext(ctx, `SELECT peer_key,kind,title,subtitle,username,selectable,selected,ad_filter,archived,pinned,selected_at,last_at FROM dialogs WHERE peer_key=?`, peerKey).
		Scan(&d.PeerKey, &d.Kind, &d.Title, &d.Subtitle, &d.Username, &d.Selectable, &d.Selected, &d.AdFilter, &d.Archived, &d.Pinned, &d.SelectedAt, &d.LastAt)
	if errors.Is(err, sql.ErrNoRows) {
		return dialog{}, false, nil
	}
	return d, err == nil, err
}

func (s *store) updateDialog(ctx context.Context, peerKey string, selected, adFilter *bool) error {
	d, ok, err := s.dialog(ctx, peerKey)
	if err != nil {
		return err
	}
	if !ok {
		return sql.ErrNoRows
	}
	if selected != nil {
		if !d.Selectable && *selected {
			return errors.New("this dialog cannot be forwarded")
		}
		d.Selected = *selected
		if *selected {
			d.SelectedAt = time.Now().Unix()
		} else {
			d.SelectedAt = 0
		}
	}
	if adFilter != nil {
		d.AdFilter = *adFilter
	}
	_, err = s.db.ExecContext(ctx, `UPDATE dialogs SET selected=?,ad_filter=?,selected_at=? WHERE peer_key=?`, d.Selected, d.AdFilter, d.SelectedAt, peerKey)
	return err
}

// updateDialogPreview keeps the chat list in sync with live traffic; the guard
// keeps an out-of-order or replayed update from overwriting a newer preview.
func (s *store) updateDialogPreview(ctx context.Context, peerKey, sender, text string, at int64) error {
	sealedSender, err := s.sealText(sender)
	if err != nil {
		return err
	}
	sealedText, err := s.sealText(text)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE dialogs SET last_sender=?,last_text=?,last_at=? WHERE peer_key=? AND last_at<=?`,
		sealedSender, sealedText, at, peerKey, at)
	return err
}

func (s *store) enqueue(ctx context.Context, message queuedMessage) error {
	title, err := s.seal([]byte(message.Title))
	if err != nil {
		return err
	}
	content, err := s.seal([]byte(message.Content))
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT OR IGNORE INTO queue(dedupe_key,title,content,template,created_at,next_attempt) VALUES(?,?,?,?,?,?)`,
		message.DedupeKey, title, content, message.Template, message.CreatedAt.Unix(), message.CreatedAt.Unix())
	return err
}

func (s *store) nextQueued(ctx context.Context, now time.Time) (queuedMessage, bool, error) {
	var q queuedMessage
	var title, content []byte
	var createdAt, nextAttempt int64
	err := s.db.QueryRowContext(ctx, `SELECT id,dedupe_key,title,content,template,created_at,attempts,next_attempt FROM queue WHERE next_attempt<=? ORDER BY id LIMIT 1`, now.Unix()).
		Scan(&q.ID, &q.DedupeKey, &title, &content, &q.Template, &createdAt, &q.Attempts, &nextAttempt)
	if errors.Is(err, sql.ErrNoRows) {
		return queuedMessage{}, false, nil
	}
	if err != nil {
		return queuedMessage{}, false, err
	}
	plainTitle, err := s.open(title)
	if err != nil {
		return queuedMessage{}, false, err
	}
	plainContent, err := s.open(content)
	if err != nil {
		return queuedMessage{}, false, err
	}
	q.Title = string(plainTitle)
	q.Content = string(plainContent)
	q.CreatedAt = time.Unix(createdAt, 0)
	q.NextAttempt = time.Unix(nextAttempt, 0)
	return q, true, nil
}

func (s *store) retryQueued(ctx context.Context, id int64, attempts int, next time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE queue SET attempts=?,next_attempt=? WHERE id=?`, attempts, next.Unix(), id)
	return err
}

func (s *store) deleteQueued(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM queue WHERE id=?`, id)
	return err
}

func (s *store) deleteExpiredQueued(ctx context.Context, cutoff time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM queue WHERE created_at<?`, cutoff.Unix())
	return err
}

func (s *store) saveJSON(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.set(ctx, key, string(data), false)
}

func (s *store) loadJSON(ctx context.Context, key string, value any) (bool, error) {
	data, ok, err := s.get(ctx, key, false)
	if err != nil || !ok {
		return ok, err
	}
	if strings.TrimSpace(data) == "" {
		return true, nil
	}
	return true, json.Unmarshal([]byte(data), value)
}
