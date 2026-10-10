# Model

A **single callable model** — `claude-sonnet-5-5`, `gpt-4o`, etc. The Model document declares what the model can do and who publishes it. How the model is served (which host, which adapter) is declared separately in a [HostBinding](#hostbinding) file.

Files live under `data/providers/<provider>/models/<model-name>.yaml`.

## Metadata

| Field | Required | Description |
|---|---|---|
| `name` | yes | The callable model handle. Matched against `Pricing.spec.targetModels`. Stable — do not rename; use `spec.aliases` instead. |
| `owner.kind` | yes | Must be `provider`. |
| `owner.name` | yes | `metadata.name` of the owning [Provider](provider.md). |
| `displayName`, `description`, `labels` | no | Standard. |

## Spec

### Taxonomy

| Field | Type | Description |
|---|---|---|
| `family` | string | Model family label (`claude`, `gpt`, `llama`, …). |
| `version` | string | Free-form version string. |

### Capabilities

`capabilities` is an object. Absent boolean fields are treated as `false`. Set only the ones the model actually supports.

`chat`, `embeddings`, `streaming`, `tools`, `parallelTools`, `vision`, `audio`, `promptCache`, `reasoning`, `jsonMode`, `structuredOutputs`, `batch`, `computerUse`, `webSearch`, `fileInput`, `audioInput`, `audioOutput`, `systemMessages`, `assistantPrefill`.

`unsupportedParams` is a string list of canonical sampling parameter names (`temperature`, `top_p`, `top_k`) the upstream rejects for this model. The pipeline strips listed params before forwarding the request.

Two non-boolean fields describe effort control on reasoning models (both require `reasoning: true`). `reasoningEfforts` lists the levels the model accepts, ordered low to high, from `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max`; omit it when the model has no effort control. `defaultReasoningEffort` is the level the vendor applies when the request sets none, and must be one of the listed levels.

### Modalities

| Field | Type | Description |
|---|---|---|
| `modalities.input` | `[]string` | Media types accepted (`text`, `image`, `audio`, `pdf`). |
| `modalities.output` | `[]string` | Media types produced. |

### Context window

| Field | Type | Description |
|---|---|---|
| `contextWindowTotal` | int | Canonical total context size in tokens. |
| `contextWindowInput` | int | Soft input cap if different from total. |
| `contextWindowOutput` | int | Soft output cap if different from total. |
| `maxOutputTokens` | int | Hard cap on tokens produced in a single response. |

### Snapshots and pointer

A model may expose multiple dated snapshots (e.g. `claude-sonnet-5-5` plus `claude-sonnet-5-5-20261001`). Each entry in `snapshots[]` has:

| Field | Required | Description |
|---|---|---|
| `name` | yes | Snapshot slug — the upstream sends this name verbatim. |
| `releasedAt` | no | ISO date the snapshot was released. |
| `originalName` | no | Override for the upstream wire name when it differs from `name`. |

`pointer` names the snapshot the bare model handle resolves to. It must match one of the `snapshots[].name` values.

### Lifecycle

| Field | Type | Description |
|---|---|---|
| `releaseDate` | string | ISO date. |
| `knowledgeCutoff` | string | Training data cutoff (ISO year-month or date). |
| `deprecation.status` | enum | `active` \| `deprecated` \| `sunset`. |
| `deprecation.sunsetDate` | string | When the upstream removes access. |
| `deprecation.replacement` | string | Name of the replacement Model. |
| `deprecationDate` | string | Legacy flat field; prefer `deprecation`. |

### Discovery

| Field | Type | Description |
|---|---|---|
| `aliases` | `[]string` | Resolution-only alternate names. A matched alias routes to this model but the alias string is sent upstream verbatim. Must be unique across all models and not equal `metadata.name`. |
| `tags` | `[]string` | Free-form. |
| `documentation` | string | Markdown or URL. |
| `license` | string | License identifier. |
| `providerModelPageURL` | string | Vendor's product page. |
| `enabled` | bool | Defaults to true. |

## Example

```yaml
# yaml-language-server: $schema=https://relay-api.wyolet.dev/schemas/v1alpha2/Model.schema.json
apiVersion: relay.wyolet.dev/v1alpha2
kind: Model
metadata:
  name: claude-sonnet-5-5
  displayName: Claude Sonnet 5.5
  owner:
    kind: provider
    name: anthropic
spec:
  family: claude
  capabilities:
    chat: true
    streaming: true
    tools: true
    vision: true
    promptCache: true
    reasoning: true
    structuredOutputs: true
    systemMessages: true
    reasoningEfforts: [low, medium, high, xhigh, max]
    defaultReasoningEffort: high
  modalities:
    input: [text, image, pdf]
    output: [text]
  contextWindowTotal: 1000000
  maxOutputTokens: 128000
  releaseDate: "2026-09-28"
  knowledgeCutoff: "2026-06"
  snapshots:
    - name: claude-sonnet-5-5
      releasedAt: "2026-09-28"
  pointer: claude-sonnet-5-5
  aliases:
    - claude-sonnet-5-5[1m]
```

## HostBinding

A **HostBinding** declares that a Model is reachable through a specific Host via a specific adapter. HostBindings live alongside the model they bind, under `data/hosts/<host>/bindings/<model>.yaml`.

```yaml
# yaml-language-server: $schema=https://relay-api.wyolet.dev/schemas/v1alpha2/HostBinding.schema.json
apiVersion: relay.wyolet.dev/v1alpha2
kind: HostBinding
metadata:
  name: claude-sonnet-5-5-on-anthropic
  displayName: "Claude Sonnet 5.5 via anthropic"
spec:
  model: claude-sonnet-5-5
  host: anthropic
  adapter: anthropic
  pricing: anthropic-claude-sonnet-5-5
```

HostBinding spec fields:

| Field | Required | Description |
|---|---|---|
| `model` | yes | `metadata.name` of the [Model](model.md). |
| `host` | yes | `metadata.name` of the [Host](host.md). |
| `adapter` | yes | Wire protocol: `openai`, `anthropic`, or `gemini`. |
| `upstreamName` | no | Override for the model name sent upstream; defaults to the snapshot's `Upstream()` value. |
| `pricing` | no | `metadata.name` of the [Pricing](pricing.md) to apply. |
| `snapshots` | no | Subset of model snapshots this binding covers. Empty means all. |
| `enabled` | no | Defaults to true. |

## Relationships

- `metadata.owner.name` → [Provider](provider.md) by name.
- Hosting is declared via [HostBinding](#hostbinding) files, not on the Model itself.
- Pricing references Models via `spec.targetModels[]` (inverse — see [Pricing](pricing.md)).
