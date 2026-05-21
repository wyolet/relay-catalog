# relay-catalog

Curated catalog data for the **relay** project. Two top-level concepts: **Providers** (vendors that publish models) and **Hosts** (API endpoints relay talks to). Everything else lives under one of these.

Relay consumes these YAML files via `relay seed --from <dir>`. User-specific and system-runtime config (users, default rate limits) stays in the relay repo — it is not curated here.

## Layout

```
data/
  providers/<provider>/
    provider.yaml                       # the vendor brand
    models/<model>.yaml                 # models the provider publishes
  hosts/<host>/
    host.yaml                           # the API endpoint
    pricing/<model>.yaml                # what this host charges for that model
    policies/<policy>.yaml              # tier policies + their rate limits
```

A name can appear in both trees (`openai` is both a Provider and a Host). They're independent entities; edit them separately.

## Docs

Start with [`docs/README.md`](docs/README.md), then drill into:

- [Host](docs/host.md)
- [Provider](docs/provider.md)
- [Model](docs/model.md)
- [Pricing](docs/pricing.md)
- [Policy & RateLimit](docs/policy.md)

Logo assets currently live in the relay frontend's `public/` folder and are referenced from YAML by path (e.g. `/provider/anthropic.svg`).

## Validation

Three layers, all enforced in CI:

1. **Structural** — JSON Schema per kind, served at `https://relay-api.wyolet.dev/schemas/v1alpha2/<Kind>.schema.json`. Every YAML file carries a `# yaml-language-server: $schema=...` directive at the top so editors (VS Code Red Hat YAML extension, neovim with yaml-language-server, IntelliJ) give live autocomplete + inline diagnostics.

2. **Per-entity** — relay's `Validate()` methods enforce intra-entity invariants (e.g. "host-owned policy must not list hostKeys"). Runs implicitly during manifest parsing.

3. **Graph + curation** — `go run ./cmd/validate ./data` runs:
   - relay's schema-generic graph linter (`catalogvalidate.ValidateGraph`) — cross-refs, owner mismatches, duplicate names, orphan warnings
   - this repo's curation rules under `cmd/validate/rules/` — content conventions specific to wyolet's catalog (icon presence, pricing-target-host-binding, tier-policy completeness, ...)

```bash
go run ./cmd/validate ./data           # default: warn on hints, fail on errors
go run ./cmd/validate --strict ./data  # promote warnings to errors (release prep)
go run ./cmd/validate --list           # list every rule with description + severity
go run ./cmd/validate --skip <name>... # suppress specific rules
```

CI runs the validator on every push + PR. Failure surfaces the exact issue with kind / source field / target.

### Adding a curation rule

```
cmd/validate/rules/<rule_name>.go
```

Each rule lives in its own file: an `init()` that appends a `catalogvalidate.Rule` to `All`, plus a pure `Check func([]manifest.Document) []Issue`. Unit-test in `<rule_name>_test.go`. See existing rules (`provider_icon.go`, `pricing_target.go`, `host_tier_policy.go`) as templates.

### Mixed-kind files

The tier-policy files (`data/hosts/<host>/policies/tier-N.yaml`) bundle a `RateLimit` and a `Policy` document in one file via `---`. yaml-language-server can't switch schemas per-document, so these files have no IDE directive — they still validate fully at CI time via the Go validator.
