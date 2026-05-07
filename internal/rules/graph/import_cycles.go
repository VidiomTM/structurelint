package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// ImportCyclesRule detects circular dependencies in the import graph.
type ImportCyclesRule struct {
	Graph *graph.ImportGraph
}

func (r *ImportCyclesRule) Name() string {
	return "disallow-import-cycles"
}

func (r *ImportCyclesRule) Check(_ []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	if r.Graph == nil {
		return nil
	}

	d := &cycleDetector{
		graph:    r.Graph,
		ruleName: r.Name(),
		visited:  make(map[string]bool),
		recStack: make(map[string]bool),
	}

	for file := range r.Graph.Dependencies {
		if !d.visited[file] {
			d.path = d.path[:0]
			d.dfs(file)
		}
	}
	return d.violations
}

type cycleDetector struct {
	graph      *graph.ImportGraph
	ruleName   string
	visited    map[string]bool
	recStack   map[string]bool
	path       []string
	violations []rules.Violation
}

func (d *cycleDetector) dfs(file string) {
	d.visited[file] = true
	d.recStack[file] = true
	d.path = append(d.path, file)

	for _, dep := range d.graph.Dependencies[file] {
		depFile := d.resolve(dep, file)
		if depFile == "" {
			continue
		}
		if !d.visited[depFile] {
			d.dfs(depFile)
		} else if d.recStack[depFile] {
			d.recordCycle(file, depFile)
		}
	}

	d.path = d.path[:len(d.path)-1]
	d.recStack[file] = false
}

func (d *cycleDetector) recordCycle(file, depFile string) {
	start := -1
	for i, p := range d.path {
		if p == depFile {
			start = i
			break
		}
	}
	if start < 0 {
		return
	}
	cyclePath := append([]string{}, d.path[start:]...)
	cyclePath = append(cyclePath, depFile)
	d.violations = append(d.violations, rules.Violation{
		Rule:    d.ruleName,
		Path:    file,
		Message: "cyclic dependency detected",
		Context: fmt.Sprintf("cycle: %s", strings.Join(cyclePath, " -> ")),
		Suggestions: []string{
			"Break the cycle with an interface or abstraction",
			"Apply dependency inversion to invert the import direction",
		},
	})
}

func (d *cycleDetector) resolve(importPath, sourceFile string) string {
	files := d.graph.AllFiles
	if len(files) == 0 {
		seen := make(map[string]bool)
		for f := range d.graph.Dependencies {
			seen[f] = true
		}
		for _, deps := range d.graph.Dependencies {
			for _, dep := range deps {
				seen[dep] = true
			}
		}
		for f := range seen {
			files = append(files, f)
		}
	}

	for _, f := range files {
		if f == importPath {
			return importPath
		}
	}

	if strings.HasPrefix(importPath, ".") {
		resolved := filepath.Join(filepath.Dir(sourceFile), importPath)
		for _, f := range files {
			if f == resolved {
				return resolved
			}
		}
		for _, ext := range []string{".go", ".py", ".ts", ".js", ".java", ".cs", ".cpp", ".hpp"} {
			candidate := resolved + ext
			for _, f := range files {
				if f == candidate {
					return candidate
				}
			}
		}
	}

	for _, f := range files {
		if strings.HasSuffix(f, importPath) {
			return f
		}
	}
	return ""
}

func NewImportCyclesRule(g *graph.ImportGraph) *ImportCyclesRule {
	return &ImportCyclesRule{Graph: g}
}
