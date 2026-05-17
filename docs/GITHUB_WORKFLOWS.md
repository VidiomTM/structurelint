# Phase 8: GitHub Workflow Enforcement

## Overview

Phase 8 adds comprehensive GitHub Actions workflow enforcement to ensure your project has proper CI/CD pipelines configured for:

1. **Test Execution** - Automated testing on pull requests and pushes
2. **Security Scanning** - CodeQL, dependency scanning, secret detection
3. **Code Quality** - Linting, formatting, static analysis, coverage checks

### Auto-Fix Support 🎯

The GitHub workflows rule now includes **automatic fix generation** that creates best-practice workflow files based on your project's detected language(s). When a required workflow is missing, structurelint will:

- ✅ Detect your project's primary language (Go, Python, TypeScript, Rust, Java, etc.)
- ✅ Generate a language-specific best-practice workflow
- ✅ Include the auto-fix in the violation's `AutoFix` field
- ✅ Provide ready-to-use workflow content that can be written to `.github/workflows/`

**Language-Specific Templates**: Each language gets optimized workflows with:
- Proper setup actions (setup-go, setup-python, setup-node, etc.)
- Dependency caching for faster builds
- Language-specific testing commands
- Linting and formatting tools
- Code coverage reporting

**Repomix Workflow**: Generated repomix workflows automatically use your project name as the output filename in plain text format.

## Why Enforce GitHub Workflows?

### The Problem

Many projects lack proper CI/CD configuration, leading to:
- **Security Vulnerabilities**: Undetected security issues in dependencies or code
- **Quality Degradation**: Code merged without linting or formatting checks
- **Broken Builds**: Tests not run automatically, allowing broken code to merge
- **Compliance Issues**: Missing security scans required for compliance
- **Technical Debt**: No automated quality gates to prevent degradation

### The Solution

structurelint validates that your project has:
- ✅ `.github/workflows/` directory with workflow files
- ✅ Test workflows that run on PRs and pushes
- ✅ Security scanning (CodeQL, dependency scans, secret detection)
- ✅ Code quality checks (linting, formatting, coverage)
- ✅ Properly configured jobs with required steps
- ✅ Appropriate triggers (pull_request, push, schedule)

## Configuration

Add to your `.structurelint.yml`:

```yaml
root: true

rules:
  # Phase 8: GitHub Workflow Enforcement
  github-workflows:
    # Require a workflow that runs tests
    require-tests: true

    # Require a workflow that performs security scanning
    require-security: true

    # Require a workflow that checks code quality
    require-quality: true

    # Optionally require workflows to commit logs to the branch
    require-log-commits: false

    # Optionally require workflows to package codebase with repomix and upload as artifact
    require-repomix-artifact: false

    # Optionally require specific jobs
    required-jobs:
      - test
      - security
      - lint

    # Optionally require specific triggers
    required-triggers:
      - pull_request
      - push
```

## Validation Checks

### 1. Workflow Presence

Checks for `.github/workflows/` directory and workflow files (`.yml` or `.yaml`).

**Violation Example**:
```
.github/workflows: GitHub workflows directory not found.
Add CI/CD workflows for testing, security, and code quality.
```

### 2. Test Workflow Detection

Looks for workflows with test-related keywords: `test`, `ci`, `build`

**Violation Example**:
```
.github/workflows: No test/CI workflow found.
Add a workflow that runs tests on pull requests and pushes.
```

**Valid Example**:
```yaml
# .github/workflows/test.yml
name: CI Tests  # ✅ Contains "test"
on: [push, pull_request]
jobs:
  test:  # ✅ Job named "test"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - run: go test ./...
```

### 3. Security Workflow Detection

Looks for security-related keywords: `security`, `scan`, `codeql`, `dependabot`

**Violation Example**:
```
.github/workflows: No security scanning workflow found.
Add CodeQL, dependency scanning, or other security checks.
```

**Valid Examples**:
```yaml
# .github/workflows/security.yml
name: Security Scan  # ✅ Contains "security"
on: [push, pull_request, schedule]
jobs:
  codeql:  # ✅ Job named "codeql"
    runs-on: ubuntu-latest
    steps:
      - uses: github/codeql-action/init@v2
      - uses: github/codeql-action/analyze@v2
```

### 4. Quality Workflow Detection

Looks for quality-related keywords: `quality`, `lint`, `format`, `coverage`

