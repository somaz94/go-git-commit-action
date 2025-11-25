# Development Guide

Guide for developers who want to contribute to or modify the Go Git Commit Action.

<br/>

## Table of Contents

- [Project Structure](#project-structure)
- [Setup](#setup)
- [Testing](#testing)
- [Building](#building)
- [Docker](#docker)
- [Contributing](#contributing)

---

## Project Structure

```
.
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── errors/                 # Custom error types
│   │   ├── errors.go
│   │   └── errors_test.go
│   ├── executor/               # Command executor interface
│   │   ├── executor.go
│   │   ├── mock_executor.go
│   │   └── executor_test.go
│   ├── git/                    # Git operations
│   │   ├── commit.go
│   │   ├── common.go
│   │   ├── common_test.go
│   │   ├── file_operations.go
│   │   ├── pr.go               # PR orchestration
│   │   ├── tag.go
│   │   └── pr/                 # PR modules
│   │       ├── branch.go       # Branch management
│   │       ├── creation.go     # PR creation & API
│   │       └── diff.go         # Change detection
│   └── gitcmd/                 # Git command builders
│       ├── commands.go
│       └── commands_test.go
├── docs/                       # Documentation
├── test/                       # Test data
├── .dockerignore              # Docker build exclusions
├── .gitignore
├── action.yml                  # GitHub Action metadata
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## Setup

<br/>

### Prerequisites

- Go 1.23 or later
- Git
- Docker (for testing containerized build)

<br/>

### Clone and Install

```bash
# Clone the repository
git clone https://github.com/somaz94/go-git-commit-action.git
cd go-git-commit-action

# Download dependencies
go mod download

# Verify installation
go version
```

---

## Testing

<br/>

### Run All Tests

```bash
# Basic test run
go test ./...

# Verbose output
go test ./... -v

# With coverage
go test ./... -cover
```

<br/>

### Run Specific Package Tests

```bash
# Test executor package
go test ./internal/executor/... -v

# Test config package
go test ./internal/config/... -v

# Test errors package
go test ./internal/errors/... -v

# Test gitcmd package
go test ./internal/gitcmd/... -v
```

<br/>

### Coverage Report

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out
```

### Current Test Coverage

| Package | Coverage |
|---------|----------|
| `config` | 97.8% |
| `errors` | 100% |
| `executor` | 82.5% |
| `gitcmd` | 100% |

---

## Building

<br/>

### Build Binary

```bash
# Build for current platform
go build -o go-git-commit-action ./cmd/main.go

# Build for Linux (for Docker)
GOOS=linux GOARCH=amd64 go build -o go-git-commit-action ./cmd/main.go
```

<br/>

### Build with Optimizations

```bash
# Build with size optimization
go build -ldflags="-s -w" -o go-git-commit-action ./cmd/main.go

# Build for specific platform
GOOS=linux GOARCH=arm64 go build -o go-git-commit-action ./cmd/main.go
```

---

## Docker

<br/>

### Build Docker Image

```bash
# Build image
docker build -t go-git-commit-action .

# Build with specific tag
docker build -t go-git-commit-action:v1.0.0 .
```

<br/>

### Test Docker Image

```bash
# Run with environment variables
docker run --rm \
  -e INPUT_USER_EMAIL="test@example.com" \
  -e INPUT_USER_NAME="Test User" \
  -e INPUT_COMMIT_MESSAGE="Test commit" \
  go-git-commit-action
```

<br/>

### Docker Optimization

The `.dockerignore` file excludes:
- Test files (`*_test.go`)
- Test data (`test/`)
- Coverage reports
- Documentation
- Development files

This reduces image size and build time.

---

## Code Organization

<br/>

### Package Responsibilities

**cmd/main.go**
- Application entry point
- Signal handling
- High-level orchestration

**internal/config**
- Environment variable parsing
- Configuration validation
- Default values

**internal/errors**
- Structured error types
- Error wrapping and unwrapping
- Error context

**internal/executor**
- Command execution abstraction
- Mock executor for testing
- Interface for dependency injection

**internal/git**
- Git operations (commit, tag)
- File operations
- PR orchestration

**internal/git/pr**
- Branch management (`branch.go`)
- PR creation and GitHub API (`creation.go`)
- Change detection (`diff.go`)

**internal/gitcmd**
- Git command building
- Argument construction
- Command constants

---

## Testing Guidelines

<br/>

### Unit Tests

- Test each function independently
- Use table-driven tests for multiple scenarios
- Mock external dependencies using `executor.MockExecutor`

<br/>

### Example Test

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  *GitConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: &GitConfig{
                CreatePR: true,
                PRBranch: "feature",
                PRBase: "main",
                GitHubToken: "token",
            },
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Contributing

<br/>

### Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

<br/>

### Code Style

- Follow standard Go formatting (`gofmt`)
- Write clear, descriptive commit messages
- Add comments for exported functions
- Keep functions focused and small

<br/>

### Pull Request Checklist

- [ ] Tests added/updated
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Code formatted (`gofmt`)
- [ ] No breaking changes (or clearly documented)

---

## Debugging

<br/>

### Enable Debug Mode

```bash
# Set debug environment variable
export INPUT_DEBUG=true

# Run the action
go run ./cmd/main.go
```

<br/>

### Common Issues

**Issue:** Import cycle errors
- **Solution:** Avoid circular dependencies between packages

**Issue:** Test failures in CI
- **Solution:** Ensure environment variables are properly mocked

**Issue:** Docker build fails
- **Solution:** Check `.dockerignore` and ensure necessary files are included

---

## Release Process

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create and push git tag
4. GitHub Actions will automatically build and publish

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```
