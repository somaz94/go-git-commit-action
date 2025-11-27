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

| Package | Coverage | Notes |
|---------|----------|-------|
| `config` | 97.8% | Configuration parsing and validation |
| `errors` | 100% | Custom error types (GitError, ConfigError, RetryError, APIError) |
| `executor` | 98.2% | Command execution abstraction and mocking |
| `gitcmd` | 100% | Git command builders |
| `cmd` | 0% | Entry point (tested via integration tests) |
| `git` | 0% | Git operations (tested via CI workflows) |
| `git/pr` | 0% | PR operations (tested via CI workflows) |

**Why some packages have 0% coverage:**
- **cmd/main.go**: Application entry point with signal handling - tested through integration tests in CI
- **internal/git**: Core git operations with heavy external dependencies (git commands, filesystem) - validated through real workflow execution
- **internal/git/pr**: GitHub API interactions requiring network calls - tested end-to-end in CI workflows

**Testing Strategy:**
- **Unit Tests** (High Coverage): Pure logic without external dependencies
- **Integration Tests** (CI/CD): Components requiring git, filesystem, or API interactions

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

---

## Improving Test Coverage

<br/>

### Coverage Analysis

<br/>

#### Finding Low Coverage Areas

```bash
# Generate detailed coverage report
go test -v -coverprofile=coverage.out ./internal/executor

# View function-level coverage
go tool cover -func=coverage.out | grep executor
```

**Example Output:**
```
executor.go:44:    ExecuteWithStreams      0.0%
mock_executor.go:54:  ExecuteWithOutput    71.4%
mock_executor.go:73:  ExecuteWithStreams   70.0%
mock_executor.go:120: GetLastCommand       66.7%
```

<br/>

#### Common Coverage Issues and Solutions

**1. Uncovered Error Paths**

```go
// Before: Only testing success case (71.4% coverage)
func TestMockExecutor_ExecuteWithOutput(t *testing.T) {
    mock := NewMockExecutor()
    mock.SetOutput([]byte("output"), "git", "log")
    
    output, err := mock.ExecuteWithOutput("git", "log")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}

// After: Test success + error + default cases (100% coverage)
func TestMockExecutor_ExecuteWithOutput(t *testing.T) {
    mock := NewMockExecutor()
    
    // Test configured output
    mock.SetOutput([]byte("output"), "git", "log")
    output, err := mock.ExecuteWithOutput("git", "log")
    // ... assertions
    
    // Test error case
    mock.Reset()
    expectedErr := errors.New("output error")
    mock.SetError(expectedErr, "git", "status")
    _, err = mock.ExecuteWithOutput("git", "status")
    // ... assertions
    
    // Test unconfigured command (default behavior)
    mock.Reset()
    output, err = mock.ExecuteWithOutput("git", "branch")
    // ... assertions
}
```

**2. Missing Nil Checks**

```go
// Before: GetLastCommand test causing panic (66.7% coverage)
func TestMockExecutor_GetLastCommand(t *testing.T) {
    mock := NewMockExecutor()
    
    // This caused panic - cmd was nil!
    cmd := mock.GetLastCommand()
    if cmd.Name != "" {  // ❌ Panic: nil pointer dereference
        t.Error("Expected empty")
    }
}

// After: Proper nil handling (100% coverage)
func TestMockExecutor_GetLastCommand(t *testing.T) {
    mock := NewMockExecutor()
    
    // Test with no commands
    cmd := mock.GetLastCommand()
    if cmd != nil {  // ✅ Check for nil first
        t.Errorf("Expected nil, got %v", cmd)
    }
    
    // Test with commands
    mock.Execute("git", "status")
    cmd = mock.GetLastCommand()
    if cmd == nil {
        t.Fatal("Expected command, got nil")
    }
    // ... assertions
}
```

**3. Untested Stream Operations**

```go
// Before: RealExecutor_ExecuteWithStreams not tested (0% coverage)
// No test existed for this function

// After: Add stream testing (100% coverage)
func TestRealExecutor_ExecuteWithStreams(t *testing.T) {
    executor := NewRealExecutor()
    var stdout, stderr bytes.Buffer
    
    // Test successful command
    err := executor.ExecuteWithStreams("echo", []string{"hello"}, &stdout, &stderr)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if stdout.Len() == 0 {
        t.Error("Expected stdout output")
    }
    
    // Test failing command
    stdout.Reset()
    stderr.Reset()
    err = executor.ExecuteWithStreams("false", []string{}, &stdout, &stderr)
    if err == nil {
        t.Error("Expected error for 'false' command")
    }
}
```

