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
    bindings/<model>.yaml         # HostBinding: which adapter routes <model> via this host
    pricing/<model>.yaml          # pricing for <model> when served via this host
    policies/<policy>.yaml        # tier policy + its rate limit (two docs per file)
```

> A name can appear in **both** trees. `openai` exists as a Provider (the brand) and as a Host (`api.openai.com`). `anthropic` likewise. That's intentional — they're different concepts that happen to share a slug. The Host's `displayName` can disambiguate (e.g. "OpenAI Direct") if it ever matters in UI.

## Shape of every document

```yaml
apiVersion: relay.wyolet.dev/v1alpha2
kind: <Host | Provider | Model | Pricing | Policy | RateLimit | HostBinding>
metadata:
  name: <stable-slug>          # required, DNS-1123, unique within its kind
  displayName: <human label>   # optional
  description: <free text>     # optional
  owner:                       # omit entirely for catalog-shipped (system-owned) entities
    kind: <provider|host|user> # required when present
    name: <name of owner>
  labels: { key: value, ... }  # optional
spec:
  ...                          # kind-specific
```

## Conventions

- **`metadata.name` is the public slug.** Cross-references between documents use names, not UUIDs. Relay's seeder resolves names to internal IDs on load. Renaming a published name is breaking — use aliases on Model, or new entries elsewhere.
- **Filenames mirror the slug** (without redundant prefixes — the directory already says which host). E.g. `data/hosts/anthropic/pricing/claude-opus-4-5-20251101.yaml`, not `anthropic-claude-opus-4-5-20251101.yaml`.
- **`enabled` defaults to true.** Set `enabled: false` to disable without deleting.
- **Icons ship with the catalog.** `icon.path: /provider/<file>` names `data/icons/<file>`. See [Icons](#icons).
- **One entity per file**, except for Policy + its RateLimit which are paired in one file (two YAML documents separated by `---`).

## Kinds

- [Host](host.md) — an upstream API endpoint relay talks to.
- [Provider](provider.md) — the vendor brand that publishes models.
- [Model](model.md) — a callable model and its capabilities.
- [HostBinding](host.md#hostbinding) — routes one Model through one Host via a named adapter.
- [Pricing](pricing.md) — billing rates per model per host.
- [Policy & RateLimit](policy.md) — usage tiers a host publishes.

## Icons

Logos are files in `data/icons/`, flat, named after the slug they mostly serve (`anthropic.png`, `z-ai.png`). `icon.path` is the URL path `/provider/<file>`, which maps to `data/icons/<file>`. Several entities may share one file.

A Host and a Provider each carry their own `spec.icon.path`. A model's logo is the first one set of:

1. the Model's `spec.icon.path`;
2. its family's `icon.path` in [`families.yaml`](../data/families.yaml);
3. its Provider's `spec.icon.path`.

A family gets an icon only when the model brand has its own mark, distinct from the publisher's: `claude` (Anthropic), `gemini` (Google), `grok` (xAI), `kimi` (Moonshot AI), `qwen` (Alibaba). It applies to every model of the family, whoever publishes or hosts it. A Model sets its own only to differ from both.

- Format: PNG only, 256×256, at most 64 KiB. Run large or gradient-heavy files through `pngquant`.
- Background, so the logo reads on both light and dark UI:
  - transparent by default, with the mark trimmed and centred at about 80% of the canvas;
  - a white square behind a mark that is mostly black ink (OpenAI, Ollama);
  - a black square behind a mark that is white ink only;
  - a logo whose brand is a filled tile (Anthropic, Groq) keeps its tile, filling the canvas.
- Source: the vendor's own brand page, site icons (SVG favicon, apple-touch-icon, web manifest), or their GitHub or Hugging Face org avatar. Rasterize vectors at 1024 px before scaling down. No third-party icon packs.

Icons ship in the release tarball with the manifests; the YAML loader skips non-YAML files. Files under `data/drafts/` are not shipped, so an icon a draft needs still goes in `data/icons/`.

## How relay consumes this

The relay seeder walks the tree, parses each YAML (dispatching on `kind`), resolves name references to UUIDs, and upserts into Postgres. The loader lives in `app/manifest/` in the relay repo; wire DTOs are in `app/manifest/dto.go`. That file is the source of truth when the docs disagree.

## Entity graph

```
Provider ◄── owner.name ── Model
                              ▲
                              │ spec.model
                           HostBinding ── spec.host ──► Host ◄─ owner ── Pricing ── targetModels[] ──► Model
                                                          ▲
                                                          ├─ spec.policies[] ──► Policy ── spec.rateLimit ──► RateLimit
                                                          │                                                      ▲
                                                          └─────────────────────── owner ───────────────────────┘
```
