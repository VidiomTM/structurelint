package export

import (
	"bytes"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGraph() *graph.ImportGraph {
	appLayer := config.Layer{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}}
	domainLayer := config.Layer{Name: "domain", Path: "src/domain/**", DependsOn: []string{}}

	return &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/application/service.go": {"src/domain/user"},
			"src/domain/user.go":         {},
		},
		AllFiles: []string{"src/application/service.go", "src/domain/user.go"},
		FileLayers: map[string]*config.Layer{
			"src/application/service.go": &appLayer,
			"src/domain/user.go":         &domainLayer,
		},
		Layers: []config.Layer{appLayer, domainLayer},
		IncomingRefs: map[string]int{
			"src/application/service.go": 0,
			"src/domain/user.go":         1,
		},
	}
}

func TestNewDOTExporter(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	require.NotNil(t, e)
	assert.Equal(t, "Dependency Graph", e.options.Title)
}

func TestNewDOTExporter_WithTitle(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{Title: "Custom"})
	assert.Equal(t, "Custom", e.options.Title)
}

func TestExport_Basic(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "digraph")
	assert.Contains(t, output, "rankdir=LR")
	assert.Contains(t, output, "}")
}

func TestExport_EmptyGraph(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: make(map[string][]string),
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No nodes to display")
}

func TestExport_WithLayers(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{ShowLayers: true})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "subgraph cluster_legend")
}

func TestExport_WithCycles(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"A": {"B"},
			"B": {"A"},
		},
		AllFiles:   []string{"A", "B"},
		FileLayers: make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{ShowCycles: true})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "digraph")
}

func TestExport_WithFilterLayer(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{FilterLayer: "domain"})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "user.go")
}

func TestExport_SimplifyPaths(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{SimplifyPaths: true})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "application/service.go")
}

func TestExport_HighlightViolations(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{HighlightViolations: true})
	var buf bytes.Buffer
	err := e.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "digraph")
}

func TestWriteDOTHeader(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{Title: "Test"})
	var buf bytes.Buffer
	err := e.writeDOTHeader(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "digraph \"Test\"")
	assert.Contains(t, output, "rankdir=LR")
	assert.Contains(t, output, "node [shape=box")
}

func TestWriteEmptyGraph(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	err := e.writeEmptyGraph(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No nodes to display")
}

func TestGetCyclesIfEnabled(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"A": {"B"},
			"B": {"A"},
		},
		AllFiles:   []string{"A", "B"},
		FileLayers: make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{ShowCycles: false})
	cycles := e.getCyclesIfEnabled()
	assert.Empty(t, cycles)

	e2 := NewDOTExporter(g, DOTOptions{ShowCycles: true})
	cycles2 := e2.getCyclesIfEnabled()
	assert.NotEmpty(t, cycles2)
}

func TestGetNodeColor(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})

	e.options.ShowLayers = false
	assert.Equal(t, "black", e.getNodeColor("anything"))

	e.options.ShowLayers = true
	assert.Equal(t, "#1565C0", e.getNodeColor("src/application/service.go"))
	assert.Equal(t, "#2E7D32", e.getNodeColor("src/domain/user.go"))
	color := e.getNodeColor("unknown/file")
	assert.Equal(t, "gray", color)
}

func TestGetNodeFillColor(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})

	e.options.ShowLayers = false
	assert.Equal(t, "#FFFFFF", e.getNodeFillColor("anything"))

	e.options.ShowLayers = true
	assert.Equal(t, "#BBDEFB", e.getNodeFillColor("src/application/service.go"))
	assert.Equal(t, "#C8E6C9", e.getNodeFillColor("src/domain/user.go"))
}

func TestGetEdgeStyle(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{ShowCycles: true, HighlightViolations: true})
	cycles := e.detectAllCycles()

	color, style, width := e.getEdgeStyle("A", "B", cycles)
	assert.NotEmpty(t, color)
	assert.NotEmpty(t, style)
	assert.NotEmpty(t, width)
}

func TestGetEdgeStyle_Cycle(t *testing.T) {
	g := &graph.ImportGraph{
		FileLayers: make(map[string]*config.Layer),
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles: []string{"A"},
	}
	cycles := map[string]map[string]bool{"A": {"B": true}}
	e := NewDOTExporter(g, DOTOptions{ShowCycles: true})
	color, style, width := e.getEdgeStyle("A", "B", cycles)
	assert.Equal(t, "orange", color)
	assert.Equal(t, "bold", style)
	assert.Equal(t, "2.0", width)
}

func TestGetEdgeStyle_Violation(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{HighlightViolations: true})
	emptyCycles := make(map[string]map[string]bool)

	appLayer := g.FileLayers["src/application/service.go"]
	domainLayer := g.FileLayers["src/domain/user.go"]
	assert.NotNil(t, appLayer)
	assert.NotNil(t, domainLayer)

	color, style, width := e.getEdgeStyle("src/application/service.go", "src/domain/user.go", emptyCycles)
	assert.Equal(t, "black", color)
	assert.Equal(t, "solid", style)
	assert.Equal(t, "1.0", width)
}

func TestGetEdgeStyle_Default(t *testing.T) {
	g := &graph.ImportGraph{
		FileLayers: make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	cycles := make(map[string]map[string]bool)
	color, style, width := e.getEdgeStyle("A", "B", cycles)
	assert.Equal(t, "black", color)
	assert.Equal(t, "solid", style)
	assert.Equal(t, "1.0", width)
}

func TestGetFilteredNodes(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	nodes := e.getFilteredNodes()
	assert.Equal(t, 2, len(nodes))
}

func TestGetFilteredNodes_FilterLayer(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{FilterLayer: "domain"})
	nodes := e.getFilteredNodes()
	assert.Equal(t, 1, len(nodes))
	assert.Contains(t, nodes[0], "domain")
}

