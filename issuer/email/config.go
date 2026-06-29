package email

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// IssuerYAML is the email driver block under ingress.issuers[].config.
type IssuerYAML struct {
	TTL            string   `yaml:"ttl,omitempty"`
	SMTPHost       string   `yaml:"smtp_host,omitempty"`
	SMTPPort       string   `yaml:"smtp_port,omitempty"`
	SMTPUser       string   `yaml:"smtp_user,omitempty"`
	SMTPPassEnv    string   `yaml:"smtp_pass_env,omitempty"`
	From           string   `yaml:"from,omitempty"`
	AllowedDomains []string `yaml:"allowed_domains,omitempty"`
	UI             bool     `yaml:"ui,omitempty"`
	UITitle        string   `yaml:"ui_title,omitempty"`
	ClientBaseURL  string   `yaml:"client_base_url,omitempty"`
	UIModels       []string `yaml:"ui_models,omitempty"`
}

// DecodeIssuerYAML parses the generic issuer config node for driver: email.
func DecodeIssuerYAML(node yaml.Node) (IssuerYAML, error) {
	if node.Kind == 0 {
		return IssuerYAML{}, fmt.Errorf("email issuer: config required")
	}
	var cfg IssuerYAML
	if err := node.Decode(&cfg); err != nil {
		return IssuerYAML{}, fmt.Errorf("email issuer config: %w", err)
	}
	return cfg, nil
}