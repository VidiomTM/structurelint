package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestSvelteKitRequiredGates(t *testing.T) {
	strat := NewSvelteKitStrategy(nil, nil)
	gates := strat.RequiredCIGates()
	if len(gates) < 4 {
		t.Fatalf("expected at least 4 gates, got %d", len(gates))
	}
}

func TestSvelteKitChecksSvelteCheck(t *testing.T) {
	strat := NewSvelteKitStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"quality": {
			Steps: []core.StepInfo{
				{Name: "svelte-check", Run: "pnpm exec svelte-check --tsconfig tsconfig.json"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	found := false
	for _, r := range results {
		if strings.Contains(r.Message, "--fail-on-warnings") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for missing --fail-on-warnings")
	}
}

func TestSvelteKitMissingRequiredGate(t *testing.T) {
	cfg := map[string]interface{}{"require-vitest-linter": true}
	strat := NewSvelteKitStrategy(nil, cfg)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "run tests", Run: "pnpm vitest run"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	found := false
	for _, r := range results {
		if strings.Contains(r.Message, "vitest-linter") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for missing vitest-linter gate")
	}
}
