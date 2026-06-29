package email

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/subosito/daigate/ingress/argonhash"
	"github.com/subosito/daigate/ingress/keyring"
)

// Config is email OTP issuer settings.
type Config struct {
	TTL             time.Duration
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPass        string
	From            string
	AllowedDomains  []string
}

// Issuer mints issued gateway keys via email OTP.
type Issuer struct {
	Store  keyring.KeyStore
	DB     *sql.DB
	Config Config
}

func (i *Issuer) ensureSchema() error {
	_, err := i.DB.Exec(`
		CREATE TABLE IF NOT EXISTS issuer_email_otp (
			email TEXT PRIMARY KEY,
			code_hash TEXT NOT NULL,
			expires_at INTEGER NOT NULL
		)`)
	return err
}

// Mount registers issuer routes on the admin mux.
func (i *Issuer) Mount(mux *http.ServeMux) error {
	if err := i.ensureSchema(); err != nil {
		return err
	}
	mux.HandleFunc("/v1/issuer/email/request", i.request)
	mux.HandleFunc("/v1/issuer/email/verify", i.verify)
	return nil
}

func (i *Issuer) request(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "email required", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		http.Error(w, "email required", http.StatusBadRequest)
		return
	}
	if !emailDomainAllowed(email, i.Config.AllowedDomains) {
		http.Error(w, "email domain not allowed", http.StatusForbidden)
		return
	}
	code, err := randomCode(6)
	if err != nil {
		http.Error(w, "otp error", http.StatusInternalServerError)
		return
	}
	ttl := i.Config.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	exp := time.Now().Add(ttl).UnixMilli()
	sealed, err := sealCode(code)
	if err != nil {
		http.Error(w, "otp error", http.StatusInternalServerError)
		return
	}
	_, err = i.DB.ExecContext(r.Context(), `
		INSERT INTO issuer_email_otp (email, code_hash, expires_at) VALUES (?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET code_hash=excluded.code_hash, expires_at=excluded.expires_at`,
		email, sealed, exp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := i.sendMail(email, code); err != nil {
		slog.Warn("email issuer: smtp failed, dev fallback", "email", email, "err", err)
		slog.Info("email issuer otp (dev)", "email", email, "code", code)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "expires_in_sec": int(ttl.Seconds())})
}

func (i *Issuer) verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !emailDomainAllowed(email, i.Config.AllowedDomains) {
		http.Error(w, "email domain not allowed", http.StatusForbidden)
		return
	}
	var stored string
	var exp int64
	err := i.DB.QueryRowContext(r.Context(), `
		SELECT code_hash, expires_at FROM issuer_email_otp WHERE email = ?`, email).Scan(&stored, &exp)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid or expired code", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if time.Now().UnixMilli() > exp || !argonhash.Verify(strings.TrimSpace(req.Code), stored) {
		http.Error(w, "invalid or expired code", http.StatusUnauthorized)
		return
	}
	if err := revokeIssuedKeysForEmail(r.Context(), i.Store, email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	keyTTL := 720 * time.Hour
	secret, id, err := i.Store.Create(r.Context(), email, keyring.KindIssued, keyTTL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = i.DB.ExecContext(r.Context(), `DELETE FROM issuer_email_otp WHERE email = ?`, email)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "key": secret, "kind": keyring.KindIssued})
}

func (i *Issuer) sendMail(to, code string) error {
	if i.Config.SMTPHost == "" {
		return fmt.Errorf("smtp not configured")
	}
	port := i.Config.SMTPPort
	if port == "" {
		port = "587"
	}
	from := i.Config.From
	if from == "" {
		from = "daigate@localhost"
	}
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: daigate login code\r\n\r\nYour code: %s\r\n", from, to, code))
	host := i.Config.SMTPHost
	addr := net.JoinHostPort(host, port)
	auth := smtp.PlainAuth("", i.Config.SMTPUser, i.Config.SMTPPass, host)
	if port == "465" {
		return sendMailSMTPS(addr, host, auth, from, []string{to}, msg)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// sendMailSMTPS submits mail over implicit TLS (SMTPS), e.g. Cloudflare port 465.
func sendMailSMTPS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("smtp: server does not support AUTH")
		}
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func revokeIssuedKeysForEmail(ctx context.Context, store keyring.KeyStore, email string) error {
	keys, err := store.List(ctx)
	if err != nil {
		return err
	}
	for _, k := range keys {
		if k.Name != email || k.Kind != keyring.KindIssued || k.Revoked {
			continue
		}
		if err := store.Revoke(ctx, k.ID); err != nil {
			return err
		}
	}
	return nil
}

func emailDomainAllowed(addr string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	addr = strings.ToLower(strings.TrimSpace(addr))
	at := strings.LastIndex(addr, "@")
	if at < 1 || at == len(addr)-1 {
		return false
	}
	domain := addr[at+1:]
	for _, entry := range allowed {
		if domain == strings.ToLower(strings.TrimSpace(entry)) {
			return true
		}
	}
	return false
}

func randomCode(n int) (string, error) {
	var sb strings.Builder
	for range n {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteByte(byte('0' + v.Int64()))
	}
	return sb.String(), nil
}

func sealCode(code string) (string, error) {
	return argonhash.Seal(code)
}

// SeedOTP stores an OTP for tests and local tooling.
func SeedOTP(db *sql.DB, addr, code string, exp time.Time) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS issuer_email_otp (
			email TEXT PRIMARY KEY,
			code_hash TEXT NOT NULL,
			expires_at INTEGER NOT NULL
		)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO issuer_email_otp (email, code_hash, expires_at) VALUES (?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET code_hash=excluded.code_hash, expires_at=excluded.expires_at`,
		addr, mustSealCode(code), exp.UnixMilli())
	return err
}

func mustSealCode(code string) string {
	sealed, err := sealCode(code)
	if err != nil {
		panic(err)
	}
	return sealed
}

// NewFromConfig builds an Issuer when driver=email is enabled.
func NewFromConfig(db *sql.DB, ks keyring.KeyStore, ttl time.Duration, smtpHost, smtpPort, smtpUser, smtpPass, from string, allowedDomains []string) *Issuer {
	return &Issuer{
		Store: ks,
		DB:    db,
		Config: Config{
			TTL: ttl, SMTPHost: smtpHost, SMTPPort: smtpPort,
			SMTPUser: smtpUser, SMTPPass: smtpPass, From: from,
			AllowedDomains: allowedDomains,
		},
	}
}