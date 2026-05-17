# GitHub Workflow Enforcement Example

This directory contains a complete example of using structurelint's Phase 8: GitHub Workflow Enforcement feature.

## Overview

The GitHub workflow enforcement feature ensures your project has proper CI/CD pipelines configured for:

1. **Test Execution** - Automated testing on PRs and pushes
2. **Security Scanning** - CodeQL, dependency scanning, secret detection
3. **Code Quality** - Linting, formatting, static analysis, coverage

## Directory Structure

```
.
├── .github/
│   └── workflows/
│       ├── test.yml       # Test/CI workflow
│       ├── security.yml   # Security scanning workflow
│       └── quality.yml    # Code quality workflow
├── .structurelint.yml     # Configuration
└── README.md              # This file
```

## Configuration

The `.structurelint.yml` file configures the workflow enforcement:

```yaml
root: true

rules:
  github-workflows:
    require-tests: true      # Require test/CI workflow
    require-security: true   # Require security scanning
    require-quality: true    # Require code quality checks
```

## Workflows

### 1. Test Workflow (`test.yml`)

**Purpose**: Run automated tests on every push and pull request

**Key Features**:
- Matrix testing across multiple Go versions
- Test coverage reporting
- Caching for faster builds
- Integration tests

**Triggers**: `push`, `pull_request`

**Jobs**:
- `test` - Unit tests with coverage
- `integration` - Integration tests

### 2. Security Workflow (`security.yml`)

**Purpose**: Scan for security vulnerabilities

**Key Features**:
- CodeQL analysis for security vulnerabilities
- Dependency vulnerability scanning with govulncheck
- Secret scanning with Trivy
- SAST with Gosec

**Triggers**: `push`, `pull_request`, `schedule` (weekly)

**Jobs**:
- `codeql` - CodeQL static analysis
- `dependency-scan` - Scan Go dependencies
- `secret-scan` - Detect secrets in code
- `sast` - Static application security testing

### 3. Quality Workflow (`quality.yml`)

**Purpose**: Enforce code quality standards

**Key Features**:
- Linting with golangci-lint
- Format checking with gofmt
- Static analysis with staticcheck
- Coverage threshold enforcement (80%)
- Complexity analysis with gocyclo

**Triggers**: `push`, `pull_request`

**Jobs**:
- `lint` - Code linting
- `format` - Format checking
- `staticcheck` - Static analysis
- `coverage` - Coverage requirements
- `complexity` - Complexity analysis

## Running structurelint

To validate your workflows:

```bash
# From the project root
structurelint examples/github-workflows

# With verbose output
structurelint examples/github-workflows --verbose
```

## Validation Checks

structurelint validates:

✅ **Workflow Presence**
- `.github/workflows/` directory exists
- Required workflow types are present (test, security, quality)

✅ **Workflow Structure**
- Workflows have `name` field
- Workflows have `on` triggers configured
- Workflows have at least one job defined

✅ **Job Configuration**
- Jobs have `runs-on` specified
- Jobs have steps defined

✅ **Trigger Requirements** (optional)
- Workflows trigger on required events (e.g., `pull_request`)

✅ **Job Requirements** (optional)
- Specific jobs are present (e.g., `test`, `security`)

## Common Violations

### Missing Workflows Directory

```
.github/workflows: GitHub workflows directory not found.
Add CI/CD workflows for testing, security, and code quality.
```

**Fix**: Create `.github/workflows/` directory with workflow files.

### Missing Test Workflow

```
.github/workflows: No test/CI workflow found.
Add a workflow that runs tests on pull requests and pushes.
```

**Fix**: Add a workflow with "test", "ci", or "build" in the name/jobs.

### Missing Security Workflow

```
.github/workflows: No security scanning workflow found.
Add CodeQL, dependency scanning, or other security checks.
```

**Fix**: Add a workflow with "security", "scan", "codeql", or "dependabot" in the name/jobs.

### Missing Quality Workflow

```
.github/workflows: No code quality workflow found.
Add linting, formatting, or coverage checks.
```

**Fix**: Add a workflow with "quality", "lint", "format", or "coverage" in the name/jobs.

### Workflow Missing Name

```
.github/workflows/test.yml: Workflow missing 'name' field.
Add a descriptive name for the workflow.
```

**Fix**: Add `name: "CI Tests"` to the workflow file.

### Workflow Missing Triggers

```
.github/workflows/test.yml: Workflow missing 'on' triggers.
Specify when the workflow should run.
```

**Fix**: Add `on: [push, pull_request]` to the workflow file.

### Job Missing Steps

```
.github/workflows/test.yml: Job 'test' has no steps defined.
Add steps to execute in this job.
```

**Fix**: Add a `steps:` section with at least one step.

## Best Practices

### Test Workflows

1. **Run on PRs and Pushes**: `on: [push, pull_request]`
2. **Use Matrix Testing**: Test multiple versions/platforms
3. **Cache Dependencies**: Speed up builds with caching
4. **Report Coverage**: Upload coverage to Codecov/Coveralls
5. **Separate Integration Tests**: Run integration tests after unit tests

### Security Workflows

1. **Run Regularly**: Use `schedule` for periodic scans
2. **Multiple Tools**: Combine CodeQL + dependency scanning + secret scanning
3. **Fail on Findings**: Use `exit-code: 1` for critical issues
4. **Upload SARIFs**: Upload results to GitHub Security tab
5. **Scan Dependencies**: Use govulncheck, Dependabot, or Snyk

### Quality Workflows

1. **Enforce Standards**: Fail builds on lint/format issues
2. **Set Thresholds**: Enforce minimum coverage (e.g., 80%)
3. **Multiple Checks**: Combine linting + formatting + static analysis
4. **Fast Feedback**: Run quality checks before expensive tests
5. **Document Exceptions**: Use inline comments for acceptable violations

## Integration with Pre-commit

Add to `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: structurelint
        name: structurelint
        entry: structurelint
        language: system
        pass_filenames: false
        always_run: true
```

## Customization

### Disable Specific Requirements

```yaml
rules:
  github-workflows:
    require-tests: true
    require-security: false  # Disable security requirement
    require-quality: true
```

### Require Specific Triggers

```yaml
rules:
  github-workflows:
    require-tests: true
    required-triggers:
      - pull_request  # Must trigger on PRs
      - push          # Must trigger on pushes
```

### Require Specific Jobs

```yaml
rules:
  github-workflows:
    require-tests: true
    required-jobs:
      - test
      - integration
      - e2e
```

## Language-Specific Examples

### Python Projects

```yaml
# .github/workflows/test.yml
name: Python Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      - run: pip install -r requirements.txt
      - run: pytest --cov=. --cov-report=xml
```

### JavaScript/TypeScript Projects

```yaml
# .github/workflows/test.yml
name: Node.js Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm test
```

### Rust Projects

```yaml
# .github/workflows/test.yml
name: Rust Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      - run: cargo test
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [CodeQL Documentation](https://codeql.github.com/docs/)
- [Security Hardening for GitHub Actions](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
- [structurelint Documentation](../../README.md)

## Support

For issues or questions:
- File an issue: https://github.com/Jonathangadeaharder/structurelint/issues
- Discussions: https://github.com/Jonathangadeaharder/structurelint/discussions
