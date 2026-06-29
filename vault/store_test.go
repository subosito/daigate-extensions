package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/subosito/daigate/credential/seal"
	"github.com/subosito/daigate/credential/store"
	vaultstore "github.com/subosito/daigate-extensions/vault"
)

func openTestSQLite(t *testing.T) *store.SQLite {
	t.Helper()
	key, err := seal.ParseKey("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.OpenSQLite(t.TempDir()+"/broker.db", key)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func vaultMockServer(t *testing.T, secrets map[string]map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for profile, data := range secrets {
			path := "/v1/secret/data/ops/" + profile
			if r.URL.Path == path {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{"data": data},
				})
				return
			}
		}
		http.NotFound(w, r)
	}))
}

func TestBackendVaultModeFallbackSQLite(t *testing.T) {
	sqlite := openTestSQLite(t)
	if _, err := sqlite.PutAPIKey(context.Background(), "local", "sk-local"); err != nil {
		t.Fatal(err)
	}

	srv := vaultMockServer(t, map[string]map[string]any{
		"remote": {"kind": "api_key", "api_key": "sk-remote"},
	})
	defer srv.Close()

	t.Setenv("VAULT_TOKEN_TEST", "tok")
	vcfg := vaultstore.VaultYAML{
		Address: srv.URL,
		Mount:   "secret",
		Paths:   map[string]string{"remote": "ops/remote"},
	}
	vcfg.Auth.TokenEnv = "VAULT_TOKEN_TEST"
	vc, err := vaultstore.NewClient(vcfg)
	if err != nil {
		t.Fatal(err)
	}
	backend := vaultstore.NewBackend(sqlite, vc, vaultstore.VaultYAML{}, "vault")

	remote, err := backend.Get(context.Background(), "remote")
	if err != nil || remote.APIKey != "sk-remote" {
		t.Fatalf("remote: %v %+v", err, remote)
	}
	local, err := backend.Get(context.Background(), "local")
	if err != nil || local.APIKey != "sk-local" {
		t.Fatalf("local: %v %+v", err, local)
	}
}

func TestBackendHybridMode(t *testing.T) {
	sqlite := openTestSQLite(t)
	if _, err := sqlite.PutAPIKey(context.Background(), "local", "sk-local"); err != nil {
		t.Fatal(err)
	}

	srv := vaultMockServer(t, map[string]map[string]any{
		"remote": {"kind": "api_key", "api_key": "sk-remote"},
	})
	defer srv.Close()

	t.Setenv("VAULT_TOKEN_TEST", "tok")
	vcfg := vaultstore.VaultYAML{
		Address:  srv.URL,
		Mount:    "secret",
		Paths:    map[string]string{"remote": "ops/remote"},
		Profiles: []string{"remote", "vault-only"},
	}
	vcfg.Auth.TokenEnv = "VAULT_TOKEN_TEST"
	vc, err := vaultstore.NewClient(vcfg)
	if err != nil {
		t.Fatal(err)
	}
	backend := vaultstore.NewBackend(sqlite, vc, vcfg, "hybrid")

	if _, err := backend.Get(context.Background(), "local"); err != nil {
		t.Fatalf("local should use sqlite: %v", err)
	}
	remote, err := backend.Get(context.Background(), "remote")
	if err != nil || remote.APIKey != "sk-remote" {
		t.Fatalf("remote: %v %+v", err, remote)
	}
	if _, err := backend.Get(context.Background(), "vault-only"); err == nil {
		t.Fatal("expected vault-only missing error")
	}

	list, err := backend.ListSummaries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var vaultOnly bool
	for _, cs := range list {
		if cs.Profile == "remote" && cs.Source != "vault" {
			t.Fatalf("remote source=%q", cs.Source)
		}
		if cs.Profile == "vault-only" {
			vaultOnly = true
			if cs.Source != "vault" {
				t.Fatalf("vault-only source=%q", cs.Source)
			}
		}
	}
	if !vaultOnly {
		t.Fatal("expected vault-only summary")
	}
}