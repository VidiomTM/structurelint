package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

func TestGoMissingGates(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewGoStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"build": {
			Steps: []ci.StepInfo{
				{Name: "build", Run: "go build ./..."},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	foundTest := false
	foundLint := false
	foundVet := false
	for _, r := range results {
		if strings.Contains(r.Message, "go test") { foundTest = true }
		if strings.Contains(r.Message, "golangci-lint") { foundLint = true }
		if strings.Contains(r.Message, "go vet") { foundVet = true }
	}
	if !foundTest || !foundLint || !foundVet {
		t.Fatal("expected violations for all 3 missing Go gates")
	}
}

func TestGoAllGatesPresent(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewGoStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"quality": {
			Steps: []ci.StepInfo{
				{Name: "lint", Run: "golangci-lint run ./..."},
				{Name: "vet", Run: "go vet ./..."},
				{Name: "test", Run: "go test -race -covermode=atomic ./..."},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	if len(results) > 0 {
		t.Fatalf("expected 0 violations, got %d: %v", len(results), results)
	}
}
