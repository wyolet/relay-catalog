#!/usr/bin/env python3
"""One-shot: stamp `featured: "true"` into the labels block of selected model
docs. Idempotent — skips docs already featured. Targets docs by metadata.name
so it works correctly inside multi-doc (--- separated) model files.

Run from the catalog repo root:  python3 scripts/apply_featured.py
"""
import pathlib
import re
import sys

# The curated featured set. Edit this list, not the YAML by hand.
FEATURED = {
    # OpenAI gpt-5*
    "gpt-5", "gpt-5-chat", "gpt-5-mini", "gpt-5-nano", "gpt-5-search-api",
    "gpt-5-1", "gpt-5-1-chat", "gpt-5-2", "gpt-5-2-chat", "gpt-5-3-chat",
    "gpt-5-4", "gpt-5-4-mini", "gpt-5-4-nano", "gpt-5-5",
    # Anthropic claude 4*
    "claude-haiku-4-5", "claude-4-sonnet", "claude-sonnet-4",
    "claude-sonnet-4-5", "claude-sonnet-4-6", "claude-4-opus", "claude-opus-4",
    "claude-opus-4-1", "claude-opus-4-5", "claude-opus-4-6", "claude-opus-4-7",
    "claude-opus-4-8",
    # Google gemini 3*
    "gemini-3-1-flash-lite-preview", "gemini-3-1-flash-lite",
    "gemini-3-1-pro-preview", "gemini-3-5-flash", "gemini-3-flash-preview",
    # Newest open models served via Ollama (flagship of each family)
    "deepseek-v4-pro", "deepseek-v4-flash", "glm-5", "glm-5-1", "kimi-k2-6",
    "minimax-m2-7", "qwen3-5-397b", "qwen3-vl-235b", "gpt-oss-120b",
    "gemma4-31b", "mistral-large-3", "nemotron-3-super",
}

NAME_RE = re.compile(r"^    name: (\S+)\s*$")
LABELS_RE = re.compile(r"^    labels:\s*$")
ROOT = pathlib.Path("data/providers")


def feature_doc(lines: list[str]) -> tuple[list[str], bool]:
    """Insert featured label into one doc's labels block if it's a target."""
    name = None
    for ln in lines:
        m = NAME_RE.match(ln)
        if m:
            name = m.group(1)
            break
    if name not in FEATURED:
        return lines, False
    if any('featured: "true"' in ln for ln in lines):
        return lines, False  # already featured
    out = []
    inserted = False
    for ln in lines:
        out.append(ln)
        if not inserted and LABELS_RE.match(ln):
            out.append('        featured: "true"')
            inserted = True
    if not inserted:
        print(f"  WARN: {name} has no labels block — skipped", file=sys.stderr)
        return lines, False
    return out, True


def main() -> int:
    changed = 0
    for path in sorted(ROOT.rglob("*.yaml")):
        text = path.read_text()
        docs = text.split("\n---\n")
        new_docs = []
        touched = False
        for doc in docs:
            new_lines, did = feature_doc(doc.split("\n"))
            if did:
                touched = True
                changed += 1
            new_docs.append("\n".join(new_lines))
        if touched:
            path.write_text("\n---\n".join(new_docs))
            print(f"featured in {path}")
    print(f"\n{changed} model doc(s) marked featured.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
