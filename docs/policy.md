# Policy & RateLimit

These two kinds live together under each host: `data/hosts/<host>/policies/<policy>.yaml`. A `RateLimit` and its consuming `Policy` are defined in the same file as separate YAML documents (separated by `---`).

A **RateLimit** is a budget: how many requests/tokens are allowed per window, and how that budget is enforced. A **Policy** is what a Host actually serves to callers — it bundles a RateLimit with model access, host-key selection, and other knobs. A [Host](host.md) publishes a list of Policies (`spec.policies`) and picks one as `defaultPolicy`.

## RateLimit

`kind: RateLimit`. Files live under `data/policies/`.

### Metadata

Standard. `owner.kind: host`, `owner.id: <host-name>`.

### Spec

| Field | Required | Type | Description |
|---|---|---|---|
| `rules` | **yes** | `[]Rule` | One or more rules evaluated together on each request. |
| `enabled` | no | bool | Defaults to true. |

### Rule

| Field | Required | Type | Description |
|---|---|---|---|
| `meter` | **yes** | string | What's being counted. See meters below. |
| `amount` | **yes** | int | Budget. |
| `window` | **yes** | duration | Go duration string (`30s`, `1m`, `24h`) or nanoseconds. For `concurrency` meter this only sets the key TTL. |
| `strategy` | **yes** | string | Enforcement algorithm. See strategies below. |

**Meters:** `requests`, `tokens`, `tokens.input`, `tokens.output`, `tokens.cache_read`, `tokens.cache_creation`, `tokens.reasoning`, `tokens.server_tool_use_input`, `tokens.server_tool_use_output`, `concurrency`.

**Strategies:**
- `token-bucket` — continuous refill at `amount`/`window`.
- `sliding-window` — two-bucket weighted approximation.
- `fixed-window` — bucket resets at `floor(now/window)`.
- `leaky-bucket` — constant drain rate.
- `session-window` — window anchored to the first request after reset.

## Policy

`kind: Policy`. Files live under `data/policies/`.

### Metadata

Standard. `owner.kind: host`, `owner.id: <host-name>`. A Policy is referenced by Host `spec.policies[]` and `spec.defaultPolicy` *by name*.

### Spec

| Field | Required | Type | Description |
|---|---|---|---|
| `rateLimit` | no | string | Name of the **single** [RateLimit](#ratelimit) this policy enforces. A Policy can reference at most one. |
| `models` | no | `[]string` | Allow-list of Model names. Empty/absent = all models the host serves. |
| `hostKeys` | no | `[]string` | Host-key names the policy may use for upstream auth. |
| `keySelection` | no | string | Strategy for picking a host-key. `prioritized` (default), `round-robin`, `least-recently-used`. |
| `skipDefaultLimits` | no | bool | Bypass host/system default rate limits for this policy. |
| `includeDeprecated` | no | bool | Allow deprecated models through this policy. |
| `enabled` | no | bool | Defaults to true. |

## Example (combined file)

`data/hosts/anthropic/policies/tier-1.yaml`:

```yaml
apiVersion: relay.wyolet.dev/v1
kind: RateLimit
metadata:
  name: anthropic-tier-1-rl
  displayName: Anthropic Tier 1 rate limits
  owner:
    kind: host
    name: anthropic
spec:
  enabled: true
  rules:
    - meter: requests
      amount: 50
      window: 1m
      strategy: sliding-window
    - meter: tokens
      amount: 30000
      window: 1m
      strategy: sliding-window
---
apiVersion: relay.wyolet.dev/v1
kind: Policy
metadata:
  name: anthropic-tier-1
  displayName: Anthropic Tier 1
  owner:
    kind: host
    name: anthropic
spec:
  rateLimit: anthropic-tier-1-rl
```

The corresponding Host then lists `anthropic-tier-1` (and its siblings) under `spec.policies`:

```yaml
# data/hosts/anthropic/host.yaml — excerpt
spec:
  policies:
    - anthropic-tier-1
    - anthropic-tier-2
    - anthropic-tier-3
    - anthropic-tier-4
  defaultPolicy: anthropic-tier-1
```

## Validation rules enforced by relay

- Every name in a Host's `spec.policies` must resolve to an **enabled, host-owned** Policy whose `owner.id` matches that Host.
- `spec.defaultPolicy` must appear in `spec.policies`.
- A Policy's `rateLimit` name must resolve to a RateLimit owned by the same Host.

## Relationships

- Policy → RateLimit (`spec.rateLimit`, at most one).
- Policy → [Model](model.md) (`spec.models[]`, optional allow-list).
- Policy ← [Host](host.md) (`Host.spec.policies[]`, `Host.spec.defaultPolicy`).
- RateLimit and Policy both → [Host](host.md) via `metadata.owner`.
