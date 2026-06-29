package vault

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// VaultYAML is HashiCorp Vault KV v2 settings under credential.backend_config.
type VaultYAML struct {
	Address    string            `yaml:"address"`
	Mount      string            `yaml:"mount"`
	PathPrefix string            `yaml:"path_prefix"`
	Paths      map[string]string `yaml:"paths,omitempty"`
	Profiles   []string          `yaml:"profiles,omitempty"`
	Auth       struct {
		Method   string `yaml:"method"`
		TokenEnv string `yaml:"token_env"`
	} `yaml:"auth"`
}

// DecodeVaultYAML parses credential.backend_config for vault/hybrid backends.
func DecodeVaultYAML(node yaml.Node) (VaultYAML, error) {
	if node.Kind == 0 {
		return VaultYAML{}, fmt.Errorf("vault backend: backend_config required")
	}
	var cfg VaultYAML
	if err := node.Decode(&cfg); err != nil {
		return VaultYAML{}, fmt.Errorf("vault backend_config: %w", err)
	}
	if cfg.Auth.TokenEnv == "" {
		cfg.Auth.TokenEnv = "VAULT_TOKEN"
	}
	if cfg.Mount == "" {
		cfg.Mount = "secret"
	}
	return cfg, nil
}