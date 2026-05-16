// Package export provides graph visualization and export functionality
package export

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
)

// DOTExporter exports dependency graphs in GraphViz DOT format
type DOTExporter struct {
	graph   *graph.ImportGraph
	options DOTOptions
}

// DOTOptions configures the DOT export
type DOTOptions struct {
	// Title of the graph
	Title string

	// ShowLayers colors nodes by their layer
	ShowLayers bool

	// HighlightViolations marks illegal dependencies in red
	HighlightViolations bool

	// FilterLayer only shows files in this layer (empty = all)
	FilterLayer string

	// MaxDepth limits dependency depth (0 = unlimited)
	MaxDepth int

	// ShowCycles highlights circular dependencies
	ShowCycles bool

	// SimplifyPaths shortens file paths for readability
	SimplifyPaths bool
}

// NewDOTExporter creates a new DOT exporter
func NewDOTExporter(g *graph.ImportGraph, options DOTOptions) *DOTExporter {
	if options.Title == "" {
		options.Title = "Dependency Graph"
	}
	return &DOTExporter{
		graph:   g,
		options: options,
	}
}

// Export writes the graph in DOT format to the writer
func (e *DOTExporter) Export(w io.Writer) error {
	if err := e.writeDOTHeader(w); err != nil {
		return err
	}

	nodes := e.getFilteredNodes()
	if len(nodes) == 0 {
		return e.writeEmptyGraph(w)
	}

	cycles := e.getCyclesIfEnabled()
	nodeIDs, err := e.writeNodes(w, nodes)
	if err != nil {
		return err
	}

	if err := e.writeEdges(w, nodes, nodeIDs, cycles); err != nil {
		return err
	}

	if e.options.ShowLayers {
		if err := e.writeLegend(w); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "}\n"); err != nil {
		return fmt.Errorf("failed to write DOT closing brace: %w", err)
	}

	return nil
}

// writeDOTHeader writes the DOT file header
func (e *DOTExporter) writeDOTHeader(w io.Writer) error {
	headers := []struct {
		content string
		errMsg  string
	}{
		{fmt.Sprintf("digraph \"%s\" {\n", e.options.Title), "failed to write DOT header"},
		{"  rankdir=LR;\n", "failed to write DOT rankdir"},
		{"  node [shape=box, style=rounded];\n", "failed to write DOT node style"},
		{"  edge [arrowhead=vee];\n\n", "failed to write DOT edge style"},
	}

	for _, h := range headers {
		if _, err := fmt.Fprintf(w, "%s", h.content); err != nil {
			return fmt.Errorf("%s: %w", h.errMsg, err)
		}
	}

	return nil
}

// writeEmptyGraph writes an empty graph and closes it
func (e *DOTExporter) writeEmptyGraph(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "  // No nodes to display\n"); err != nil {
		return fmt.Errorf("failed to write DOT comment: %w", err)
	}
	if _, err := fmt.Fprintf(w, "}\n"); err != nil {
		return fmt.Errorf("failed to write DOT closing: %w", err)
	}
	return nil
}

// getCyclesIfEnabled returns cycles if cycle detection is enabled
func (e *DOTExporter) getCyclesIfEnabled() map[string]map[string]bool {
	if e.options.ShowCycles {
		return e.detectAllCycles()
	}
	return make(map[string]map[string]bool)
}

// writeNodes writes all nodes and returns the node ID mapping
func (e *DOTExporter) writeNodes(w io.Writer, nodes []string) (map[string]string, error) {
	nodeIDs := make(map[string]string)

	for i, node := range nodes {
		nodeID := fmt.Sprintf("n%d", i)
		nodeIDs[node] = nodeID

		if err := e.writeSingleNode(w, node, nodeID); err != nil {
			return nil, err
		}
	}

	if _, err := fmt.Fprintf(w, "\n"); err != nil {
		return nil, fmt.Errorf("failed to write node separator: %w", err)
	}
	return nodeIDs, nil
}

