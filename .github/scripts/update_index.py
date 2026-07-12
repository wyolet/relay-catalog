"""Point a schema channel's `latest` at a new release tag in index.yaml.

Usage: update_index.py <channel> <tag>

Never downgrades: a semver tag lower than the channel's current latest is
a no-op (re-pushing an old tag must not move the channel backwards).
"""
import sys

import yaml

HEADER = """\
# Channel index — maps each schema channel (manifest apiVersion suffix) to
# the newest catalog release published for it. Relays started with
# RELAY_CATALOG_VERSION=latest resolve through this file, so an old binary
# never fetches a catalog schema it can't parse. Updated by the
# release-index workflow on tag push; do not edit by hand.
"""


def semver_key(v):
    return [int(p) for p in v.lstrip("v").split(".")]


def main():
    channel, tag = sys.argv[1], sys.argv[2]
    if not channel:
        sys.exit("update_index: empty channel")

    with open("index.yaml") as f:
        doc = yaml.safe_load(f) or {}
    channels = doc.setdefault("channels", {})

    cur = (channels.get(channel) or {}).get("latest", "")
    if cur:
        try:
            if semver_key(tag) <= semver_key(cur):
                print(f"{channel}: {cur} already >= {tag}; no update")
                return
        except ValueError:
            pass  # non-semver stamp — take the new tag

    channels[channel] = {"latest": tag}
    with open("index.yaml", "w") as f:
        f.write(HEADER)
        yaml.safe_dump({"channels": channels}, f, indent=4, sort_keys=True)
    print(f"{channel}: latest -> {tag}")


if __name__ == "__main__":
    main()
