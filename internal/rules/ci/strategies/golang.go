package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

type GoStrategy struct {
	reader   core.FileReader
	coverage core.CoverageThresholds
}

func NewGoStrategy(reader core.FileReader, cfg map[string]interface{}) *GoStrategy {
	s := &GoStrategy{reader: reader}
	s.coverage = core.CoverageThresholds{Lines: 90}
	if cfg != nil {
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
		}
	}
	return s
}

func (s *GoStrategy) ProjectType() core.ProjectType { return core.Go }
func (s *GoStrategy) RequiredCoverage() core.CoverageThresholds { return s.coverage }
func (s *GoStrategy) RequiredCIGates() []core.CIGate {
	return []core.CIGate{
		{Name: "go test -race", Required: true, Hint: "Add go test -race -covermode=atomic"},
		{Name: "golangci-lint", Required: true, Hint: "Add golangci-lint to CI"},
		{Name: "go vet", Required: true, Hint: "Add go vet to CI"},
	}
}
func (s *GoStrategy) RequiredLinters() []core.LinterTool {
	return []core.LinterTool{
		{Name: "golangci-lint", Required: true, Hint: "Configure .golangci.yml"},
	}
}
func (s *GoStrategy) CheckProjectConfig(files []core.FileInfo, reader core.FileReader) []core.CheckResult { return nil }
func (s *GoStrategy) CheckWorkflowSteps(jobs map[string]core.JobInfo) []core.CheckResult {
	var results []core.CheckResult
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
		results = append(results, core.CheckResult{Message: "Missing go test in CI", Fix: "Add go test -race -covermode=atomic to CI."})
	}
	if !foundLint {
		results = append(results, core.CheckResult{Message: "Missing golangci-lint in CI", Fix: "Add golangci-lint run to CI."})
	}
	if !foundVet {
		results = append(results, core.CheckResult{Message: "Missing go vet in CI", Fix: "Add go vet to CI."})
	}
	return results
}
func (s *GoStrategy) CheckSuppressions(files []core.FileInfo, reader core.FileReader) []core.CheckResult { return nil }