// writeSingleNode writes a single node definition
func (e *DOTExporter) writeSingleNode(w io.Writer, node, nodeID string) error {
	label := node
	if e.options.SimplifyPaths {
		label = e.simplifyPath(node)
	}

	color := e.getNodeColor(node)
	fillColor := e.getNodeFillColor(node)

	if _, err := fmt.Fprintf(w, "  %s [label=\"%s\", color=\"%s\", fillcolor=\"%s\", style=\"rounded,filled\"];\n",
		nodeID, label, color, fillColor); err != nil {
		return fmt.Errorf("failed to write DOT node: %w", err)
	}

	return nil
}

// writeEdges writes all edges between nodes
func (e *DOTExporter) writeEdges(w io.Writer, nodes []string, nodeIDs map[string]string, cycles map[string]map[string]bool) error {
	for _, fromNode := range nodes {
		if err := e.writeNodeEdges(w, fromNode, nodeIDs, cycles); err != nil {
			return err
		}
	}
	return nil
}

// writeNodeEdges writes all edges from a single node
func (e *DOTExporter) writeNodeEdges(w io.Writer, fromNode string, nodeIDs map[string]string, cycles map[string]map[string]bool) error {
	fromID := nodeIDs[fromNode]
	deps := e.graph.GetDependencies(fromNode)

	for _, toNode := range deps {
		toID, exists := nodeIDs[toNode]
		if !exists {
			continue
		}

		if err := e.writeSingleEdge(w, fromNode, toNode, fromID, toID, cycles); err != nil {
			return err
		}
	}

	return nil
}

// writeSingleEdge writes a single edge with appropriate styling
func (e *DOTExporter) writeSingleEdge(w io.Writer, fromNode, toNode, fromID, toID string, cycles map[string]map[string]bool) error {
	edgeColor, edgeStyle, edgeWidth := e.getEdgeStyle(fromNode, toNode, cycles)

	if _, err := fmt.Fprintf(w, "  %s -> %s [color=\"%s\", style=\"%s\", penwidth=%s];\n",
		fromID, toID, edgeColor, edgeStyle, edgeWidth); err != nil {
		return fmt.Errorf("failed to write DOT edge: %w", err)
	}

	return nil
}

// getEdgeStyle determines the styling for an edge
func (e *DOTExporter) getEdgeStyle(fromNode, toNode string, cycles map[string]map[string]bool) (color, style, width string) {
	color = "black"
	style = "solid"
	width = "1.0"

	isCycle := cycles[fromNode] != nil && cycles[fromNode][toNode]
	isViolation := e.isViolation(fromNode, toNode)

	if isCycle && e.options.ShowCycles {
		return "orange", "bold", "2.0"
	}

	if isViolation && e.options.HighlightViolations {
		return "red", "bold", "2.0"
	}

	return color, style, width
}

// getFilteredNodes returns nodes to display based on filter options
func (e *DOTExporter) getFilteredNodes() []string {
	var nodes []string

	// Get all files from graph
	allFiles := e.graph.AllFiles
	if len(allFiles) == 0 {
		// Fallback: collect from dependencies map
		for file := range e.graph.Dependencies {
			allFiles = append(allFiles, file)
		}
	}

	// Apply layer filter
	for _, file := range allFiles {
		if e.options.FilterLayer != "" {
			layer := e.graph.GetLayerForFile(file)
			if layer == nil || layer.Name != e.options.FilterLayer {
				continue
			}
		}
		nodes = append(nodes, file)
	}

	// Apply depth filter if specified
	if e.options.MaxDepth > 0 {
		nodes = e.filterByDepth(nodes, e.options.MaxDepth)
	}

	return nodes
}

