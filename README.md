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

## Verified-only ships; everything else lives in `drafts/`

The live `data/` tree is **verified-routable only**: an entry ships there once someone has held a real key for that host and confirmed a binding actually routes end-to-end. The published catalog's promise is "every model in here works."

`data/drafts/` is the staging area for entries we **can't confirm yet** — typically a host with no free tier (Bedrock, Vertex, Azure, and most of the imported list), so we can author the YAML but can't test it. `manifest.LoadDir` skips `drafts/` entirely, so nothing there reaches a running relay.

Promoting a host out of drafts:

1. Get a key for the host.
2. Confirm a binding routes (a real request returns through relay).
3. `git mv data/drafts/hosts/<host> data/hosts/<host>` (and any provider models it needs).
4. Validate clean (see [Validation](#validation)), then PR.

Demoting works the reverse way — if an entry can't be verified, it belongs in `drafts/`, not live. Don't ship a host that routes nothing.

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

3. **Graph + curation** — relay's `catalog-validate --curation ./data` runs:
   - relay's schema-generic graph linter (`catalogvalidate.ValidateGraph`) — cross-refs, owner mismatches, duplicate names, orphan warnings
   - the curation rules — content conventions for a published catalog (icon presence, pricing-target-host-binding, tier-policy completeness, ...)
   - every Model's `spec.family` is listed in `families.yaml`; a new lineage is added there first
   - a set `spec.productLine` is one of its family's `lines` in `families.yaml`
   - `featured.yaml` matches the Models' `featured` labels

The tools run from relay's `catalog-server` image, pinned in [`.github/catalog-tools.env`](.github/catalog-tools.env). From the repo root:

```bash
source .github/catalog-tools.env
docker run --rm -v "$PWD:/src" -w /src "$CATALOG_TOOLS_IMAGE" catalog-validate --curation ./data                    # warn on hints, fail on errors
docker run --rm -v "$PWD:/src" -w /src "$CATALOG_TOOLS_IMAGE" catalog-validate --strict --curation ./data           # promote warnings to errors (release prep)
docker run --rm "$CATALOG_TOOLS_IMAGE" catalog-validate --list                                                      # list every rule with description + severity
docker run --rm -v "$PWD:/src" -w /src "$CATALOG_TOOLS_IMAGE" catalog-validate --curation --skip <name> ./data      # suppress a rule (repeatable)
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD:/src" -w /src "$CATALOG_TOOLS_IMAGE" catalog-validate --write-featured ./data  # regenerate featured.yaml
```

The image is linux/amd64 only: on arm64 hosts add `--platform linux/amd64`, and `-e GOGC=off` if the tool crashes under qemu emulation. From a relay checkout next to this one, the same tool runs as `go run ./cmd/catalog-validate --curation ../relay-catalog/data`.

CI runs the validator on every push + PR. Failure surfaces the exact issue with kind / source field / target.

### Adding a curation rule

Curation rules live in relay, in `app/catalogvalidate/curation`: one file per rule defining a `catalogvalidate.Rule`, listed in `Rules` in `curation.go`, with tests beside it. A new rule reaches this repo's CI when [`.github/catalog-tools.env`](.github/catalog-tools.env) is bumped to a relay release that has it.

### Mixed-kind files

The tier-policy files (`data/hosts/<host>/policies/<policy>.yaml`) bundle a `RateLimit` and a `Policy` document in one file via `---`. yaml-language-server can't switch schemas per-document, so these files have no IDE directive — they still validate fully in CI via `catalog-validate`.

## Release asset

Every `v*` tag attaches the flattened SDK catalog to that tag's GitHub release:

- `https://github.com/wyolet/relay-catalog/releases/download/<tag>/catalog.json.gz`
- `https://github.com/wyolet/relay-catalog/releases/download/<tag>/catalog.json.gz.sha256` — `<hex>  catalog.json.gz`, checkable with `sha256sum -c`.

To fetch the newest catalog, read `channels.<channel>.latest` from `https://raw.githubusercontent.com/wyolet/relay-catalog/main/index.yaml` for the schema channel your SDK reads (e.g. `v1alpha2`) and use that tag in the URL above. Avoid `/releases/latest/download/`: it ignores schema channels. The release workflow updates `index.yaml` only after the tag's assets are published.

Verify where an asset was built with `gh attestation verify <file> --repo wyolet/relay-catalog`. Consumers pin the published sha256s, so a published asset is never replaced: re-running the release workflow with identical bytes uploads nothing, and different bytes fail the run.

Both release files are built by relay's `catalog-server build` from the image pinned in [`.github/catalog-tools.env`](.github/catalog-tools.env). To reproduce a release, from a checkout that has `<tag>`:

```bash
source .github/catalog-tools.env
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD:/src" -w /src "$CATALOG_TOOLS_IMAGE" \
  catalog-server build --repo . --ref "$(git rev-parse <tag>^{commit})" --version <tag> --out out
```

This writes both files and their `.sha256` files to `out/`; compare those with the release's (`gh release download <tag> -R wyolet/relay-catalog -p '*.sha256'`). The catalog's format comes from the pinned relay version, so a newer image may compose an old tag differently; the data tarball depends only on git. The reproduce workflow runs this check for the latest tag whenever the pin changes.

### Data tree

Every `v*` tag also attaches the live data tree, the YAML relay seeds from (`data/` without `drafts/`):

- `https://github.com/wyolet/relay-catalog/releases/download/<tag>/relay-catalog-data-<tag>.tar.gz`
- `https://github.com/wyolet/relay-catalog/releases/download/<tag>/relay-catalog-data-<tag>.tar.gz.sha256`

The tarball is `git archive --format=tar.gz <tag> -- data ':(exclude)data/drafts'`, reproducible from the tag with git alone on any platform (`git archive` uses its built-in gzip); the `catalog-server build` command above writes it too.

`.gitattributes` limits `git archive` and GitHub's source archives to the same tree, so they never contain tooling. Release tooling keeps the explicit pathspec because older tags have no `.gitattributes`; CI fails when the two disagree.
