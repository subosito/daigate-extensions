# AGENTS.md — daigate-extensions

`github.com/subosito/daigate-extensions` — library extras for daigate operator binaries (Go 1.26).

## Repo

Link from an operator binary that embeds `github.com/subosito/daigate`. Ships credential backends, issuers, and optional translate adapters.

## Components

| Package | Doc |
|---------|-----|
| `compose` | [docs/compose.md](docs/compose.md) |
| `adapters/elevenlabs` | [docs/adapters/elevenlabs.md](docs/adapters/elevenlabs.md) |
| `vault` | [docs/vault.md](docs/vault.md) |
| `issuer/email` | [docs/issuer-email.md](docs/issuer-email.md) |
| `vault/register`, `issuer/email/register` | [docs/register.md](docs/register.md) |

Hub: [docs/README.md](docs/README.md)

## Layout

```text
compose/
adapters/elevenlabs/
vault/ + vault/register/
issuer/email/ + issuer/email/register/
docs/
```

## Wiring

**Backends and issuers** — blank-import a `*/register` package so `init()` calls daigate hooks:

```go
_ "github.com/subosito/daigate-extensions/issuer/email/register"
_ "github.com/subosito/daigate-extensions/vault/register"
```

Keep `register` in a subpackage so importing `vault` or `issuer/email` for types/tests does not auto-wire the binary.

**Translate adapters** — register via `compose.ExtAdapters()` / `compose.FromConfig()` at serve time.

## Commands

```bash
just          # go vet + go test -race ./...
```

Local monorepo: gitignored `go.work` with `replace` to sibling checkouts.

## Making changes

- **New translate adapter** → `adapters/<name>/`, add to `compose/ExtAdapters()`, add `docs/adapters/<name>.md`
- **New backend or issuer** → main package + `register/` with `init()` hook, add `docs/<name>.md`
- **Yaml snippets** → live in the component doc; keep `docs/README.md` as a short index

## Verify

```bash
just
```