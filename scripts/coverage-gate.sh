#!/usr/bin/env bash
# {project_root}/scripts/coverage-gate.sh
#
# The coverage floor that used to live in two GitHub workflows.
#
# ⭐ The floor is DERIVED from the parked workflows, not retyped here: they are
# the specification this local gate replaces, and a number typed twice is a
# number that drifts. Both workflows state it; if they ever disagree, that is a
# finding and this script refuses rather than picking one.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

WORKFLOW_CI=".github/workflows-disabled/ci.yml"
WORKFLOW_RELEASE=".github/workflows-disabled/release.yml"
COVERAGE_PROFILE="coverage.out"

floor_from_workflow() {
  local file="$1"
  grep -oE '\$COVERAGE < [0-9]+(\.[0-9]+)?' "$file" | head -1 | awk '{print $3}'
}

for f in "$WORKFLOW_CI" "$WORKFLOW_RELEASE"; do
  if [ ! -f "$f" ]; then
    echo "coverage-gate: $f is missing; the floor cannot be derived from the workflow it came from." >&2
    exit 1
  fi
done

FLOOR_CI="$(floor_from_workflow "$WORKFLOW_CI")"
FLOOR_RELEASE="$(floor_from_workflow "$WORKFLOW_RELEASE")"

if [ -z "$FLOOR_CI" ] || [ -z "$FLOOR_RELEASE" ]; then
  echo "coverage-gate: could not derive the coverage floor from the parked workflows." >&2
  exit 1
fi
if [ "$FLOOR_CI" != "$FLOOR_RELEASE" ]; then
  echo "coverage-gate: $WORKFLOW_CI says $FLOOR_CI%, $WORKFLOW_RELEASE says $FLOOR_RELEASE%. Reconcile them first." >&2
  exit 1
fi

FLOOR="$FLOOR_CI"
echo "coverage-gate: floor ${FLOOR}% (derived from $WORKFLOW_RELEASE)"

go test -race -coverprofile="$COVERAGE_PROFILE" -covermode=atomic ./...
go tool cover -html="$COVERAGE_PROFILE" -o coverage.html
COVERAGE="$(go tool cover -func="$COVERAGE_PROFILE" | grep total | awk '{print $3}' | sed 's/%//')"
echo "coverage-gate: measured ${COVERAGE}%"

if (( $(echo "$COVERAGE < $FLOOR" | bc -l) )); then
  echo "coverage-gate: FAIL — ${COVERAGE}% is below the ${FLOOR}% floor" >&2
  exit 1
fi
echo "coverage-gate: OK — ${COVERAGE}% meets the ${FLOOR}% floor"
