# CI Standardization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans.

**Goal:** Every VidiomTM project gets standardized 10-gate CI with structurelint enforcement.

**Architecture:** Two-phase rollout — (1) structurelint rules for ongoing enforcement, (2) one-time config/workflow generation for existing repos.

**Tech Stack:** Go (structurelint), GitHub Actions (CI), pyproject.toml, tsconfig.json

---

### Task 1: Revive `linter-config` rule in structurelint

**Files:**
- Modify: `internal/rules/structure/init.go`
- Create: `internal/rules/structure/linter_config.go`
- Create: `internal/rules/structure/linter_config_test.go`
- Modify: `internal/linter/factory.go`

- [ ] **Step 1: Read existing state**

Read `internal/rules/structure/init.go` to see how rules are registered. Read `internal/linter/factory.go` to see where the old `linter-config` was removed and what error message it now prints.

```bash
cd ~/projects/linters/structurelint
grep -n 'linter-config\|linter_config\|LinterConfig' internal/linter/factory.go
grep -rn 'linter-config\|linter_config\|LinterConfig' internal/rules/
```

- [ ] **Step 2: Create the rule file**

Write `internal/rules/structure/linter_config.go` with a `LinterConfigRule` that implements the `Rule` interface:

```go
package structure

import (
	"fmt"
	"path/filepath"

	"github.com/structurelint/structurelint/internal/lang"
	"github.com/structurelint/structurelint/internal/linter"
)

type LinterConfigRule struct {
	ctx *linter.RuleContext
}

func init() {
	linter.RegisterRule("linter-config", func(ctx *linter.RuleContext) (linter.Rule, error) {
		return &LinterConfigRule{ctx: ctx}, nil
	})
}

func (r *LinterConfigRule) Name() string { return "linter-config" }

func (r *LinterConfigRule) Check(files []string, dirs []string) []linter.Violation {
	var violations []linter.Violation

	// Determine project language from manifest
	lang := lang.Detect(r.ctx.Root)
	hasPyproject := fileExists(filepath.Join(r.ctx.Root, "pyproject.toml"))
	hasTsconfig := fileExists(filepath.Join(r.ctx.Root, "tsconfig.json"))
	hasSvelteConfig := fileExists(filepath.Join(r.ctx.Root, "svelte.config.js")) ||
		fileExists(filepath.Join(r.ctx.Root, "svelte.config.ts"))

	switch {
	case hasPyproject && hasSvelteConfig:
		// Polyglot: check both
		violations = append(violations, checkPyproject(r.ctx.Root)...)
		violations = append(violations, checkTsconfig(r.ctx.Root)...)
	case hasPyproject:
		violations = append(violations, checkPyproject(r.ctx.Root)...)
	case hasSvelteConfig || hasTsconfig:
		violations = append(violations, checkTsconfig(r.ctx.Root)...)
	}

	// All projects need .semgrep.yml
	if !fileExists(filepath.Join(r.ctx.Root, ".semgrep.yml")) {
		violations = append(violations, linter.Violation{
			File:    ".semgrep.yml",
			Message: "missing .semgrep.yml — required for Semgrep security gate",
			Severity: "error",
		})
	}

	return violations
}
```

- [ ] **Step 3: Write helper functions**

Add helper functions at the bottom of `linter_config.go`:

```go
func checkPyproject(root string) []linter.Violation {
	// Check pyproject.toml has [tool.pyright] with typeCheckingMode = "strict"
	// Check it has [tool.ruff] section
	return nil // TODO: implement with TOML parsing
}

func checkTsconfig(root string) []linter.Violation {
	// Check tsconfig.json has "strict": true
	return nil // TODO: implement with JSON parsing
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```

- [ ] **Step 4: Write the failing test**

Write `internal/rules/structure/linter_config_test.go`:

