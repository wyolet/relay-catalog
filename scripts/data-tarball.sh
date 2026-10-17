#!/usr/bin/env bash
# Writes relay-catalog-data-<tag>.tar.gz and its .sha256 into <outdir>: the
# live data tree at <commit>, without drafts/. git archive stamps every entry
# with the commit's time and compresses with its built-in gzip, so the same
# commit gives the same bytes on every run and platform; an external gzip
# would tie the bytes to that gzip's version.
set -euo pipefail

tag=${1:?usage: scripts/data-tarball.sh <tag> <commit> <outdir>}
commit=${2:?usage: scripts/data-tarball.sh <tag> <commit> <outdir>}
out=${3:?usage: scripts/data-tarball.sh <tag> <commit> <outdir>}

tarball="relay-catalog-data-${tag}.tar.gz"
mkdir -p "$out"
git archive --format=tar.gz -o "$out/$tarball" "$commit" -- data ':(exclude)data/drafts'
(cd "$out" && sha256sum "$tarball" >"$tarball.sha256")
echo "$out/$tarball"
