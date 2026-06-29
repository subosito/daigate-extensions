package email_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	emailissuer "github.com/subosito/daigate-extensions/issuer/email"
)

func TestDecodeIssuerYAML(t *testing.T) {
	var root struct {
		Config yaml.Node `yaml:"config"`
	}
	if err := yaml.Unmarshal([]byte(`
config:
  ttl: 5m
  smtp_host: smtp.example.com
  ui: true
  client_base_url: https://aigate.example.com/v1
`), &root); err != nil {
		t.Fatal(err)
	}
	cfg, err := emailissuer.DecodeIssuerYAML(root.Config)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TTL != "5m" || cfg.SMTPHost != "smtp.example.com" || !cfg.UI {
		t.Fatalf("cfg=%+v", cfg)
	}
	if cfg.ClientBaseURL != "https://aigate.example.com/v1" {
		t.Fatalf("client_base_url=%q", cfg.ClientBaseURL)
	}
}