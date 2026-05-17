# Using structurelint with pre-commit

structurelint can be integrated with [pre-commit](https://pre-commit.com/) to automatically check your project structure before every commit.

## Installation

1. **Install pre-commit**:

   ```bash
   # Using pip
   pip install pre-commit

   # Using homebrew (macOS)
   brew install pre-commit

   # Using apt (Debian/Ubuntu)
   sudo apt install pre-commit
   ```

2. **Create `.pre-commit-config.yaml`** in your project root:

   ```yaml
   repos:
     - repo: https://github.com/Jonathangadeaharder/structurelint
       rev: v0.1.0  # Use the latest release tag
       hooks:
         - id: structurelint
   ```

3. **Install the git hook**:

   ```bash
   pre-commit install
   ```

## Configuration

structurelint will automatically look for `.structurelint.yml` in your project root.

Example configuration:

```yaml
# .structurelint.yml
root: true

rules:
  max-depth:
    max: 5

  max-files-in-dir:
    max: 15

  naming-convention:
    pattern: snake_case

  enforce-layer-boundaries:
    enabled: true

layers:
  - name: domain
    path: src/domain/**
    dependsOn: []

  - name: application
    path: src/application/**
    dependsOn: [domain]

  - name: presentation
    path: src/presentation/**
    dependsOn: [application, domain]
```

## Usage

Once installed, structurelint will run automatically before every commit:

```bash
git commit -m "Add new feature"
# structurelint runs automatically
```

### Manual Run

Run pre-commit on all files manually:

```bash
pre-commit run --all-files
```

### Skip Hook (when needed)

To skip the hook for a specific commit:

```bash
git commit --no-verify -m "WIP: incomplete feature"
```

**Note**: Use `--no-verify` sparingly. It's better to fix violations or update your configuration.

## Advanced Configuration

### Running only on changed files

By default, structurelint runs on the entire project. To run only on changed files, modify the hook:

```yaml
repos:
  - repo: https://github.com/Jonathangadeaharder/structurelint
    rev: v0.1.0
    hooks:
      - id: structurelint
        pass_filenames: true
```

**Note**: Since structural linting checks the entire project architecture, running on all files is usually more accurate.

### Custom arguments

Pass additional arguments to structurelint:

```yaml
repos:
  - repo: https://github.com/Jonathangadeaharder/structurelint
    rev: v0.1.0
    hooks:
      - id: structurelint
        args: ['--config', 'custom-config.yml']
```

### Running in CI

pre-commit can also run in CI environments:

```yaml
# .github/workflows/pre-commit.yml
name: pre-commit

on:
  pull_request:
  push:
    branches: [main]

jobs:
  pre-commit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-python@v4
        with:
          python-version: '3.x'
      - uses: pre-commit/action@v3.0.0
```

## Troubleshooting

### Hook fails with "command not found"

Make sure structurelint is built and available:

```bash
# Manual installation
go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest

# Or build from source
git clone https://github.com/Jonathangadeaharder/structurelint.git
cd structurelint
go build -o structurelint ./cmd/structurelint
sudo cp structurelint /usr/local/bin/
```

### Hook is slow

structurelint analyzes the entire project structure, which can take time for large projects. Consider:

1. **Excluding directories** in `.structurelint.yml`:
   ```yaml
   exclude:
     - vendor/**
     - node_modules/**
     - .git/**
   ```

2. **Running less frequently**: Use pre-commit only in CI, not locally

3. **Skipping for WIP commits**: Use `git commit --no-verify` for work-in-progress

### Configuration not found

Ensure `.structurelint.yml` is in your project root, or specify the path:

```yaml
repos:
  - repo: https://github.com/Jonathangadeaharder/structurelint
    rev: v0.1.0
    hooks:
      - id: structurelint
        args: ['--config', 'path/to/config.yml']
```

## Best Practices

1. **Start with basic rules**: Don't enable all rules at once. Start small and gradually add more.

2. **Run manually first**: Test your configuration before committing:
   ```bash
   structurelint
   ```

3. **Document exceptions**: If you need to skip the hook, document why in the commit message.

4. **Keep configuration in version control**: Commit `.structurelint.yml` and `.pre-commit-config.yaml` so the team uses the same rules.

5. **Update regularly**: Keep structurelint and pre-commit updated:
   ```bash
   pre-commit autoupdate
   ```

## Example Workflow

```bash
# One-time setup
pip install pre-commit
pre-commit install

# Daily workflow (hook runs automatically)
git add .
git commit -m "Add user authentication"
# structurelint runs, commit proceeds if checks pass

# If violations occur
# Fix the violations or update .structurelint.yml
git add .
git commit -m "Add user authentication"
# Commit succeeds
```

## Resources

- [pre-commit documentation](https://pre-commit.com/)
- [structurelint documentation](https://github.com/Jonathangadeaharder/structurelint)
- [Example configurations](https://github.com/Jonathangadeaharder/structurelint/tree/main/testdata/fixtures)

---

For more information, see the [main README](../README.md) or [CONTRIBUTING.md](../CONTRIBUTING.md).
