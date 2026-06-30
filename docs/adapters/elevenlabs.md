# ElevenLabs speech adapter

Translate adapter: OpenAI `POST /v1/audio/speech` → ElevenLabs `POST /v1/text-to-speech/{voice}`.

**Package:** `github.com/subosito/daigate-extensions/adapters/elevenlabs`

---

## Enable

Add to `daigate.yaml`:

```yaml
adapters:
  enable: [passthrough, elevenlabs]
```

Register in your operator binary via `compose.FromConfig` (included in `ExtAdapters()`). No blank-import `register` package needed — adapters register through the compose registry at serve time.

---

## Catalog

```yaml
providers:
  elevenlabs:
    credential_profile: elevenlabs
    # inject optional — adapter defaults to xi-api-key
    surfaces:
      speech:
        adapter: elevenlabs
        base_url: https://api.elevenlabs.io

models:
  eleven-multilingual-v2:
    modalities:
      speech:
        wire: openai-audio-speech
        providers:
          - provider_ref: elevenlabs
            surface: speech
            model: eleven_multilingual_v2
```

Import the upstream API key: `daigate credential import elevenlabs --api-key <key>`.

---

## Wire mapping

| OpenAI field | ElevenLabs |
|--------------|------------|
| `input` | `text` (JSON body) |
| `voice` | path segment `/v1/text-to-speech/{voice}` |
| `response_format` | `output_format` query + `Accept` header (`mp3`, `opus`, `wav`, …) |
| catalog `model` | `model_id` (JSON body) |

Credential inject uses `xi-api-key` by default (`DefaultInject` in adapter code). Override in yaml:

```yaml
inject:
  xi-api-key: "${key}"
```

See [daigate catalog-inject.md](https://github.com/subosito/daigate/blob/main/docs/catalog-inject.md).