# daigate-extensions

Library extras for **daigate operator binaries** — HashiCorp Vault credential backends, email OTP issuer, optional vendor translate adapters.

**Module:** `github.com/subosito/daigate-extensions` · **Go:** 1.26

Link from a binary that already embeds `github.com/subosito/daigate`. No standalone CLI or example operator ships here — yaml and deployment live in your app repo.

---

## Components

| Package | Doc | Purpose |
|---------|-----|---------|
| `compose` | [docs/compose.md](docs/compose.md) | `ExtAdapters()`, `AllAdapters()`, `FromConfig()` |
| `adapters/elevenlabs` | [docs/adapters/elevenlabs.md](docs/adapters/elevenlabs.md) | OpenAI speech → ElevenLabs TTS |
| `vault` | [docs/vault.md](docs/vault.md) | KV v2 backend (`vault`, `hybrid`) |
| `issuer/email` | [docs/issuer-email.md](docs/issuer-email.md) | Admin-plane email OTP + optional setup UI |
| `*/register` | [docs/register.md](docs/register.md) | Link-time hooks (`init()` → daigate gateway) |

---

## Wiring

**Translate adapters** — via `compose` at serve time:

```go
import "github.com/subosito/daigate-extensions/compose"

reg, err := compose.FromConfig(cfg.Adapters.Enable)
```

**Backends and issuers** — blank-import only what you need:

```go
import _ "github.com/subosito/daigate-extensions/issuer/email/register"
import _ "github.com/subosito/daigate-extensions/vault/register"
```

Full operator `main` also calls `gateway.Serve` with `compose.FromConfig` and passthrough from daigate core. See [docs/compose.md](docs/compose.md).

Gateway yaml (`daigate.yaml`, `providers.yaml`) and secrets belong in the **operator repo**. Each component doc has the yaml snippets for its config blocks.

---

## Layout

```text
compose/
adapters/elevenlabs/
vault/ + vault/register/
issuer/email/ + issuer/email/register/
docs/
```

---

## Development

```bash
just          # go vet + go test -race
```

CI: [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

Doc hub: **[docs/README.md](docs/README.md)** · **Agents:** [AGENTS.md](AGENTS.md)