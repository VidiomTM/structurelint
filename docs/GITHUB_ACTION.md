# Using structurelint in GitHub Actions

Integrate structurelint into your CI/CD pipeline to automatically check project structure on every pull request and push.

## Quick Start

Create `.github/workflows/structurelint.yml` in your repository:

```yaml
name: structurelint

on:
  pull_request:
  push:
    branches: [main, master]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v6

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install structurelint
        run: go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest

      - name: Run structurelint
        run: structurelint
```

## Configuration

### Basic Configuration

Ensure you have `.structurelint.yml` in your repository root:

```yaml
root: true

rules:
  max-depth:
    max: 5

  naming-convention:
    pattern: snake_case
```

### Advanced Configuration

#### Run on specific events

```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened]
  push:
    branches:
      - main
      - develop
      - 'release/**'
```

#### Run only when relevant files change

```yaml
on:
  pull_request:
    paths:
      - 'src/**'
      - 'internal/**'
      - '.structurelint.yml'
```

#### Use specific structurelint version

```yaml
- name: Install structurelint
  run: go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@v0.1.0
```

#### Cache Go modules for faster builds

```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version: '1.24'
    cache: true

- name: Cache structurelint
  uses: actions/cache@v5
  with:
    path: ~/go/bin/structurelint
    key: ${{ runner.os }}-structurelint-${{ hashFiles('**/go.sum') }}
```

## Examples

### Example 1: Basic Workflow

```yaml
name: Structure Linting

on: [pull_request, push]

jobs:
  structurelint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install and run structurelint
        run: |
          go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
          structurelint
```

### Example 2: Multiple Checks

```yaml
name: Code Quality

on:
  pull_request:
  push:
    branches: [main]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install tools
        run: |
          go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

      - name: Run linters
        run: |
          golangci-lint run
          structurelint

      - name: Run tests
        run: go test ./...
```

### Example 3: Matrix Strategy (Multiple Projects)

```yaml
name: Multi-project Linting

on: [pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        project: [frontend, backend, shared]
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install structurelint
        run: go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest

      - name: Lint ${{ matrix.project }}
        working-directory: ./${{ matrix.project }}
        run: structurelint
```

### Example 4: Custom Configuration Path

```yaml
name: structurelint

on: [pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install structurelint
        run: go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest

      - name: Run structurelint with custom config
        run: structurelint --config .structurelint.custom.yml
```

### Example 5: Fail on Warnings

```yaml
name: Strict Linting

on: [pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install structurelint
        run: go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest

      - name: Run structurelint (fail on any violation)
        run: |
          structurelint
          if [ $? -ne 0 ]; then
            echo "❌ Structure violations detected"
            exit 1
          fi
```

### Example 6: PR Comments (Advanced)

Post violations as PR comments:

```yaml
name: structurelint with PR comments

on:
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Install structurelint
        run: go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest

      - name: Run structurelint
        id: lint
        continue-on-error: true
        run: |
          structurelint > lint-output.txt 2>&1
          echo "exit_code=$?" >> $GITHUB_OUTPUT

      - name: Comment on PR
        if: steps.lint.outputs.exit_code != '0'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const output = fs.readFileSync('lint-output.txt', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## ⚠️ structurelint violations\n\n\`\`\`\n${output}\n\`\`\``
            });

      - name: Fail if violations
        if: steps.lint.outputs.exit_code != '0'
        run: exit 1
```

## Integration with Existing Workflows

### Add to existing CI workflow

If you already have a CI workflow, add structurelint as an additional step:

```yaml
# Existing workflow
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      # Existing steps...
      - name: Run tests
        run: npm test

      # Add structurelint
      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'

      - name: Check project structure
        run: |
          go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
          structurelint
```

## Troubleshooting

### structurelint not found

Ensure Go is set up and the binary is in PATH:

```yaml
- name: Install structurelint
  run: |
    go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
    export PATH=$PATH:$(go env GOPATH)/bin
    structurelint
```

### Config file not found

Ensure `.structurelint.yml` is checked into version control and in the repository root.

### Slow builds

Cache the structurelint binary:

```yaml
- name: Cache structurelint
  uses: actions/cache@v5
  with:
    path: ~/go/bin/structurelint
    key: structurelint-${{ runner.os }}-${{ hashFiles('go.sum') }}

- name: Install structurelint if not cached
  run: |
    if [ ! -f ~/go/bin/structurelint ]; then
      go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
    fi
```

## Best Practices

1. **Run on pull requests**: Catch violations before merging
2. **Use specific versions**: Pin to a version tag for consistency
3. **Cache dependencies**: Speed up builds with caching
4. **Fail fast**: Run structurelint early in the pipeline
5. **Document configuration**: Keep `.structurelint.yml` well-commented
6. **Gradual adoption**: Start with basic rules, add more over time

## Status Badges

Add a badge to your README:

```markdown
[![structurelint](https://github.com/YOUR_ORG/YOUR_REPO/actions/workflows/structurelint.yml/badge.svg)](https://github.com/YOUR_ORG/YOUR_REPO/actions/workflows/structurelint.yml)
```

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [structurelint Documentation](https://github.com/Jonathangadeaharder/structurelint)
- [Example workflows](https://github.com/Jonathangadeaharder/structurelint/tree/main/.github/workflows)

---

For more information, see the [main README](../README.md) or [CONTRIBUTING.md](../CONTRIBUTING.md).
