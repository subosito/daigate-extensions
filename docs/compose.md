# compose

Wiring helpers for linking extension translate adapters into a daigate operator binary.

**Package:** `github.com/subosito/daigate-extensions/compose`

---

## API

| Function | Returns |
|----------|---------|
| `ExtAdapters()` | Optional vendor translate adapters shipped in this module (currently none) |
| `AllAdapters()` | Core `passthrough` + `ExtAdapters()` |
| `FromConfig(enable []string)` | Filtered `*adaptersdk.Registry` from `daigate.yaml` `adapters.enable` |

`FromConfig` delegates to `daigate/compose.FromConfig`, matching adapter names against `AllAdapters()`.

---

## Usage

Pass the registry into `gateway.Serve` or `gateway.EmbedServe` from your operator binary:

```go
import (
    "github.com/subosito/daigate/gateway"
    "github.com/subosito/daigate-extensions/compose"
    _ "github.com/subosito/daigate-extensions/issuer/email/register"
    _ "github.com/subosito/daigate-extensions/vault/register"
)

reg, err := compose.FromConfig(cfg.Adapters.Enable)
```

Additional translate adapters can be passed alongside `ExtAdapters()` in the operator binary's adapter list.

---

## Config

```yaml
adapters:
  enable: [passthrough]              # required for chat/embed/media relay
```