```go
package structure

import (
	"testing"
	"os"
	"path/filepath"
)

func TestLinterConfigMissingSemgrep(t *testing.T) {
	// Create temp dir with pyproject.toml but no .semgrep.yml
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte("[project]\nname=\"test\"\n"), 0644)
	
	ctx := &linter.RuleContext{Root: tmpDir}
	rule := &LinterConfigRule{ctx: ctx}
	
	violations := rule.Check(nil, nil)
	
	found := false
	for _, v := range violations {
		if v.File == ".semgrep.yml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected violation for missing .semgrep.yml, got none")
	}
}
```

- [ ] **Step 5: Run test to verify it compiles and fails correctly**

```bash
cd ~/projects/linters/structurelint
go test ./internal/rules/structure/ -run TestLinterConfigMissingSemgrep -v 2>&1 | head -20
```

Expected: either PASS or compile errors about missing `linter.RuleContext` type. Fix imports accordingly.

- [ ] **Step 6: Fix the test linter.RuleContext import path**

Check the actual `RuleContext` type path:

```bash
grep -rn 'type RuleContext' internal/linter/
```

Fix import path in test to match actual package. Then re-run.

```bash
go test ./internal/rules/structure/ -run TestLinterConfigMissingSemgrep -v
```

Expected: PASS

- [ ] **Step 7: Add pyproject.toml pyright check**

Add the pyproject validation helper:

```go
import (
	"os"
	"path/filepath"
	"github.com/BurntSushi/toml"
)

type pyprojectToml struct {
	Tool struct {
		Pyright *struct {
			TypeCheckingMode string `toml:"typeCheckingMode"`
		} `toml:"pyright"`
		Ruff *struct{} `toml:"ruff"`
	} `toml:"tool"`
}

func checkPyproject(root string) []linter.Violation {
	path := filepath.Join(root, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg pyprojectToml
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	var v []linter.Violation
	if cfg.Tool.Pyright == nil {
		v = append(v, linter.Violation{
			File:    "pyproject.toml",
			Message: "missing [tool.pyright] section — required for type checking gate",
		})
	} else if cfg.Tool.Pyright.TypeCheckingMode != "strict" {
		v = append(v, linter.Violation{
			File:    "pyproject.toml",
			Message: fmt.Sprintf("[tool.pyright] typeCheckingMode should be \"strict\", got %q", cfg.Tool.Pyright.TypeCheckingMode),
		})
	}
	if cfg.Tool.Ruff == nil {
		v = append(v, linter.Violation{
			File:    "pyproject.toml",
			Message: "missing [tool.ruff] section — required for linting gate",
		})
	}
	return v
}
```

- [ ] **Step 8: Add tsconfig.json strict check**

```go
func checkTsconfig(root string) []linter.Violation {
	path := filepath.Join(root, "tsconfig.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg struct {
		CompilerOptions struct {
			Strict *bool `json:"strict"`
		} `json:"compilerOptions"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	var v []linter.Violation
	if cfg.CompilerOptions.Strict == nil || !*cfg.CompilerOptions.Strict {
		v = append(v, linter.Violation{
			File:    "tsconfig.json",
			Message: "\"strict\": true required in compilerOptions for TypeScript type checking gate",
		})
	}
	return v
}
```

- [ ] **Step 9: Write comprehensive tests**

```go
func TestLinterConfigValidPyproject(t *testing.T) {
	tmpDir := t.TempDir()
	data := `
[tool.pyright]
typeCheckingMode = "strict"
[tool.ruff]
`
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(data), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".semgrep.yml"), []byte("rules: []"), 0644)

	rule := &LinterConfigRule{}
	// Need to set ctx with root
	violations := rule.Check(nil, nil)
	if len(violations) > 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestLinterConfigPyprojectMissingPyright(t *testing.T) {
	tmpDir := t.TempDir()
	data := `[project]
name = "test"
`
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(data), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".semgrep.yml"), []byte("rules: []"), 0644)

	violations := checkPyproject(tmpDir)
	found := false
	for _, v := range violations {
		if v.File == "pyproject.toml" && contains(v.Message, "[tool.pyright]") {
			found = true
		}
	}
	if !found {
		t.Error("expected violation about missing [tool.pyright]")
	}
}

