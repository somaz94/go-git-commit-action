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
| `auto_branch`       | No       | Whether to create automatic branch | true                          |
| `pr_title`          | No       | Pull request title             | Auto PR by Go Git Commit Action   |
| `pr_base`           | No       | Base branch for pull request   | main                             |
| `pr_branch`         | No       | Branch to create pull request from | -                            |
| `delete_source_branch` | No    | Whether to delete source branch after PR | false                   |
| `github_token`      | No       | GitHub token for PR creation   | -                                |

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

#### Auto Branch (Recommended)
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: true  # Creates timestamped branch automatically
    pr_base: main
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Custom Branch
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    branch: feature/my-branch
    pr_branch: feature/my-branch  # Branch to create PR from
    pr_base: main
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### With Custom PR Title
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_title: feat: my custom PR title
    pr_base: main
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Delete Source Branch After PR
```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: true
    pr_base: main
    delete_source_branch: true
    github_token: ${{ secrets.PAT_TOKEN }}
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

<br/>

## Notes

### General
- The action automatically handles git configuration
- If `repository_path` is not specified, it uses the current directory
- `file_pattern` supports standard git pattern matching
- The action will skip the commit if there are no changes to commit

### Tag Operations
- Tag operations are optional and only executed when `tag_name` is provided
- Use `tag_message` to create annotated tags
- Set `delete_tag: 'true'` to delete a tag both locally and remotely
- Use `tag_reference` to create tags pointing to specific commits, other tags, or branches

### Pull Request Operations
- Set `create_pr: 'true'` to create a pull request
- When `auto_branch` is true (default), it creates a timestamped branch automatically
- `pr_title` can be customized (defaults to "Auto PR: %s to %s (Run ID: %s)", sourceBranch, config.PRBase, runID)
- `pr_base` specifies the target branch for the PR (defaults to main)
- Use `pr_branch` to specify a custom source branch when `auto_branch` is false
- Set `delete_source_branch: 'true'` to automatically delete the source branch after PR is created (only works when `auto_branch` is true)

<br/>

### Authentication
- When working with workflow files or tags, you need to use a Personal Access Token (PAT) with appropriate permissions:
  ```yaml
  - uses: actions/checkout@v4
    with:
      token: ${{ secrets.PAT_TOKEN }}  # Required for tag operations
  ```
  This is especially important when you need to push tags or modify workflow files, as the default GITHUB_TOKEN may not have sufficient permissions.
- The `github_token` input is required for creating pull requests

<br/>

### Branch Operations
- The `branch` input specifies the target branch for commits (defaults to main)
- When using `auto_branch`, a new branch is created with a timestamp format

<br/>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

<br/>

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.