package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// LayerBoundariesRule enforces architectural layer boundaries
type LayerBoundariesRule struct {
	Graph *graph.ImportGraph
}

// Name returns the rule name
func (r *LayerBoundariesRule) Name() string {
	return "enforce-layer-boundaries"
}

// Check validates that imports respect layer boundaries
func (r *LayerBoundariesRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	if r.Graph == nil {
		return []rules.Violation{}
	}

	var violations []rules.Violation

	for sourceFile, dependencies := range r.Graph.Dependencies {
		violations = append(violations, r.checkFileDependencies(sourceFile, dependencies, files)...)
	}

	return violations
}

func (r *LayerBoundariesRule) checkFileDependencies(sourceFile string, dependencies []string, files []walker.FileInfo) []rules.Violation {
	sourceLayer := r.Graph.GetLayerForFile(sourceFile)
	if sourceLayer == nil {
		return nil
	}

	var violations []rules.Violation
	for _, dep := range dependencies {
		violation := r.checkSingleDependency(sourceFile, dep, sourceLayer, files)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}
	return violations
}

func (r *LayerBoundariesRule) checkSingleDependency(sourceFile, dep string, sourceLayer *config.Layer, files []walker.FileInfo) *rules.Violation {
	targetFile := r.resolveDependencyToFile(dep, files)
	if targetFile == "" {
		return nil
	}
	targetLayer := r.Graph.GetLayerForFile(targetFile)
	if r.Graph.CanLayerDependOn(sourceLayer, targetLayer) {
		return nil
	}
	targetLayerName := "unknown"
	if targetLayer != nil {
		targetLayerName = targetLayer.Name
	}
	return &rules.Violation{
		Rule: r.Name(),
		Path: sourceFile,
		Message: fmt.Sprintf(
			"layer '%s' cannot import from layer '%s' (imported: %s)",
			sourceLayer.Name,
			targetLayerName,
			targetFile,
		),
	}
}

// resolveDependencyToFile attempts to resolve an import path to an actual file in the project
func (r *LayerBoundariesRule) resolveDependencyToFile(dep string, files []walker.FileInfo) string {
	// Try exact match
	for _, file := range files {
		if file.Path == dep {
			return file.Path
		}
	}

	// Try with common extensions
	extensions := []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py"}
	for _, ext := range extensions {
		testPath := dep + ext
		for _, file := range files {
			if file.Path == testPath {
				return file.Path
			}
		}
	}

	// Try as directory with index file
	indexFiles := []string{"index.ts", "index.tsx", "index.js", "index.jsx"}
	for _, indexFile := range indexFiles {
		testPath := filepath.Join(dep, indexFile)
		for _, file := range files {
			if file.Path == testPath {
				return file.Path
			}
		}
	}

	// Try matching any file in the dependency path (for Go packages)
	for _, file := range files {
		if strings.HasPrefix(file.Path, dep) {
			return file.Path
		}
	}

	return ""
}

// NewLayerBoundariesRule creates a new LayerBoundariesRule
func NewLayerBoundariesRule(importGraph *graph.ImportGraph) *LayerBoundariesRule {
	return &LayerBoundariesRule{
		Graph: importGraph,
	}
}