func TestLinterConfigTsconfigNotStrict(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{"compilerOptions": {}}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".semgrep.yml"), []byte("rules: []"), 0644)

	violations := checkTsconfig(tmpDir)
	found := false
	for _, v := range violations {
		if contains(v.Message, "\"strict\": true") {
			found = true
		}
	}
	if !found {
		t.Error("expected violation about missing strict: true")
	}
}
```

- [ ] **Step 10: Run all tests**

```bash
cd ~/projects/linters/structurelint
go test ./internal/rules/structure/ -run TestLinterConfig -v
```

Expected: all PASS

- [ ] **Step 11: Remove old error message from factory.go**

The old `linter-config` rule was removed and replaced with an error message. Find and remove it:

```bash
grep -n 'linter-config' internal/linter/factory.go
```

Comment out or remove the case that prints an error for removed linter-config rule, since it's now revived.

- [ ] **Step 12: Commit**

```bash
cd ~/projects/linters/structurelint
git add internal/rules/structure/linter_config.go internal/rules/structure/linter_config_test.go internal/linter/factory.go
git commit -m "feat: revive linter-config rule with pyproject/tsconfig/semgrep validation"
```

---

### Task 2: Add `ci-gates` rule to structurelint

**Files:**
- Create: `internal/rules/structure/ci_gates.go`
- Create: `internal/rules/structure/ci_gates_test.go`
- Modify: `internal/rules/structure/init.go`

- [ ] **Step 1: Create ci-gates rule**

Write `internal/rules/structure/ci_gates.go`:

```go
package structure

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/linter"
)

type CIGatesRule struct {
	ctx *linter.RuleContext
}

func init() {
	linter.RegisterRule("ci-gates", func(ctx *linter.RuleContext) (linter.Rule, error) {
		return &CIGatesRule{ctx: ctx}, nil
	})
}

func (r *CIGatesRule) Name() string { return "ci-gates" }

func (r *CIGatesRule) Check(files []string, dirs []string) []linter.Violation {
	var violations []linter.Violation
	workflowDir := filepath.Join(r.ctx.Root, ".github", "workflows")

	hasPrGate := false
	hasMergeGate := false

	// Check if workflow files exist
	if entries, err := os.ReadDir(workflowDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "pr-gate") {
				hasPrGate = true
			}
			if strings.HasPrefix(e.Name(), "merge-gate") {
				hasMergeGate = true
			}
		}
	}

	if !hasPrGate {
		violations = append(violations, linter.Violation{
			File:    ".github/workflows/pr-gate.yml",
			Message: "missing pr-gate.yml workflow — required for PR CI gates",
		})
	}
	if !hasMergeGate {
		violations = append(violations, linter.Violation{
			File:    ".github/workflows/merge-gate.yml",
			Message: "missing merge-gate.yml workflow — required for merge CI gates",
		})
	}

	return violations
}
```

- [ ] **Step 2: Write tests**

```go
func TestCIGatesMissing(t *testing.T) {
	tmpDir := t.TempDir()
	rule := &CIGatesRule{ctx: &linter.RuleContext{Root: tmpDir}}
	violations := rule.Check(nil, nil)
	if len(violations) != 2 {
		t.Errorf("expected 2 violations (pr-gate + merge-gate), got %d", len(violations))
	}
}

