# Contributing to go-cuserr

Thank you for your interest in contributing to go-cuserr! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## How to Contribute

### 🐛 Reporting Bugs

Before creating bug reports, please check the [existing issues](https://github.com/itsatony/go-cuserr/issues) to avoid duplicates.

When creating a bug report, please include:

- **Clear title and description**
- **Go version** (`go version`)
- **Operating system and version**
- **Minimal reproducible example**
- **Expected vs actual behavior**
- **Error messages** (if any)
- **Relevant logs** (if applicable)

### 💡 Suggesting Features

Feature suggestions are welcome! Please:

1. Check existing [issues](https://github.com/itsatony/go-cuserr/issues) and [discussions](https://github.com/itsatony/go-cuserr/discussions)
2. Create a new issue with the `enhancement` label
3. Provide a clear description of the feature
4. Explain the use case and benefits
5. Consider backward compatibility implications

### 🔧 Code Contributions

#### Getting Started

1. **Fork the repository**
   ```bash
   gh repo fork itsatony/go-cuserr --clone
   cd go-cuserr
   ```

2. **Set up your development environment**
   ```bash
   go version  # Ensure Go 1.21+
   make deps   # Install development dependencies
   ```

3. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

#### Development Guidelines

##### Code Standards

We follow vAudience.AI "Excellence. Always." standards:

- **Zero external dependencies** - Use only Go standard library
- **Thread-safe by design** - All public APIs must be concurrent-safe
- **Comprehensive testing** - Maintain >85% test coverage
- **Performance focused** - Benchmark critical paths
- **Production ready** - Consider real-world usage scenarios

##### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Use meaningful variable and function names
- Write clear, concise comments for public APIs
- Keep functions focused and testable
- Prefer composition over inheritance

##### File Organization

Follow the existing pattern:
- `cuserr.*.go` - Core functionality files
- `cuserr_*_test.go` - Test files
- `examples/` - Usage examples
- `.github/` - CI/CD and templates

##### Testing Requirements

All contributions must include:

1. **Unit tests** for new functionality
2. **Thread safety tests** for concurrent operations
3. **Benchmark tests** for performance-critical code
4. **Integration tests** for end-to-end scenarios

```bash
# Run all tests
make test

# Run tests with race detection
go test -race -v ./...

# Run benchmarks
make bench

# Check coverage
make test-coverage
```

##### Performance Considerations

- **Error creation**: Target <2μs per operation
- **Metadata operations**: Target <100ns per operation
- **JSON serialization**: Target <1μs per operation
- **Memory allocation**: Minimize heap allocations

#### Pull Request Process

1. **Ensure all tests pass**
   ```bash
   make pre-commit
   ```

2. **Update documentation**
   - Add/update Go doc comments
   - Update README.md if needed
   - Add examples for new features
   - Update CHANGELOG.md

3. **Create pull request**
   - Use a clear, descriptive title
   - Reference related issues
   - Describe what changes were made and why
   - Include testing information

4. **Code review process**
   - Address reviewer feedback promptly
   - Keep discussions focused and respectful
   - Update tests and documentation as requested

#### PR Checklist

- [ ] Code follows Go conventions and project style
- [ ] Tests written and passing (including race detection)
- [ ] Benchmarks added for performance-critical changes
- [ ] Documentation updated (Go docs, README, examples)
- [ ] CHANGELOG.md updated with changes
- [ ] No breaking changes (or clearly documented)
- [ ] Thread safety maintained
- [ ] Performance regression tests pass

### 📚 Documentation Contributions

Documentation improvements are highly valued:

- **API documentation** - Improve Go doc comments
- **Usage examples** - Add to `examples/` directory
- **README updates** - Clarify usage patterns
- **Tutorial content** - Help new users get started

### 🏗️ Development Workflow

#### Local Development

```bash
# Run tests continuously during development
make test

# Auto-format code
make fmt

# Run linter
make lint

# Full pre-commit checks
make pre-commit
```

#### Makefile Targets

- `make help` - Show available targets
- `make test` - Run all tests with race detection
- `make bench` - Run benchmarks
- `make lint` - Run golangci-lint
- `make fmt` - Format code and imports
- `make coverage` - Generate coverage report
- `make examples` - Run example code
- `make pre-commit` - Run all pre-commit checks
- `make release-check` - Validate release readiness

#### CI/CD Pipeline

Our GitHub Actions pipeline runs:

- **Quality checks** - formatting, linting, vetting
- **Cross-platform tests** - Linux, Windows, macOS
- **Go version matrix** - 1.21, 1.22, 1.23
- **Security scanning** - gosec analysis
- **Performance tests** - benchmark regression detection
- **Coverage validation** - >85% threshold
- **Integration tests** - example compilation and execution

### 🔒 Security

#### Reporting Security Issues

**Do not open public issues for security vulnerabilities.**

Instead, email security concerns to: [security@vaudience.ai](mailto:security@vaudience.ai)

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if known)

We will respond within 48 hours and work with you to address the issue.

#### Security Best Practices

When contributing:
- Never log or expose sensitive data
- Validate all inputs thoroughly
- Use secure defaults
- Consider timing attack vulnerabilities
- Review error messages for information leakage

### 📋 Release Process

Releases follow semantic versioning (SemVer):

- **Patch** (0.0.X) - Bug fixes, performance improvements
- **Minor** (0.X.0) - New features, backward compatible
- **Major** (X.0.0) - Breaking changes

#### Release Checklist (Maintainers)

1. Update VERSION file
2. Update CHANGELOG.md
3. Run `make release-check`
4. Create and push git tag: `git tag v0.1.0 && git push origin v0.1.0`
5. GitHub Actions handles the rest

### 🤝 Community

#### Getting Help

- **GitHub Discussions** - General questions and ideas
- **GitHub Issues** - Bug reports and feature requests
- **Documentation** - Check README.md and examples/

#### Recognition

Contributors are recognized in:
- CHANGELOG.md for their contributions
- GitHub contributors list
- Special mentions for significant contributions

### 💡 Development Tips

#### Common Patterns

```go
// Thread-safe metadata access
func (e *CustomError) GetMetadata(key string) (string, bool) {
    e.mu.RLock()
    defer e.mu.RUnlock()
    value, exists := e.metadata[key]
    return value, exists
}

// Chainable methods for fluent API
func (e *CustomError) WithMetadata(key, value string) *CustomError {
    e.mu.Lock()
    defer e.mu.Unlock()
    if e.metadata == nil {
        e.metadata = make(map[string]string)
    }
    e.metadata[key] = value
    return e
}
```

#### Testing Patterns

```go
// Test thread safety
func TestConcurrentAccess(t *testing.T) {
    err := NewCustomError(ErrInternal, nil, "test")
    
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            err.WithMetadata(fmt.Sprintf("key_%d", id), "value")
        }(i)
    }
    wg.Wait()
}

// Benchmark critical operations
func BenchmarkErrorCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewCustomError(ErrInternal, nil, "test error")
    }
}
```

#### Performance Optimization

- Use `sync.Pool` for frequently allocated objects
- Prefer stack allocation over heap when possible
- Minimize string concatenation in hot paths
- Use `strings.Builder` for building strings
- Profile with `go test -cpuprofile` and `go tool pprof`

### 📝 Questions?

If you have questions not covered in this guide:

1. Check the [README.md](README.md)
2. Browse [existing issues](https://github.com/itsatony/go-cuserr/issues)
3. Start a [discussion](https://github.com/itsatony/go-cuserr/discussions)
4. Open a new issue with the `question` label

---

**Thank you for contributing to go-cuserr!** 

*Excellence. Always.* - vAudience.AI Team