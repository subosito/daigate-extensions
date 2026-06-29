# Vault credential backend

HashiCorp Vault KV v2 backends for upstream provider credentials.

**Packages:** `github.com/subosito/daigate-extensions/vault`, `…/vault/register`

---

## Register

```go
import _ "github.com/subosito/daigate-extensions/vault/register"
```

Registers `vault` and `hybrid` credential backends with daigate core.

---

## Vault-only (`backend: vault`)

```yaml
credential:
  backend: vault
  broker: broker.db
  encryption:
    key_env: DAIGATE_BROKER_KEY
  backend_config:
    address: https://vault.example:8200
    mount: secret
    path_prefix: daigate/credentials
    auth:
      method: token
      token_env: VAULT_TOKEN
    paths:
      acme: ops/llm/acme   # optional per-profile KV path override
```

**KV value shape** (JSON at path):

```json
{
  "kind": "api_key",
  "api_key": "sk-…"
}
```

OAuth profiles can use the same envelope (`kind`, `access_token`, `refresh_token`, `expires_at`) when Vault is the source of truth. Typical deployments use Vault for static `api_key` paths; OAuth login/refresh writes encrypted sqlite unless the deployment uses Vault-only static secrets.

---

## Hybrid (`backend: hybrid`)

OAuth/metadata stays in encrypted sqlite; selected profiles read `api_key` from Vault KV at forward time:

```yaml
credential:
  backend: hybrid
  broker: broker.db
  encryption:
    key_env: DAIGATE_BROKER_KEY
  backend_config:
    address: https://vault.example:8200
    mount: secret
    path_prefix: daigate/credentials
    profiles: [acme, team-beta]   # KV read; others use sqlite
```

| Concern | Behavior |
|---------|----------|
| `Store.Get` | Read KV v2; map to `Material` for inject |
| `credential login` / refresh | sqlite (encrypted) |
| `credential list` | sqlite summaries; Vault-backed profiles show `source=vault` |
| `DAIGATE_BROKER_KEY` | Still required for sqlite encryption and gateway key hashes |