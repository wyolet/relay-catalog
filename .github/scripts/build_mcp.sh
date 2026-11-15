#!/usr/bin/env bash
# Builds mcp-catalog-data-<tag>.tar.gz and its .sha256 in <out> exactly as relay's
# `catalog-server build` does: git archive of mcp/ without mcp/drafts at the commit, git's
# built-in gzip, the commit's time as mtime, no GIT_* variables and no system or global config.
# The catalog server rebuilds the same bytes, so any difference here is a mismatch it reports.
# Writes nothing when the commit has no mcp/ tree or its live part holds no YAML file.
#
# Usage: build_mcp.sh <commit> <tag> <out>   (run from the repository)
set -euo pipefail

commit=$1 tag=$2 out=$3
name="mcp-catalog-data-${tag}.tar.gz"
[[ $commit =~ ^[0-9a-f]{40}$ ]] || { echo "build_mcp: want a full commit id, got $commit" >&2; exit 2; }

for v in $(compgen -e | grep '^GIT_' || true); do
  unset "$v"
done
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null GIT_TERMINAL_PROMPT=0 LC_ALL=C
# Like catalog-server: trust this checkout, which a container may see under another uid.
git() { command git -c safe.directory="$PWD" "$@"; }

if [ -z "$(git ls-tree -d --name-only "$commit" -- mcp)" ]; then
  echo "build_mcp: $commit has no mcp/ tree; nothing to build"
  exit 0
fi
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
git archive --format=tar.gz "$commit" -- mcp ':(exclude)mcp/drafts' >"$tmp"
if ! tar -tzf "$tmp" | grep -qE '\.ya?ml$'; then
  echo "build_mcp: mcp/ has no live documents; nothing to build"
  exit 0
fi
mkdir -p "$out"
cp "$tmp" "$out/$name"
(cd "$out" && sha256sum "$name" >"$name.sha256")
echo "build_mcp: wrote $name"
