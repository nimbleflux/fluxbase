#!/usr/bin/env bash
# =============================================================================
# Bump the pinned pgschema version + checksums in the Dockerfile.
#
# Renovate bumps PGSCHEMA_VERSION via a regexManager but cannot update the
# pinned sha256 ARGs, so auto-bumps always fail the image build's checksum
# verification. Renovate therefore no longer proposes pgschema updates; run
# this script instead — it downloads both release binaries, recomputes the
# checksums, and rewrites the Dockerfile ARGs in one step.
#
# Usage: scripts/bump-pgschema.sh <version>     e.g. scripts/bump-pgschema.sh 1.14.0
# =============================================================================
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

VERSION="$1"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DOCKERFILE="$REPO_ROOT/Dockerfile"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

for ARCH in amd64 arm64; do
  URL="https://github.com/pgplex/pgschema/releases/download/v${VERSION}/pgschema-${VERSION}-linux-${ARCH}"
  echo "==> downloading ${ARCH}"
  curl -fsSL "$URL" -o "$TMP/pgschema-${ARCH}"
  SHA="$(sha256sum "$TMP/pgschema-${ARCH}" | cut -d' ' -f1)"
  echo "    sha256($ARCH) = $SHA"
  declare "PG_SHA_${ARCH}=$SHA"
done

echo "==> rewriting ${DOCKERFILE}"
sed -i \
  -e "s|^ARG PGSCHEMA_VERSION=.*|ARG PGSCHEMA_VERSION=${VERSION}|" \
  -e "s|^ARG PGSCHEMA_SHA256_AMD64=.*|ARG PGSCHEMA_SHA256_AMD64=${PG_SHA_amd64}|" \
  -e "s|^ARG PGSCHEMA_SHA256_ARM64=.*|ARG PGSCHEMA_SHA256_ARM64=${PG_SHA_arm64}|" \
  "$DOCKERFILE"

echo "==> verifying current versions in ${DOCKERFILE}"
grep -n '^ARG PGSCHEMA' "$DOCKERFILE"

echo
echo "Done. Review 'git diff $DOCKERFILE' and commit."
