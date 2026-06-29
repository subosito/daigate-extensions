package email_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/subosito/daigate-extensions/issuer/email"
	"github.com/subosito/daigate/ingress/keyring"
	_ "modernc.org/sqlite"
)

func TestEmailIssuerVerifyMintsKey(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ks := keyring.NewMemoryStore()
	iss := email.NewFromConfig(db, ks, 15*time.Minute, "", "", "", "", "", nil)

	mux := http.NewServeMux()
	if err := iss.Mount(mux); err != nil {
		t.Fatal(err)
	}

	if err := email.SeedOTP(db, "user@example.com", "123456", time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	verifyBody, _ := json.Marshal(map[string]string{"email": "user@example.com", "code": "123456"})
	vreq := httptest.NewRequest(http.MethodPost, "/v1/issuer/email/verify", bytes.NewReader(verifyBody))
	vrec := httptest.NewRecorder()
	mux.ServeHTTP(vrec, vreq)
	if vrec.Code != http.StatusOK {
		t.Fatalf("verify status=%d %s", vrec.Code, vrec.Body.String())
	}
	var out struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(vrec.Body.Bytes(), &out); err != nil || !strings.HasPrefix(out.Key, "sk-dg-") {
		t.Fatalf("verify body=%s", vrec.Body.Bytes())
	}
}

func TestEmailIssuerVerifyRevokesPreviousKey(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ks := keyring.NewMemoryStore()
	iss := email.NewFromConfig(db, ks, 15*time.Minute, "", "", "", "", "", nil)
	mux := http.NewServeMux()
	if err := iss.Mount(mux); err != nil {
		t.Fatal(err)
	}

	const addr = "user@dkatalis.com"
	verify := func(code string) string {
		if err := email.SeedOTP(db, addr, code, time.Now().Add(time.Minute)); err != nil {
			t.Fatal(err)
		}
		body, _ := json.Marshal(map[string]string{"email": addr, "code": code})
		req := httptest.NewRequest(http.MethodPost, "/v1/issuer/email/verify", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("verify status=%d %s", rec.Code, rec.Body.String())
		}
		var out struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out.Key
	}

	first := verify("111111")
	second := verify("222222")
	if first == second {
		t.Fatal("expected distinct keys")
	}
	if _, err := ks.Verify(context.Background(), first); err == nil {
		t.Fatal("previous key should be revoked")
	}
	if p, err := ks.Verify(context.Background(), second); err != nil || p.ID != addr {
		t.Fatalf("new key verify err=%v principal=%+v", err, p)
	}
}

func TestEmailIssuerRejectsDisallowedDomain(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	iss := email.NewFromConfig(db, keyring.NewMemoryStore(), 15*time.Minute, "", "", "", "", "", []string{"dkatalis.com"})
	mux := http.NewServeMux()
	if err := iss.Mount(mux); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{"email": "user@gmail.com"})
	req := httptest.NewRequest(http.MethodPost, "/v1/issuer/email/request", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("request status=%d %s", rec.Code, rec.Body.String())
	}
}

func TestEmailIssuerAllowsWhitelistedDomain(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	iss := email.NewFromConfig(db, keyring.NewMemoryStore(), 15*time.Minute, "", "", "", "", "", []string{"dkatalis.com"})
	mux := http.NewServeMux()
	if err := iss.Mount(mux); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{"email": "User@DKatalis.COM"})
	req := httptest.NewRequest(http.MethodPost, "/v1/issuer/email/request", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("request status=%d %s", rec.Code, rec.Body.String())
	}
}