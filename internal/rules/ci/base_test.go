package ci

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestCheckCommandMasking(t *testing.T) {
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "Run tests", Run: "go test ./... || true"},
			},
		},
	}
	results := checkCommandMasking(jobs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCheckCommandMaskingClean(t *testing.T) {
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "Run tests", Run: "go test ./..."},
			},
		},
	}
	results := checkCommandMasking(jobs)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestCheckContinueOnError(t *testing.T) {
	jobs := map[string]core.JobInfo{
		"quality": {
			Steps: []core.StepInfo{
				{Name: "lint", Run: "ruff check", ContinueOnError: "true"},
			},
		},
	}
	results := checkContinueOnError(jobs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCheckRequiredChecksAggregator(t *testing.T) {
	jobs := map[string]core.JobInfo{
		"test":            {},
		"required-checks": {},
	}
	results := checkRequiredChecksAggregator(jobs)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestCheckRequiredChecksAggregatorMissing(t *testing.T) {
	jobs := map[string]core.JobInfo{
		"test": {},
		"lint": {},
	}
	results := checkRequiredChecksAggregator(jobs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
