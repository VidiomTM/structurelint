package export

import (
	"bytes"
	"os"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMermaidExport_EmptyGraph(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{})
	var buf bytes.Buffer
	err := m.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No nodes to display")
}

func TestMermaidExport_EmptyGraph_WriteEmptyError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{})
	err := m.writeEmptyGraph(&badWriter{})
	assert.Error(t, err)
}

func TestMermaidExport_WithCycles(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}, "B": {"A"}},
		AllFiles:     []string{"A", "B"},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{ShowCycles: true})
	var buf bytes.Buffer
	err := m.Export(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "cycle")
}

func TestMermaidExport_WithViolations(t *testing.T) {
	appLayer := config.Layer{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}}
	domainLayer := config.Layer{Name: "domain", Path: "src/domain/**", DependsOn: []string{}}
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/domain/user.go": {"src/application/service.go"},
		},
		AllFiles: []string{"src/application/service.go", "src/domain/user.go"},
		FileLayers: map[string]*config.Layer{
			"src/application/service.go": &appLayer,
			"src/domain/user.go":         &domainLayer,
		},
		Layers: []config.Layer{appLayer, domainLayer},
	}
	m := NewMermaidExporter(g, MermaidOptions{HighlightViolations: true})
	var buf bytes.Buffer
	err := m.Export(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "violation")
}

func TestMermaidExport_IsolatedNode(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {}},
		AllFiles:     []string{"A"},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{})
	var buf bytes.Buffer
	err := m.Export(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "n0")
}

func TestMermaidExport_IsolatedNodeWriteError(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {}},
		AllFiles:     []string{"A"},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{})
	n := m.createNodeIDMap([]string{"A"})
	err := m.writeIsolatedNode(&badWriter{}, n["A"], "A")
	assert.Error(t, err)
}

func TestMermaidExport_ExportWithWrapper(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A"},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{})
	var buf bytes.Buffer
	err := m.ExportWithWrapper(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "```mermaid")
	assert.Contains(t, output, "```")
}

func TestMermaidExport_ExportHTML(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A"},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{Title: "Test"})
	var buf bytes.Buffer
	err := m.ExportHTML(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "mermaid.min.js")
	assert.Contains(t, output, "Test")
}

func TestMermaidExport_ExportWithWrapperWriteError(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{})
	err := m.ExportWithWrapper(&badWriter{})
	assert.Error(t, err)
}

func TestMermaidExport_GetMermaidEdgeStyle_Cycle(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{ShowCycles: true})
	d := NewDOTExporter(g, DOTOptions{})
	style := m.getMermaidEdgeStyle("A", "B", "n0", "n1", "A", "B",
		map[string]map[string]bool{"A": {"B": true}}, d)
	assert.Contains(t, style, "cycle")
}

func TestMermaidExport_GetMermaidEdgeStyle_Violation(t *testing.T) {
	appLayer := config.Layer{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}}
	domainLayer := config.Layer{Name: "domain", Path: "src/domain/**", DependsOn: []string{}}
	g := &graph.ImportGraph{
		FileLayers: map[string]*config.Layer{
			"src/domain/user.go": &domainLayer,
			"src/application/service.go": &appLayer,
		},
	}
	m := NewMermaidExporter(g, MermaidOptions{HighlightViolations: true})
	d := NewDOTExporter(g, DOTOptions{})
	style := m.getMermaidEdgeStyle("src/domain/user.go", "src/application/service.go",
		"n0", "n1", "user", "service",
		make(map[string]map[string]bool), d)
	assert.Contains(t, style, "violation")
}

func TestMermaidExport_GetMermaidEdgeStyle_Default(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{})
	d := NewDOTExporter(g, DOTOptions{})
	style := m.getMermaidEdgeStyle("A", "B", "n0", "n1", "A", "B",
		make(map[string]map[string]bool), d)
	assert.Contains(t, style, "-->")
}

func TestDOTExport_UnknownLayerColors(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{ShowLayers: true})
	color := e.getNodeColor("unknown")
	assert.Equal(t, "gray", color)
	fill := e.getNodeFillColor("unknown")
	assert.Equal(t, "#F5F5F5", fill)
}

func TestDOTExport_UnknownLayerNameColor(t *testing.T) {
	layer := config.Layer{Name: "custom", Path: "src/custom/**"}
	g := &graph.ImportGraph{
		FileLayers: map[string]*config.Layer{"file": &layer},
	}
	e := NewDOTExporter(g, DOTOptions{ShowLayers: true})
	color := e.getNodeColor("file")
	assert.Equal(t, "black", color)
	fill := e.getNodeFillColor("file")
	assert.Equal(t, "#FFFFFF", fill)
}

func TestDOTExport_WriteEmptyGraph_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	err := e.writeEmptyGraph(&badWriter{})
	assert.Error(t, err)
}

func TestDOTExport_WriteDOTHeaderError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	err := e.writeDOTHeader(&badWriter{})
	assert.Error(t, err)
}

func TestDOTExport_WriteNodes_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	_, err := e.writeNodes(&badWriter{}, []string{"A"})
	assert.Error(t, err)
}

func TestDOTExport_WriteNodeEdges_WriteError(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A", "B"},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	err := e.writeNodeEdges(&badWriter{}, "A", map[string]string{"A": "n0", "B": "n1"}, nil)
	assert.Error(t, err)
}