// filterByDepth limits nodes by dependency depth
func (e *DOTExporter) filterByDepth(nodes []string, maxDepth int) []string {
	roots := e.findRootNodes(nodes)
	if len(roots) == 0 && len(nodes) > 0 {
		roots = []string{nodes[0]}
	}
	visited := e.bfsWithinDepth(roots, maxDepth)
	filtered := make([]string, 0, len(visited))
	for _, node := range nodes {
		if visited[node] {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func (e *DOTExporter) findRootNodes(nodes []string) []string {
	var roots []string
	for _, node := range nodes {
		if e.graph.IncomingRefs[node] == 0 {
			roots = append(roots, node)
		}
	}
	return roots
}

func (e *DOTExporter) bfsWithinDepth(roots []string, maxDepth int) map[string]bool {
	visited := make(map[string]bool)
	depthMap := make(map[string]int)
	queue := make([]string, len(roots))
	copy(queue, roots)

	for _, root := range roots {
		depthMap[root] = 0
		visited[root] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if depthMap[current] >= maxDepth {
			continue
		}
		for _, dep := range e.graph.GetDependencies(current) {
			if !visited[dep] {
				visited[dep] = true
				depthMap[dep] = depthMap[current] + 1
				queue = append(queue, dep)
			}
		}
	}

	return visited
}

// getNodeColor returns the border color for a node based on its layer
func (e *DOTExporter) getNodeColor(node string) string {
	if !e.options.ShowLayers {
		return "black"
	}

	layer := e.graph.GetLayerForFile(node)
	if layer == nil {
		return "gray"
	}

	// Color scheme for different layers
	colors := map[string]string{
		"domain":         "#2E7D32", // Dark green
		"application":    "#1565C0", // Dark blue
		"infrastructure": "#C62828", // Dark red
		"presentation":   "#F57C00", // Dark orange
		"api":            "#6A1B9A", // Dark purple
		"cmd":            "#424242", // Dark gray
		"internal":       "#00695C", // Dark teal
	}

	if color, ok := colors[layer.Name]; ok {
		return color
	}

	return "black"
}

// getNodeFillColor returns the fill color for a node based on its layer
func (e *DOTExporter) getNodeFillColor(node string) string {
	if !e.options.ShowLayers {
		return "#FFFFFF"
	}

	layer := e.graph.GetLayerForFile(node)
	if layer == nil {
		return "#F5F5F5"
	}

	// Light fill colors for different layers
	colors := map[string]string{
		"domain":         "#C8E6C9", // Light green
		"application":    "#BBDEFB", // Light blue
		"infrastructure": "#FFCDD2", // Light red
		"presentation":   "#FFE0B2", // Light orange
		"api":            "#E1BEE7", // Light purple
		"cmd":            "#E0E0E0", // Light gray
		"internal":       "#B2DFDB", // Light teal
	}

	if color, ok := colors[layer.Name]; ok {
		return color
	}

	return "#FFFFFF"
}

// isViolation checks if a dependency violates layer rules
func (e *DOTExporter) isViolation(from, to string) bool {
	fromLayer := e.graph.GetLayerForFile(from)
	toLayer := e.graph.GetLayerForFile(to)

	return !e.graph.CanLayerDependOn(fromLayer, toLayer)
}

// detectAllCycles finds all circular dependencies in the graph
func (e *DOTExporter) detectAllCycles() map[string]map[string]bool {
	detector := &cycleDetector{
		exporter: e,
		cycles:   make(map[string]map[string]bool),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
		path:     make([]string, 0),
	}

	// Check all nodes
	for _, node := range e.graph.AllFiles {
		if !detector.visited[node] {
			detector.dfs(node)
		}
	}

	return detector.cycles
}

// cycleDetector manages state for cycle detection
type cycleDetector struct {
	exporter *DOTExporter
	cycles   map[string]map[string]bool
	visited  map[string]bool
	recStack map[string]bool
	path     []string
}

// dfs performs depth-first search to detect cycles
func (cd *cycleDetector) dfs(node string) bool {
	cd.visited[node] = true
	cd.recStack[node] = true
	cd.path = append(cd.path, node)

	for _, dep := range cd.exporter.graph.GetDependencies(node) {
		if cd.processDependency(dep) {
			return true
		}
	}

	cd.path = cd.path[:len(cd.path)-1]
	cd.recStack[node] = false
	return false
}

// processDependency handles a single dependency during DFS
func (cd *cycleDetector) processDependency(dep string) bool {
	if !cd.visited[dep] {
		return cd.dfs(dep)
	}

	if cd.recStack[dep] {
		cd.recordCycle(dep)
		return true
	}

	return false
}

// recordCycle marks all edges in the detected cycle
func (cd *cycleDetector) recordCycle(dep string) {
	cycleStart := cd.findCycleStart(dep)
	if cycleStart < 0 {
		return
	}

	for i := cycleStart; i < len(cd.path); i++ {
		from := cd.path[i]
		to := cd.getNextNode(i, dep)
		cd.addCycleEdge(from, to)
	}
}

// findCycleStart finds where the cycle begins in the path
func (cd *cycleDetector) findCycleStart(dep string) int {
	for i, n := range cd.path {
		if n == dep {
			return i
		}
	}
	return -1
}

// getNextNode returns the next node in the cycle path
func (cd *cycleDetector) getNextNode(i int, dep string) string {
	if i+1 < len(cd.path) {
		return cd.path[i+1]
	}
	return dep
}

// addCycleEdge adds an edge to the cycles map
func (cd *cycleDetector) addCycleEdge(from, to string) {
	if cd.cycles[from] == nil {
		cd.cycles[from] = make(map[string]bool)
	}
	cd.cycles[from][to] = true
}

// simplifyPath shortens a file path for display
func (e *DOTExporter) simplifyPath(path string) string {
	// Remove common prefixes
	path = strings.TrimPrefix(path, "./")

	// Use only filename if in same directory
	if !strings.Contains(path, "/") {
		return path
	}

	// Show only last 2 path components
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return filepath.Join(parts[len(parts)-2:]...)
	}

	return path
}

// writeLegend adds a legend showing layer colors
func (e *DOTExporter) writeLegend(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "\n  // Legend\n"); err != nil {
		return fmt.Errorf("failed to write legend comment: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  subgraph cluster_legend {\n"); err != nil {
		return fmt.Errorf("failed to write legend subgraph: %w", err)
	}
	if _, err := fmt.Fprintf(w, "    label=\"Layers\";\n"); err != nil {
		return fmt.Errorf("failed to write legend label: %w", err)
	}
	if _, err := fmt.Fprintf(w, "    style=filled;\n"); err != nil {
		return fmt.Errorf("failed to write legend style: %w", err)
	}
	if _, err := fmt.Fprintf(w, "    fillcolor=\"#F0F0F0\";\n"); err != nil {
		return fmt.Errorf("failed to write legend fillcolor: %w", err)
	}

	// Collect unique layers
	layerMap := make(map[string]*config.Layer)
	for _, layer := range e.graph.Layers {
		layerMap[layer.Name] = &layer
	}

	i := 0
	for _, layer := range e.graph.Layers {
		// Temporarily set layer for color lookup
		tempFile := "temp"
		e.graph.FileLayers[tempFile] = &layer
		color := e.getNodeColor(tempFile)
		fillColor := e.getNodeFillColor(tempFile)
		delete(e.graph.FileLayers, tempFile)

		if _, err := fmt.Fprintf(w, "    legend%d [label=\"%s\", color=\"%s\", fillcolor=\"%s\", style=\"rounded,filled\"];\n",
			i, layer.Name, color, fillColor); err != nil {
			return fmt.Errorf("failed to write legend node: %w", err)
		}
		i++
	}

	if _, err := fmt.Fprintf(w, "  }\n"); err != nil {
		return fmt.Errorf("failed to write legend closing: %w", err)
	}
	return nil
}