**Violation Example**:
```
.github/workflows: No code quality workflow found.
Add linting, formatting, or coverage checks.
```

**Valid Example**:
```yaml
# .github/workflows/quality.yml
name: Code Quality  # ✅ Contains "quality"
on: [push, pull_request]
jobs:
  lint:  # ✅ Job named "lint"
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: golangci/golangci-lint-action@v3
```

### 5. Workflow Structure

Validates individual workflow files have:
- `name` field (descriptive workflow name)
- `on` triggers (when workflow runs)
- At least one `job` defined
- Jobs have `runs-on` specified
- Jobs have `steps` defined

**Violations**:
```
.github/workflows/test.yml: Workflow missing 'name' field.
.github/workflows/test.yml: Workflow missing 'on' triggers.
.github/workflows/test.yml: Job 'test' missing 'runs-on' field.
.github/workflows/test.yml: Job 'test' has no steps defined.
```

### 6. Required Triggers

Optionally enforce specific triggers:

```yaml
rules:
  github-workflows:
    required-triggers:
      - pull_request  # Must run on PRs
      - push          # Must run on pushes
```

**Violation**:
```
.github/workflows/test.yml: Required trigger 'pull_request' not found.
Add to workflow triggers.
```

### 7. Required Jobs

Optionally enforce specific job names:

```yaml
rules:
  github-workflows:
    required-jobs:
      - test
      - integration
      - security-scan
```

**Violation**:
```
.github/workflows/test.yml: Required job 'integration' not found in workflow.
```

### 8. Execution Log Commits

Optionally enforce that workflows commit execution logs back to the triggering branch, allowing agents and developers to review results:

```yaml
rules:
  github-workflows:
    require-log-commits: true  # Require workflows to commit logs to the branch
```

This ensures workflows:
- Create log files (using `tee` or redirection to `.log` files)
- Commit logs using `git add` and `git commit`
- Push logs back to the triggering branch using `git push`

**Violation**:
```
.github/workflows/test.yml: Job 'test' missing log commit steps.
Add steps to commit and push execution logs to the triggering branch
so agents can pull and review results.
Example: use 'git add *.log && git commit -m "Add logs" && git push'.
```

**Valid Example**:
```yaml
name: Tests
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: Run tests
        run: go test ./... 2>&1 | tee test.log
      - name: Commit logs
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add *.log
          git commit -m "Add test execution logs [skip ci]"
          git push
```

### 9. Repomix Codebase Artifact

Optionally require workflows to package the codebase with repomix and upload as an artifact for agent context:

```yaml
rules:
  github-workflows:
    require-repomix-artifact: true  # Require workflows to upload repomix codebase summary
```

This ensures workflows:
- Run repomix to create a comprehensive codebase summary
- Upload the repomix output as an artifact for agent analysis

**Violation**:
```
.github/workflows/test.yml: Job 'test' missing repomix artifact steps.
Add steps to run repomix and upload the codebase summary as an artifact for agent context.
Example: run 'npx repomix' and use 'actions/upload-artifact@v4' to upload the output.
```

**Valid Example**:
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: Run tests
        run: go test ./...
      - name: Generate repomix summary
        run: npx repomix
      - name: Upload repomix artifact
        uses: actions/upload-artifact@v4
        with:
          name: repomix-output-${{ github.sha }}
          path: repomix-output.txt
          retention-days: 7
```

**Note**: Repomix creates a comprehensive codebase summary that agents can use to understand the full context of your project, making it easier for automated tools to analyze failures and suggest fixes.

## Example Workflows

See [examples/github-workflows/](../examples/github-workflows/) for complete working examples.

### Minimal Test Workflow

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'
      - run: go test -v ./...
```

### Minimal Security Workflow

```yaml
name: Security
on: [push, pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: github/codeql-action/init@v2
      - uses: github/codeql-action/analyze@v2
```

### Minimal Quality Workflow

```yaml
name: Quality
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: golangci/golangci-lint-action@v3
```

## Best Practices

### Test Workflows

✅ **DO**:
- Run tests on `pull_request` and `push` events
- Use matrix testing for multiple versions/platforms
- Cache dependencies for faster builds
- Report test coverage to Codecov/Coveralls
- Separate unit tests from integration tests
- Fail fast on test failures

❌ **DON'T**:
- Skip tests on certain branches
- Run tests only on schedule
- Ignore test failures
- Run tests without proper checkout

### Security Workflows

