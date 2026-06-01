# Contributing to pgecloudctl

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch: `git checkout -b feat/my-feature`
4. Install pre-commit hooks: `pre-commit install`

## Development

```bash
make build    # Build the binary
make test     # Run tests with race detector
make lint     # Run golangci-lint
make generate # Regenerate API client from OpenAPI spec
```

## Pull Requests

- Use conventional commit messages (`feat:`, `fix:`, `docs:`,
  `chore:`, etc.)
- Ensure `make test` and `make lint` pass
- Add tests for new functionality
- Keep PRs focused — one feature or fix per PR

## Code Style

- `gofmt` mandatory for all Go code
- Follow existing patterns in the codebase
- Table-driven tests preferred

## License

By contributing, you agree that your contributions will be
licensed under the PostgreSQL License.
