package vault

import (
	"fmt"
	"strconv"
	"time"

	"github.com/subosito/daigate/credential/store"
)

// MaterialFromKV maps a Vault KV data map to upstream Material.
func MaterialFromKV(profile string, data map[string]any) (store.Material, error) {
	kind, _ := data["kind"].(string)
	if kind == "" {
		kind, _ = data["type"].(string)
	}
	switch store.Kind(kind) {
	case store.KindAPIKey:
		key, _ := data["api_key"].(string)
		if key == "" {
			key, _ = data["key"].(string)
		}
		if key == "" {
			return store.Material{}, fmt.Errorf("vault profile %q: api_key missing", profile)
		}
		return store.Material{Profile: profile, Kind: store.KindAPIKey, APIKey: key}, nil
	case store.KindOAuth:
		access, _ := data["access_token"].(string)
		if access == "" {
			access, _ = data["access"].(string)
		}
		refresh, _ := data["refresh_token"].(string)
		if refresh == "" {
			refresh, _ = data["refresh"].(string)
		}
		if access == "" {
			return store.Material{}, fmt.Errorf("vault profile %q: access_token missing", profile)
		}
		mat := store.Material{
			Profile:      profile,
			Kind:         store.KindOAuth,
			AccessToken:  access,
			RefreshToken: refresh,
		}
		if v, ok := data["expires_at"]; ok {
			mat.ExpiresAt = parseTime(v)
		} else if v, ok := data["expires"]; ok {
			mat.ExpiresAt = parseTime(v)
		}
		if email, ok := data["email"].(string); ok {
			mat.Email = email
		}
		if account, ok := data["account_id"].(string); ok {
			mat.AccountID = account
		}
		if project, ok := data["project_id"].(string); ok {
			mat.ProjectID = project
		}
		return mat, nil
	default:
		return store.Material{}, fmt.Errorf("vault profile %q: unknown kind %q", profile, kind)
	}
}

func parseTime(v any) time.Time {
	switch t := v.(type) {
	case string:
		if n, err := strconv.ParseInt(t, 10, 64); err == nil {
			return timeFromEpoch(n)
		}
		if ts, err := time.Parse(time.RFC3339, t); err == nil {
			return ts
		}
	case float64:
		return timeFromEpoch(int64(t))
	case int64:
		return timeFromEpoch(t)
	case int:
		return timeFromEpoch(int64(t))
	}
	return time.Time{}
}

func timeFromEpoch(n int64) time.Time {
	if n > 1_000_000_000_000 {
		return time.UnixMilli(n)
	}
	return time.Unix(n, 0)
}