package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/subosito/daigate-extensions/vault"
)

func TestReadKV(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/data/daigate/credentials/mock" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("X-Vault-Token"); got != "tok" {
			t.Fatalf("token=%q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"data": map[string]any{"kind": "api_key", "api_key": "sk-vault"},
			},
		})
	}))
	defer srv.Close()

	t.Setenv("VAULT_TOKEN_TEST", "tok")
	cfg := vault.VaultYAML{
		Address:    srv.URL,
		Mount:      "secret",
		PathPrefix: "daigate/credentials",
	}
	cfg.Auth.TokenEnv = "VAULT_TOKEN_TEST"
	vc, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	data, err := vc.ReadKV(context.Background(), "mock")
	if err != nil {
		t.Fatal(err)
	}
	mat, err := vault.MaterialFromKV("mock", data)
	if err != nil || mat.APIKey != "sk-vault" {
		t.Fatalf("material: %v %+v", err, mat)
	}
}

func TestReadKVNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	t.Setenv("VAULT_TOKEN_TEST", "tok")
	cfg := vault.VaultYAML{Address: srv.URL, Mount: "secret"}
	cfg.Auth.TokenEnv = "VAULT_TOKEN_TEST"
	vc, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vc.ReadKV(context.Background(), "missing")
	if err != vault.ErrNotFound {
		t.Fatalf("err=%v", err)
	}
}