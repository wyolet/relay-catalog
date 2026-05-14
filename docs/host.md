# Host

An **upstream API endpoint** relay forwards requests to. One Host per API root — `api.openai.com`, `api.anthropic.com`, a self-hosted Ollama, etc. A single vendor can have multiple Hosts (e.g. OpenAI direct vs. Azure OpenAI), and a single Model can be served by multiple Hosts.

Each host lives in its own folder: `data/hosts/<name>/host.yaml`. The folder also contains the host's `pricing/` and `policies/` subdirectories — everything the host owns.

> A host can share a name with a [Provider](provider.md) (e.g. `openai` is both). They're separate entities — the host represents the *endpoint*, the provider represents the *brand*.

## Metadata

| Field | Required | Description |
|---|---|---|
| `name` | yes | DNS-1123 slug. Referenced by Model (`spec.hosts[].host`) and Pricing (`metadata.owner.id`). |
| `displayName` | no | Human-readable label shown in UIs. |
| `description` | no | Free text. |
| `owner` | no | Omit entirely for catalog-shipped (system-owned) Hosts. Relay defaults absent owners to system ownership. |
| `labels` | no | Arbitrary `map[string]string` for selectors. |

## Spec

| Field | Required | Type | Description |
|---|---|---|---|
| `baseURL` | **yes** | string | Upstream API root, e.g. `https://api.openai.com`. No trailing slash. |
| `backend` | no | `map[string]string` | Opaque backend hints passed through to relay's transport layer. |
| `enabled` | no | bool | Defaults to true. |
| `homepageURL` | no | string | Vendor homepage. |
| `docsURL` | no | string | Developer documentation. |
| `consoleURL` | no | string | Web dashboard / console. |
| `statusPageURL` | no | string | Incident/status page. |
| `icon.path` | no | string | Relative path to logo asset under the frontend's `public/`. |
| `policies` | no | `[]string` | Names of [Policies](policy.md) this host publishes. Each must be enabled and owned by this host. |
| `defaultPolicy` | no | string | Fallback policy name. Must appear in `policies`. |

## Example

```yaml
apiVersion: relay.wyolet.dev/v1
kind: Host
metadata:
  name: anthropic
  displayName: Anthropic
spec:
  baseURL: https://api.anthropic.com
  homepageURL: https://www.anthropic.com
  docsURL: https://docs.anthropic.com
  consoleURL: https://console.anthropic.com
  statusPageURL: https://status.anthropic.com
  icon:
    path: /provider/anthropic.svg
```

## Relationships

- Referenced by **Model** via `spec.hosts[].host` (Models declare which Hosts can serve them).
- Referenced by **Pricing** via `metadata.owner.id` (Pricing always attaches to the Host that bills it — different Hosts can charge different rates for the same Model).
- References **[Policy](policy.md)** by name via `spec.policies[]` and `spec.defaultPolicy`.
