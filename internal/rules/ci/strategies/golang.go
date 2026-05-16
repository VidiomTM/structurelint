package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type GoStrategy struct {
	reader   ci.FileReader
	coverage ci.CoverageThresholds
}

func NewGoStrategy(reader ci.FileReader, cfg map[string]interface{}) *GoStrategy {
	s := &GoStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{Lines: 90}
	if cfg != nil {
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
		}
	}
	return s
}

func (s *GoStrategy) ProjectType() ci.ProjectType { return ci.Go }
func (s *GoStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *GoStrategy) RequiredCIGates() []ci.CIGate {
	return []ci.CIGate{
		{Name: "go test -race", Required: true, Hint: "Add go test -race -covermode=atomic"},
		{Name: "golangci-lint", Required: true, Hint: "Add golangci-lint to CI"},
		{Name: "go vet", Required: true, Hint: "Add go vet to CI"},
	}
}
func (s *GoStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "golangci-lint", Required: true, Hint: "Configure .golangci.yml"},
	}
}
func (s *GoStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
func (s *GoStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	foundTest := false
	foundLint := false
	foundVet := false

	for _, job := range jobs {
		for _, step := range job.Steps {
			combined := strings.ToLower(step.Run + " " + step.Name)
			if strings.Contains(combined, "go test") || strings.Contains(combined, "gotest") {
				foundTest = true
			}
			if strings.Contains(combined, "golangci") && strings.Contains(combined, "lint") {
				foundLint = true
			}
			if strings.Contains(combined, "go vet") {
				foundVet = true
			}
		}
	}
	if !foundTest {
		results = append(results, ci.CheckResult{Message: "Missing go test in CI", Fix: "Add go test -race -covermode=atomic to CI."})
	}
	if !foundLint {
		results = append(results, ci.CheckResult{Message: "Missing golangci-lint in CI", Fix: "Add golangci-lint run to CI."})
	}
	if !foundVet {
		results = append(results, ci.CheckResult{Message: "Missing go vet in CI", Fix: "Add go vet to CI."})
	}
	return results
}
func (s *GoStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
