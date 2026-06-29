package vault

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/subosito/daigate/credential/store"
)

// Backend composes sqlite (writes, OAuth, gateway metadata) with Vault KV reads.
type Backend struct {
	sqlite   *store.SQLite
	client   *Client
	mode     string
	profiles map[string]struct{}
}

// NewBackend wraps sqlite with Vault KV reads.
// mode is "vault" (try vault then sqlite) or "hybrid" (listed profiles vault-only).
func NewBackend(sqlite *store.SQLite, client *Client, v VaultYAML, mode string) *Backend {
	profiles := make(map[string]struct{}, len(v.Profiles))
	for _, p := range v.Profiles {
		profiles[p] = struct{}{}
	}
	return &Backend{sqlite: sqlite, client: client, mode: mode, profiles: profiles}
}

// BrokerDB exposes the local sqlite broker for keyring/admin.
func (b *Backend) BrokerDB() *sql.DB { return b.sqlite.DB() }

func (b *Backend) Close() error { return b.sqlite.Close() }

func (b *Backend) useVault(profile string) bool {
	switch b.mode {
	case "hybrid":
		_, ok := b.profiles[profile]
		return ok
	default:
		return true
	}
}

func (b *Backend) Get(ctx context.Context, profile string) (store.Material, error) {
	if b.useVault(profile) {
		data, err := b.client.ReadKV(ctx, profile)
		if err == nil {
			return MaterialFromKV(profile, data)
		}
		if err != ErrNotFound {
			return store.Material{}, err
		}
		if b.mode == "hybrid" {
			return store.Material{}, fmt.Errorf("credential profile %q not found in vault", profile)
		}
	}
	return b.sqlite.Get(ctx, profile)
}

func (b *Backend) PutAPIKey(ctx context.Context, profile, key string) (int64, error) {
	return b.sqlite.PutAPIKey(ctx, profile, key)
}

func (b *Backend) PutOAuth(ctx context.Context, profile string, m store.Material) (int64, error) {
	return b.sqlite.PutOAuth(ctx, profile, m)
}

func (b *Backend) UpdateOAuth(ctx context.Context, profile string, m store.Material) error {
	return b.sqlite.UpdateOAuth(ctx, profile, m)
}

func (b *Backend) ListSummaries(ctx context.Context) ([]store.CredentialSummary, error) {
	list, err := b.sqlite.ListSummaries(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(list))
	for i := range list {
		seen[list[i].Profile] = struct{}{}
		if b.useVault(list[i].Profile) {
			list[i].Source = "vault"
		}
	}
	for profile := range b.profiles {
		if _, ok := seen[profile]; ok {
			continue
		}
		now := time.Now().UnixMilli()
		list = append(list, store.CredentialSummary{
			Profile:   profile,
			Kind:      store.KindAPIKey,
			Status:    "active",
			Source:    "vault",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return list, nil
}

func (b *Backend) GetSummary(ctx context.Context, id int64) (store.CredentialSummary, error) {
	cs, err := b.sqlite.GetSummary(ctx, id)
	if err != nil {
		return cs, err
	}
	if b.useVault(cs.Profile) {
		cs.Source = "vault"
	}
	return cs, nil
}

func (b *Backend) Disable(ctx context.Context, id int64, cause string) error {
	return b.sqlite.Disable(ctx, id, cause)
}

func (b *Backend) SnapshotMeta(ctx context.Context) (store.SnapshotMeta, error) {
	return b.sqlite.SnapshotMeta(ctx)
}

func (b *Backend) BumpGeneration(ctx context.Context) error {
	return b.sqlite.BumpGeneration(ctx)
}