✅ **DO**:
- Run CodeQL analysis weekly via `schedule`
- Scan dependencies with govulncheck/Snyk/Dependabot
- Scan for secrets with Trivy/GitLeaks
- Upload SARIF results to GitHub Security tab
- Fail builds on critical vulnerabilities
- Use multiple security tools (defense in depth)

❌ **DON'T**:
- Run security scans only on releases
- Ignore security findings
- Skip dependency scanning
- Use outdated security tools

### Quality Workflows

✅ **DO**:
- Enforce linting with golangci-lint/ESLint/Ruff
- Check formatting with gofmt/prettier/black
- Require minimum test coverage (e.g., 80%)
- Run static analysis (staticcheck/mypy)
- Check complexity (gocyclo/radon)
- Fail builds on quality violations

❌ **DON'T**:
- Make quality checks optional
- Allow unformatted code to merge
- Skip linting errors
- Ignore coverage drops

## Language-Specific Examples

### Go

```yaml
name: Go CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'
      - run: go test -v -race -coverprofile=coverage.txt ./...
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v3
```

### Python

```yaml
name: Python CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      - run: pip install -r requirements.txt
      - run: pytest --cov=. --cov-report=xml
      - run: ruff check .
      - run: black --check .
```

### Node.js/TypeScript

```yaml
name: Node.js CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm test
      - run: npm run lint
      - run: npm run type-check
```

### Rust

```yaml
name: Rust CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      - run: cargo test
      - run: cargo clippy -- -D warnings
      - run: cargo fmt -- --check
```

## Advanced Configuration

### Per-Directory Overrides

```yaml
root: true

rules:
  github-workflows:
    require-tests: true
    require-security: false
    require-quality: true

overrides:
  # Stricter requirements for production code
  - files: ['src/**']
    rules:
      github-workflows:
        require-security: true
        required-triggers:
          - pull_request
          - push
        required-jobs:
          - test
          - integration
          - security-scan
```

### Monorepo Configuration

```yaml
# Disable at root (monorepo may have central workflows)
root: true

rules:
  github-workflows: false

# Enable for specific packages
overrides:
  - files: ['packages/*/']
    rules:
      github-workflows:
        require-tests: true
```

## Integration with Pre-commit

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: structurelint-workflows
        name: Check GitHub Workflows
        entry: structurelint .
        language: system
        pass_filenames: false
        always_run: false
        files: ^\.github/workflows/.*\.ya?ml$
```

## Troubleshooting

### False Positives

**Issue**: structurelint reports missing workflow but it exists

**Solution**: Ensure workflow name or job names contain keywords:
- Test: `test`, `ci`, `build`
- Security: `security`, `scan`, `codeql`
- Quality: `quality`, `lint`, `format`, `coverage`

### Monorepo Configuration

**Issue**: Monorepo has workflows at root, not in packages

**Solution**: Configure at root level:
```yaml
# At monorepo root
root: true
rules:
  github-workflows:
    require-tests: true
    require-security: true
    require-quality: true
```

### Existing Non-Compliant Workflows

**Issue**: Legacy workflows don't match conventions

**Options**:
1. Rename workflows/jobs to match keywords
2. Use `allow-missing` (if supported)
3. Disable rule during migration period
4. Add custom job name requirements

## Security Considerations

### Supply Chain Security

Validate workflow files to prevent:
- ❌ Unverified third-party actions
- ❌ Hardcoded secrets
- ❌ Privilege escalation
- ❌ Code injection via untrusted inputs

**Best Practices**:
- ✅ Pin actions to specific commit SHAs
- ✅ Use GitHub's official actions
- ✅ Review action permissions
- ✅ Use OIDC for cloud credentials
- ✅ Enable Dependabot for actions

### Workflow Permissions

Use minimal permissions:

```yaml
name: Tests
on: [push, pull_request]

permissions:
  contents: read  # Only read access

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - run: go test ./...
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Security Hardening for GitHub Actions](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
- [CodeQL Documentation](https://codeql.github.com/docs/)
- [GitHub Actions Best Practices](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)

## Examples

See complete working examples:
- [examples/github-workflows/](../examples/github-workflows/) - Complete setup
- [examples/github-workflows/README.md](../examples/github-workflows/README.md) - Detailed guide

## Support

For issues or questions:
- File an issue: [GitHub Issues](https://github.com/Jonathangadeaharder/structurelint/issues)
- Read the docs: [README.md](../README.md)
