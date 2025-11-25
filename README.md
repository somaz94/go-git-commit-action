# Go Git Commit Action

[![License](https://img.shields.io/github/license/somaz94/go-git-commit-action)](https://github.com/somaz94/go-git-commit-action)
![Latest Tag](https://img.shields.io/github/v/tag/somaz94/go-git-commit-action)
![Top Language](https://img.shields.io/github/languages/top/somaz94/go-git-commit-action?color=green&logo=go&logoColor=b)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Go%20Git%20Commit%20Action-blue?logo=github)](https://github.com/marketplace/actions/go-git-commit-action)

A GitHub Action that automates git commit, push, tag, and pull request operations. Written in Go for better performance and reliability.

<br/>

## Features

- **Fast & Reliable** - Written in Go
- **Commit & Push** - Automated git operations
- **Tag Management** - Create and delete tags
- **Pull Requests** - Automated PR creation
- **Flexible Patterns** - Multiple file pattern support
- **Secure** - Built-in authentication handling

<br/>

## Documentation

- **[Examples](docs/EXAMPLES.md)** - Workflow examples and use cases
- **[Configuration](docs/CONFIGURATION.md)** - Complete input reference
- **[Authentication](docs/AUTHENTICATION.md)** - Token setup and best practices
- **[Development](docs/DEVELOPMENT.md)** - Contributing guidelines

<br/>

## Quick Start

<br/>

### Basic Commit

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    commit_message: Auto commit
    branch: main
    github_token: ${{ secrets.GITHUB_TOKEN }}
```

<br/>

### Create Tag

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    tag_name: v1.0.0
    tag_message: Release version 1.0.0
    github_token: ${{ secrets.PAT_TOKEN }}    
```

<br/>

### Create Pull Request

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_branch: feature/my-branch
    pr_base: main
    github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

## Inputs

| Input                | Required | Description                    | Default                           |
|---------------------|----------|--------------------------------|-----------------------------------|
| `user_email`        | Yes      | Git user email                 | -                                 |
| `user_name`         | Yes      | Git user name                  | -                                 |
| `commit_message`    | No       | Commit message                 | Auto commit by Go Git Commit Action |
| `branch`            | No       | Branch to push to              | main                              |
| `repository_path`   | No       | Path to the repository         | .                                 |
| `file_pattern`      | No       | File pattern to add            | .                                 |
| `tag_name`          | No       | Tag name to create or delete   | -                                 |
| `tag_message`       | No       | Tag message (for annotated tags)| -                                |
| `delete_tag`        | No       | Whether to delete the tag      | false                            |
| `tag_reference`     | No       | Git reference for the tag      | -                                |
| `create_pr`         | No       | Whether to create a pull request | false                           |
| `auto_branch`       | No       | Whether to create automatic branch | false                         |
| `pr_title`          | No       | Pull request title             | Auto PR by Go Git Commit Action   |
| `pr_base`           | No       | Base branch for pull request   | main                             |
| `pr_branch`         | No       | Branch to create pull request from | -                            |
| `delete_source_branch` | No    | Whether to delete source branch after PR | false                   |
| `github_token`      | No       | GitHub token for PR creation   | -                                |
| `pr_labels`         | No       | Labels to add to pull request (comma-separated) | -               |
| `pr_body`           | No       | Custom body message for pull request | -                          |
| `skip_if_empty`     | No       | Skip the action if there are no changes | false                   |
| `pr_closed`         | No       | Whether to close the pull request after creation | false          |
| `pr_dry_run`        | No       | Simulate PR creation without actually creating one | false         |

**See [Configuration](docs/CONFIGURATION.md) for detailed descriptions and validation rules.**

<br/>

## Authentication

<br/>

### For checkout@v6 users

```yaml
- uses: actions/checkout@v6

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    github_token: ${{ secrets.PAT_TOKEN }}  # Required!
```

**See [Authentication Guide](docs/AUTHENTICATION.md) for detailed setup.**

<br/>

## Common Use Cases

- Auto-commit generated files
- Create release tags
- Automated documentation updates
- Sync configuration files
- Create pull requests from workflows

**See [Examples](docs/EXAMPLES.md) for more use cases.**

<br/>

## Development

<br/>

### Running Tests

```bash
go test ./...              # Run all tests
go test ./... -cover       # With coverage
```

**See [Development Guide](docs/DEVELOPMENT.md) for contributing.**

<br/>

## License

MIT License - see [LICENSE](LICENSE) file for details.

<br/>

## Contributing

Contributions welcome! See [Development Guide](docs/DEVELOPMENT.md) for details.