func TestFilterByDepth(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	nodes := e.getFilteredNodes()
	filtered := e.filterByDepth(nodes, 1)
	assert.NotEmpty(t, filtered)
}

func TestIsViolation(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	assert.False(t, e.isViolation("src/application/service.go", "src/domain/user.go"))
}

func TestSimplifyPath(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	assert.Equal(t, "file.go", e.simplifyPath("file.go"))
	assert.Equal(t, "a/b.go", e.simplifyPath("a/b.go"))
	assert.Equal(t, "b/c.go", e.simplifyPath("a/b/c.go"))
	assert.Equal(t, "c/d.go", e.simplifyPath("a/b/c/d.go"))
	assert.Equal(t, "b/c.go", e.simplifyPath("./a/b/c.go"))
}

func TestDetectAllCycles_NoCycle(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	cycles := e.detectAllCycles()
	assert.Empty(t, cycles)
}

func TestDetectAllCycles_WithCycle(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"A": {"B"},
			"B": {"A"},
		},
		AllFiles:   []string{"A", "B"},
		FileLayers: make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	cycles := e.detectAllCycles()
	assert.NotEmpty(t, cycles)
}

func TestWriteLegend(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	err := e.writeLegend(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "cluster_legend")
	assert.Contains(t, output, "Layers")
}

func TestWriteNodes(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	nodes := e.getFilteredNodes()
	var buf bytes.Buffer
	nodeIDs, err := e.writeNodes(&buf, nodes)
	require.NoError(t, err)
	assert.Equal(t, 2, len(nodeIDs))
}

func TestWriteSingleNode(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{SimplifyPaths: true})
	var buf bytes.Buffer
	err := e.writeSingleNode(&buf, "src/application/service.go", "n0")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "n0")
}

func TestWriteEdges(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	nodes := e.getFilteredNodes()
	nodeIDs := map[string]string{"src/application/service.go": "n0", "src/domain/user.go": "n1"}
	var buf bytes.Buffer
	err := e.writeEdges(&buf, nodes, nodeIDs, nil)
	require.NoError(t, err)
}

func TestWriteSingleEdge(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	err := e.writeSingleEdge(&buf, "A", "B", "n0", "n1", nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "n0 -> n1")
}

func TestProcessDependency_VisitedNotInStack(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A", "B"},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	cd := &cycleDetector{
		exporter: e,
		cycles:   make(map[string]map[string]bool),
		visited:  map[string]bool{"B": true},
		recStack: make(map[string]bool),
		path:     []string{"A"},
	}
	result := cd.processDependency("B")
	assert.False(t, result)
}

func TestProcessDependency_NotVisited(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A", "B"},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	cd := &cycleDetector{
		exporter: e,
		cycles:   make(map[string]map[string]bool),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
		path:     []string{},
	}
	result := cd.processDependency("A")
	assert.False(t, result)
}

func TestFindCycleStart(t *testing.T) {
	cd := &cycleDetector{path: []string{"A", "B", "C", "D"}}
	assert.Equal(t, 1, cd.findCycleStart("B"))
	assert.Equal(t, -1, cd.findCycleStart("X"))
}

func TestGetNextNode(t *testing.T) {
	cd := &cycleDetector{path: []string{"A", "B", "C"}}
	assert.Equal(t, "B", cd.getNextNode(0, ""))
	assert.Equal(t, "dep", cd.getNextNode(5, "dep"))
}

func TestAddCycleEdge(t *testing.T) {
	cd := &cycleDetector{cycles: make(map[string]map[string]bool)}
	cd.addCycleEdge("A", "B")
	assert.True(t, cd.cycles["A"]["B"])

	cd.addCycleEdge("A", "C")
	assert.True(t, cd.cycles["A"]["C"])
}

func TestRecordCycle(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	cd := &cycleDetector{
		exporter: e,
		cycles:   make(map[string]map[string]bool),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
		path:     []string{"A", "B", "C"},
	}
	cd.recordCycle("B")
	assert.True(t, cd.cycles["B"]["C"])
}

func TestDfs_WithCycle(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}, "B": {"A"}},
		AllFiles:     []string{"A", "B"},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	cd := &cycleDetector{
		exporter: e,
		cycles:   make(map[string]map[string]bool),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
		path:     []string{},
	}
	result := cd.dfs("A")
	assert.True(t, result)
}

func BenchmarkExport(b *testing.B) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		buf.Reset()
		e.Export(&buf)
	}
}

func TestDfs_NoCycle(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}, "B": {"C"}, "C": {}},
		AllFiles:     []string{"A", "B", "C"},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	cd := &cycleDetector{
		exporter: e,
		cycles:   make(map[string]map[string]bool),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
		path:     []string{},
	}
	cd.dfs("A")
	assert.Empty(t, cd.cycles)
}

func TestWriteNodeEdges(t *testing.T) {
	g := newTestGraph()
	e := NewDOTExporter(g, DOTOptions{})
	nodeIDs := map[string]string{
		"src/application/service.go": "n0",
		"src/domain/user":            "n1",
	}
	var buf bytes.Buffer
	err := e.writeNodeEdges(&buf, "src/application/service.go", nodeIDs, nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "n0 -> n1")
}

func TestWriteNodeEdges_NoTarget(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"nonexistent"}},
		AllFiles:     []string{"A"},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	nodeIDs := map[string]string{"A": "n0"}
	var buf bytes.Buffer
	err := e.writeNodeEdges(&buf, "A", nodeIDs, nil)
	require.NoError(t, err)
	assert.Empty(t, buf.String())
}
