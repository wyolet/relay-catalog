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
| `family` | string | The base model lineage the model ships under (`claude`, `gpt`, `llama`, …), from [`families.yaml`](../families.yaml). No version, size or product line: `gemini-3-flash` is `gemini`, `qwen2.5-coder` is `qwen`. A fine-tune takes its base model's family (a Llama fine-tune is `llama`), whoever publishes it. |
| `productLine` | string | Optional. The vendor's named tier or variant inside the family (`opus`, `flash-lite`, `coder`), from that family's `lines` in [`families.yaml`](../families.yaml). Leave it out when the model has none. See [Family and product line](#family-and-product-line). |
| `version` | string | Free-form version string. |

#### Family and product line

| | `family` | `productLine` |
|---|---|---|
| Names | the base model lineage | the vendor's named tier or variant inside that family |
| Allowed values | `families.yaml` `families[].name` | the family's `lines[].name`; empty is always allowed |
| Never contains | a version, size or line | a version or size |
| Fine-tunes | the base model's family | a vendor fine-tuning target keeps its base's line (`ft-o4-mini` → `o`); a third-party fine-tune takes its publisher's brand, or stays empty under a descriptive name |

Two tests settle the unclear cases:

- **A suffix that composes is a modifier, not a line.** A model sits in exactly one line. `mini` attaches to `gpt-4o-mini`, `o4-mini` and `phi4-mini`, so it is a size and never a `productLine`. A word can still be a line in another family: Gemini `pro` never combines with `flash`, so it is one.
- **The slug is the vendor's name, lowercased.** Drop the family word when it stands apart ("Claude Opus" → `opus`, "Gemini Flash-Lite" → `flash-lite`, "Mistral Large" → `large`); keep the whole word when the vendor fuses it (`codegemma`, `codellama`, `devstral`, `qwq`, `codegeex`).

| Model | `family` | `productLine` | Why |
|---|---|---|---|
| `claude-opus-4-8` | `claude` | `opus` | named tier |
| `gemini-3-1-flash-lite` | `gemini` | `flash-lite` | named tier |
| `ministral-3-8b` | `mistral` | `ministral` | fused vendor name; `8b` is a size |
| `qwen3-coder-plus` | `qwen` | `coder` | the variant wins over the tier |
| `o4-mini` | `gpt` | `o` | `mini` is a size |
| `hermes3-8b` | `llama` | `hermes` | fine-tune, publisher's brand |
| `gpt-4o-mini` | `gpt` | (empty) | `4o` is a version, `mini` a size |
| `llama2-uncensored-7b` | `llama` | (empty) | descriptive fine-tune name, not a brand |

The validator rejects a `productLine` that is not one of its family's lines, so `opus` under `gpt` fails. Line `aliases` in `families.yaml` only map importer family names onto a line and are not valid values.

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
  productLine: sonnet
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