func TestCIGatesPresent(t *testing.T) {
	tmpDir := t.TempDir()
	workDir := filepath.Join(tmpDir, ".github", "workflows")
	os.MkdirAll(workDir, 0755)
	os.WriteFile(filepath.Join(workDir, "pr-gate.yml"), []byte("name: PR Gate"), 0644)
	os.WriteFile(filepath.Join(workDir, "merge-gate.yml"), []byte("name: Merge Gate"), 0644)

	rule := &CIGatesRule{ctx: &linter.RuleContext{Root: tmpDir}}
	violations := rule.Check(nil, nil)
	if len(violations) > 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd ~/projects/linters/structurelint
go test ./internal/rules/structure/ -run TestCIGates -v
```

Expected: all PASS

- [ ] **Step 4: Update presets**

Read the existing presets and add `linter-config` and `ci-gates` to the
`sveltekit` and `python-monorepo` presets:

```bash
grep -rn 'sveltekit\|python-monorepo' internal/ --include='*.go' | head -20
```

Find where presets are defined and add the new rules to both presets.

- [ ] **Step 5: Commit**

```bash
cd ~/projects/linters/structurelint
git add internal/rules/structure/ci_gates.go internal/rules/structure/ci_gates_test.go internal/rules/structure/init.go
# also add preset changes if any
git commit -m "feat: add ci-gates rule for workflow existence validation"
```

---

### Task 3: Generate Python configs for all Python repos

**Files:** Per repo pyproject.toml, .semgrep.yml, .pre-commit-config.yaml

Repos: Gretel, gh-wait, sitcom-pilot, timelog, mac-control-mcp, scriptforge

For each Python repo:

- [ ] **Step 1: Generate pyproject.toml config**

Check if `pyproject.toml` exists. If yes, append/add the relevant sections (`[tool.ruff]`, `[tool.pyright]`, `[tool.vulture]`, `[tool.pytest.ini_options]`). If no, create it.

Use the local clone if available. For each repo:

```bash
cd ~/projects/{repo}
# Add tool sections to pyproject.toml
```

Config to add:

```toml
[tool.ruff]
target-version = "py312"
line-length = 100

[tool.ruff.lint]
select = ["E", "F", "I", "N", "W", "UP", "B", "SIM", "ARG", "PL"]

[tool.pyright]
typeCheckingMode = "strict"

[tool.vulture]
paths = ["src/"]

[tool.pytest.ini_options]
addopts = "--cov --cov-branch --cov-fail-under=90 --cov-report=term-missing"
```

- [ ] **Step 2: Generate .semgrep.yml**

```yaml
rules:
  - id: no-hardcoded-secrets
    pattern-either:
      - pattern: password = "..."
      - pattern: api_key = "..."
      - pattern: secret = "..."
      - pattern: token = "..."
    severity: ERROR
    message: "Hardcoded secret detected"
```

- [ ] **Step 3: Generate .pre-commit-config.yaml**

```yaml
repos:
  - repo: https://github.com/astral-sh/ty
    rev: v0.0.36
    hooks:
      - id: ty
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.11.0
    hooks:
      - id: ruff
      - id: ruff-format
```

- [ ] **Step 4: Commit per repo**

```bash
git add pyproject.toml .semgrep.yml .pre-commit-config.yaml
git commit -m "chore: add CI gate configs (ruff, pyright, vulture, semgrep, pre-commit)"
git push
```

---

### Task 4: Generate SvelteKit/JS/TS configs

Repos: ObstetraScroll (SvelteKit), Vidiom, svelteuml, scriptforge (JS side)

- [ ] **Step 1: Ensure tsconfig.json has strict mode**

For SvelteKit repos, ensure `tsconfig.json` extends the SvelteKit base and sets strict:

```json
{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true
  }
}
```

- [ ] **Step 2: Add check script to package.json**

```json
{
  "scripts": {
    "check": "svelte-check --tsconfig ./tsconfig.json",
    "knip": "knip",
    "audit": "pnpm audit --audit-level=high"
  }
}
```

- [ ] **Step 3: Generate .semgrep.yml** (same as Python version)

- [ ] **Step 4: Commit per repo**

```bash
git add tsconfig.json package.json .semgrep.yml
git commit -m "chore: add CI gate configs (svelte-check, knip, semgrep)"
git push
```

---

### Task 5: Generate pr-gate.yml + merge-gate.yml per project

For each project type, generate the standard PR gate and merge gate workflows.

- [ ] **Step 1: Generate Python pr-gate.yml**

Template for `{repo}/.github/workflows/pr-gate.yml`:

```yaml
name: PR Gate

