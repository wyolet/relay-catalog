# Catalog data model

The catalog has exactly two top-level concepts:

- **Providers** — the vendors/brands that publish models. Models live under their provider.
- **Hosts** — the API endpoints relay forwards to. Pricing and Policies live under their host because the host owns them: it bills the request and serves the tier.

```
data/
  providers/<provider>/
    provider.yaml
    models/<model>.yaml
  hosts/<host>/
    host.yaml
    pricing/<model>.yaml          # pricing for <model> when served via this host
    policies/<policy>.yaml        # tier policy + its rate limit (two docs per file)
```

> A name can appear in **both** trees. `openai` exists as a Provider (the brand) and as a Host (`api.openai.com`). `anthropic` likewise. That's intentional — they're different concepts that happen to share a slug. The Host's `displayName` can disambiguate (e.g. "OpenAI Direct") if it ever matters in UI.

## Shape of every document

```yaml
apiVersion: relay.wyolet.dev/v1
kind: <Host | Provider | Model | Pricing | Policy | RateLimit>
metadata:
  name: <stable-slug>          # required, DNS-1123, unique within its kind
  displayName: <human label>   # optional
  description: <free text>     # optional
  owner:                       # omit entirely for catalog-shipped (system-owned) entities
    kind: <provider|host|user> # required when present
    id:   <name of owner>
  labels: { key: value, ... }  # optional
spec:
  ...                          # kind-specific
```

## Conventions

- **`metadata.name` is the public slug.** Cross-references between documents use names, not UUIDs. Relay's seeder resolves names to internal IDs on load. Renaming a published name is breaking — use aliases on Model, or new entries elsewhere.
- **Filenames mirror the slug** (without redundant prefixes — the directory already says which host). E.g. `data/hosts/anthropic/pricing/claude-opus-4-5-20251101.yaml`, not `anthropic-claude-opus-4-5-20251101.yaml`.
- **`enabled` defaults to true.** Set `enabled: false` to disable without deleting.
- **Icons reference frontend assets by path** (e.g. `/provider/anthropic.svg`). The SVGs live in the relay frontend's `public/`, not here.
- **One entity per file**, except for Policy + its RateLimit which are paired in one file (two YAML documents separated by `---`).

## Kinds

- [Host](host.md) — an upstream API endpoint relay talks to.
- [Provider](provider.md) — the vendor brand that publishes models.
- [Model](model.md) — a callable model and how it's served.
- [Pricing](pricing.md) — billing rates per model per host.
- [Policy & RateLimit](policy.md) — usage tiers a host publishes.

## How relay consumes this

`relay seed --from <dir>` walks the tree, parses each YAML, dispatches on `kind`, resolves name references to UUIDs, and upserts into Postgres. The loader lives in `app/manifest/` in the relay repo; wire DTOs are in `app/manifest/dto.go`. That file is the source of truth when the docs disagree.

## Entity graph

```
Provider ◄── owner.id ── Model ── spec.hosts[].host ──► Host ◄─ owner ── Pricing ── targetModels[] ──► Model
                                                          ▲
                                                          ├─ spec.policies[] ──► Policy ── spec.rateLimit ──► RateLimit
                                                          │                       │                            ▲
                                                          │                       └─ spec.models[] ─► Model    │
                                                          └─────────────────────── owner ──────────────────────┘
```
