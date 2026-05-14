# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

`relay-catalog` is **not a code project** — it is curated configuration data for the downstream **relay** project. YAML files describe providers, their models, the hosts (API endpoints) those models are served from, the pricing each host charges, and the tier policies each host publishes.

Treat every change as a data edit: correctness, schema-conformance, and reviewable diffs matter more than cleverness.

## Layout

The repo has exactly two top-level data concepts. Everything else nests under one of them based on ownership:

```
data/
  providers/<provider>/
    provider.yaml          # the vendor brand
    models/<model>.yaml    # models published by this provider
  hosts/<host>/
    host.yaml              # the API endpoint
    pricing/<model>.yaml   # billing rates for that model on this host
    policies/<policy>.yaml # tier policy + its RateLimit (two docs per file)
```

Ownership rules behind the layout:
- A **Model** is owned by a Provider (the brand that publishes it) → lives under `providers/`.
- **Pricing** and **Policies** are owned by a Host (the endpoint that bills and enforces them) → live under `hosts/`. The same Model can have different pricing/policies on different Hosts (Anthropic direct vs. Bedrock, etc.).

A name can appear in **both** trees (`openai`, `anthropic` are both a Provider and a Host). That's expected; they're separate concepts that happen to share a slug.

## Docs

[`docs/`](docs/) describes every kind: Host, Provider, Model, Pricing, Policy, RateLimit. **Read the relevant doc before editing or adding fields** — they document every field, allowed values, and cross-entity relationships. Source of truth for wire schema is `app/manifest/dto.go` in the relay repo.

## What is NOT in this repo

- `users/` — user accounts and credentials. Lives in relay.
- `ratelimits/system.yaml` — system-wide runtime policy. Lives in relay.

Don't move user-specific or per-deployment runtime config here.

## Conventions

- All files use `apiVersion: relay.wyolet.dev/v1` with `kind`, `metadata`, `spec`.
- `metadata.name` is the stable slug. Once published, do not rename — relay references it. Add aliases on Model, or new entries elsewhere.
- One entity per file, **except** Policy + RateLimit which are paired in one file as two YAML docs.
- **Filenames mirror the slug minus the host prefix** — the directory already says which host. `data/hosts/anthropic/pricing/claude-opus-4-5-20251101.yaml`, not `anthropic-claude-opus-4-5-20251101.yaml`. The `metadata.name` inside still carries the full slug for cross-refs.
- Logos referenced by path (e.g. `/provider/anthropic.svg`); the image lives in the relay frontend's `public/` folder.
- Deprecation: don't delete entries. Mark deprecated, point at a replacement where applicable.
- Pricing edits: append a new file or new rate entry rather than mutating historical pricing in place — relay attributes historical usage to the rate that was active at the time. Only edit existing files to correct genuine data-entry errors.

## When adding a new model

1. Read [`docs/model.md`](docs/model.md).
2. Add the model file: `data/providers/<provider>/models/<model-id>.yaml`.
3. For each host that serves it, add a pricing file: `data/hosts/<host>/pricing/<model-id>.yaml`.
4. If it's a new provider, add `data/providers/<provider>/provider.yaml` ([docs](docs/provider.md)).
5. If it's a new host, add `data/hosts/<host>/host.yaml` ([docs](docs/host.md)) plus any policies under `data/hosts/<host>/policies/` ([docs](docs/policy.md)).
6. Use a sibling file as a template — keep field ordering consistent for review.
