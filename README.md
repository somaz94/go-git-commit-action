# Go Git Commit Action

[![License](https://img.shields.io/github/license/somaz94/go-git-commit-action)](https://github.com/somaz94/go-git-commit-action)
![Latest Tag](https://img.shields.io/github/v/tag/somaz94/go-git-commit-action)
![Top Language](https://img.shields.io/github/languages/top/somaz94/go-git-commit-action?color=green&logo=go&logoColor=b)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Go%20Git%20Commit%20Action-blue?logo=github)](https://github.com/marketplace/actions/go-git-commit-action)

## Overview

The **Go Git Commit Action** is a GitHub Action that automates git commit, push, and tag operations. Written in Go, it provides a reliable and efficient way to commit changes and manage tags in your GitHub Actions workflows. This action is particularly useful for workflows that need to automatically commit changes and manage tags in repositories.

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

<br/>

## Example Workflows

<br/>

### Basic Commit Example

Below is an example of how to use the **Go Git Commit Action** in your GitHub Actions workflow:

```yaml
name: Auto Commit Workflow
on: [push]

permissions:
  contents: write

jobs:
  commit:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Auto Commit Changes
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          commit_message: Auto commit by GitHub Actions
          branch: main
          repository_path: path/to/repo  # Optional
          file_pattern: '*.md'             # Example: commit only markdown files
```

<br/>

### Creating a Tag

```yaml
name: Create Tag
on: [workflow_dispatch]

permissions:
  contents: write

jobs:
  tag:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.PAT_TOKEN }}
      
      - name: Create Git Tag
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: v1.0.0
          tag_message: Release version 1.0.0  # Optional for annotated tags
```

<br/>

### Deleting a Tag

```yaml
name: Delete Tag
on: [workflow_dispatch]

permissions:
  contents: write

jobs:
  delete-tag:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.PAT_TOKEN }}
      
      - name: Delete Git Tag
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: v1.0.0
          delete_tag: true
```

<br/>

### Creating a Tag with Reference

```yaml
name: Create Tag with Reference
on: [workflow_dispatch]

permissions:
  contents: write

jobs:
  tag:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      # Create tag pointing to specific commit
      - name: Create Git Tag at Commit
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: v1
          tag_reference: ${{ github.sha }}  # Points to specific commit SHA
          
      # Create tag pointing to another tag
      - name: Create Git Tag from Tag
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: latest
          tag_reference: v1.0.2  # Points to existing tag
          
      # Create tag pointing to branch
      - name: Create Git Tag from Branch
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: stable
          tag_reference: main  # Points to branch
```

<br/>

### Create Pull Request

#### Using Custom Branch (Default)
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false  # Default value
    branch: feature/my-branch # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Source branch for PR
    pr_base: main                 # Required: Target branch for PR
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Using Auto Branch (Creates timestamped branch from pr_branch)
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: true   # Will create 'update-files-{timestamp}' branch from pr_branch
    branch: main # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Base branch for auto-generated branch
    pr_base: main                 # Required: Target branch for PR
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### With Custom PR Title and Labels
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    branch: feature/my-branch # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Source branch for PR
    pr_base: main                 # Required: Target branch for PR
    pr_title: "feat: my custom PR title"
    pr_labels: "enhancement,automated,test"
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### With Auto Branch and Delete Source Branch
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: true   # Will create 'update-files-{timestamp}' branch
    branch: main # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Base branch for auto-generated branch
    pr_base: main                 # Required: Target branch for PR
    delete_source_branch: true    # Will delete the auto-generated branch after PR
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### With Labels and Custom Body
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    branch: main # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Source branch for PR
    pr_base: main                 # Required: Target branch for PR
    pr_labels: "enhancement,automated,test"
    pr_body: |
      ## Custom Pull Request
      This PR was automatically created with custom labels and body.
      
      ### Changes
      - Feature 1
      - Feature 2
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Skip If Empty Changes
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    branch: feature/my-branch # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Source branch for PR
    pr_base: main                 # Required: Target branch for PR
    skip_if_empty: true          # Skips PR creation if no changes detected
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Auto Close Pull Request
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    branch: main # Not Required: Push branch for commit and tag (Default: main)
    pr_branch: feature/my-branch  # Required: Source branch for PR
    pr_base: main                 # Required: Target branch for PR
    pr_closed: true              # Automatically closes PR after creation
    pr_title: "Auto Close PR Example"
    pr_body: "This PR will be automatically closed after creation"
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### PR Dry Run (Simulation Mode)
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    branch: main
    pr_branch: feature/my-branch  # Required: Source branch for PR
    pr_base: main                 # Required: Target branch for PR
    pr_title: "Test PR Dry Run"
    pr_labels: "test,automated,dry-run-test"
    repository_path: "path/to/repo"
    file_pattern: "*.txt"
    github_token: ${{ secrets.PAT_TOKEN }}
    pr_dry_run: true             # Simulates PR creation without executing it
```

#### Using Multiple File Patterns
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    commit_message: "chore: update multiple file types"
    branch: main
    repository_path: "."
    file_pattern: "docs/*.md src/*.go *.yaml"  # Multiple file patterns separated by spaces
```

##### Complex Example with Different Directories
```yaml
- name: Commit Changes to Specific File Types
  uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    commit_message: "chore: update configuration and documentation"
    branch: main
    repository_path: "."
    file_pattern: "config/*.yaml docs/*.md src/util/*.js"  # Multiple patterns for different directories
```

<br/>

## Features

- Written in Go for better performance and reliability
- Supports custom git user configuration
- Flexible file pattern matching
- Optional repository path specification
- Automatic branch pushing
- Detailed error reporting
- Git tag management (create and delete)
- Support for both lightweight and annotated tags
- Automatic tag pushing to remote
- Support for creating tags pointing to specific commits, tags, or branches
- Flexible tag reference system
- Support for PR labels and custom body messages
- Skip action when no changes are detected
- Auto-close PR functionality
- Detailed change detection between branches

<br/>

## Notes

### General
- The action automatically handles git configuration
- If `repository_path` is not specified, it uses the current directory
- `file_pattern` supports standard git pattern matching
- The action will skip the commit if there are no changes to commit

### Branch Operations
- The `branch` parameter is used for simple commit and tag operations (without PR)
- For pull request operations:
  - `pr_branch`: Specifies the source branch for the PR
  - `pr_base`: Specifies the target branch for the PR
  - When `auto_branch: true`, a new branch is created with format 'update-files-{timestamp}' based on `pr_branch`
  - When `auto_branch: false` (default), the PR is created directly from `pr_branch` to `pr_base`

### Tag Operations
- Tag operations are optional and only executed when `tag_name` is provided
- Use `tag_message` to create annotated tags
- Set `delete_tag: 'true'` to delete a tag both locally and remotely
- Use `tag_reference` to create tags pointing to specific commits, other tags, or branches
- The `branch` parameter is used to specify which branch to tag

### Pull Request Operations
- Set `create_pr: 'true'` to create a pull request
- `pr_branch` and `pr_base` are required when creating a PR:
  - `pr_branch`: The source branch containing your changes
  - `pr_base`: The target branch where changes will be merged
- When `auto_branch: true`:
  - Creates a new timestamped branch from `pr_branch`
  - The new branch name format is 'update-files-{timestamp}'
  - This is useful for automated updates that need unique branch names
- `pr_title` can be customized (defaults to "Auto PR: %s to %s (Run ID: %s)")
- Use `pr_labels` to add labels to the PR (comma-separated)
- Use `pr_body` to set a custom PR description
- Enable `skip_if_empty` to skip the action when no changes are detected
- Set `pr_closed: 'true'` to automatically close the PR after creation
- Set `delete_source_branch: 'true'` to automatically delete the source branch after PR is created (only works with `auto_branch: true`)
- Use `pr_dry_run: 'true'` to simulate PR creation without actually creating one, useful for testing or previewing what will be submitted

### Multiple File Pattern Support
- The action supports space-separated file patterns 
- Example: `file_pattern: "*.md *.txt"` will add all markdown and text files
- This is useful when you need to commit specific file types from different directories

<br/>

## Authentication

- When working with workflow files or tags, you need to use a Personal Access Token (PAT) with appropriate permissions:
  ```yaml
  - uses: actions/checkout@v4
    with:
      token: ${{ secrets.PAT_TOKEN }}  # Required for tag operations
  ```
  This is especially important when you need to push tags or modify workflow files, as the default GITHUB_TOKEN may not have sufficient permissions.
- The `github_token` input is required for creating pull requests

### Important: GitHub Token Required for checkout@v6

Starting from `actions/checkout@v6`, the action stores Git credentials in `$RUNNER_TEMP` directory, which is not accessible from Docker containers. To ensure proper authentication when using this action, you **must** provide the `github_token` input:

```yaml
- uses: actions/checkout@v6
  with:
    token: ${{ secrets.PAT_TOKEN }}  # or ${{ secrets.GITHUB_TOKEN }}

- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    github_token: ${{ secrets.PAT_TOKEN }}  # Required for checkout@v6 compatibility
```

**Why is this needed?**
- `checkout@v6` changed how credentials are stored for security improvements
- Docker containers cannot access the `$RUNNER_TEMP` directory where credentials are stored
- This action configures Git credentials using the provided token to enable push operations

**Recommendation:**
- Always include `github_token` when using `actions/checkout@v6`
- Use `PAT_TOKEN` for operations requiring elevated permissions (tags, workflow modifications)
- Use `GITHUB_TOKEN` for standard commit/push operations within the same repository

<br/>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

<br/>

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.