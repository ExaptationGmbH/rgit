#!/usr/bin/env bash
#
# update-formula.sh — pin Formula/rgit.rb to a tagged release.
#
# Given a git tag (e.g. v0.1.0), this downloads the GitHub-generated source
# tarball for that tag, computes its SHA-256, and rewrites the formula's
# `url` and `sha256` fields to match. Run it after you push the tag /
# publish the release.
#
# Usage:
#   scripts/update-formula.sh v0.1.0
#
# Requires: curl, shasum (or sha256sum).
set -euo pipefail

OWNER="ExaptationGmbH"
REPO="rgit"
FORMULA="$(cd "$(dirname "$0")/.." && pwd)/Formula/rgit.rb"

tag="${1:-}"
if [[ -z "$tag" ]]; then
  echo "usage: $0 <tag>   e.g. $0 v0.1.0" >&2
  exit 2
fi
# Constrain the tag to a semver-ish shape so it can't inject into the URL or
# the sed replacement below.
if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.]+)?$ ]]; then
  echo "error: tag '$tag' is not a valid vMAJOR.MINOR.PATCH tag" >&2
  exit 2
fi

url="https://github.com/${OWNER}/${REPO}/archive/refs/tags/${tag}.tar.gz"
echo "Fetching $url"

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp"

if command -v sha256sum >/dev/null 2>&1; then
  sha="$(sha256sum "$tmp" | awk '{print $1}')"
else
  sha="$(shasum -a 256 "$tmp" | awk '{print $1}')"
fi
echo "sha256: $sha"

# Rewrite the url and sha256 lines in the formula.
# BSD sed (macOS) and GNU sed differ on -i; write to a temp and move.
out="$(mktemp)"
sed -E \
  -e "s|^(  url ).*|\1\"${url}\"|" \
  -e "s|^(  sha256 ).*|\1\"${sha}\"|" \
  "$FORMULA" >"$out"
mv "$out" "$FORMULA"

echo "Updated $FORMULA -> $tag"
