package analysis

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCycle_String(t *testing.T) {
	c := &Cycle{Nodes: []string{"A", "B", "C"}, Length: 3}
	assert.Equal(t, "A -> B -> C -> A", c.String())
}

func TestCycle_String_Empty(t *testing.T) {
	c := &Cycle{Nodes: []string{}, Length: 0}
	assert.Equal(t, "(empty cycle)", c.String())
}

func newTestGraph(deps map[string][]string, files []string) *graph.ImportGraph {
	g := &graph.ImportGraph{
		Dependencies: deps,
		AllFiles:     files,
		FileLayers:   make(map[string]*config.Layer),
	}
	if g.AllFiles == nil {
		for k := range deps {
			g.AllFiles = append(g.AllFiles, k)
		}
	}
	return g
}

func TestNewCycleDetector(t *testing.T) {
	g := newTestGraph(nil, nil)
	d := NewCycleDetector(g)
	require.NotNil(t, d)
	assert.Equal(t, g, d.graph)
}

func TestHasCycle_NoCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {},
	}
	d := NewCycleDetector(newTestGraph(deps, nil))
	assert.False(t, d.HasCycle())
}

func TestHasCycle_DirectCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}
	d := NewCycleDetector(newTestGraph(deps, nil))
	assert.True(t, d.HasCycle())
}

func TestHasCycle_SelfLoop(t *testing.T) {
	deps := map[string][]string{
		"A": {"A"},
	}
	d := NewCycleDetector(newTestGraph(deps, nil))
	assert.True(t, d.HasCycle())
}

func TestHasCycle_IndirectCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	}
	d := NewCycleDetector(newTestGraph(deps, nil))
	assert.True(t, d.HasCycle())
}

func TestFindAllCycles_NoCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {},
	}
	d := NewCycleDetector(newTestGraph(deps, nil))
	cycles := d.FindAllCycles()
	assert.Empty(t, cycles)
}

func TestFindAllCycles_OneCycle(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}
	d := NewCycleDetector(newTestGraph(deps, nil))
	cycles := d.FindAllCycles()
	require.Equal(t, 1, len(cycles))
	assert.Equal(t, 2, cycles[0].Length)
}

func TestFindAllCycles_MultipleCycles(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"A"},
		"C": {"D"},
		"D": {"C"},
	}
	files := []string{"A", "B", "C", "D"}
	d := NewCycleDetector(newTestGraph(deps, files))
	cycles := d.FindAllCycles()
	assert.Equal(t, 2, len(cycles))
}

func TestFindCyclesInLayer_NoSuchLayer(t *testing.T) {
	g := newTestGraph(nil, nil)
	d := NewCycleDetector(g)
	cycles := d.FindCyclesInLayer("nonexistent")
	assert.Nil(t, cycles)
}

func TestFindCyclesInLayer_WithCycle(t *testing.T) {
	layer := config.Layer{Name: "app", Path: "src/app/**"}
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/app/a.go": {"src/app/b.go"},
			"src/app/b.go": {"src/app/a.go"},
		},
		AllFiles: []string{"src/app/a.go", "src/app/b.go"},
		FileLayers: map[string]*config.Layer{
			"src/app/a.go": &layer,
			"src/app/b.go": &layer,
		},
		Layers: []config.Layer{layer},
	}
	d := NewCycleDetector(g)
	cycles := d.FindCyclesInLayer("app")
	require.Equal(t, 1, len(cycles))
}

func TestGetStronglyConnectedComponents(t *testing.T) {
	deps := map[string][]string{
		"A": {"B"},
		"B": {"A"},
		"C": {},
	}
	files := []string{"A", "B", "C"}
	d := NewCycleDetector(newTestGraph(deps, files))
	sccs := d.GetStronglyConnectedComponents()
	require.Equal(t, 1, len(sccs))
	assert.Equal(t, 2, len(sccs[0]))
}

func TestHasSelfLoop(t *testing.T) {
	g := newTestGraph(map[string][]string{
		"A": {"A", "B"},
	}, nil)
	d := NewCycleDetector(g)
	assert.True(t, d.hasSelfLoop("A"))
	assert.False(t, d.hasSelfLoop("B"))
}

func TestExtractCycle(t *testing.T) {
	d := NewCycleDetector(newTestGraph(nil, nil))
	cycle := d.extractCycle([]string{"A", "B", "C", "D"}, "B")
	require.NotNil(t, cycle)
	assert.Equal(t, []string{"B", "C", "D"}, cycle.Nodes)
	assert.Equal(t, 3, cycle.Length)
}

func TestExtractCycle_NotFound(t *testing.T) {
	d := NewCycleDetector(newTestGraph(nil, nil))
	cycle := d.extractCycle([]string{"A", "B"}, "X")
	assert.Nil(t, cycle)
}

func TestCyclesEqual(t *testing.T) {
	d := NewCycleDetector(newTestGraph(nil, nil))
	assert.True(t, d.cyclesEqual(
		&Cycle{Nodes: []string{"A", "B", "C"}, Length: 3},
		&Cycle{Nodes: []string{"B", "C", "A"}, Length: 3},
	))
	assert.False(t, d.cyclesEqual(
		&Cycle{Nodes: []string{"A", "B"}, Length: 2},
		&Cycle{Nodes: []string{"A", "B", "C"}, Length: 3},
	))
	assert.False(t, d.cyclesEqual(
		&Cycle{Nodes: []string{"A", "B"}, Length: 2},
		&Cycle{Nodes: []string{"A", "C"}, Length: 2},
	))
}

func TestIsDuplicate(t *testing.T) {
	d := NewCycleDetector(newTestGraph(nil, nil))
	cycles := []Cycle{
		{Nodes: []string{"A", "B"}, Length: 2},
	}
	assert.True(t, d.isDuplicate(cycles, &Cycle{Nodes: []string{"B", "A"}, Length: 2}))
	assert.False(t, d.isDuplicate(cycles, &Cycle{Nodes: []string{"A", "C"}, Length: 2}))
}

func TestGetAllNodes(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A", "C"},
	}
	d := NewCycleDetector(g)
	nodes := d.getAllNodes()
	assert.ElementsMatch(t, []string{"A", "C"}, nodes)
}
