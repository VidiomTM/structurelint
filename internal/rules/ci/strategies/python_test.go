package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestPythonCheckPytestCoverage(t *testing.T) {
	strat := NewPythonStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "run tests", Run: "pytest"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	covBranch := false
	covFailUnder := false
	for _, r := range results {
		if strings.Contains(r.Message, "--cov-branch") {
			covBranch = true
		}
		if strings.Contains(r.Message, "--cov-fail-under") {
			covFailUnder = true
		}
	}
	if !covBranch || !covFailUnder {
		t.Fatal("expected violations for missing --cov-branch and --cov-fail-under")
	}
}

func TestPythonCheckPytestCoveragePass(t *testing.T) {
	strat := NewPythonStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "test", Run: "pytest --cov --cov-branch --cov-fail-under=90"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	for _, r := range results {
		if strings.Contains(r.Message, "pytest") {
			t.Fatalf("unexpected violation: %s", r.Message)
		}
	}
}

func TestPythonMissingRuff(t *testing.T) {
	strat := NewPythonStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "run tests", Run: "pytest"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	found := false
	for _, r := range results {
		if strings.Contains(r.Message, "ruff") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for missing ruff")
	}
}
