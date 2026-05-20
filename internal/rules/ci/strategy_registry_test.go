package ci

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

type mockStrategy struct {
	pt core.ProjectType
}

func (m mockStrategy) ProjectType() core.ProjectType                      { return m.pt }
func (m mockStrategy) RequiredCoverage() core.CoverageThresholds          { return core.CoverageThresholds{} }
func (m mockStrategy) RequiredCIGates() []core.CIGate                     { return nil }
func (m mockStrategy) RequiredLinters() []core.LinterTool                 { return nil }
func (m mockStrategy) CheckProjectConfig(files []core.FileInfo, reader core.FileReader) []core.CheckResult {
	return nil
}
func (m mockStrategy) CheckWorkflowSteps(jobs map[string]core.JobInfo) []core.CheckResult { return nil }
func (m mockStrategy) CheckSuppressions(files []core.FileInfo, reader core.FileReader) []core.CheckResult {
	return nil
}

func TestStrategyRegistry(t *testing.T) {
	r := NewStrategyRegistry()
	s := mockStrategy{pt: core.SvelteKit}
	r.Register(s)

	got := r.StrategiesFor([]core.ProjectType{core.SvelteKit})
	if len(got) != 1 {
		t.Fatalf("expected 1 strategy, got %d", len(got))
	}

	got = r.StrategiesFor([]core.ProjectType{core.Python})
	if len(got) != 0 {
		t.Fatalf("expected 0 strategies for unregistered type, got %d", len(got))
	}
}
