#!/usr/bin/env python3
"""List every model doc carrying `labels.featured: "true"`, grouped by provider.

Run from the catalog repo root:  python3 scripts/list_featured.py
"""
import glob
import yaml

total = 0
by_provider: dict[str, list[str]] = {}
for path in sorted(glob.glob("data/providers/**/models/*.yaml", recursive=True)):
    for doc in yaml.safe_load_all(open(path)):
        if not doc:
            continue
        meta = doc.get("metadata", {})
        if (meta.get("labels") or {}).get("featured") == "true":
            prov = meta.get("owner", {}).get("name", "?")
            by_provider.setdefault(prov, []).append(meta.get("name", "?"))
            total += 1

for prov in sorted(by_provider):
    names = sorted(by_provider[prov])
    print(f"\n{prov} ({len(names)})")
    for n in names:
        print(f"  - {n}")
print(f"\n{total} featured model(s) across {len(by_provider)} provider(s).")
