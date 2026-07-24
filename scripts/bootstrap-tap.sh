#!/usr/bin/env bash
#
# bootstrap-tap.sh — first-time publish so `brew install ExaptationGmbH/tap/rgit`
# works. Run this once from a clean rgit checkout on your own machine.
#
# It will:
#   1. push main to the rgit remote
#   2. create + push a release tag (default v0.1.0)
#   3. cut a GitHub release for that tag
#   4. pin Formula/rgit.rb to the tag's tarball sha256 and push that
#   5. create the ExaptationGmbH/homebrew-tap repo (if missing) and publish
#      the formula into it
#
# Requires: git, gh (authenticated: `gh auth login`), curl, shasum/sha256sum.
#
# Usage:
#   scripts/bootstrap-tap.sh            # tag v0.1.0
#   scripts/bootstrap-tap.sh v0.2.0     # a different tag
set -euo pipefail

OWNER="ExaptationGmbH"
REPO="rgit"
TAP_REPO="homebrew-tap"
TAG="${1:-v0.1.0}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

command -v gh >/dev/null || { echo "error: gh CLI not found (brew install gh)"; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "error: run 'gh auth login' first"; exit 1; }

cd "$ROOT"

echo "==> 1/5 push main"
git push -u origin main

echo "==> 2/5 tag $TAG"
if ! git rev-parse "$TAG" >/dev/null 2>&1; then
  git tag -a "$TAG" -m "rgit $TAG"
fi
git push origin "$TAG"

echo "==> 3/5 GitHub release $TAG"
if ! gh release view "$TAG" --repo "$OWNER/$REPO" >/dev/null 2>&1; then
  gh release create "$TAG" --repo "$OWNER/$REPO" --title "rgit $TAG" --generate-notes
fi

echo "==> 4/5 pin formula to $TAG"
scripts/update-formula.sh "$TAG"
if ! git diff --quiet Formula/rgit.rb; then
  git add Formula/rgit.rb
  git commit -m "chore: pin homebrew formula to $TAG"
  git push
fi

echo "==> 5/5 publish tap $OWNER/$TAP_REPO"
if ! gh repo view "$OWNER/$TAP_REPO" >/dev/null 2>&1; then
  gh repo create "$OWNER/$TAP_REPO" --public \
    --description "Homebrew tap for ExaptationGmbH tools"
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
git clone "https://github.com/$OWNER/$TAP_REPO.git" "$work/tap"
mkdir -p "$work/tap/Formula"
cp "$ROOT/Formula/rgit.rb" "$work/tap/Formula/rgit.rb"
git -C "$work/tap" add Formula/rgit.rb
if ! git -C "$work/tap" diff --cached --quiet; then
  git -C "$work/tap" commit -m "rgit $TAG"
  git -C "$work/tap" push
fi

echo
echo "Done. Install with:"
echo "    brew install $OWNER/tap/$REPO"
