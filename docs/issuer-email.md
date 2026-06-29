# Email OTP issuer

Admin-plane self-service gateway keys via email one-time password.

**Packages:** `github.com/subosito/daigate-extensions/issuer/email`, `…/issuer/email/register`

---

## Register

```go
import _ "github.com/subosito/daigate-extensions/issuer/email/register"
```

---

## Config

```yaml
ingress:
  client_auth: keyring
  issuers:
    - driver: email
      config:
        ttl: 15m
        smtp_host: smtp.example.com
        smtp_pass_env: SMTP_PASS
        from: noreply@example.com
        allowed_domains: [example.com]
        ui: true
        ui_title: my-gateway
        client_base_url: https://aigate.example.com/v1
        ui_models: [gpt-5.4-mini]
```

Core stores `driver` + opaque `config`; `issuer/email` decodes the fields above.

Without SMTP configured, OTP is logged to stderr (dev fallback).

---

## Admin routes

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/v1/issuer/email/request` | Send OTP |
| POST | `/v1/issuer/email/verify` | Exchange OTP for `issued` gateway key |

Minted keys are **`issued`** kind with TTL from issuer config. Public OTP flow — no admin token required for request/verify.

Optional `ui: true` serves a small HTML form on the admin listener.

---

## Tenant-facing env vars

```text
OPENAI_BASE_URL=https://llm.example.com/v1
OPENAI_API_KEY=<tenant key from email OTP>
OPENAI_MODEL=<catalog model allowed for tenant>
```