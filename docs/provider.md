# Provider

The **vendor brand** that publishes a family of models — Anthropic, OpenAI, Mistral, etc. A Provider is a grouping/branding entity; it does not carry an API URL (that lives on [Host](host.md)) or pricing (that lives on [Pricing](pricing.md)).

Files live under `data/providers/<name>/provider.yaml`. The provider's models live as sibling files under `data/providers/<name>/models/`.

> A provider can share a name with a [Host](host.md) (e.g. `openai` is both). They're separate entities — the provider is the brand, the host is the endpoint. Edit them independently.

## Metadata

| Field | Required | Description |
|---|---|---|
| `name` | yes | DNS-1123 slug. Referenced by Model via `metadata.owner.id`. |
| `displayName` | no | Human-readable label. |
| `description` | no | Free text. |
| `owner` | no | Omit for catalog-shipped (system-owned) Providers. |
| `labels` | no | Arbitrary map. |

## Spec

All fields optional — Provider exists primarily as a referenceable owner for Models.

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Defaults to true. |
| `homepageURL` | string | Vendor homepage. |
| `docsURL` | string | Developer documentation root. |
| `statusPageURL` | string | Status page. |
| `icon.path` | string | Relative path to logo asset in the frontend's `public/`. |

Note: Provider has **no** `baseURL`, `consoleURL`, or `backend` — those are Host concerns.

## Example

```yaml
apiVersion: relay.wyolet.dev/v1
kind: Provider
metadata:
  name: anthropic
  displayName: Anthropic
spec:
  homepageURL: https://www.anthropic.com
  docsURL: https://docs.anthropic.com
  statusPageURL: https://status.anthropic.com
  icon:
    path: /provider/anthropic.svg
```

## Relationships

- Referenced by **Model** via `metadata.owner.id` (every Model belongs to exactly one Provider).
- Provider references nothing else.

## Provider vs. Host

These are deliberately separate concepts:

- **Provider** = who built the model (the brand).
- **Host** = where relay sends the HTTP request (the endpoint).

Anthropic models, for example, can be served by the `anthropic` Host (direct API) and by an `amazon-bedrock` Host (same models, different endpoint, different pricing). The Provider is `anthropic` in both cases; the Host differs.