on:
  pull_request:
    branches: [main, master, feat/quality-gates]

jobs:
  format:
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uvx ruff format --check .

  lint:
    needs: [format]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uvx ruff check .

  typecheck:
    needs: [lint]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uvx pyright

  build:
    needs: [typecheck]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uv build

  unit:
    needs: [build]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uv run pytest --cov --cov-branch --cov-fail-under=90

  mutation:
    needs: [build]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    if: steps.changed-files.outputs.any_changed == 'true'
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uv sync
      - uses: tj-actions/changed-files@v44
        id: changed-files
        with:
          files: 'src/**/*.py'
          separator: ','
      - run: uv run mutmut run --paths-to-mutate ${{ steps.changed-files.outputs.all_changed_files }}

  deadcode:
    needs: [build]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uvx vulture src/

  deps:
    needs: [build]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.12' }
      - run: pip install uv && uvx pip-audit

  semgrep:
    needs: [build]
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - uses: actions/checkout@v4
      - run: pip install semgrep && semgrep --config=auto .

  required-checks:
    needs: [unit, mutation, deadcode, deps, semgrep]
    if: always()
    runs-on: ["self-hosted", "macos", "arm64", "VidiomTM"]
    steps:
      - run: |
          echo "${{ toJSON(needs) }}" | jq -e '
            to_entries | all(
              .value.result == "success" or .value.result == "skipped"
            )
          '
```

- [ ] **Step 2: Generate Python merge-gate.yml**

Same as pr-gate but runs on push to main, and doesn't need mutation (already done in PR).

- [ ] **Step 3: Generate SvelteKit pr-gate.yml + merge-gate.yml**

Similar structure but with:
- Format: `prettier --check`
- Lint: `eslint`
- Type: `svelte-check --tsconfig`
- Build: `pnpm build`
- Unit: `vitest run --coverage`
- Mutation: Stryker
- Dead: `knip`
- Deps: `pnpm audit`
- Semgrep: same
- E2E (merge gate only): Playwright

- [ ] **Step 4: Generate JS/TS pr-gate.yml + merge-gate.yml**

Same as SvelteKit but without svelte-check (use tsc instead) and without E2E.

- [ ] **Step 5: Commit per repo**

```bash
git add .github/workflows/pr-gate.yml .github/workflows/merge-gate.yml
git commit -m "chore: add standardized CI gate workflows"
git push
```

---

### Task 6: Push structurelint changes and update self-check

- [ ] **Step 1: Run full test suite**

```bash
cd ~/projects/linters/structurelint
go test ./... 2>&1 | tail -20
```

Expected: all PASS

- [ ] **Step 2: Push structurelint changes**

```bash
git push origin main
```

- [ ] **Step 3: Run structurelint self-check**

```bash
cd ~/projects/linters/structurelint
./structurelint .
```

Expected: no violations from new rules (structurelint should pass its own checks)

---

### Task 7: Verify CI on a real PR

- [ ] **Step 1: Create a test PR in one repo**

Pick a simple repo (e.g., timelog), create a branch with a trivial change, and verify the PR gate runs with all 10 gates.

```bash
cd ~/projects/timelog
git checkout -b test/verify-ci-gates
# Add comment to a file
echo "# CI gate verification" >> README.md
git add README.md && git commit -m "chore: verify CI gates" && git push origin test/verify-ci-gates
```

Check the PR on GitHub for workflow run results.

- [ ] **Step 2: Fix any issues and iterate**

If CI fails, diagnose and fix. Report results.
