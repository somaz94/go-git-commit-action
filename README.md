# Go Git Commit Action

[![License](https://img.shields.io/github/license/somaz94/go-git-commit-action)](https://github.com/somaz94/go-git-commit-action)
![Latest Tag](https://img.shields.io/github/v/tag/somaz94/go-git-commit-action)
![Top Language](https://img.shields.io/github/languages/top/somaz94/go-git-commit-action?color=green&logo=go&logoColor=b)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Go%20Git%20Commit%20Action-blue?logo=github)](https://github.com/marketplace/actions/go-git-commit-action)

## Overview

The **Go Git Commit Action** is a GitHub Action that automates git commit, push, and tag operations. Written in Go, it provides a reliable and efficient way to commit changes and manage tags in your GitHub Actions workflows. This action is particularly useful for workflows that need to automatically commit changes and manage tags in repositories.

## Inputs

| Input             | Required | Description                    | Default                           |
|-------------------|----------|--------------------------------|-----------------------------------|
| `user_email`      | Yes      | Git user email                 | -                                 |
| `user_name`       | Yes      | Git user name                  | -                                 |
| `commit_message`  | No       | Commit message                 | Auto commit by Go Git Commit Action |
| `branch`          | No       | Branch to push to              | main                              |
| `repository_path` | No       | Path to the repository         | .                                 |
| `file_pattern`    | No       | File pattern to add            | .                                 |
| `tag_name`        | No       | Tag name to create or delete   | -                                 |
| `tag_message`     | No       | Tag message (for annotated tags)| -                                |
| `delete_tag`      | No       | Whether to delete the tag      | false                            |

## Example Workflows

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
          user_email: 'github-actions@github.com'
          user_name: 'GitHub Actions'
          commit_message: 'Auto commit by GitHub Actions'
          branch: 'main'
          repository_path: 'path/to/repo'  # Optional
          file_pattern: '*.md'             # Example: commit only markdown files

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
      
      - name: Create Git Tag
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: 'github-actions@github.com'
          user_name: 'GitHub Actions'
          tag_name: 'v1.0.0'
          tag_message: 'Release version 1.0.0'  # Optional for annotated tags

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
      
      - name: Delete Git Tag
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: 'github-actions@github.com'
          user_name: 'GitHub Actions'
          tag_name: 'v1.0.0'
          delete_tag: 'true'

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

## Notes

- The action automatically handles git configuration
- If `repository_path` is not specified, it uses the current directory
- `file_pattern` supports standard git pattern matching
- The action will skip the commit if there are no changes to commit
- Tag operations are optional and only executed when `tag_name` is provided
- Use `tag_message` to create annotated tags
- Set `delete_tag: 'true'` to delete a tag both locally and remotely

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.