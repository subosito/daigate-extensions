package vault_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	vaultstore "github.com/subosito/daigate-extensions/vault"
)

func TestDecodeVaultYAML(t *testing.T) {
	var root struct {
		BackendConfig yaml.Node `yaml:"backend_config"`
	}
	if err := yaml.Unmarshal([]byte(`
backend_config:
  address: https://vault.example:8200
  mount: secret
  path_prefix: daigate/credentials
  profiles: [acme]
  auth:
    token_env: VAULT_TOKEN_TEST
`), &root); err != nil {
		t.Fatal(err)
	}
	cfg, err := vaultstore.DecodeVaultYAML(root.BackendConfig)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "https://vault.example:8200" || cfg.Mount != "secret" {
		t.Fatalf("cfg=%+v", cfg)
	}
	if cfg.Auth.TokenEnv != "VAULT_TOKEN_TEST" {
		t.Fatalf("token_env=%q", cfg.Auth.TokenEnv)
	}
}