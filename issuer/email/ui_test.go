package email

import "testing"

func TestResolveDataBaseURLExplicitPublic(t *testing.T) {
	cfg := UIConfig{
		ClientBaseURL: "https://aigate.subosito.id/v1",
	}
	got := resolveDataBaseURL(cfg)
	if got != cfg.ClientBaseURL {
		t.Fatalf("got %q want %q", got, cfg.ClientBaseURL)
	}
}

func TestResolveDataBaseURLLocalFallback(t *testing.T) {
	got := resolveDataBaseURL(UIConfig{})
	if got != "http://127.0.0.1:9420/v1" {
		t.Fatalf("got %q", got)
	}
}