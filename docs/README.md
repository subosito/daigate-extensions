# daigate-extensions documentation

Library module: Vault credential backends, email OTP issuer, and optional vendor translate adapters.

Link from an operator binary that already embeds `github.com/subosito/daigate`.

**Module:** `github.com/subosito/daigate-extensions` · **Go:** 1.26

**Catalog routing:** ambiguous model+wire pairs use **`X-Catalog-Modality`** (yaml `modalities.<key>` — operator-defined; see `daigate/docs/catalog.md`). Hosts map product-specific headers before ingress. `gateway.WrapDataHandler` for middleware.

---

## Components

| Doc | Package | Purpose |
|-----|---------|---------|
| [compose.md](compose.md) | `compose` | `ExtAdapters()`, `AllAdapters()`, `FromConfig()` |
| [vault.md](vault.md) | `vault` | KV v2 backend (`vault`, `hybrid`) |
| [issuer-email.md](issuer-email.md) | `issuer/email` | Admin-plane email OTP issuer |
| [register.md](register.md) | `vault/register`, `issuer/email/register` | Link-time hooks into daigate core |

---

## Quick start

```go
import (
    "github.com/subosito/daigate/gateway"
    "github.com/subosito/daigate-extensions/compose"
    _ "github.com/subosito/daigate-extensions/issuer/email/register"
    _ "github.com/subosito/daigate-extensions/vault/register"
)

// In gateway.Serve / EmbedServe registry callback:
reg, err := compose.FromConfig(cfg.Adapters.Enable)
```

Gateway yaml (`daigate.yaml`, `providers.yaml`) lives in the operator repo. Component-specific snippets are in each doc below.

---

## Layout

```text
daigate-extensions/
  compose/
  vault/
    register/
  issuer/email/
    register/
  docs/
```

---

## Development

```bash
just          # go vet + go test -race
```

CI: [`.github/workflows/ci.yml`](../.github/workflows/ci.yml).