# Pricing

**Billing rates** for one or more [Models](model.md) when served through a specific [Host](host.md). Pricing is attached to the Host because the same Model can cost different amounts depending on where it's served (Anthropic direct vs. Bedrock vs. Vertex).

Files live under `data/hosts/<host>/pricing/<model-name>.yaml`. The owning host is the parent directory — no host prefix in the filename.

## Metadata

| Field | Required | Description |
|---|---|---|
| `name` | yes | Conventionally `<host>-<model-name>` (the file path encodes the host, the slug retains it for stable cross-refs). |
| `owner.kind` | yes | `host`. |
| `owner.id` | yes | `metadata.name` of the [Host](host.md) charging these rates. |
| `displayName`, `description`, `labels` | no | Standard. |

## Spec

| Field | Required | Type | Description |
|---|---|---|---|
| `currency` | **yes** | string | ISO-4217 code. In practice always `USD`. |
| `targetModels` | **yes** | `[]string` | Model `metadata.name` values this pricing applies to. One Pricing doc can cover multiple aliases/dated versions of the same model. |
| `rates` | **yes** | `[]Rate` | One entry per billable meter. |
| `enabled` | no | bool | Defaults to true. |

Each `rates[]` entry:

| Field | Required | Type | Description |
|---|---|---|---|
| `meter` | **yes** | string | Billing dimension. One of `tokens.input`, `tokens.output`, `tokens.cache_read`, `tokens.cache_creation`, `tokens.reasoning`, `tokens.audio_input`, `tokens.audio_output`, `tokens.accepted_prediction`, `tokens.rejected_prediction`, `tokens.server_tool_use_input`, `tokens.server_tool_use_output`. A meter for a part of another (`tokens.reasoning` inside output, `tokens.audio_input` inside input) charges that part at its own rate; without one the part bills inside its whole. |
| `unit` | **yes** | string | `per_million` (tokens) or `per_unit`. |
| `amount` | **yes** | float | Cost in `currency` per `unit`. |
| `aboveTokens` | no | int | Context-length tier: the rate applies when the request's prompt length (input + cache reads + cache writes) reaches this many tokens, and then bills the whole meter. Omit for flat pricing. |

## Example

```yaml
apiVersion: relay.wyolet.dev/v1alpha2
kind: Pricing
metadata:
  name: anthropic-claude-3-haiku-20240307
  owner:
    kind: host
    id: anthropic
spec:
  currency: USD
  targetModels:
    - claude-3-haiku-20240307
  rates:
    - meter: tokens.input
      unit: per_million
      amount: 0.25
    - meter: tokens.output
      unit: per_million
      amount: 1.25
    - meter: tokens.cache_creation
      unit: per_million
      amount: 0.30
    - meter: tokens.cache_read
      unit: per_million
      amount: 0.03
```

Tiered example (different rate above a token threshold):

```yaml
spec:
  currency: USD
  targetModels: [some-long-context-model]
  rates:
    - meter: tokens.input
      unit: per_million
      amount: 3.00
    - meter: tokens.input
      unit: per_million
      amount: 6.00
      aboveTokens: 200000
```

## Relationships

- `metadata.owner.id` → [Host](host.md) by name.
- `spec.targetModels[]` → [Model](model.md) by name (one-to-many).

## Editing pricing

When a vendor changes prices, the cleanest pattern is to add a new Pricing file with the new rates rather than mutating in place. Relay attributes historical usage to whichever Pricing was active at the time; rewriting old rates retroactively rewrites the bill. Only edit an existing file when correcting a genuine data-entry error.
