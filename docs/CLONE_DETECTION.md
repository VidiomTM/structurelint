# Code Clone Detection

## Overview

structurelint includes a state-of-the-art code clone detection system that identifies duplicated code across your codebase. The system uses **Strategy A: Syntactic Detection** based on AST normalization and Rabin-Karp rolling hash to detect Type-1, Type-2, and Type-3 code clones.

## What Are Code Clones?

Code clones are segments of code that are similar or identical. They are classified into four types:

- **Type-1 (Exact Clones)**: Identical code fragments except for whitespace, comments, and layout
- **Type-2 (Renamed Clones)**: Syntactically identical code with renamed variables, functions, or types
- **Type-3 (Near-Miss Clones)**: Copied code with minor modifications (added/deleted/modified statements)
- **Type-4 (Semantic Clones)**: Code with different syntax but functionally equivalent behavior

### Current Support

The current implementation detects **Type-1, Type-2, and Type-3 clones** using syntactic analysis. Type-4 semantic clone detection using GraphCodeBERT embeddings is planned for a future release (Strategy B).

## Why Detect Clones?

Code clones indicate:

1. **Bug Propagation Risk**: A bug in one clone must be fixed in all copies
2. **Maintenance Burden**: Changes must be synchronized across all clones
3. **Refactoring Opportunities**: Clones can often be extracted into shared functions/modules
4. **Technical Debt**: Accumulated duplication increases codebase complexity

## Quick Start

### Basic Usage

```bash
# Detect clones in current directory
structurelint clones

# Detect clones with custom thresholds
structurelint clones --min-tokens 30 --min-lines 5

# Scan specific directory
structurelint clones internal/rules/
```

### Example Output

```
🔍 Detecting code clones in internal/rules/...

Found 18 Go files
Normalized 18 files
Index: 1278 unique hashes, 4498 total shingles, 816 collisions
Found 149 hash collisions
Expanded to 555 clone pairs
After filtering: 475 clones (min 30 tokens, 5 lines)

🔍 Found 475 code clone pairs:
================================================================================

Clone Pair #1 (256 tokens, ~123 lines) [Type-2 (renamed)]
--------------------------------------------------------------------------------
  Location A: internal/rules/max_cognitive_complexity.go:8-130
  Location B: internal/rules/max_halstead_effort.go:11-133
  Similarity: 100.0%

...

Total: 475 clone pairs detected
Clones: 475 total (Type-1: 0, Type-2: 475, Type-3: 0) | Tokens: 16530 | Lines: ~7200
```

## Configuration Options

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--min-tokens` | 20 | Minimum clone size in tokens (higher = fewer, larger clones) |
| `--min-lines` | 3 | Minimum clone size in lines (approximate) |
| `--k-gram` | 20 | Window size for shingling (matching sensitivity) |
| `--format` | console | Output format: `console`, `json`, `sarif` |
| `--cross-file-only` | true | Only report cross-file clones (ignore within-file) |
| `--workers` | 4 | Number of parallel workers for processing |

### Configuration File

Add clone detection settings to `.structurelint.yml`:

```yaml
clone-detection:
  min-tokens: 30
  min-lines: 5
  k-gram-size: 20
  cross-file-only: true
  exclude-patterns:
    - "**/*_test.go"      # Exclude test files
    - "**/*_gen.go"       # Exclude generated code
    - "**/vendor/**"      # Exclude vendored dependencies
    - "**/node_modules/**"
```

## Output Formats

### Console (Default)

Human-readable output with clone pairs grouped and formatted:

```bash
structurelint clones --format console
```

### JSON

Machine-readable JSON for CI/CD integration:

```bash
structurelint clones --format json > clones.json
```

Example JSON structure:

```json
{
  "total_clones": 2,
  "clones": [
    {
      "type": "Type-2 (renamed)",
      "token_count": 256,
      "line_count": 123,
      "similarity": 1.0,
      "locations": [
        {
          "FilePath": "internal/rules/max_cognitive_complexity.go",
          "StartLine": 8,
          "EndLine": 130,
          "StartToken": 10,
          "EndToken": 266
        },
        {
          "FilePath": "internal/rules/max_halstead_effort.go",
          "StartLine": 11,
          "EndLine": 133,
          "StartToken": 12,
          "EndToken": 268
        }
      ]
    }
  ]
}
```

### SARIF

SARIF format for IDE integration (VS Code, GitHub Code Scanning):

```bash
structurelint clones --format sarif > clones.sarif
```

## Detection Algorithm

The clone detection system uses a multi-stage pipeline:

### 1. AST Normalization

- Parse Go source files using `go/ast` and `go/token`
- Traverse the AST and extract tokens
- Normalize identifiers to `_ID_` (e.g., `myVar` → `_ID_`)
- Normalize literals to `_LIT_` (e.g., `"hello"`, `42` → `_LIT_`)
- Keep keywords and operators as-is (`if`, `for`, `+`, `==`)

**Example**:
```go
// Original code
func calculateSum(a int, b int) int {
    result := a + b
    return result
}

