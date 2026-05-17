# Structurelint Enhancements

This document describes the latest enhancements to structurelint, focused on developer experience, advanced analysis, and ecosystem adoption.

## Table of Contents

1. [Generalized Inline Ignore Directives](#generalized-inline-ignore-directives)
2. [Multiple Output Formats](#multiple-output-formats)
3. [Shareable Configurations](#shareable-configurations)
4. [Cyclomatic Complexity Analysis](#cyclomatic-complexity-analysis)

---

## Generalized Inline Ignore Directives

### Overview

Previously, structurelint only supported the `@structurelint:no-test` directive. We've now generalized this to support ignoring any rule on a per-file basis.

### Usage

**Ignore all rules for a file:**
```go
// @structurelint:ignore This is a legacy file being phased out
package legacy

// ... rest of file ...
```

**Ignore specific rules:**
```go
// @structurelint:ignore max-depth max-files-in-dir Legacy code, refactor planned
package legacy

// ... rest of file ...
```

**Legacy support:**
The original `@structurelint:no-test` directive is still fully supported:
```go
// @structurelint:no-test Interface definitions only, tested through implementations
package interfaces
```

### How It Works

- Directives must be placed in the first 100 lines of the file (preferably at the top)
- Directives must be in comments (`//`, `#`, `/*`, or `*`)
- Multiple rules can be specified, space-separated
- A reason should be provided after the rule names

### Supported Comment Styles

- **Go:** `// @structurelint:ignore max-depth Reason here`
- **Python:** `# @structurelint:ignore max-depth Reason here`
- **JavaScript/TypeScript:** `// @structurelint:ignore max-depth Reason here`
- **Block comments:** `/* @structurelint:ignore max-depth Reason here */`

---

## Multiple Output Formats

### Overview

Structurelint now supports multiple output formats for better integration with CI/CD pipelines and tooling.

### Available Formats

#### Text (default)
Human-readable text output:
```bash
$ structurelint .
internal/linter/linter.go: function has cyclomatic complexity 25, exceeds max 15
```

#### JSON
Machine-readable JSON for parsing and integration:
```bash
$ structurelint --format json .
{
  "version": "1.0.0",
  "timestamp": "2025-11-12T23:28:01Z",
  "violations": 1,
  "results": [
    {
      "rule": "max-cyclomatic-complexity",
      "path": "internal/linter/linter.go",
      "message": "function has cyclomatic complexity 25, exceeds max 15"
    }
  ]
}
```

#### JUnit XML
JUnit XML format for Jenkins, GitHub Actions, and other CI systems:
```bash
$ structurelint --format junit . > results.xml
```

The JUnit format organizes violations by rule, making it easy to see patterns in CI reports:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="structurelint" tests="1" failures="1" errors="0">
  <testsuite name="max-cyclomatic-complexity" tests="1" failures="1">
    <testcase name="internal/linter/linter.go" classname="max-cyclomatic-complexity">
      <failure message="function has cyclomatic complexity 25, exceeds max 15" type="StructureLintViolation">...</failure>
    </testcase>
  </testsuite>
</testsuites>
```

### CI/CD Integration Example

**GitHub Actions:**
```yaml
- name: Run Structurelint
  run: structurelint --format junit . > structurelint-results.xml

- name: Publish Test Results
  uses: EnricoMi/publish-unit-test-result-action@v2
  if: always()
  with:
    files: structurelint-results.xml
```

**Jenkins:**
```groovy
sh 'structurelint --format junit . > results.xml'
junit 'results.xml'
```

---

## Shareable Configurations

### Overview

The `extends` feature allows you to inherit configuration from base configs, making it easy to share best practices across projects and teams.

### Basic Usage

**Extending a local file:**
```yaml
# .structurelint.yml
extends: ./path/to/base-config.yml

# Override or add rules
rules:
  max-depth: 6  # Override the base config
```

**Extending multiple configs:**
```yaml
extends:
  - ./configs/base.yml
  - ./configs/team-standards.yml

rules:
  max-depth: 6
```

### Example: Go Project with Standard Preset

**1. Create a team standard config (`configs/go-standard.yml`):**
```yaml
rules:
  max-depth: 5
  max-files-in-dir: 15
  test-adjacency:
    pattern: "adjacent"
    file-patterns: ["**/*.go"]
```

**2. Extend it in your project (`.structurelint.yml`):**
```yaml
extends: ./configs/go-standard.yml

# Project-specific overrides
rules:
  max-depth: 6  # This project needs slightly more depth
```

### Merge Behavior

- Extended configs are applied first
- Local config rules override extended rules
- `layers` and `entrypoints` are merged (not replaced)
- `exclude` patterns are combined

### Available Presets

See `examples/presets/` for community presets:
- `go-standard.yml` - Sensible defaults for Go projects
- `typescript-react.yml` - TypeScript/React with feature-sliced architecture

### Future: Package-based Presets

In a future release, you'll be able to extend published presets:
```yaml
extends: "@structurelint/preset-go-standard"
```

---

## Cyclomatic Complexity Analysis

### Overview

The new `max-cyclomatic-complexity` rule analyzes Go code using AST parsing to detect overly complex functions.

### Configuration

```yaml
rules:
  max-cyclomatic-complexity:
    max: 15
    file-patterns:
      - "**/*.go"
```

### What It Measures

Cyclomatic complexity = 1 + number of decision points:
- `if` statements
- `for` and `range` loops
- `switch` cases (except `default`)
- `select` cases
- Logical operators (`&&`, `||`)

### Example

```go
// Complexity = 1 + 3 = 4
func ProcessOrder(order Order) error {
    if order.Total > 1000 {  // +1
        if order.Customer.IsPremium {  // +1
            return applyDiscount(order)
        }
    }

    if order.IsInternational {  // +1
        return applyShipping(order)
    }

    return nil
}
```

### Benefits

- **Catches complex code early** - Before it becomes technical debt
- **Encourages refactoring** - High complexity is a smell
- **Objective metric** - Not subjective like "this feels complex"
- **Works with directives** - Can ignore specific files:
  ```go
  // @structurelint:ignore max-cyclomatic-complexity Legacy algorithm, refactor planned
  func ComplexLegacyFunction() { ... }
  ```

### Output

When a function exceeds the threshold:
```
internal/linter/linter.go: function 'createRules' has cyclomatic complexity 37, exceeds max 15
internal/rules/layer_boundaries.go: function '(r *LayerBoundariesRule).Check' has cyclomatic complexity 18, exceeds max 15
```

---

## Feature Comparison

| Feature | Before | After |
|---------|--------|-------|
| **Ignore directives** | `@structurelint:no-test` only | Any rule, with `@structurelint:ignore` |
| **Output formats** | Text only | Text, JSON, JUnit XML |
| **Config reuse** | Copy/paste configs | `extends` feature |
| **Complexity analysis** | Not available | AST-based cyclomatic complexity |

---

## Backward Compatibility

All enhancements are **100% backward compatible**:
- Existing configs work without changes
- `@structurelint:no-test` still works
- Default output is still text
- No breaking changes to existing rules

---

## Implementation Details

### Directive Parser

- Scans first 100 lines of each file
- Only parses directives in comments (not strings)
- Supports all major comment styles
- Zero performance impact when not used

### Output Formatters

- Clean separation: formatters don't affect linting logic
- Extensible: new formats can be added easily
- Type-safe: Uses Go interfaces

### Config Extends

- Recursive: Extended configs can extend other configs
- Relative paths: `./path/to/config.yml`
- Absolute paths: `/absolute/path/config.yml`
- Future: npm package resolution

### Complexity Analyzer

- Uses Go's standard `go/ast` package
- Only analyzes Go files
- Skips test files (`*_test.go`)
- Zero dependencies on external tools

---

## Next Steps & Future Enhancements

The following enhancements were discussed but not yet implemented:

### Autofixing
- `--fix` flag to automatically fix certain violations
- Starting with simple rules (file-content, naming-convention)
- Dry-run mode for safety

### AST Enhancements
- Symbol-level unused exports (not just file-level)
- More granular layer boundary rules
- Support for TypeScript/JavaScript AST parsing

### IDE Extensions
- VSCode extension for real-time linting
- Visual violations in file explorer
- Quick fixes for supported rules

### Package Ecosystem
- NPM wrapper package for Node.js projects
- PyPI wrapper for Python projects
- Published preset packages (`@structurelint/preset-*`)

---

## Contributing

Have ideas for more enhancements? Open an issue or PR at:
[https://github.com/Jonathangadeaharder/structurelint](https://github.com/Jonathangadeaharder/structurelint)

---

## Credits

These enhancements were designed based on real-world feedback and best practices from tools like:
- ESLint (config extends, inline directives)
- gocyclo (complexity analysis)
- JUnit (XML output format)
