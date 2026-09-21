# Parked GitHub workflows — inert on purpose

Hosted GitHub Actions is **retired for this repository** (operator decision,
2026-09-19; tracked as [issue #10](https://github.com/itsatony/go-cuserr/issues/10)).
Every check now runs locally, in a gate that is at least as complete as the CI
it replaces.

These two files are **kept, not deleted**: they are the *specification* the local
gate is compared against. `TestLocalCIReproducesTheDisabledWorkflows`
(`cuserr_ci_parity_test.go`) reads them on every test run and fails if a job, a
matrix axis or a `run:` command in here has no local counterpart and no declared
exception. A deleted file cannot be compared against anything.

Actions is also disabled at the **repository level**
(`gh api -X PATCH repos/itsatony/go-cuserr/actions/permissions -F enabled=false`).
Moving a workflow file out of `.github/workflows/` does *not* disable Actions —
any workflow restored or newly added would start reporting again, with no signal
that this repo had decided against it.

## Where each job went

| Workflow | Job | Local answer |
|---|---|---|
| `ci.yml` | `quality` | `make ci-preflight ci-fmt ci-vet ci-lint` |
| `ci.yml` | `test` | `make ci-build ci-test ci-bench` |
| `ci.yml` | `coverage` | `make ci-coverage` (floor **derived** from these files) |
| `ci.yml` | `security` | *declared exception* — see below |
| `ci.yml` | `integration` | `make ci-examples` |
| `ci.yml` | `performance` | `make ci-perf` |
| `ci.yml` | `build-matrix` | `make build-matrix` (GOOS/GOARCH loop) |
| `release.yml` | `validate` | `make release-check` → `scripts/release-check.sh` |
| `release.yml` | `release` | *declared exception* — publishing, not checking |
| `release.yml` | `notify` | *declared exception* — an echo |

Everything runs at once with **`make ci-local`** (alias: `make ci`), and
**`make release-ready`** adds the release guards.

## Declared exceptions, with reasons

1. **The 3 OS × Go 1.21/1.22/1.23 test matrix is NOT reproduced.** One host runs
   one operating system and one toolchain; that cannot be faked. The decision
   (issue #10) is to **accept single-host coverage** and say so, rather than let
   a local gate claim a parity it does not have. What carries the compatibility
   promise instead: the cross-compile `build-matrix` arm (linux/windows/darwin ×
   amd64/arm64, which *is* reproduced) and the Go version floor in `go.mod`.
   ⚠ Note the parked matrix names 1.21–1.23 while `go.mod` already requires a
   newer Go — the matrix had gone stale, unnoticed, which is itself the argument
   for not keeping arms nobody reads.
2. **`ci.yml`'s `security` job ran no scanner.** Its entire body is an `echo`
   saying gosec was unavailable. There is a `make security` target if a real
   scanner is ever wanted; reproducing a placeholder would be reproducing
   nothing.
3. **`release.yml`'s `release` and `notify` jobs are publishing, not checking** —
   creating the GitHub Release, uploading its body, pinging `proxy.golang.org`.
   Releases are now cut by hand with `gh release create` after
   `make release-ready` is green.

The load-bearing parts of `release.yml` — the **tag ↔ VERSION guard** and the
**≥ 84 % coverage floor** — did *not* get parked with it. They live in
`scripts/release-check.sh` and `scripts/coverage-gate.sh`, and the floor is
**derived by parsing these workflow files** rather than retyped, so the two can
never disagree.

## Re-enabling hosted CI

It is a two-step decision, in this order:

1. `git mv .github/workflows-disabled/*.yml .github/workflows/` — and update
   `TestHostedCIIsDisabledDeliberately`, which fails while an active workflow
   exists.
2. `gh api -X PATCH repos/itsatony/go-cuserr/actions/permissions -F enabled=true`
   — without this the file is present and Actions still refuses to run it.

Before doing either, fix what parking exposed: the workflows as written were
**red on a clean tree** when they were parked (see the `0.3.1` CHANGELOG entry) —
`gofmt -s -l` reported three files and the `golangci-lint` config was in the v1
format that current golangci-lint refuses to load. An arm that never executes is
indistinguishable from one that passes.
