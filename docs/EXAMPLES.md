# Example Workflows

This document contains detailed examples of how to use the Go Git Commit Action in various scenarios.

<br/>

## Table of Contents

- [Basic Commit](#basic-commit)
- [Tag Management](#tag-management)
  - [Creating Tags](#creating-tags)
  - [Deleting Tags](#deleting-tags)
  - [Tags with References](#tags-with-references)
- [Pull Requests](#pull-requests)
  - [Custom Branch PR](#custom-branch-pr)
  - [Auto Branch PR](#auto-branch-pr)
  - [PR with Labels and Custom Body](#pr-with-labels-and-custom-body)
  - [Advanced PR Options](#advanced-pr-options)
- [File Patterns](#file-patterns)

---

## Basic Commit

Simple commit and push workflow:

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

---

## Tag Management

<br/>

### Creating Tags

#### Simple Tag

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

### Deleting Tags

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

### Tags with References

Create tags pointing to specific commits, other tags, or branches:

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
      
      # Tag specific commit
      - name: Create Git Tag at Commit
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: v1
          tag_reference: ${{ github.sha }}
          
      # Tag from another tag
      - name: Create Git Tag from Tag
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: latest
          tag_reference: v1.0.2
          
      # Tag from branch
      - name: Create Git Tag from Branch
        uses: somaz94/go-git-commit-action@v1
        with:
          user_email: actions@github.com
          user_name: GitHub Actions
          tag_name: stable
          tag_reference: main
```

---

## Pull Requests

<br/>

### Custom Branch PR

Use an existing branch for PR:

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: false
    pr_branch: feature/my-branch
    pr_base: main
    github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### Auto Branch PR

Automatically create a timestamped branch:

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: true
    pr_branch: feature/my-branch  # Base for auto-generated branch
    pr_base: main
    github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### PR with Labels and Custom Body

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_branch: feature/my-branch
    pr_base: main
    pr_title: "feat: my custom PR title"
    pr_labels: "enhancement,automated,test"
    pr_body: |
      ## Custom Pull Request
      This PR was automatically created.
      
      ### Changes
      - Feature 1
      - Feature 2
    github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### Advanced PR Options

#### With Auto Branch and Delete Source Branch

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    auto_branch: true
    pr_branch: feature/my-branch
    pr_base: main
    delete_source_branch: true
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Skip If No Changes

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_branch: feature/my-branch
    pr_base: main
    skip_if_empty: true
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### Auto Close PR

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_branch: feature/my-branch
    pr_base: main
    pr_closed: true
    pr_title: "Auto Close PR Example"
    pr_body: "This PR will be automatically closed"
    github_token: ${{ secrets.PAT_TOKEN }}
```

#### PR Dry Run

Test PR creation without actually creating one:

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    create_pr: true
    pr_branch: feature/my-branch
    pr_base: main
    pr_title: "Test PR Dry Run"
    pr_labels: "test,automated"
    pr_dry_run: true
    github_token: ${{ secrets.PAT_TOKEN }}
```

---

## File Patterns

<br/>

### Multiple File Patterns

Commit multiple file types:

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    commit_message: "chore: update multiple file types"
    branch: main
    file_pattern: "docs/*.md src/*.go *.yaml"
```

<br/>

### Complex Patterns with Different Directories

```yaml
- uses: somaz94/go-git-commit-action@v1
  with:
    user_email: actions@github.com
    user_name: GitHub Actions
    commit_message: "chore: update configuration and documentation"
    branch: main
    repository_path: "."
    file_pattern: "config/*.yaml docs/*.md src/util/*.js"
```
