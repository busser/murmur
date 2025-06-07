# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Development
- **Format code**: `make fmt` or `go fmt ./...`
- **Vet code**: `make vet` or `go vet ./...`  
- **Run unit tests**: `make test` or `go test ./...`
- **Run all tests (including e2e)**: `make test-e2e` or `go test -tags=e2e ./...`
- **Build binary**: `make build` (outputs to `bin/murmur`)

### Testing
- Run specific test: `go test ./path/to/package -run TestName`
- Run e2e tests for specific provider: `go test -tags=e2e ./pkg/murmur/providers/awssm`

## Architecture

Murmur is a secrets injection tool that fetches secrets from various cloud providers and injects them as environment variables for subprocesses.

### Core Components

**Command Structure**: 
- `main.go` → `pkg/cmd/` → `pkg/murmur/`
- Primary commands: `run` (preferred) and `exec` (deprecated)

**Provider System** (`pkg/murmur/providers/`):
- Each provider implements the `Provider` interface with `Resolve()` and `Close()` methods
- Supported providers: `awssm` (AWS Secrets Manager), `azkv` (Azure Key Vault), `gcpsm` (GCP Secret Manager), `scwsm` (Scaleway Secret Manager), `passthrough` (testing)
- Provider registration in `provider.go` via `ProviderFactories` map

**Query Processing Pipeline** (`resolve.go`):
1. **Parse**: Environment variables → query objects (`provider_id:secret_ref|filter_id:filter_rule`)
2. **Resolve**: Fetch secrets from providers (concurrent by provider, cached for duplicates)
3. **Filter**: Apply transformations like JSONPath parsing
4. **Output**: Final environment variables for subprocess

**Filter System** (`filter.go`, `filters/`):
- Currently supports `jsonpath` filter using Kubernetes JSONPath syntax
- Filters transform raw secret values (e.g., extract fields from JSON secrets)

### Key Design Patterns

- **Concurrent processing**: Secrets are fetched concurrently per provider to minimize latency
- **Caching**: Duplicate secret references are cached to avoid redundant API calls  
- **Pipeline architecture**: Variable processing flows through parse → resolve → filter stages
- **Provider isolation**: Each cloud provider is completely isolated in its own package

### Testing Structure

- Unit tests alongside source files (`*_test.go`)
- E2E tests require `-tags=e2e` flag and real cloud credentials
- Mock providers available for testing (`providers/mock/`, `providers/jsonmock/`)
- Test data in `pkg/murmur/testdata/`

## Release Process

Murmur uses a structured release process with organized release notes and automated tooling.

### Release Commands

- **Release Dry Run**: `make release-dry-run` - Test the complete release process without publishing
- **Release**: `make release` - Create and publish a new release (requires `gh` CLI authentication)

### Release Workflow

1. **Prepare Release Notes**: Create `docs/release-notes/vX.Y.Z.md` with comprehensive release notes
2. **Create Release Branch**: `git checkout -b release/vX.Y.Z`
3. **Update VERSION**: Change `VERSION` file to target version (e.g., `v0.7.0`)
4. **Commit and PR**: `git commit -m "release vX.Y.Z"` and open PR
5. **Merge and Release**: After PR merge, checkout main, pull, and run `make release`

### Release Notes Format

Release notes in `docs/release-notes/` follow this structure:
- **Emoji-prefixed sections** (🔥, 📋, 📚, 🔧)
- **Brief descriptions** with **code examples**
- **Links to documentation** for detailed information
- **User-focused language** highlighting benefits

### Automated Release Features

- **Multi-platform binaries**: Linux, macOS, Windows (amd64, arm64, 386)
- **Container images**: Automatic Docker image publishing to GitHub Container Registry
- **GitHub integration**: Automated release creation with binaries and checksums
- **Release notes**: Automatically included from `docs/release-notes/$(VERSION).md`
- **Backward compatibility**: Continues publishing both `murmur` and `whisper` binaries

### Dependencies

- **GoReleaser v2**: Handles cross-compilation and publishing
- **GitHub CLI (`gh`)**: Provides authentication token for releases
- **Git tags**: Version tags trigger GoReleaser's release process