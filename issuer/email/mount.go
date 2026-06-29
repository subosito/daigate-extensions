package email

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/subosito/daigate/ingress/keyring"
)

// MountIssuer registers email OTP routes and optional setup UI.
func MountIssuer(mux *http.ServeMux, db *sql.DB, ks keyring.KeyStore, cfg IssuerYAML) error {
	var ttl time.Duration
	if cfg.TTL != "" {
		d, err := time.ParseDuration(cfg.TTL)
		if err != nil {
			return err
		}
		ttl = d
	}
	smtpPass := ""
	if cfg.SMTPPassEnv != "" {
		smtpPass = os.Getenv(cfg.SMTPPassEnv)
	}
	i := NewFromConfig(db, ks, ttl, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, smtpPass, cfg.From, cfg.AllowedDomains)
	if err := i.Mount(mux); err != nil {
		return err
	}
	if cfg.UI {
		MountUI(mux, UIConfig{
			Title:          cfg.UITitle,
			ClientBaseURL:  cfg.ClientBaseURL,
			AllowedDomains: cfg.AllowedDomains,
			Models:         cfg.UIModels,
		})
	}
	return nil
}