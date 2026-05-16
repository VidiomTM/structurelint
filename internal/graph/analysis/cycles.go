// Package analysis provides graph analysis utilities
package analysis

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
)

// Cycle represents a circular dependency
type Cycle struct {
	// Nodes in the cycle, in order
	Nodes []string

	// Length of the cycle
	Length int
}

// String returns a human-readable representation of the cycle
func (c *Cycle) String() string {
	if len(c.Nodes) == 0 {
		return "(empty cycle)"
	}
	// Show cycle as: A -> B -> C -> A
	path := strings.Join(c.Nodes, " -> ")
	return fmt.Sprintf("%s -> %s", path, c.Nodes[0])
}

// CycleDetector finds circular dependencies in a graph
type CycleDetector struct {
	graph *graph.ImportGraph
}

// NewCycleDetector creates a new cycle detector
func NewCycleDetector(g *graph.ImportGraph) *CycleDetector {
	return &CycleDetector{graph: g}
}

// FindAllCycles finds all cycles in the graph
func (d *CycleDetector) FindAllCycles() []Cycle {
	var cycles []Cycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range d.graph.GetDependencies(node) {
			if !visited[dep] {
				dfs(dep)
			} else if recStack[dep] {
				// Found a cycle - extract it
				cycle := d.extractCycle(path, dep)
				if cycle != nil && !d.isDuplicate(cycles, cycle) {
					cycles = append(cycles, *cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	// Check all nodes
	allNodes := d.getAllNodes()
	for _, node := range allNodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// HasCycle checks if the graph contains any cycles
func (d *CycleDetector) HasCycle() bool {
	state := &dfsCycleState{
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
	}
	for _, node := range d.getAllNodes() {
		if !state.visited[node] && d.hasCycleDFS(node, state) {
			return true
		}
	}
	return false
}

type dfsCycleState struct {
	visited  map[string]bool
	recStack map[string]bool
}

func (d *CycleDetector) hasCycleDFS(node string, state *dfsCycleState) bool {
	state.visited[node] = true
	state.recStack[node] = true

	for _, dep := range d.graph.GetDependencies(node) {
		if !state.visited[dep] {
			if d.hasCycleDFS(dep, state) {
				return true
			}
		} else if state.recStack[dep] {
			return true
		}
	}

	state.recStack[node] = false
	return false
}

// FindCyclesInLayer finds cycles within a specific layer
func (d *CycleDetector) FindCyclesInLayer(layerName string) []Cycle {
	layer := d.graph.FindLayerByName(layerName)
	if layer == nil {
		return nil
	}

	// Get nodes in this layer
	var layerNodes []string
	for file, fileLayer := range d.graph.FileLayers {
		if fileLayer.Name == layerName {
			layerNodes = append(layerNodes, file)
		}
	}

	// Build subgraph containing only layer nodes
	subgraphDeps := make(map[string][]string)
	for _, node := range layerNodes {
		deps := d.graph.GetDependencies(node)
		for _, dep := range deps {
			depLayer := d.graph.GetLayerForFile(dep)
			if depLayer != nil && depLayer.Name == layerName {
				subgraphDeps[node] = append(subgraphDeps[node], dep)
			}
		}
	}

	// Find cycles in subgraph
	return d.findCyclesInSubgraph(layerNodes, subgraphDeps)
}

// findCyclesInSubgraph finds cycles in a subgraph
func (d *CycleDetector) findCyclesInSubgraph(nodes []string, deps map[string][]string) []Cycle {
	var cycles []Cycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range deps[node] {
			if !visited[dep] {
				dfs(dep)
			} else if recStack[dep] {
				cycle := d.extractCycle(path, dep)
				if cycle != nil && !d.isDuplicate(cycles, cycle) {
					cycles = append(cycles, *cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	for _, node := range nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// extractCycle extracts a cycle from the current path
func (d *CycleDetector) extractCycle(path []string, cycleNode string) *Cycle {
	// Find where the cycle starts
	cycleStart := -1
	for i, node := range path {
		if node == cycleNode {
			cycleStart = i
			break
		}
	}

	if cycleStart == -1 {
		return nil
	}

	// Extract cycle nodes
	cycleNodes := make([]string, 0, len(path)-cycleStart)
	for i := cycleStart; i < len(path); i++ {
		cycleNodes = append(cycleNodes, path[i])
	}

	return &Cycle{
		Nodes:  cycleNodes,
		Length: len(cycleNodes),
	}
}

// isDuplicate checks if a cycle is already in the list
func (d *CycleDetector) isDuplicate(cycles []Cycle, newCycle *Cycle) bool {
	for _, existing := range cycles {
		if d.cyclesEqual(&existing, newCycle) {
			return true
		}
	}
	return false
}

// cyclesEqual checks if two cycles are equivalent (same nodes in same order, allowing rotation)
func (d *CycleDetector) cyclesEqual(c1, c2 *Cycle) bool {
	if c1.Length != c2.Length {
		return false
	}

	// Try all rotations
	for offset := 0; offset < c1.Length; offset++ {
		matches := true
		for i := 0; i < c1.Length; i++ {
			if c1.Nodes[i] != c2.Nodes[(i+offset)%c2.Length] {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}

	return false
}

// getAllNodes returns all nodes in the graph
func (d *CycleDetector) getAllNodes() []string {
	nodeSet := make(map[string]bool)

	// Add all files
	for _, file := range d.graph.AllFiles {
		nodeSet[file] = true
	}

	// Add all nodes from dependencies map
	for node := range d.graph.Dependencies {
		nodeSet[node] = true
	}

	// Convert to slice
	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}

	return nodes
}

// GetStronglyConnectedComponents finds all strongly connected components
// (maximal sets of nodes where every node is reachable from every other)
func (d *CycleDetector) GetStronglyConnectedComponents() [][]string {
	state := &tarjanState{
		indices: make(map[string]int),
		lowlinks: make(map[string]int),
		onStack: make(map[string]bool),
		stack: make([]string, 0),
		sccs: make([][]string, 0),
	}

	allNodes := d.getAllNodes()
	for _, node := range allNodes {
		if _, visited := state.indices[node]; !visited {
			d.tarjanStrongConnect(node, state)
		}
	}

	return state.sccs
}

type tarjanState struct {
	index    int
	stack    []string
	indices  map[string]int
	lowlinks map[string]int
	onStack  map[string]bool
	sccs     [][]string
}

func (d *CycleDetector) tarjanStrongConnect(node string, state *tarjanState) {
	// Initialize node
	state.indices[node] = state.index
	state.lowlinks[node] = state.index
	state.index++
	state.stack = append(state.stack, node)
	state.onStack[node] = true

	// Process dependencies
	d.processTarjanDependencies(node, state)

	// If node is a root node, pop the stack to form SCC
	d.extractTarjanSCC(node, state)
}

func (d *CycleDetector) processTarjanDependencies(node string, state *tarjanState) {
	for _, dep := range d.graph.GetDependencies(node) {
		if _, visited := state.indices[dep]; !visited {
			d.tarjanStrongConnect(dep, state)
			if state.lowlinks[dep] < state.lowlinks[node] {
				state.lowlinks[node] = state.lowlinks[dep]
			}
		} else if state.onStack[dep] {
			if state.indices[dep] < state.lowlinks[node] {
				state.lowlinks[node] = state.indices[dep]
			}
		}
	}
}

func (d *CycleDetector) extractTarjanSCC(node string, state *tarjanState) {
	if state.lowlinks[node] == state.indices[node] {
		var scc []string
		for {
			w := state.stack[len(state.stack)-1]
			state.stack = state.stack[:len(state.stack)-1]
			state.onStack[w] = false
			scc = append(scc, w)
			if w == node {
				break
			}
		}
		if len(scc) > 1 || d.hasSelfLoop(scc[0]) {
			state.sccs = append(state.sccs, scc)
		}
	}
}

// hasSelfLoop checks if a node has a dependency on itself
func (d *CycleDetector) hasSelfLoop(node string) bool {
	deps := d.graph.GetDependencies(node)
	for _, dep := range deps {
		if dep == node {
			return true
		}
	}
	return false
}
