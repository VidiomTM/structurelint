# ADR-001: Project Architecture — Go-Based Repository Structure Linter

**Status:** Accepted  
**Date:** 2026-05-17  
**Deciders:** Project team  
**Authors:** Jonathan Gadea Harder
**Reviewers:** Jonathan Gadea Harder
**References:** [DESIGN.md](../DESIGN.md), [README.md](../README.md), [IMPLEMENTATION_SUMMARY.md](../IMPLEMENTATION_SUMMARY.md)

---

## Context

structurelint fills a gap in the linting ecosystem: existing tools (ls-lint, folderslint) enforce naming conventions but fail to address quantitative metrics of project topology — directory depth, file counts, subdirectory limits, architectural layer boundaries, and dead code detection.

The market has no single tool that unifies filesystem metrics, import-graph analysis, and orphan detection. structurelint aims to be that tool.

## Decision

Adopt a **modular Go monolith** with four architectural layers:

```
cmd/structurelint/          # CLI entry point
internal/
  config/                   # YAML-based cascading configuration
  walker/                   # Filesystem traversal + metric collection
  rules/                    # Pluggable Rule interface + implementations
  linter/                   # Orchestration: config → walker → rules → violations
  parser/                   # Multi-language import/export parsing (TS, JS, Go, Python)
  graph/                    # Import graph construction and analysis
  clones/                   # Syntactic clone detection pipeline
```

### Key Architectural Decisions

1. **Go as implementation language.** Rationale: compiled single binary, goroutine-based concurrency for parallel rule execution, fast startup for pre-commit hooks, strong stdlib for filesystem operations. See DESIGN.md §Performance.

2. **ESLint-style cascading configuration.** `.structurelint.yml` files merge from parent to child directories. `root: true` stops upward search. Overrides provide surgical rule application. See DESIGN.md §Configuration System Design.

3. **Filesystem walker collects typed metrics.** `FileInfo` and `DirInfo` structs capture depth, file count, subdir count — separate from rule execution. This enables future optimizations like incremental checking and caching.

4. **Three-phase feature scope:**
   - Phase 0 (Core): Filesystem metric rules, naming conventions, pattern matching
   - Phase 1 (Layers): Import graph parsing, architectural boundary enforcement
   - Phase 2 (Orphans): Dead code detection, unused export identification

5. **Backward compatibility is preserved across all phases.** Breaking changes are explicitly documented and gated behind configuration migration paths.

## Consequences

- **Positive:** Single binary distribution, fast execution (<1s for typical projects), clean separation of concerns, easy to add new rules via the Rule interface.
- **Negative:** Go-only runtime means multi-language parsing requires CGo (tree-sitter) or hand-written parsers — increases build complexity and limits cross-compilation.
- **Trade-off:** The cascading config model adds complexity for monorepo setups but provides fine-grained control that flat configs cannot.
- **Risk:** tree-sitter CGo dependency limits coverage on some platforms and inflates CI build times.

## Compliance

Architecture is enforced via `internal/` package boundaries. No package outside `cmd/` imports from another top-level external package except through the linter orchestrator.