func TestDOTExport_WriteSingleEdge_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	err := e.writeSingleEdge(&badWriter{}, "A", "B", "n0", "n1", nil)
	assert.Error(t, err)
}

func TestDOTExport_WriteLegend_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	err := e.writeLegend(&badWriter{})
	assert.Error(t, err)
}

func TestDOTExport_WriteSingleNode_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	e := NewDOTExporter(g, DOTOptions{})
	err := e.writeSingleNode(&badWriter{}, "A", "n0")
	assert.Error(t, err)
}

func TestDOTExport_Export_WriteClosingError(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles:   []string{"A"},
		FileLayers: make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	w := &writeOnceWriter{}
	err := e.Export(w)
	require.Error(t, err)
}

func TestMermaidExport_Export_HeaderWriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{})
	err := m.writeMermaidHeader(&badWriter{})
	assert.Error(t, err)
}

func TestMermaidExport_writeMermaidEdges_Error(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		FileLayers:   make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{})
	d := NewDOTExporter(g, DOTOptions{})
	nodes := m.createNodeIDMap([]string{"A", "B"})
	err := m.writeMermaidEdges(&badWriter{}, []string{"A"}, nodes,
		make(map[string]map[string]bool), d)
	assert.Error(t, err)
}

func TestMermaidExport_WriteSingleEdge_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{})
	d := NewDOTExporter(g, DOTOptions{})
	err := m.writeSingleEdge(&badWriter{}, "A", "B", "n0", "n1", "A",
		make(map[string]map[string]bool), d)
	assert.Error(t, err)
}

func TestMermaidExport_ApplyLayerStyles_WriteError(t *testing.T) {
	layer := config.Layer{Name: "domain", Path: "src/domain/**"}
	g := &graph.ImportGraph{
		FileLayers: map[string]*config.Layer{"src/domain/user.go": &layer},
	}
	m := NewMermaidExporter(g, MermaidOptions{ShowLayers: true})
	err := m.applyLayerStyles(&badWriter{}, map[string]string{"src/domain/user.go": "n0"}, []string{"src/domain/user.go"})
	assert.Error(t, err)
}

func TestMermaidExport_WriteStyles_WriteError(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{ShowLayers: true})
	err := m.writeStyles(&badWriter{}, map[string]string{}, nil)
	assert.Error(t, err)
}

func TestDOTExport_FilterByDepth_NoRoots(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		AllFiles:     []string{"A", "B"},
		IncomingRefs: map[string]int{"A": 1, "B": 0},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	filtered := e.filterByDepth([]string{"A", "B"}, 1)
	assert.NotEmpty(t, filtered)
}

func TestDOTExport_GetFilteredNodes_Fallback(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{"A": {"B"}},
		FileLayers:   make(map[string]*config.Layer),
	}
	e := NewDOTExporter(g, DOTOptions{})
	nodes := e.getFilteredNodes()
	assert.Contains(t, nodes, "A")
}

func TestExport_WriteLegend_NoLayers(t *testing.T) {
	g := newTestGraph()
	g.Layers = nil
	e := NewDOTExporter(g, DOTOptions{})
	var buf bytes.Buffer
	err := e.writeLegend(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "cluster_legend")
}

func TestNewMermaidExporter_Defaults(t *testing.T) {
	g := &graph.ImportGraph{}
	m := NewMermaidExporter(g, MermaidOptions{Title: "Custom", Direction: "TB"})
	assert.Equal(t, "TB", m.options.Direction)
	assert.Equal(t, "Custom", m.options.Title)
}

func TestMermaidExport_ApplySpecialStyles_HighlightViolations(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{HighlightViolations: true})
	var buf bytes.Buffer
	err := m.applySpecialStyles(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "linkStyle")
}

func TestMermaidExport_WriteStyles_Empty(t *testing.T) {
	g := &graph.ImportGraph{FileLayers: make(map[string]*config.Layer)}
	m := NewMermaidExporter(g, MermaidOptions{})
	var buf bytes.Buffer
	err := m.writeStyles(&buf, map[string]string{}, []string{})
	require.NoError(t, err)
}

func TestDOTExport_WriteLegend_WithLayers(t *testing.T) {
	appLayer := config.Layer{Name: "application", Path: "src/application/**", DependsOn: []string{}}
	domainLayer := config.Layer{Name: "domain", Path: "src/domain/**", DependsOn: []string{}}
	customLayer := config.Layer{Name: "customname", Path: "src/custom/**", DependsOn: []string{}}
	g := &graph.ImportGraph{
		AllFiles: []string{"src/application/service.go", "src/domain/user.go", "src/custom/x.go"},
		FileLayers: map[string]*config.Layer{
			"src/application/service.go": &appLayer,
			"src/domain/user.go":         &domainLayer,
			"src/custom/x.go":            &customLayer,
		},
		Layers: []config.Layer{appLayer, domainLayer, customLayer},
	}
	e := NewDOTExporter(g, DOTOptions{ShowLayers: true})
	var buf bytes.Buffer
	err := e.writeLegend(&buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "customname")
}

// badWriter always fails on Write
type badWriter struct{}

func (b *badWriter) Write(p []byte) (int, error) {
	return 0, os.ErrInvalid
}

// writeOnceWriter only allows one write
type writeOnceWriter struct {
	written bool
	buf     bytes.Buffer
}

func (w *writeOnceWriter) Write(p []byte) (int, error) {
	if w.written {
		return 0, os.ErrInvalid
	}
	w.written = true
	return w.buf.Write(p)
}
