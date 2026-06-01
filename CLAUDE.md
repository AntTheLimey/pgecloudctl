# pgecloudctl

Go CLI for managing pgEdge Cloud resources via the REST API.

## Build & Test

- `make build` — build the binary
- `make test` — run all tests with race detector
- `make lint` — run golangci-lint
- `make generate` — regenerate API client from OpenAPI spec

## Project Structure

- `cmd/pgecloudctl/` — entry point
- `internal/api/` — generated HTTP client (do not edit by hand)
- `internal/auth/` — token fetch, cache, refresh
- `internal/cmd/` — Cobra command tree
- `internal/config/` — config file read/write
- `internal/output/` — table/JSON/YAML formatters
- `openapi/` — vendored OpenAPI spec + codegen config

## Standards

- gofmt mandatory
- golangci-lint must pass
- Tests required for new functionality
- Table-driven tests preferred
- 4-space indentation in non-Go files
- 79-character line wrapping in markdown
