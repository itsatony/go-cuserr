# Go-CUserr Makefile
# Following vAudience.AI "Excellence. Always." standards

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run all tests
	go test -v -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bench
bench: ## Run benchmark tests
	go test -bench=. -benchmem -run=^$$ ./...

.PHONY: bench-compare
bench-compare: ## Run benchmarks and save for comparison
	go test -bench=. -benchmem -run=^$$ ./... | tee bench.out

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	goimports -w -local github.com/itsatony/go-cuserr .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy and verify module dependencies
	go mod tidy
	go mod verify

.PHONY: build
build: ## Build the package (check compilation)
	go build ./...

.PHONY: clean
clean: ## Clean build artifacts and test cache
	go clean -cache -testcache -modcache
	rm -f coverage.out coverage.html bench.out

.PHONY: examples
examples: ## Run all examples
	@echo "Running basic usage example..."
	go run examples/basic_usage.go
	@echo "\nExamples completed. HTTP service and middleware examples need to be run separately:"
	@echo "  go run examples/http_service.go    # Runs on :8080"
	@echo "  go run examples/middleware.go      # Runs on :8081"

.PHONY: deps
deps: ## Install development dependencies
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || (echo "Please install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development dependencies ready"

.PHONY: pre-commit
pre-commit: fmt vet lint test ## Run all pre-commit checks
	@echo "✅ All pre-commit checks passed!"

# ---------------------------------------------------------------------------
# Local CI gate — the replacement for the parked GitHub workflows.
#
# Hosted GitHub Actions is retired (operator decision, 2026-09-19). The
# workflows are kept, inert, in .github/workflows-disabled/ because they are the
# SPECIFICATION this gate is compared against — see the README there, and
# TestLocalCIReproducesTheDisabledWorkflows, which fails if the two drift.
#
# ⛔ One arm of the hosted CI is NOT reproducible here and is a declared
# exception: the 3 OS x Go 1.21/1.22/1.23 test matrix. One host runs one OS and
# one toolchain. Saying so is the point — a gate that silently dropped it would
# claim a parity it does not have.
# ---------------------------------------------------------------------------

.PHONY: ci-preflight
ci-preflight: ## CI arm: the environment guarantees the hosted runner used to provide
	@command -v go > /dev/null || { echo "go toolchain not found"; exit 1; }
	@command -v golangci-lint > /dev/null || { echo "golangci-lint not found: see make deps"; exit 1; }
	@command -v bc > /dev/null || { echo "bc not found: the coverage floor needs it"; exit 1; }
	go mod download
	go mod verify

.PHONY: ci-fmt
ci-fmt: ## CI arm: gofmt -s check (never rewrites)
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted correctly:"; \
		gofmt -s -l .; \
		exit 1; \
	fi

.PHONY: ci-vet
ci-vet: ## CI arm: go vet
	go vet ./...

.PHONY: ci-lint
ci-lint: ## CI arm: golangci-lint
	golangci-lint run --timeout=5m

.PHONY: ci-build
ci-build: ## CI arm: compile the packages
	go build -v ./...

.PHONY: ci-test
ci-test: ## CI arm: the race test suite (single host — see the matrix exception)
	go test -v -race -timeout 10m ./...

.PHONY: ci-bench
ci-bench: ## CI arm: benchmark smoke test
	go test -bench=. -benchtime=100ms -run=^$$ ./...

.PHONY: ci-coverage
ci-coverage: ## CI arm: coverage floor, derived from the parked workflows
	scripts/coverage-gate.sh

.PHONY: ci-examples
ci-examples: ## CI arm: the integration job — every tagged example must build
	timeout 10s go run -tags basic examples/basic_usage.go || [ $$? -eq 124 ]
	@# The tag set is DERIVED from the files, not hand-listed: the hosted job named
	@# four examples and the directory has eight, so three of them could rot unseen
	@# (and did — the API they used was removed in b362b56).
	@set -e; for f in examples/*.go; do \
		tag=$$(sed -n 's|^//go:build ||p' $$f | head -1); \
		if [ -z "$$tag" ]; then echo "--> $$f has no build tag, skipping"; continue; fi; \
		echo "--> building $$f (tag $$tag)"; \
		go build -tags $$tag -o /dev/null $$f; \
	done

.PHONY: ci-perf
ci-perf: ## CI arm: the named performance benchmarks
	go test -bench=BenchmarkNewCustomError -benchtime=1s -run=^$$ ./...
	go test -bench=BenchmarkWithMetadata -benchtime=1s -run=^$$ ./...
	go test -bench=BenchmarkToJSON -benchtime=1s -run=^$$ ./...

.PHONY: build-matrix
build-matrix: ## CI arm: cross-compile for every GOOS/GOARCH the workflow built
	@set -e; for goos in linux windows darwin; do \
		for goarch in amd64 arm64; do \
			if [ "$$goos" = "windows" ] && [ "$$goarch" = "arm64" ]; then continue; fi; \
			echo "--> building $$goos/$$goarch"; \
			GOOS=$$goos GOARCH=$$goarch go build -v ./...; \
		done; \
	done

.PHONY: ci-local
ci-local: ci-preflight ci-fmt ci-vet ci-lint ci-build ci-test ci-bench ci-coverage ci-examples ci-perf build-matrix ## Run the whole local CI gate
	@echo "✅ Local CI gate passed (single-host; the OS/Go matrix is a declared exception)"

.PHONY: ci
ci: ci-local ## Alias for ci-local, kept because muscle memory and docs say `make ci`

.PHONY: version-check
version-check: ## Check if VERSION file matches git tags
	@VERSION=$$(cat VERSION) && \
	echo "VERSION file: $$VERSION" && \
	if git tag -l | grep -q "^v$$VERSION$$"; then \
		echo "✅ Version $$VERSION matches git tag"; \
	else \
		echo "⚠️  Version $$VERSION not found in git tags"; \
		echo "Available tags:"; \
		git tag -l | head -10; \
	fi

.PHONY: release-check
release-check: ## The load-bearing half of the parked release.yml: tag <-> VERSION guard + coverage floor
	scripts/release-check.sh

.PHONY: release-ready
release-ready: ci-local release-check ## Full pre-release gate: local CI plus the release guards
	@echo "🚀 Release gates passed. Tag with: git tag v$$(cat VERSION) && git push origin v$$(cat VERSION)"

.PHONY: install-hooks
install-hooks: ## Install git pre-commit hooks
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "make pre-commit" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"

.PHONY: security
security: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...
	@echo "✅ Security checks completed"

.PHONY: doc
doc: ## Generate and serve documentation
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "📚 Documentation server started at http://localhost:6060/pkg/github.com/itsatony/go-cuserr/"
	@echo "Press Ctrl+C to stop the server"

.PHONY: all
all: deps pre-commit ## Install dependencies and run all checks

# Default target
.DEFAULT_GOAL := help

# Development shortcuts
.PHONY: t
t: test ## Shortcut for test

.PHONY: b
b: bench ## Shortcut for bench

.PHONY: l
l: lint ## Shortcut for lint

.PHONY: f
f: fmt ## Shortcut for fmt