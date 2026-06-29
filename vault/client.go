package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

)

const (
	vaultRequestTimeout = 5 * time.Second
	vaultCacheTTL       = 30 * time.Second
	breakerThreshold    = 3
	breakerCooldown     = 10 * time.Second
)

// Client reads upstream material from HashiCorp Vault KV v2.
type Client struct {
	baseURL    string
	mount      string
	pathPrefix string
	paths      map[string]string
	token      string
	http       *http.Client

	mu        sync.Mutex
	cache     map[string]cacheEntry
	failures  int
	openUntil time.Time
}

type cacheEntry struct {
	data    map[string]any
	expires time.Time
}

// NewClient builds a KV reader from decoded vault backend_config.
func NewClient(v VaultYAML) (*Client, error) {
	addr := strings.TrimSpace(v.Address)
	if addr == "" {
		return nil, fmt.Errorf("vault backend_config.address required")
	}
	token := strings.TrimSpace(os.Getenv(v.Auth.TokenEnv))
	if token == "" {
		return nil, fmt.Errorf("vault token required: set %s", v.Auth.TokenEnv)
	}
	if v.Auth.Method != "" && v.Auth.Method != "token" {
		return nil, fmt.Errorf("unsupported vault auth method %q (token only)", v.Auth.Method)
	}
	prefix := strings.Trim(v.PathPrefix, "/")
	return &Client{
		baseURL:    strings.TrimRight(addr, "/"),
		mount:      strings.Trim(v.Mount, "/"),
		pathPrefix: prefix,
		paths:      v.Paths,
		token:      token,
		http:       &http.Client{Timeout: vaultRequestTimeout},
		cache:      make(map[string]cacheEntry),
	}, nil
}

func (c *Client) pathFor(profile string) string {
	if p, ok := c.paths[profile]; ok && strings.TrimSpace(p) != "" {
		return strings.Trim(p, "/")
	}
	if c.pathPrefix == "" {
		return profile
	}
	return c.pathPrefix + "/" + profile
}

// ReadKV fetches and decodes the secret data map at the profile path.
func (c *Client) ReadKV(ctx context.Context, profile string) (map[string]any, error) {
	if err := c.checkBreaker(); err != nil {
		return nil, err
	}
	if data, ok := c.cached(profile); ok {
		return data, nil
	}
	data, err := c.readKVHTTP(ctx, profile)
	if err != nil {
		c.recordFailure(err)
		return nil, err
	}
	c.recordSuccess(profile, data)
	return data, nil
}

func (c *Client) checkBreaker() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Now().Before(c.openUntil) {
		return fmt.Errorf("vault circuit open")
	}
	return nil
}

func (c *Client) cached(profile string) (map[string]any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.cache[profile]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return cloneMap(entry.data), true
}

func (c *Client) recordSuccess(profile string, data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
	c.openUntil = time.Time{}
	c.cache[profile] = cacheEntry{data: cloneMap(data), expires: time.Now().Add(vaultCacheTTL)}
}

func (c *Client) recordFailure(err error) {
	if err == ErrNotFound {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.failures >= breakerThreshold {
		c.openUntil = time.Now().Add(breakerCooldown)
		c.failures = 0
	}
}

func (c *Client) readKVHTTP(ctx context.Context, profile string) (map[string]any, error) {
	path := c.pathFor(profile)
	url := fmt.Sprintf("%s/v1/%s/data/%s", c.baseURL, c.mount, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("vault read %s: %s", path, strings.TrimSpace(string(body)))
	}

	var envelope struct {
		Data struct {
			Data map[string]any `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data.Data == nil {
		return nil, fmt.Errorf("vault read %s: empty data", path)
	}
	return envelope.Data.Data, nil
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}