// Normalized token stream
func _ID_ ( _ID_ _ID_ , _ID_ _ID_ ) _ID_ { _ID_ := _ID_ + _ID_ return _ID_ }
```

### 2. K-Gram Shingling

- Generate overlapping k-gram windows (default: 20 tokens)
- Apply Rabin-Karp rolling hash to each window
- Store hash → location mappings in inverted index

**Example** (k=3):
```
Tokens:     [func, _ID_, (, _ID_, ), {, return, _ID_, }]
Shingles:   [func _ID_ (] [_ID_ ( _ID_] [( _ID_ )] [_ID_ ) {] ...
```

### 3. Hash Index Construction

- Build inverted index: `hash → [location1, location2, ...]`
- Thread-safe concurrent access
- In-memory for fast lookup

### 4. Collision Detection

- Find all hash values with multiple locations (collisions)
- Filter to cross-file collisions (optional)
- Each collision is a potential clone seed

### 5. Greedy Match Expansion

For each collision:

1. **Verify Seed**: Token-by-token comparison to rule out hash collisions
2. **Expand Backward**: Match tokens backward until mismatch
3. **Expand Forward**: Match tokens forward until mismatch
4. **Report**: Full expanded clone with accurate line numbers

### 6. Filtering

- Filter by minimum token count (`--min-tokens`)
- Filter by minimum line count (`--min-lines`)
- Deduplicate overlapping reports (partial - improvement pending)

## Tuning Parameters

### Minimum Tokens (`--min-tokens`)

Controls the minimum clone size to report:

- **10-15**: Finds many small clones (may include boilerplate)
- **20-30**: Balanced (default: 20)
- **50+**: Only large, significant clones

**Recommendation**: Start with 30-50 for your first scan to avoid noise.

### Minimum Lines (`--min-lines`)

Filters clones by approximate line count:

- **3-5**: Balanced (default: 3)
- **10+**: Only substantial clones

### K-Gram Size (`--k-gram`)

Window size for hashing:

- **10-15**: More sensitive, finds smaller similarities
- **20**: Balanced (default)
- **30+**: Less sensitive, only finds larger exact matches

**Trade-off**: Smaller k → more clones found, higher false positives

## CI/CD Integration

### GitHub Actions

```yaml
name: Clone Detection

on: [push, pull_request]

jobs:
  clones:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - name: Install structurelint
        run: |
          go install github.com/Jonathangadeaharder/structurelint/cmd/structurelint@latest
      - name: Detect clones
        run: |
          structurelint clones --format json --min-tokens 50 > clones.json
          if [ -s clones.json ]; then
            echo "⚠️ Code clones detected"
            structurelint clones --min-tokens 50
            exit 1
          fi
```

### Pre-Commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

structurelint clones --min-tokens 50 --min-lines 10
if [ $? -ne 0 ]; then
    echo "❌ Clone detection failed. Please refactor duplicated code."
    exit 1
fi
```

## Limitations & Future Work

### Current Limitations

1. **Go Only**: Currently only supports Go source files
2. **Overlapping Reports**: Same clone may be reported multiple times due to overlapping shingles
3. **No Type-4 Detection**: Semantic clones (different implementation, same function) not yet supported
4. **No Cross-Language**: Cannot detect clones across languages (Go vs TypeScript)

### Future Enhancements

#### Strategy B: Semantic Clone Detection (Type-4)

Planned enhancement using GraphCodeBERT for semantic understanding:

- Extract functions/classes as code blocks
- Generate 768-dimensional embeddings using transformer model
- Index embeddings in vector database (FAISS/Milvus)
- Detect functionally equivalent code with different implementations

**Timeline**: Phase 2 (after validation of Strategy A)

#### Multi-Language Support

Extend to Python, TypeScript, JavaScript using:

- Tree-sitter for language-agnostic AST parsing
- Per-language normalization queries
- Unified token representation

#### Clone Deduplication

Improve reporting by:

- Merging overlapping clones into single reports
- Clustering similar clones
- Ranking by severity (size, spread across files)

## Troubleshooting

### High False Positive Rate

**Problem**: Many small, irrelevant clones reported

**Solution**:
- Increase `--min-tokens` to 30-50
- Increase `--min-lines` to 5-10
- Add boilerplate patterns to exclude list

### Missing Known Clones

**Problem**: Known clones not detected

**Solution**:
- Decrease `--min-tokens` threshold
- Check if files are excluded by pattern
- Verify Go files are syntactically valid

### Performance Issues

**Problem**: Detection takes too long

**Solution**:
- Increase `--workers` for parallel processing
- Increase `--min-tokens` to reduce result set
- Exclude large directories (vendor, node_modules)

## References

### Academic Background

This implementation is based on the architectural specification:

> "A Hybrid Architecture for Multi-Language Semantic Code Clone Detection"

**Key techniques**:
- **AST Normalization**: Section II (Creating Canonical Representation)
- **Rabin-Karp Hashing**: Section III.B (Rolling Hash for Shingling)
- **Greedy Expansion**: Section III.C (Collision Investigation and Match Expansion)

### Related Tools

- **PMD CPD**: Copy-Paste Detector (text-based)
- **jscpd**: JavaScript Copy-Paste Detector (uses Rabin-Karp)
- **Deckard**: Tree-based clone detection (academic)
- **SourcererCC**: Token-based clone detection at scale

## Examples

### Example 1: Find Large Clones

```bash
# Find only significant clones (50+ tokens, 10+ lines)
structurelint clones --min-tokens 50 --min-lines 10
```

### Example 2: CI/CD Integration

```bash
# Generate JSON report for analysis
structurelint clones --format json --min-tokens 30 > clones-report.json

# Count total clones
jq '.total_clones' clones-report.json

# Extract file paths with clones
jq -r '.clones[].locations[].FilePath' clones-report.json | sort -u
```

### Example 3: Exclude Test Files

```bash
# Test files often have similar structure (not true duplication)
# They are excluded by default, but can be included:
structurelint clones --min-tokens 30 internal/
```

## Support

For issues, questions, or feature requests:

- GitHub Issues: https://github.com/Jonathangadeaharder/structurelint/issues
- Documentation: https://github.com/Jonathangadeaharder/structurelint/tree/main/docs
