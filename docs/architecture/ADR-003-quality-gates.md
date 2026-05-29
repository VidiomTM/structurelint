# ADR-003: Quality Gates — CI/CD Pipeline with Self-Hosted Runner

**Status:** Accepted  
**Date:** 2026-05-17  
**Deciders:** Project team  
**Authors:** Jonathan Gadea Harder
**Reviewers:** Jonathan Gadea Harder
**References:** [ci.yml](../../.github/workflows/ci.yml), [quality.yml](../../.github/workflows/quality.yml), [pr-agent.yml](../../.github/workflows/pr-agent.yml), [sonar-project.properties](../../sonar-project.properties)

---

## Context

The project requires automated quality enforcement on every pull request. Given that this is a linting tool (dogfooding is essential), the CI must run structurelint on itself as a self-test.

The project uses a self-hosted macOS runner for most jobs (performance, consistency) and matrix builds on GitHub-hosted runners for cross-platform verification.

## Decision

### CI Pipeline (`ci.yml`)

Six job types run in parallel on every push/PR to main/master:

| Job | Runner | Purpose |
|-----|--------|---------|
| `test` | self-hosted | Unit + integration tests with branch coverage (threshold: 58%) |
| `lint` | self-hosted | golangci-lint v2.1.0 |
| `complexity` | self-hosted | gocyclo standalone check (threshold: 26) |
| `build` | ubuntu/macos/windows | Cross-platform build + self-lint (dogfooding) |
| `pbt` | self-hosted | Property-based tests (rapid: 1000 checks) |
| `fuzz-smoke` | self-hosted | 30s fuzz cycles on 3 fuzz targets |
| `self-lint` | self-hosted | structurelint runs on itself, verifies zero violations |

### Coverage Threshold

Current threshold: **58% total coverage**. This is intentionally low because:
- `internal/parser` (tree-sitter CGo) — 37.3% coverage (CGo code is hard to unit test)
- `internal/config` (file I/O heavy) — 50.6%
- `cmd/structurelint` — 0% (thin CLI wrapper)

Target: **90%** once parser and config packages have proper tests (pending tree-sitter CGo compatibility improvements with Go 1.26+).

### SonarCloud

- **Project key:** `Jonathangadeaharder_structurelint`
- **Quality gate:** Enforced on PRs via SonarCloud GitHub integration.
- **Coverage reports:** `coverage.out` uploaded after test run.
- **Exclusions:** `**/*_test.go`, `**/testdata/**`, `**/fuzz/**`, `**/examples/**`

### PR-Agent (`pr-agent.yml`)

- Pattern: Slash-command only (`/review`, `/describe`, `/improve`)
- Model: `openai/qwen3.6-27b-mlx` via LM Studio (self-hosted at localhost:1234/v1)
- Disabled by default (`ENABLE_PR_AGENT` repo variable)
- Runs on self-hosted runner

### Dogfooding

Self-lint job builds structurelint and runs it against its own codebase. Zero violations required. This ensures every change that breaks structurelint's own rules is caught before merge.

## Consequences

- **Positive:** Comprehensive quality assurance (lint + type check + test + coverage + fuzz + PBT + dogfooding). Self-hosted runner avoids GitHub Actions minute quotas for heavy jobs. SonarCloud provides historical quality trends.
- **Negative:** Self-hosted runner is a single point of failure — if the macOS machine is down, lint/test/pbt/fuzz jobs cannot run. No merge-gate workflow yet (separate ADR needed for merge gates).
- **Trade-off:** Coverage threshold of 58% is low but honest. Raising it prematurely would incentivize shallow tests over meaningful coverage.
- **Risk:** tree-sitter CGo incompatibility with Go 1.26+ could break the parser package and cascade into test failures across the pipeline.

## Compliance

PRs must pass all CI jobs before merge. SonarCloud quality gate must be green. Self-lint must report zero violations.
