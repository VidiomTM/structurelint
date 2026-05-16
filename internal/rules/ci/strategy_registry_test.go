package ci

import (
	"testing"
)

type mockStrategy struct {
	pt ProjectType
}

func (m mockStrategy) ProjectType() ProjectType { return m.pt }
func (m mockStrategy) RequiredCoverage() CoverageThresholds { return CoverageThresholds{} }
func (m mockStrategy) RequiredCIGates() []CIGate { return nil }
func (m mockStrategy) RequiredLinters() []LinterTool { return nil }
func (m mockStrategy) CheckProjectConfig(files []FileInfo, reader FileReader) []CheckResult { return nil }
func (m mockStrategy) CheckWorkflowSteps(jobs map[string]JobInfo) []CheckResult { return nil }
func (m mockStrategy) CheckSuppressions(files []FileInfo, reader FileReader) []CheckResult { return nil }

func TestStrategyRegistry(t *testing.T) {
	r := NewStrategyRegistry()
	s := mockStrategy{pt: SvelteKit}
	r.Register(s)

	got := r.StrategiesFor([]ProjectType{SvelteKit})
	if len(got) != 1 {
		t.Fatalf("expected 1 strategy, got %d", len(got))
	}

	got = r.StrategiesFor([]ProjectType{Python})
	if len(got) != 0 {
		t.Fatalf("expected 0 strategies for unregistered type, got %d", len(got))
	}
}
