# ADR-004: Testing Strategy — Go Test with Branch Coverage

**Status:** Accepted  
**Date:** 2026-05-17  
**Deciders:** Project team  
**Authors:** Jonathan Gadea Harder
**Reviewers:** Jonathan Gadea Harder
**References:** [DESIGN.md §Testing Strategy](../DESIGN.md), [ci.yml](../../.github/workflows/ci.yml), [TEST_AAA_PATTERN.md](../TEST_AAA_PATTERN.md), [TEST_GWT_NAMING.md](../TEST_GWT_NAMING.md), [TEST_VALIDATION.md](../TEST_VALIDATION.md)

---

## Context

structurelint's correctness is critical — it is a linting tool whose false positives would erode developer trust. The testing strategy must cover multiple dimensions: unit tests for individual rules, integration tests for end-to-end linting, property-based tests for invariant verification, and fuzz testing for parser robustness.

Multi-language parsing (via tree-sitter CGo) introduces a special challenge: CGo code cannot be easily unit tested with standard Go patterns.

## Decision

### Testing Dimensions

| Dimension | Tool | Scope | Location |
|-----------|------|-------|----------|
| Unit tests | `go test` + testify | Per-package, per-rule | `internal/*/*_test.go` |
| Integration tests | `go test` | End-to-end with temp dirs | `tests/`, `internal/*/*_test.go` |
| Property-based tests | rapid (PBT) | Invariant verification | `internal/ptest/` |
| Fuzz tests | `go test -fuzz` | Parser robustness | `fuzz/` |
| Mutation tests | go-mutesting (disabled) | Test suite quality | — |

### Testing Conventions

- **AAA Pattern** (Arrange-Act-Assert) — mandatory for all tests. See TEST_AAA_PATTERN.md.
- **GWT Naming** (Given-When-Then) — test function names follow `TestRule_GWT_Scenario`. See TEST_GWT_NAMING.md.
- **Test fixtures** in `testdata/` — real directory structures used for integration tests.

### Coverage Requirements

- **Threshold:** 58% overall branch coverage (CI enforced).
- **Per-package targets** (future): 90% for pure-logic packages (rules, walker, config validation), lower for CGo-dependent packages (parser).
- **Excluded from coverage:** `cmd/structurelint` (thin CLI), `fuzz/` (fuzz harnesses), `testdata/` (fixtures).

### Special Cases

1. **tree-sitter CGo (`internal/parser/`):** Currently at 37.3% coverage. CGo code paths are hard to exercise from Go tests. Mitigation: integration tests that exercise parser through the linter orchestrator rather than directly.

2. **Fuzz tests:** Three targets run 30s each in CI:
   - `FuzzParse` — parser robustness against malformed input
   - `FuzzLintRule` — rule engine stability
   - `FuzzJSONOutput` — output formatting

3. **Mutation testing:** Disabled due to `go-mutesting` incompatibility with Go 1.26+ (StdSizes nil pointer). Re-enable when tool is updated or switch to a maintained alternative.

### SonarCloud Integration

Coverage reports from `go test -coverprofile=coverage.out` are uploaded to SonarCloud (project key: `Jonathangadeaharder_structurelint`). SonarCloud quality gate is enforced via GitHub Actions integration.

## Consequences

- **Positive:** Multiple testing dimensions catch different bug classes. Property-based tests find edge cases that unit tests miss. Fuzz testing ensures parser doesn't crash on adversarial input. SonarCloud provides trend analysis.
- **Negative:** CI runs 7 test-related job types — slow feedback loop on full pipeline (~10-15 min). Mutation testing gap means test suite quality is unmeasured.
- **Trade-off:** 58% coverage threshold is low but honest about CGo limitations. Raising it prematurely risks incentivizing low-value tests.
- **Risk:** tree-sitter CGo could break with Go runtime changes, taking down the parser and its tests simultaneously.

## Compliance

All tests must pass before merge. Coverage must meet 58% threshold. Fuzz targets must not crash within 30s. Property-based tests must pass with 1000 rapid checks.
