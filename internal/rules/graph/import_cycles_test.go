package graph

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestImportCyclesRule_DetectsCycle(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"a.go"},
		},
	}
	rule := NewImportCyclesRule(g)
	violations := rule.Check([]walker.FileInfo{}, nil)
	if len(violations) == 0 {
		t.Fatal("expected cycle violations")
	}
	if !strings.Contains(violations[0].Context, "->") {
		t.Errorf("expected cycle path in Context, got: %q", violations[0].Context)
	}
}

func TestImportCyclesRule_NoCycle(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {},
		},
	}
	rule := NewImportCyclesRule(g)
	if v := rule.Check([]walker.FileInfo{}, nil); len(v) != 0 {
		t.Errorf("expected no violations, got %d", len(v))
	}
}

func TestImportCyclesRule_NilGraph(t *testing.T) {
	rule := &ImportCyclesRule{Graph: nil}
	if v := rule.Check([]walker.FileInfo{}, nil); len(v) != 0 {
		t.Errorf("expected no violations on nil graph, got %d", len(v))
	}
}

func TestImportCyclesRule_SelfLoop(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"a.go"},
		},
	}
	rule := NewImportCyclesRule(g)
	if v := rule.Check([]walker.FileInfo{}, nil); len(v) == 0 {
		t.Errorf("expected self-loop to count as cycle")
	}
}
