# ADR-005: Rule Engine Architecture — Tree-Sitter Integration and Mixed-Language Support

**Status:** Accepted  
**Date:** 2026-05-17  
**Deciders:** Project team  
**Authors:** Jonathan Gadea Harder
**Reviewers:** Jonathan Gadea Harder
**References:** [DESIGN.md §Rules Engine](../DESIGN.md), [RULES.md](../RULES.md), [CLONE_DETECTION_ARCHITECTURE.md](../CLONE_DETECTION_ARCHITECTURE.md), [MIXED_LANGUAGE_ARCHITECTURE.md](../MIXED_LANGUAGE_ARCHITECTURE.md)

---

## Context

structurelint must analyze codebases across 9 languages (Go, Python, TypeScript, JavaScript, Java, C++, C#, Rust, Ruby). Each language has different import syntax, naming conventions, and test patterns. The rule engine must support:

1. **Filesystem-level rules** (depth, file counts) — language-agnostic
2. **Import-graph rules** (layer boundaries) — language-specific parsing
3. **Code quality rules** (cognitive complexity, Halstead metrics) — AST-dependent
4. **Clone detection** — AST normalization + rolling hash
5. **Test validation rules** — language-specific test naming

## Decision

### Rule Interface

All rules implement a common interface:

```go
type Rule interface {
    Name() string
    Check(files []FileInfo, dirs map[string]*DirInfo) []Violation
}
```

This enables:
- Dynamic rule instantiation from YAML config
- Parallel rule execution (goroutine per rule)
- Consistent violation reporting

### Multi-Language Parsing via Tree-Sitter

Tree-sitter provides AST parsing for all 9 supported languages via CGo bindings (`github.com/smacker/go-tree-sitter`).

**Parser layer (`internal/parser/`):**
- Language detection from file extension
- Import statement extraction (per-language grammar queries)
- Export symbol extraction (per-language grammar queries)
- Unified token representation for metrics

**Parser coverage by language:**

| Language | Imports | Exports | Metrics |
|----------|---------|---------|---------|
| Go | go/ast (native) | go/ast (native) | go/ast (native) |
| Python | tree-sitter | tree-sitter | tree-sitter |
| TypeScript | tree-sitter | tree-sitter | tree-sitter |
| JavaScript | tree-sitter | tree-sitter | tree-sitter |
| Java | tree-sitter | tree-sitter | tree-sitter |
| C++ | tree-sitter | — | tree-sitter |
| C# | tree-sitter | — | tree-sitter |
| Rust | — | — | — |
| Ruby | — | — | — |

Go uses native `go/ast` for better performance and zero CGo overhead. Other languages use tree-sitter with language-specific `.scm` query files.

### Clone Detection Pipeline

The clone detection system (`internal/clones/`) implements syntactic clone detection via:

1. **AST normalization** — identifiers → `_ID_`, literals → `_LIT_`
2. **K-gram shingling** — Rabin-Karp rolling hash (window size k=20)
3. **Inverted index** — `map[uint64][]Shingle` for hash→location lookup
4. **Greedy match expansion** — bidirectional token matching
5. **Filtering** — min tokens, min lines thresholds

Output formats: console, JSON, SARIF (IDE integration).

Currently implemented for Go only (using `go/ast`). Multi-language support (tree-sitter integration) is planned for Phase 2 of clone detection.

### Mixed-Language Architecture

For rules that need language-scoped validation (test-location, naming-convention), the engine supports:

- **File pattern filtering** — `file-patterns: ["**/*_test.go"]` scopes rules to specific language files
- **Language auto-detection** — from manifest files (go.mod, package.json, pyproject.toml, Cargo.toml, etc.)
- **Language-specific defaults** — naming conventions, test patterns, infrastructure exemptions

See MIXED_LANGUAGE_ARCHITECTURE.md for the detailed evolution from Phase 1 (pattern filtering) to Phase 2 (per-language config) to Phase 3 (auto-detection).

### Rule Categories

| Category | Rules | Language Requirement | AST Needed |
|----------|-------|---------------------|-----------|
| Filesystem metrics | max-depth, max-files, max-subdirs | None | No |
| Naming | naming-convention, regex-match | Per-language patterns | No |
| File existence | file-existence, disallowed-patterns | None | No |
| Architecture | enforce-layer-boundaries, path-based-layers | Import parsing | Yes |
| Dead code | disallow-orphaned-files, disallow-unused-exports | Import + export parsing | Yes |
| Code quality | max-cognitive-complexity, max-halstead-effort | Full AST | Yes |
| Test validation | test-adjacency, test-location | Test naming patterns | No |
| Clone detection | clone-detection | AST normalization | Yes |
| GitHub workflows | github-workflows | None | No |
| Linter config | linter-config | None | No |

### Path-Based as 50x Faster Alternative

For layer validation, path-based rules (Priority 3 feature) provide 50x faster validation than import-graph analysis by using glob/regex patterns instead of parsing. This works even when code doesn't compile — critical during refactoring.

## Consequences

- **Positive:** Unified Rule interface makes adding new rules trivial. Tree-sitter provides broad language coverage from a single dependency. Path-based layer rules provide fast validation when import graphs aren't needed. Clone detection pipeline follows academic best practices (AST normalization + rolling hash).
- **Negative:** Tree-sitter CGo dependency complicates builds and limits cross-compilation. Rust and Ruby have limited parser support (test validation only, no metrics). Clone detection is Go-only for now.
- **Trade-off:** Go uses native `go/ast` for Go files (faster, no CGo) but tree-sitter for everything else — two parser pathways to maintain.
- **Risk:** Tree-sitter grammar updates may break query files silently. Lock grammar versions in CI.

## Compliance

All rules must implement the `Rule` interface. New languages added via tree-sitter require (a) grammar query files, (b) language detection in walker, (c) test fixtures in `testdata/`. Clone detection must pass on the project's own codebase (dogfooding).
