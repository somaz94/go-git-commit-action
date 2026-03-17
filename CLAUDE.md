# CLAUDE.md - go-git-commit-action

GitHub Action to automatically commit, push, tag, and create pull requests using Go.

## Commit Guidelines

- Do not include `Co-Authored-By` lines in commit messages.

## Project Structure

```
cmd/main.go                  # Entrypoint
internal/
  config/                    # Input parsing & validation (GitConfig struct)
  errors/                    # Custom error types (GitError, APIError)
  executor/                  # Command orchestration
  git/                       # Git operations (commit, tag, file ops)
    pr/                      # PR operations (branch, diff, creation)
  gitcmd/                    # Git command argument builders
test/                        # Integration test data (test/ not tests/)
Makefile                     # Build, test, lint commands
Dockerfile                   # Multi-stage (golang:1.26-alpine → alpine:latest)
action.yml                   # GitHub Action definition (17+ inputs)
cliff.toml                   # git-cliff config for release notes
```

## Build & Test

```bash
make test            # Run unit tests (alias for test-unit)
make test-unit       # go test ./internal/... ./cmd/... -v -cover
make test-all        # Run all tests
make cover           # Generate coverage report
make cover-html      # Open coverage in browser
make bench           # Run benchmarks
make lint            # go vet
make fmt             # gofmt
make build           # Build binary
make clean           # Remove artifacts
```

## Key Inputs

- **Required**: `user_email`, `user_name`
- **Commit**: `commit_message`, `branch`, `repository_path`, `file_pattern`, `skip_if_empty`
- **Tag**: `tag_name`, `tag_message`, `tag_reference`, `delete_tag`
- **PR**: `create_pr`, `auto_branch`, `pr_title`, `pr_base`, `pr_branch`, `pr_body`, `pr_labels`, `pr_closed`, `pr_dry_run`, `delete_source_branch`
- **Auth**: `github_token`

## Workflow Structure

| Workflow | Name | Trigger |
|----------|------|---------|
| `ci.yml` | `Continuous Integration` | push(main), PR, dispatch |
| `release.yml` | `Create release` | tag push `v*` |
| `changelog-generator.yml` | `Generate changelog` | after release, PR merge, issue close |
| `use-action.yml` | `Smoke Test (Released Action)` | after release, dispatch |
| `linter.yml` | `Lint Codebase` | push(main), PR |

### Workflow Chain
```
tag push v* → Create release
                ├→ Smoke Test (Released Action)
                └→ Generate changelog
```

### CI Structure
```
unit-tests → build-and-push-docker → test-auto-branch-false → test-auto-branch-true
               → test-skip-if-empty → test-pr-auto-close → test-pr-dry-run → ci-result
```

## Testing Notes

- Many functions use `exec.Command` for git/curl calls, limiting pure unit test coverage
- Dry run paths and config validation are the most testable areas
- Integration tests in ci.yml use `uses: ./` (local action)
- Smoke tests in use-action.yml use `somaz94/go-git-commit-action@v1` (released)
- Test directory is `test/` (singular), used for integration test data

## Conventions

- **Commits**: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`, `chore:`)
- **Branches**: `main` (production), `test` (integration tests)
- **Secrets**: `PAT_TOKEN` (cross-repo ops, tag/branch operations), `GITHUB_TOKEN` (changelog, releases)
- **Docker**: Multi-stage build, alpine base
- **Comments**: English only
- **Release**: `git switch` (not `git checkout`), git-cliff for RELEASE.md
- **cliff.toml**: Skip `^Merge`, `^Update changelog`, `^Auto commit`
- **paths-ignore**: `.github/workflows/**`, `**/*.md`, `test/**`, `backup/**`
- Do NOT commit directly - recommend commit messages only

## Language

- Communicate with the user in Korean.
