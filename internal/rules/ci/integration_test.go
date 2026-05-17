package ci

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/strategies"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestIntegrationSvelteKitProject(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json":    `{"devDependencies": {"svelte": "^5.0.0"}}`,
		"/project/svelte.config.ts": `import adapter from '@sveltejs/adapter-auto'`,
		"/project/.github/workflows/ci.yml": `
name: CI
on: [push]
jobs:
  required-checks:
    runs-on: ubuntu-latest
    steps:
      - run: echo "ok"
  quality:
    runs-on: ubuntu-latest
    steps:
      - name: Lint
        run: pnpm dlx @biomejs/biome check src/
      - name: Type check
        run: pnpm exec svelte-check --tsconfig tsconfig.json --fail-on-warnings --compiler-warnings "state_referenced_locally:ignore"
      - name: Test
        run: pnpm vitest run --coverage
      - name: Build
        run: pnpm build
`,
	}}

	registry := NewStrategyRegistry()
	registry.Register(strategies.NewSvelteKitStrategy(reader, nil))

	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, map[string]interface{}{
		"global": map[string]interface{}{
			"disallow-command-masking":               true,
			"disallow-continue-on-error-on-quality":  true,
			"require-required-checks-aggregator":     true,
		},
	})

	files := []walker.FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "svelte.config.ts", AbsPath: "/project/svelte.config.ts"},
		{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"},
	}

	violations := rule.Check(files, nil)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  %s: %s", v.Path, v.Message)
		}
	}
}

func TestIntegrationSvelteKitMissingSvelteCheckFlags(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json":    `{"devDependencies": {"svelte": "^5.0.0"}}`,
		"/project/svelte.config.ts": `import adapter from '@sveltejs/adapter-auto'`,
		"/project/.github/workflows/ci.yml": `
name: CI
on: [push]
jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - name: Type check
        run: pnpm exec svelte-check --tsconfig tsconfig.json
`,
	}}

	registry := NewStrategyRegistry()
	registry.Register(strategies.NewSvelteKitStrategy(reader, nil))

	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, map[string]interface{}{
		"global": map[string]interface{}{
			"require-required-checks-aggregator": true,
		},
	})

	files := []walker.FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "svelte.config.ts", AbsPath: "/project/svelte.config.ts"},
		{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"},
	}

	violations := rule.Check(files, nil)
	if len(violations) < 2 {
		t.Fatalf("expected at least 2 violations (missing --fail-on-warnings + missing aggregator), got %d", len(violations))
	}
}

func TestIntegrationPythonProjectMissingGates(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/pyproject.toml": `[project]`,
		"/project/.github/workflows/ci.yml": `
name: CI
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        run: git checkout
`,
	}}

	registry := NewStrategyRegistry()
	registry.Register(strategies.NewPythonStrategy(reader, nil))

	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, map[string]interface{}{
		"global": map[string]interface{}{
			"require-required-checks-aggregator": true,
		},
	})

	files := []walker.FileInfo{
		{Path: "pyproject.toml", AbsPath: "/project/pyproject.toml"},
		{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"},
	}

	violations := rule.Check(files, nil)
	if len(violations) < 1 {
		t.Fatal("expected violations for missing quality gates")
	}
}

func TestIntegrationNoWorkflowsNoViolations(t *testing.T) {
	// No .github/workflows should produce no violations
	// (project type won't be detected without marker files either)
	reader := MockFileReader{Files: map[string]string{}}
	registry := NewStrategyRegistry()
	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, nil)

	files := []walker.FileInfo{
		{Path: "README.md", AbsPath: "/project/README.md"},
	}

	violations := rule.Check(files, nil)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations for no-detect project, got %d", len(violations))
	}
}

func TestIntegrationCommandMasking(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/go.mod":           `module test`,
		"/project/.github/workflows/ci.yml": `
name: CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Run tests
        run: go test ./... || true
`,
	}}

	registry := NewStrategyRegistry()
	registry.Register(strategies.NewGoStrategy(reader, nil))

	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, map[string]interface{}{
		"global": map[string]interface{}{
			"disallow-command-masking": true,
		},
	})

	files := []walker.FileInfo{
		{Path: "go.mod", AbsPath: "/project/go.mod"},
		{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"},
	}

	violations := rule.Check(files, nil)
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "Command masking") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for command masking")
	}
}
