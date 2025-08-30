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

.PHONY: ci
ci: mod-tidy fmt vet lint test bench ## Run all CI checks
	@echo "✅ All CI checks passed!"

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
release-check: version-check ci ## Check if ready for release
	@echo "🚀 Release readiness check completed"
	@echo ""
	@echo "Before releasing:"
	@echo "1. Ensure CHANGELOG.md is updated"
	@echo "2. Ensure VERSION file is correct"
	@echo "3. Create git tag: git tag v\$$(cat VERSION)"
	@echo "4. Push tag: git push origin v\$$(cat VERSION)"

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