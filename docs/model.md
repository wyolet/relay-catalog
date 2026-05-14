# Model

A **single callable model** — `claude-opus-4-5-20251101`, `gpt-4o`, etc. The Model document declares what the model can do, who publishes it, and which [Hosts](host.md) can serve it via which wire-protocol adapter.

Files live under `data/providers/<provider>/models/<model-name>.yaml`.

## Metadata

| Field | Required | Description |
|---|---|---|
| `name` | yes | The callable model handle. Matched against `Pricing.spec.targetModels`. Stable — do not rename; use `spec.aliases` instead. |
| `owner.kind` | yes | Must be `provider`. |
| `owner.id` | yes | `metadata.name` of the owning [Provider](provider.md). |
| `displayName`, `description`, `labels` | no | Standard. |

## Spec

### Serving (required)

| Field | Required | Description |
|---|---|---|
| `hosts` | **yes, min 1** | Array of host bindings — see below. |

Each `hosts[]` entry:

| Field | Required | Description |
|---|---|---|
| `host` | **yes** | `metadata.name` of a [Host](host.md). |
| `upstreamName` | **yes** | The exact model ID string the upstream API expects (may differ from `metadata.name`). |
| `adapter` | **yes** | Wire protocol: `openai` or `anthropic`. Tells relay how to translate requests/responses. |
| `enabled` | no | Defaults to true. Disable a single host binding without removing the model. |

### Taxonomy

| Field | Type | Description |
|---|---|---|
| `family` | string | Model family label (`claude`, `gpt`, `llama`, …). |
| `version` | string | Free-form version string. |

### Capabilities

`capabilities` is an object of booleans. Absent fields are treated as `false`. Set only the ones the model actually supports.

`chat`, `embeddings`, `streaming`, `tools`, `parallelTools`, `vision`, `audio`, `promptCache`, `reasoning`, `jsonMode`, `structuredOutputs`, `batch`, `computerUse`, `webSearch`, `fileInput`, `audioInput`, `audioOutput`, `systemMessages`, `assistantPrefill`.

### Modalities

| Field | Type | Description |
|---|---|---|
| `modalities.input` | `[]string` | Media types accepted (`text`, `image`, `audio`, `file`). |
| `modalities.output` | `[]string` | Media types produced. |

### Context window

| Field | Type | Description |
|---|---|---|
| `contextWindowTotal` | int | Canonical total context size in tokens. |
| `contextWindowInput` | int | Soft input cap if different from total. |
| `contextWindowOutput` | int | Soft output cap if different from total. |
| `maxOutputTokens` | int | Hard cap on tokens produced in a single response. |

### Lifecycle

| Field | Type | Description |
|---|---|---|
| `releaseDate` | string | ISO date. |
| `knowledgeCutoff` | string | Training data cutoff date. |
| `deprecation.status` | enum | `active` \| `deprecated` \| `sunset`. |
| `deprecation.sunsetDate` | string | When the upstream removes access. |
| `deprecation.replacement` | string | UUID of the replacement Model. |
| `deprecationDate` | string | Legacy flat field; prefer `deprecation`. |

### Discovery

| Field | Type | Description |
|---|---|---|
| `aliases` | `[]string` | Alternate callable names. Must be unique across all models and not equal `metadata.name`. |
| `tags` | `[]string` | Free-form. |
| `documentation` | string | Markdown or URL. |
| `license` | string | License identifier. |
| `providerModelPageURL` | string | Vendor's product page (validated as HTTP URL). |
| `enabled` | bool | Defaults to true. |

## Example

```yaml
apiVersion: relay.wyolet.dev/v1
kind: Model
metadata:
  name: claude-opus-4-5-20251101
  displayName: Claude Opus 4.5
  owner:
    kind: provider
    id: anthropic
spec:
  family: claude
  hosts:
    - host: anthropic
      upstreamName: claude-opus-4-5-20251101
      adapter: anthropic
    - host: amazon-bedrock
      upstreamName: anthropic.claude-opus-4-5-v1:0
      adapter: anthropic
  capabilities:
    chat: true
    streaming: true
    tools: true
    vision: true
    promptCache: true
    reasoning: true
    systemMessages: true
  modalities:
    input: [text, image]
    output: [text]
  contextWindowTotal: 200000
  maxOutputTokens: 32000
  releaseDate: "2025-11-01"
```

## Relationships

- `metadata.owner.id` → [Provider](provider.md) by name.
- `spec.hosts[].host` → [Host](host.md) by name (one Model, many Hosts).
- `spec.deprecation.replacement` → another Model **by UUID** (not name).
- Pricing references Models via `spec.targetModels[]` (inverse — see [Pricing](pricing.md)).