**4. Missing Edge Cases**

```go
// Before: Only testing main flow (70% coverage)
func TestMockExecutor_ExecuteWithStreams(t *testing.T) {
    mock := NewMockExecutor()
    var stdout bytes.Buffer
    
    mock.SetStreamOutput("output", "git", "diff")
    err := mock.ExecuteWithStreams("git", []string{"diff"}, &stdout, nil)
    // ... assertions
}

// After: Test all branches (100% coverage)
func TestMockExecutor_ExecuteWithStreams(t *testing.T) {
    mock := NewMockExecutor()
    var stdout bytes.Buffer
    
    // Test configured stream output
    mock.SetStreamOutput("output", "git", "diff")
    err := mock.ExecuteWithStreams("git", []string{"diff"}, &stdout, nil)
    // ... assertions
    
    // Test error case
    mock.Reset()
    expectedErr := errors.New("stream error")
    mock.SetError(expectedErr, "git", "push")
    err = mock.ExecuteWithStreams("git", []string{"push"}, &stdout, nil)
    // ... assertions
    
    // Test unconfigured command
    mock.Reset()
    stdout.Reset()
    err = mock.ExecuteWithStreams("git", []string{"status"}, &stdout, nil)
    // ... assertions
}
```

<br/>

### Coverage Improvement Workflow

1. **Identify low coverage areas:**
   ```bash
   go test -coverprofile=coverage.out ./internal/executor
   go tool cover -func=coverage.out | grep -E "^.*\s+[0-8][0-9]\.[0-9]%$"
   ```

2. **Analyze uncovered lines:**
   ```bash
   go tool cover -html=coverage.out
   # Opens browser showing covered (green) vs uncovered (red) lines
   ```

3. **Add missing test cases:**
   - Error paths
   - Edge cases (nil, empty, boundary values)
   - Default behaviors
   - All conditional branches

4. **Verify improvement:**
   ```bash
   go test -cover ./internal/executor
   # Should show higher percentage
   ```

<br/>

### Example: Improving executor Coverage (82.5% → 98.2%)

#### Steps taken:

1. **Identified gaps** using `go tool cover -func`:
   - `ExecuteWithStreams` (RealExecutor): 0% → Added test
   - `ExecuteWithOutput` (MockExecutor): 71.4% → Added error/default cases
   - `ExecuteWithStreams` (MockExecutor): 70% → Added error/unconfigured cases
   - `GetLastCommand`: 66.7% → Fixed nil handling

2. **Added comprehensive tests:**
   - Created `TestRealExecutor_ExecuteWithStreams`
   - Enhanced `TestMockExecutor_ExecuteWithOutput` with 3 scenarios
   - Enhanced `TestMockExecutor_ExecuteWithStreams` with 3 scenarios
   - Fixed `TestMockExecutor_GetLastCommand` nil check
   - Added `TestMockExecutor_MultipleCommands` for integration

3. **Result:** 82.5% → 98.2% (+15.7 percentage points)

---

## Code Quality Improvements

<br/>

### Refactoring for Better Testability

#### 1. Error Handling Consistency

```go
// Before: Mixed error handling styles
return fmt.Errorf("operation failed: %v", err)
return errors.New("failed", err)

// After: Consistent custom error types
return errors.New("operation", err)
return errors.NewWithPath("operation", path, err)
return errors.NewConfig("validation message")
return errors.NewWithContext("retry failed", attempts, lastErr)
```

#### 2. Magic Numbers → Constants

```go
// Before: Hard-coded values
time.Sleep(time.Second * time.Duration(i+1))
os.MkdirAll(dir, 0755)
os.WriteFile(path, content, 0644)

// After: Named constants
const (
    permDir  = 0755
    permFile = 0644
    retryBaseDelay = time.Second
)

time.Sleep(retryBaseDelay * time.Duration(i+1))
os.MkdirAll(dir, permDir)
os.WriteFile(path, content, permFile)
```

#### 3. Error Types Enhancement

Added new error types for better error handling:

```go
// RetryError for retry failures
type RetryError struct {
    Attempts int
    LastErr  error
    Message  string
}

// NewConfig for configuration errors without field
func NewConfig(message string) *ConfigError {
    return &ConfigError{Message: message}
}

// NewWithContext for retry context
func NewWithContext(message string, attempts int, lastErr error) *RetryError {
    return &RetryError{
        Message:  message,
        Attempts: attempts,
        LastErr:  lastErr,
    }
}
```
