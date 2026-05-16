package export

import (
	"bytes"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMermaidExporter(t *testing.T) {
	g := &graph.ImportGraph{}
	m := NewMermaidExporter(g, MermaidOptions{})
	require.NotNil(t, m)
	assert.Equal(t, "LR", m.options.Direction)
}

func TestMermaidExport_Basic(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/main.go": {"src/util"},
		},
		AllFiles:   []string{"src/main.go"},
		FileLayers: make(map[string]*config.Layer),
	}
	m := NewMermaidExporter(g, MermaidOptions{})
	var buf bytes.Buffer
	err := m.Export(&buf)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "graph LR")
	assert.Contains(t, output, "n0")
}

func TestMermaidExport_ShowLayers(t *testing.T) {
	appLayer := config.Layer{Name: "application", Path: "src/application/**"}
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/application/service.go": {"src/domain/user"},
		},
		AllFiles: []string{"src/application/service.go"},
		FileLayers: map[string]*config.Layer{
			"src/application/service.go": &appLayer,
		},
	}
	m := NewMermaidExporter(g, MermaidOptions{ShowLayers: true})
	var buf bytes.Buffer
	err := m.Export(&buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}
