# Link-time registration

Vault backends and the email issuer hook into daigate core via `init()` in small `register` subpackages. Blank-import only what your operator binary needs.

---

## Why separate `register` packages?

daigate core exposes link-time hooks:

| Core API | Registered by |
|----------|---------------|
| `gateway.RegisterCredentialBackend` | `vault/register` → `vault`, `hybrid` |
| `gateway.RegisterAdminIssuer` | `issuer/email/register` → `email` |

Translate adapters do **not** use this pattern — wire them through [`compose`](compose.md) at serve time.

Keeping `register` in a subpackage (not the main `vault` or `issuer/email` package) means importing types/helpers for tests does not pull backend/issuer wiring into your binary.

---

## Imports

**Vault only:**

```go
import _ "github.com/subosito/daigate-extensions/vault/register"
```

**Email issuer only:**

```go
import _ "github.com/subosito/daigate-extensions/issuer/email/register"
```

**Both:**

```go
import (
    _ "github.com/subosito/daigate-extensions/issuer/email/register"
    _ "github.com/subosito/daigate-extensions/vault/register"
)
```