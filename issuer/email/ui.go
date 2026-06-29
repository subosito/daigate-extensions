package email

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed static/index.html
var uiHTML []byte

// UIConfig drives the optional email OTP setup page.
type UIConfig struct {
	Title          string
	ClientBaseURL  string
	AllowedDomains []string
	Models         []string
}

// MountUI registers GET / and GET /setup.json on the admin listener.
func MountUI(mux *http.ServeMux, cfg UIConfig) {
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		serveUIIndex(w)
	})
	mux.HandleFunc("GET /setup.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(buildUISetup(cfg))
	})
}

type uiSetup struct {
	Title          string   `json:"title"`
	DataBaseURL    string   `json:"data_base_url"`
	AllowedDomains []string `json:"allowed_domains"`
	Models         []string `json:"models"`
	RequestPath    string   `json:"request_path"`
	VerifyPath     string   `json:"verify_path"`
}

func buildUISetup(cfg UIConfig) uiSetup {
	title := strings.TrimSpace(cfg.Title)
	if title == "" {
		title = "daigate"
	}
	return uiSetup{
		Title:          title,
		DataBaseURL:    resolveDataBaseURL(cfg),
		AllowedDomains: cfg.AllowedDomains,
		Models:         cfg.Models,
		RequestPath:    "/v1/issuer/email/request",
		VerifyPath:     "/v1/issuer/email/verify",
	}
}

func resolveDataBaseURL(cfg UIConfig) string {
	configured := strings.TrimRight(strings.TrimSpace(cfg.ClientBaseURL), "/")
	if configured != "" {
		return configured
	}
	return "http://127.0.0.1:9420/v1"
}

func serveUIIndex(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(uiHTML)
}