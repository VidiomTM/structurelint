package export

import (
	"fmt"
	"io"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
)

// MermaidExporter exports dependency graphs in Mermaid format
type MermaidExporter struct {
	graph   *graph.ImportGraph
	options MermaidOptions
}

// MermaidOptions configures the Mermaid export
type MermaidOptions struct {
	// Title of the graph
	Title string

	// ShowLayers colors nodes by their layer
	ShowLayers bool

	// HighlightViolations marks illegal dependencies with red edges
	HighlightViolations bool

	// FilterLayer only shows files in this layer (empty = all)
	FilterLayer string

	// MaxDepth limits dependency depth (0 = unlimited)
	MaxDepth int

	// ShowCycles highlights circular dependencies
	ShowCycles bool

	// SimplifyPaths shortens file paths for readability
	SimplifyPaths bool

	// Direction of the graph (LR, RL, TB, BT)
	Direction string
}

// NewMermaidExporter creates a new Mermaid exporter
func NewMermaidExporter(g *graph.ImportGraph, options MermaidOptions) *MermaidExporter {
	if options.Title == "" {
		options.Title = "Dependency Graph"
	}
	if options.Direction == "" {
		options.Direction = "LR"
	}
	return &MermaidExporter{
		graph:   g,
		options: options,
	}
}

// Export writes the graph in Mermaid format to the writer
func (e *MermaidExporter) Export(w io.Writer) error {
	if err := e.writeMermaidHeader(w); err != nil {
		return err
	}

	dotExporter, nodes := e.prepareExport()
	if len(nodes) == 0 {
		return e.writeEmptyGraph(w)
	}

	cycles := e.getCyclesIfEnabled(dotExporter)
	nodeIDs := e.createNodeIDMap(nodes)

	edgeCtx := &mermaidEdgeCtx{
		w:           w,
		nodeIDs:     nodeIDs,
		cycles:      cycles,
		dotExporter: dotExporter,
	}
	if err := e.writeMermaidEdges(edgeCtx, nodes); err != nil {
		return err
	}

	if e.options.ShowLayers {
		if err := e.writeStyles(w, nodeIDs, nodes); err != nil {
			return err
		}
	}

	return nil
}

// writeMermaidHeader writes the Mermaid graph header
func (e *MermaidExporter) writeMermaidHeader(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "graph %s\n", e.options.Direction); err != nil {
		return fmt.Errorf("failed to write Mermaid header: %w", err)
	}
	return nil
}

// prepareExport prepares the DOT exporter and retrieves filtered nodes
func (e *MermaidExporter) prepareExport() (*DOTExporter, []string) {
	dotExporter := NewDOTExporter(e.graph, DOTOptions{
		FilterLayer:   e.options.FilterLayer,
		MaxDepth:      e.options.MaxDepth,
		SimplifyPaths: e.options.SimplifyPaths,
	})
	return dotExporter, dotExporter.getFilteredNodes()
}

// writeEmptyGraph writes an empty graph message
func (e *MermaidExporter) writeEmptyGraph(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "  %% No nodes to display\n"); err != nil {
		return fmt.Errorf("failed to write Mermaid comment: %w", err)
	}
	return nil
}

// getCyclesIfEnabled returns cycles if cycle detection is enabled
func (e *MermaidExporter) getCyclesIfEnabled(dotExporter *DOTExporter) map[string]map[string]bool {
	if e.options.ShowCycles {
		return dotExporter.detectAllCycles()
	}
	return make(map[string]map[string]bool)
}

// createNodeIDMap creates a mapping of node names to IDs
func (e *MermaidExporter) createNodeIDMap(nodes []string) map[string]string {
	nodeIDs := make(map[string]string)
	for i, node := range nodes {
		nodeIDs[node] = fmt.Sprintf("n%d", i)
	}
	return nodeIDs
}

// writeMermaidEdges writes all edges in Mermaid format
func (e *MermaidExporter) writeMermaidEdges(
	ctx *mermaidEdgeCtx,
	nodes []string,
) error {
	for _, fromNode := range nodes {
		if err := e.writeNodeEdges(ctx, fromNode); err != nil {
			return err
		}
	}
	return nil
}

// writeNodeEdges writes edges for a single node
func (e *MermaidExporter) writeNodeEdges(
	ctx *mermaidEdgeCtx,
	fromNode string,
) error {
	fromID := ctx.nodeIDs[fromNode]
	fromLabel := e.getNodeLabel(fromNode, ctx.dotExporter)
	deps := e.graph.GetDependencies(fromNode)
	hasEdges := false

	for _, toNode := range deps {
		toID, exists := ctx.nodeIDs[toNode]
		if !exists {
			continue
		}

		hasEdges = true
		if err := e.writeSingleEdge(ctx, fromNode, toNode, fromID, toID, fromLabel); err != nil {
			return err
		}
	}

	if !hasEdges {
		return e.writeIsolatedNode(ctx.w, fromID, fromLabel)
	}

	return nil
}

// getNodeLabel returns the label for a node
func (e *MermaidExporter) getNodeLabel(node string, dotExporter *DOTExporter) string {
	if e.options.SimplifyPaths {
		return dotExporter.simplifyPath(node)
	}
	return node
}

type mermaidEdgeCtx struct {
	w           io.Writer
	nodeIDs     map[string]string
	cycles      map[string]map[string]bool
	dotExporter *DOTExporter
}

// writeSingleEdge writes a single Mermaid edge
func (e *MermaidExporter) writeSingleEdge(
	ctx *mermaidEdgeCtx,
	fromNode, toNode, fromID, toID, fromLabel string,
) error {
	toLabel := e.getNodeLabel(toNode, ctx.dotExporter)
	edgeStyle := e.getMermaidEdgeStyle(ctx, fromNode, toNode, fromID, toID, fromLabel, toLabel)

	if _, err := fmt.Fprintf(ctx.w, "%s", edgeStyle); err != nil {
		return fmt.Errorf("failed to write Mermaid edge: %w", err)
	}

	return nil
}

