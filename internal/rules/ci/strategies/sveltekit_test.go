package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

func TestSvelteKitRequiredGates(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewSvelteKitStrategy(reader, nil)
	gates := strat.RequiredCIGates()
	if len(gates) < 4 {
		t.Fatalf("expected at least 4 gates, got %d", len(gates))
	}
}

func TestSvelteKitChecksSvelteCheck(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewSvelteKitStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"quality": {
			Steps: []ci.StepInfo{
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
	reader := ci.MockFileReader{}
	cfg := map[string]interface{}{"require-vitest-linter": true}
	strat := NewSvelteKitStrategy(reader, cfg)
	jobs := map[string]ci.JobInfo{
		"test": {
			Steps: []ci.StepInfo{
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
