# Contributing to go-opds

Thank you for your interest in contributing to `go-opds`! We welcome bug reports, feature suggestions, documentation improvements, and pull requests.

## Philosophy & Guidelines

- **Zero External Dependencies**: The core packages (`opds1`, `opds2`, `convert`, `handler`, `internal/common`) rely exclusively on the Go standard library. We do not accept third-party dependencies into the core library.
- **Protocol Compliance**: Invariants defined in the [OPDS 1.2](https://specs.opds.io/opds-1.2) and [OPDS 2.0](https://drafts.opds.io/opds-2.0.html) specifications must be respected.
- **Defensive Safety & Security**: All user-facing APIs enforce input bounds, path safety, and nil-pointer guards.
- **Executable Tests**: Every public feature, builder option, and conversion pathway must have table-driven unit tests, subtests (`t.Run`), and parallel execution (`t.Parallel()`).

---

## Development Setup

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or higher
- [golangci-lint](https://golangci-lint.run/) (v1.60+ recommended)
- `git` and `make`

### Building & Testing

Clone the repository:

```bash
git clone https://github.com/littfed/go-opds.git
cd go-opds
```

Run test suite with race detector:

```bash
make test
# or: go test -v -race ./...
```

Run linters:

```bash
make lint
# or: golangci-lint run ./...
```

Run benchmarks:

```bash
make bench
# or: go test -bench=. -benchmem ./...
```

Run test coverage:

```bash
make test-coverage
```

Build example binaries:

```bash
make build
```

---

## Submitting a Pull Request

1. **Fork the repository** and create your branch from `main`:
   ```bash
   git checkout -b feature/my-new-feature
   ```
2. **Write clean, idiomatic Go**:
   - Format with `gofmt` / `goimports`.
   - Add doc comments on all exported types, functions, and constants.
   - Follow standard Go naming conventions (`ID`, `URL`, no `Get` prefixes on getters).
3. **Add tests**:
   - Include unit tests for new functionality.
   - If adding a major public API, include a godoc `Example*()` test in `*_test.go`.
4. **Verify before submitting**:
   - Ensure `make test` passes with zero race conditions.
   - Ensure `make lint` exits cleanly with zero warnings.
5. **Open a Pull Request**:
   - Describe what changed and why.
   - Link any related issues.