// getMermaidEdgeStyle determines the Mermaid edge style
func (e *MermaidExporter) getMermaidEdgeStyle(
	ctx *mermaidEdgeCtx,
	fromNode, toNode, fromID, toID, fromLabel, toLabel string,
) string {
	isCycle := ctx.cycles[fromNode] != nil && ctx.cycles[fromNode][toNode]
	isViolation := ctx.dotExporter.isViolation(fromNode, toNode)

	if isCycle && e.options.ShowCycles {
		return fmt.Sprintf("  %s[\"%s\"] -.->|cycle| %s[\"%s\"]\n",
			fromID, fromLabel, toID, toLabel)
	}

	if isViolation && e.options.HighlightViolations {
		return fmt.Sprintf("  %s[\"%s\"] -.->|violation| %s[\"%s\"]\n",
			fromID, fromLabel, toID, toLabel)
	}

	return fmt.Sprintf("  %s[\"%s\"] --> %s[\"%s\"]\n",
		fromID, fromLabel, toID, toLabel)
}

// writeIsolatedNode writes a node with no edges
func (e *MermaidExporter) writeIsolatedNode(w io.Writer, nodeID, label string) error {
	if _, err := fmt.Fprintf(w, "  %s[\"%s\"]\n", nodeID, label); err != nil {
		return fmt.Errorf("failed to write Mermaid node: %w", err)
	}
	return nil
}

// writeStyles adds CSS styling for layer colors
func (e *MermaidExporter) writeStyles(w io.Writer, nodeIDs map[string]string, nodes []string) error {
	if _, err := fmt.Fprintf(w, "\n  %% Layer styling\n"); err != nil {
		return fmt.Errorf("failed to write style comment: %w", err)
	}

	// Apply layer styles
	if err := e.applyLayerStyles(w, nodeIDs, nodes); err != nil {
		return err
	}

	// Apply violation and cycle styles
	if err := e.applySpecialStyles(w); err != nil {
		return err
	}

	return nil
}

func (e *MermaidExporter) applyLayerStyles(w io.Writer, nodeIDs map[string]string, nodes []string) error {
	layerStyles := map[string]string{
		"domain":         "fill:#C8E6C9,stroke:#2E7D32,stroke-width:2px",
		"application":    "fill:#BBDEFB,stroke:#1565C0,stroke-width:2px",
		"infrastructure": "fill:#FFCDD2,stroke:#C62828,stroke-width:2px",
		"presentation":   "fill:#FFE0B2,stroke:#F57C00,stroke-width:2px",
		"api":            "fill:#E1BEE7,stroke:#6A1B9A,stroke-width:2px",
		"cmd":            "fill:#E0E0E0,stroke:#424242,stroke-width:2px",
		"internal":       "fill:#B2DFDB,stroke:#00695C,stroke-width:2px",
	}

	// Group nodes by layer
	layerNodes := make(map[string][]string)
	for _, node := range nodes {
		layer := e.graph.GetLayerForFile(node)
		if layer != nil {
			layerNodes[layer.Name] = append(layerNodes[layer.Name], node)
		}
	}

	// Apply styles to each layer's nodes
	for layerName, layerNodeList := range layerNodes {
		style, ok := layerStyles[layerName]
		if !ok {
			continue
		}

		for _, node := range layerNodeList {
			nodeID := nodeIDs[node]
			if _, err := fmt.Fprintf(w, "  style %s %s\n", nodeID, style); err != nil {
				return fmt.Errorf("failed to write node style: %w", err)
			}
		}
	}

	return nil
}

func (e *MermaidExporter) applySpecialStyles(w io.Writer) error {
	if e.options.HighlightViolations {
		if _, err := fmt.Fprintf(w, "  linkStyle default stroke:#333,stroke-width:1px\n"); err != nil {
			return fmt.Errorf("failed to write violation link style: %w", err)
		}
	}

	if e.options.ShowCycles {
		if _, err := fmt.Fprintf(w, "  linkStyle default stroke:#333,stroke-width:1px\n"); err != nil {
			return fmt.Errorf("failed to write cycle link style: %w", err)
		}
	}

	return nil
}

// ExportWithWrapper wraps the Mermaid graph in markdown code fences
func (e *MermaidExporter) ExportWithWrapper(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "```mermaid\n"); err != nil {
		return fmt.Errorf("failed to write markdown opening: %w", err)
	}
	if err := e.Export(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "```\n"); err != nil {
		return fmt.Errorf("failed to write markdown closing: %w", err)
	}
	return nil
}

// ExportHTML generates an HTML file with embedded Mermaid
func (e *MermaidExporter) ExportHTML(w io.Writer) error {
	if _, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <script src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      margin: 20px;
      background: #f5f5f5;
    }
    .container {
      max-width: 100%%;
      background: white;
      padding: 20px;
      border-radius: 8px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    h1 {
      color: #333;
      margin-top: 0;
    }
    .mermaid {
      text-align: center;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>%s</h1>
    <div class="mermaid">
`, e.options.Title, e.options.Title); err != nil {
		return fmt.Errorf("failed to write HTML header: %w", err)
	}

	if err := e.Export(w); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, `    </div>
  </div>
  <script>
    mermaid.initialize({ startOnLoad: true, theme: 'default' });
  </script>
</body>
</html>
`); err != nil {
		return fmt.Errorf("failed to write HTML footer: %w", err)
	}

	return nil
}
