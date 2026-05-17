package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

type RustStrategy struct {
	reader              core.FileReader
	coverage            core.CoverageThresholds
	requireCargoTestLint bool
}

func NewRustStrategy(reader core.FileReader, cfg map[string]interface{}) *RustStrategy {
	s := &RustStrategy{reader: reader}
	s.coverage = core.CoverageThresholds{Lines: 90}
	if cfg != nil {
		if v, ok := cfg["require-cargo-test-lint"].(bool); ok {
			s.requireCargoTestLint = v
		}
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
		}
	}
	return s
}

func (s *RustStrategy) ProjectType() core.ProjectType { return core.Rust }
func (s *RustStrategy) RequiredCoverage() core.CoverageThresholds { return s.coverage }
func (s *RustStrategy) RequiredCIGates() []core.CIGate {
	gates := []core.CIGate{
		{Name: "cargo clippy", Required: true, Hint: "Add cargo clippy to CI"},
		{Name: "cargo fmt --check", Required: true, Hint: "Add cargo fmt --check to CI"},
		{Name: "cargo test", Required: true, Hint: "Add cargo test to CI"},
		{Name: "coverage", Required: true, Hint: "Add cargo llvm-cov or tarpaulin for coverage"},
	}
	if s.requireCargoTestLint {
		gates = append(gates, core.CIGate{Name: "cargo test-lint", Required: true, Hint: "Add cargo-test-lint to CI"})
	}
	return gates
}
func (s *RustStrategy) RequiredLinters() []core.LinterTool {
	return []core.LinterTool{
		{Name: "clippy", Required: true, Hint: "Configure clippy in Cargo.toml"},
		{Name: "rustfmt", Required: true, Hint: "Configure rustfmt"},
	}
}
func (s *RustStrategy) CheckProjectConfig(files []core.FileInfo, reader core.FileReader) []core.CheckResult { return nil }
func (s *RustStrategy) CheckWorkflowSteps(jobs map[string]core.JobInfo) []core.CheckResult {
	var results []core.CheckResult
	foundClippy := false
	foundFmt := false
	foundTest := false
	foundCoverage := false
	foundTestLint := false

	for _, job := range jobs {
		for _, step := range job.Steps {
			combined := strings.ToLower(step.Run + " " + step.Name)
			if strings.Contains(combined, "clippy") {
				foundClippy = true
			}
			if strings.Contains(combined, "cargo fmt") || strings.Contains(combined, "rustfmt") {
				foundFmt = true
			}
			if strings.Contains(combined, "cargo test") && !strings.Contains(combined, "test-lint") {
				foundTest = true
			}
			if strings.Contains(combined, "llvm-cov") || strings.Contains(combined, "tarpaulin") || strings.Contains(combined, "--fail-under") {
				foundCoverage = true
			}
			if s.requireCargoTestLint && strings.Contains(combined, "test-lint") {
				foundTestLint = true
			}
		}
	}
	if !foundClippy {
		results = append(results, core.CheckResult{Message: "Missing cargo clippy in CI", Fix: "Add cargo clippy to CI workflow."})
	}
	if !foundFmt {
		results = append(results, core.CheckResult{Message: "Missing cargo fmt --check in CI", Fix: "Add cargo fmt --check to CI."})
	}
	if !foundTest {
		results = append(results, core.CheckResult{Message: "Missing cargo test in CI", Fix: "Add cargo test to CI."})
	}
	if !foundCoverage {
		results = append(results, core.CheckResult{Message: "Missing coverage gate in CI", Fix: "Add cargo llvm-cov or tarpaulin."})
	}
	if s.requireCargoTestLint && !foundTestLint {
		results = append(results, core.CheckResult{Message: "Missing cargo test-lint in CI", Fix: "Add cargo-test-lint to CI."})
	}
	return results
}
func (s *RustStrategy) CheckSuppressions(files []core.FileInfo, reader core.FileReader) []core.CheckResult { return nil }
