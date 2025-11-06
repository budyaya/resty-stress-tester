
# Contributing to Resty-Stress-Tester

Thank you for your interest in contributing to Resty-Stress-Tester! We welcome contributions from the community.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [GitHub Issues](https://github.com/budyaya/resty-stress-tester/issues)
2. If not, create a new issue with the following information:
   - Clear description of the problem
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Check if the feature has already been suggested
2. Create a new issue with:
   - Clear description of the feature
   - Use cases and benefits
   - Proposed implementation (if any)

### Code Contributions

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass: `make test`
6. Run linter: `make lint`
7. Commit your changes: `git commit -m 'Add some feature'`
8. Push to the branch: `git push origin feature/your-feature-name`
9. Create a Pull Request

### Pull Request Guidelines

- Follow the existing code style
- Include tests for new functionality
- Update documentation as needed
- Ensure CI passes
- Keep PRs focused and manageable

## Development Setup

### Prerequisites

- Go 1.19 or later
- Git

### Building from Source

```bash
git clone https://github.com/budyaya/resty-stress-tester
cd resty-stress-tester
make build
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
go test ./test/unit/...

# Run integration tests
go test -tags=integration ./test/integration/...

# Run with coverage
make test-coverage
```

### Code Style

We use `gofmt` for code formatting and `golangci-lint` for linting.

```bash
# Format code
make fmt

# Run linter
make lint
```

## Project Structure

```
internal/
├── config/     # Configuration management
├── engine/     # Core stress testing engine
├── parser/     # Data parsing (CSV, templates)
├── reporter/   # Report generation
└── util/       # Utility functions

pkg/
├── types/      # Shared type definitions
└── version/    # Version information
```

## Documentation

- Update relevant documentation when making changes
- Add comments for public functions and types
- Update examples if needed

## Release Process

1. Update version in `pkg/version/version.go`
2. Update `CHANGELOG.md`
3. Create a git tag: `git tag v0.1.0`
4. Push tag: `git push origin v0.1.0`
5. GitHub Actions will automatically create a release

## Questions?

Feel free to open an issue or contact the maintainers if you have any questions.

Thank you for contributing! 🎉
