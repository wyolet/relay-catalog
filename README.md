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
