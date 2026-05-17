# Getting Started with structurelint

This guide will help you get started with structurelint, from installation to writing your first configuration.

## Table of Contents

- [What is structurelint?](#what-is-structurelint)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Your First Configuration](#your-first-configuration)
- [Understanding Rules](#understanding-rules)
- [Working with Layers](#working-with-layers)
- [Common Use Cases](#common-use-cases)
- [Next Steps](#next-steps)

## What is structurelint?

structurelint is a linter for project structure and architecture. Unlike traditional linters that check code syntax and style, structurelint validates:

- **Directory structure** (depth, file counts)
- **Naming conventions** (files and directories)
- **Architectural boundaries** (layer dependencies)
- **Code organization** (complexity, exports)

Think of it as ESLint or Pylint, but for your project's architecture.

## Installation

### Option 1: Go Install (Recommended)

```bash
go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
```

Verify installation:
```bash
structurelint --version
```

### Option 2: Download Binary

Download from the [releases page](https://github.com/Jonathangadeaharder/structurelint/releases):

```bash
# Linux (amd64)
curl -L https://github.com/Jonathangadeaharder/structurelint/releases/latest/download/structurelint-linux-amd64 -o structurelint
chmod +x structurelint
sudo mv structurelint /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/Jonathangadeaharder/structurelint/releases/latest/download/structurelint-darwin-arm64 -o structurelint
chmod +x structurelint
sudo mv structurelint /usr/local/bin/

# Windows
# Download structurelint-windows-amd64.exe and add to PATH
```

### Option 3: Build from Source

```bash
git clone https://github.com/Jonathangadeaharder/structurelint.git
cd structurelint
go build -o structurelint ./cmd/structurelint
sudo mv structurelint /usr/local/bin/
```

## Quick Start

### 1. Initialize Configuration

Create a `.structurelint.yml` file in your project root:

```yaml
root: true

rules:
  max-depth:
    max: 5

  max-files-in-dir:
    max: 15
```

### 2. Run structurelint

```bash
cd your-project
structurelint
```

### 3. Understand the Output

**No violations:**
```
✓ No violations found
```

**With violations:**
```
src/deeply/nested/structure/file.ts:1:1
  max-depth: Directory depth exceeds maximum (6 > 5)

src/bloated-directory/file.ts:1:1
  max-files-in-dir: Directory contains too many files (20 > 15)

Found 2 violations
```

## Your First Configuration

Let's build a practical configuration step by step.

### Step 1: Basic Structure Rules

Start with simple constraints:

```yaml
root: true

rules:
  # Limit directory nesting
  max-depth:
    max: 5

  # Prevent overcrowded directories
  max-files-in-dir:
    max: 15

  # Limit subdirectories per directory
  max-subdirs:
    max: 10
```

Run `structurelint` to see violations. If there are too many, adjust the numbers.

### Step 2: Add Naming Conventions

Enforce consistent naming:

```yaml
root: true

rules:
  max-depth:
    max: 5

  max-files-in-dir:
    max: 15

  # Enforce naming convention
  naming-convention:
    pattern: snake_case  # or camelCase, PascalCase, kebab-case
    paths:
      - src/**
```

### Step 3: Define Required Files

Ensure important files exist:

```yaml
root: true

rules:
  max-depth:
    max: 5

  file-existence:
    required:
      - path: README.md
      - path: LICENSE
      - path: .gitignore
    min_files:
      - path: src/**/*.test.ts
        count: 1
```

### Step 4: Exclude Patterns

Exclude generated or vendor code:

```yaml
root: true

exclude:
  - node_modules/**
  - dist/**
  - build/**
  - vendor/**
  - .git/**

rules:
  max-depth:
    max: 5
```

## Understanding Rules

### Phase 0: Basic Structure Rules

These rules check fundamental project organization:

#### max-depth
Limits directory nesting depth.

```yaml
rules:
  max-depth:
    max: 5  # No more than 5 levels deep
```

**Example violation:**
```
src/app/features/users/components/forms/inputs/text/TextInput.tsx
└── Depth: 8 (exceeds max: 5)
```

#### max-files-in-dir
Limits files per directory.

```yaml
rules:
  max-files-in-dir:
    max: 15
```

#### max-subdirs
Limits subdirectories per directory.

```yaml
rules:
  max-subdirs:
    max: 10
```

#### naming-convention
Enforces naming patterns.

```yaml
rules:
  naming-convention:
    pattern: snake_case  # Options: snake_case, camelCase, PascalCase, kebab-case
    paths:
      - src/**
```

#### disallowed-patterns
Prevents specific patterns.

```yaml
rules:
  disallowed-patterns:
    patterns:
      - "**/*.backup"
      - "**/TODO.txt"
      - "**/.DS_Store"
```

#### file-existence
Ensures files exist or have minimum counts.

```yaml
rules:
  file-existence:
    required:
      - path: README.md
      - path: package.json
    min_files:
      - path: src/**/*.test.ts
        count: 5  # At least 5 test files
    max_files:
      - path: src/**/*.config.js
        count: 3  # No more than 3 config files
```

### Phase 1: Architectural Rules

#### enforce-layer-boundaries
Validates architectural layer dependencies.

```yaml
rules:
  enforce-layer-boundaries:
    enabled: true

layers:
  - name: domain
    path: src/domain/**
    dependsOn: []  # No dependencies

  - name: application
    path: src/application/**
    dependsOn: [domain]  # Can depend on domain

  - name: presentation
    path: src/presentation/**
    dependsOn: [application]  # Can depend on application (and transitively domain)
```

### Phase 2: Code Quality Rules

#### unused-exports
Detects exports that are never imported.

```yaml
rules:
  unused-exports:
    enabled: true
```

#### regex-match
Advanced pattern matching with wildcards.

```yaml
rules:
  regex-match:
    paths:
      "src/*/models/*.ts": "^[A-Z][a-z]+Model$"
      "src/*/services/*.ts": "^[A-Z][a-z]+Service$"
```

## Working with Layers

Layers help enforce clean architecture patterns.

### Example: Hexagonal Architecture

```yaml
root: true

rules:
  enforce-layer-boundaries:
    enabled: true

layers:
  # Core domain (no external dependencies)
  - name: domain
    path: src/domain/**
    dependsOn: []

  # Application logic (uses domain)
  - name: application
    path: src/application/**
    dependsOn: [domain]

  # Infrastructure (adapters)
  - name: infrastructure
    path: src/infrastructure/**
    dependsOn: [domain, application]

  # API/UI (presentation)
  - name: presentation
    path: src/presentation/**
    dependsOn: [application]
```

### Example: Feature-Based Architecture

```yaml
root: true

rules:
  enforce-layer-boundaries:
    enabled: true

layers:
  # Shared utilities
  - name: shared
    path: src/shared/**
    dependsOn: []

  # Feature modules (can use shared)
  - name: auth
    path: src/features/auth/**
    dependsOn: [shared]

  - name: users
    path: src/features/users/**
    dependsOn: [shared, auth]

  - name: products
    path: src/features/products/**
    dependsOn: [shared]
```

### Wildcard Dependencies

Allow a layer to depend on anything:

```yaml
layers:
  - name: app
    path: src/app/**
    dependsOn: ["*"]  # Can import from any layer
```

## Common Use Cases

### Use Case 1: Monorepo Structure

```yaml
root: true

rules:
  max-depth:
    max: 6

  naming-convention:
    pattern: kebab-case

  enforce-layer-boundaries:
    enabled: true

layers:
  - name: shared
    path: packages/shared/**
    dependsOn: []

  - name: frontend
    path: packages/frontend/**
    dependsOn: [shared]

  - name: backend
    path: packages/backend/**
    dependsOn: [shared]
```

### Use Case 2: React Project

```yaml
root: true

rules:
  max-depth:
    max: 5

  max-files-in-dir:
    max: 20

  naming-convention:
    pattern: PascalCase
    paths:
      - src/components/**

  file-existence:
    required:
      - path: src/App.tsx
      - path: src/index.tsx

  enforce-layer-boundaries:
    enabled: true

layers:
  - name: components
    path: src/components/**
    dependsOn: []

  - name: hooks
    path: src/hooks/**
    dependsOn: [components]

  - name: pages
    path: src/pages/**
    dependsOn: [components, hooks]
```

### Use Case 3: Go Microservice

```yaml
root: true

exclude:
  - vendor/**

rules:
  max-depth:
    max: 4

  naming-convention:
    pattern: snake_case

  enforce-layer-boundaries:
    enabled: true

layers:
  - name: domain
    path: internal/domain/**
    dependsOn: []

  - name: repository
    path: internal/repository/**
    dependsOn: [domain]

  - name: service
    path: internal/service/**
    dependsOn: [domain, repository]

  - name: api
    path: internal/api/**
    dependsOn: [service]

  - name: cmd
    path: cmd/**
    dependsOn: [api]
```

## Next Steps

### 1. Integrate with CI/CD

- [GitHub Actions](./GITHUB_ACTION.md)
- [Pre-commit hooks](./PRE_COMMIT.md)

### 2. Customize Rules

Read the [full rule documentation](../README.md#rules) to understand all available options.

### 3. Gradual Adoption

Start with basic rules and gradually add architectural constraints:

**Week 1:** Basic structure rules (depth, file count)
**Week 2:** Naming conventions
**Week 3:** Layer boundaries
**Week 4:** Advanced rules (unused exports, complexity)

### 4. Team Alignment

- Share `.structurelint.yml` in version control
- Document exceptions and reasoning
- Review violations in code reviews
- Update rules as the project evolves

### 5. Contribute

Found a bug or want a new feature? See [CONTRIBUTING.md](../CONTRIBUTING.md).

## Troubleshooting

### "Config file not found"

Ensure `.structurelint.yml` exists in your project root, or specify the path:
```bash
structurelint --config path/to/config.yml
```

### "Too many violations"

Start with relaxed limits and tighten gradually:
```yaml
rules:
  max-depth:
    max: 10  # Start high, reduce over time
```

### "Layer violations I don't understand"

Run with verbose mode (if available) or check the import graph:
```bash
structurelint --verbose
```

### "Performance is slow"

Exclude unnecessary directories:
```yaml
exclude:
  - node_modules/**
  - dist/**
  - coverage/**
```

## Resources

- [README](../README.md) - Full documentation
- [Example configurations](../testdata/fixtures)
- [GitHub repository](https://github.com/Jonathangadeaharder/structurelint)
- [Issue tracker](https://github.com/Jonathangadeaharder/structurelint/issues)

## Get Help

- 📖 Read the docs
- 🐛 Report bugs on GitHub
- 💬 Ask questions in issues
- 🤝 Join discussions

---

Happy linting! 🎉
