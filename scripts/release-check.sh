#!/usr/bin/env bash
# {project_root}/scripts/release-check.sh
#
# The load-bearing half of the parked release.yml: the tag <-> VERSION guard and
# the coverage floor. Without it, "tagged v1.4.0 while VERSION says 1.3.0" is a
# silent possibility again.
#
# Usage: scripts/release-check.sh [vX.Y.Z]
#   With no argument the tag is taken from the tags pointing at HEAD; if there
#   are none, the check reports the tag that WOULD be correct and still verifies
#   everything else.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

FILE_VERSION="$(tr -d '[:space:]' < VERSION)"
TAG="${1:-$(git tag --points-at HEAD | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)}"

echo "release-check: VERSION file says $FILE_VERSION"

if [ -n "$TAG" ]; then
  TAG_VERSION="${TAG#v}"
  if [ "$FILE_VERSION" != "$TAG_VERSION" ]; then
    echo "release-check: FAIL — VERSION file ($FILE_VERSION) does not match tag ($TAG_VERSION)" >&2
    exit 1
  fi
  echo "release-check: OK — VERSION file matches tag $TAG"
else
  echo "release-check: HEAD carries no v* tag; the tag to create is v$FILE_VERSION"
fi

# A tag that already exists must name a commit whose VERSION file agrees with it.
# This is the case the workflow could never catch: it only ever saw the tag being
# pushed, never one that was already wrong.
if git rev-parse -q --verify "refs/tags/v$FILE_VERSION" >/dev/null; then
  TAGGED_VERSION="$(git show "v$FILE_VERSION:VERSION" 2>/dev/null | tr -d '[:space:]' || true)"
  if [ "$TAGGED_VERSION" != "$FILE_VERSION" ]; then
    echo "release-check: FAIL — tag v$FILE_VERSION points at a commit whose VERSION file says '$TAGGED_VERSION'" >&2
    exit 1
  fi
  echo "release-check: OK — existing tag v$FILE_VERSION agrees with its own VERSION file"
fi

# release.yml extracts the release notes from this section; an absent one made it
# publish a release body of "Release vX.Y.Z" and nobody noticed.
if ! grep -qE "^## \[$FILE_VERSION\]" CHANGELOG.md; then
  echo "release-check: FAIL — CHANGELOG.md has no '## [$FILE_VERSION]' section" >&2
  exit 1
fi
echo "release-check: OK — CHANGELOG.md has a section for $FILE_VERSION"

go mod download
go test -race ./...
scripts/coverage-gate.sh

echo "release-check: all release gates passed for $FILE_VERSION"
