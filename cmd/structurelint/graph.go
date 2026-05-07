package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/graph/analysis"
	"github.com/Jonathangadeaharder/structurelint/internal/graph/export"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func runGraph(args []string) (err error) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)

	// Output options
	outputPath := fs.String("output", "", "Output file path (default: stdout)")
	outputFormat := fs.String("format", "dot", "Output format: dot, mermaid, mermaid-html (default: dot)")

	// Filtering options
	filterLayer := fs.String("layer", "", "Show only files in this layer")
	maxDepth := fs.Int("depth", 0, "Limit dependency depth (0 = unlimited)")

	// Analysis options
	showCycles := fs.Bool("cycles", false, "Highlight circular dependencies")
	cyclesOnly := fs.Bool("cycles-only", false, "Only detect and report cycles (no graph output)")
	showLayers := fs.Bool("show-layers", true, "Color nodes by their layer")
	highlightViolations := fs.Bool("violations", true, "Highlight layer violations in red")
	simplifyPaths := fs.Bool("simplify", true, "Shorten file paths for readability")

	// Help
	helpFlag := fs.Bool("help", false, "Show help message")
	helpFlagShort := fs.Bool("h", false, "Show help message (shorthand)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Handle help flag
	if *helpFlag || *helpFlagShort {
		printGraphHelp()
		return nil
	}

	// Get path argument
	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	// Load configuration
	configs, _, err := config.FindConfigsWithGitignore(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := config.Merge(configs...)

	// Walk the project
	w := walker.New(path).WithExclude(cfg.Exclude)
	if err := w.Walk(); err != nil {
		return fmt.Errorf("failed to walk project: %w", err)
	}
	files := w.GetFiles()
	_ = w.GetDirs() // Get dirs but don't use for now

	// Build dependency graph
	builder := graph.NewBuilder(path, cfg.Layers)
	depGraph, err := builder.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	// If cycles-only mode, just detect and report cycles
	if *cyclesOnly {
		return reportCycles(depGraph)
	}

	// Create output writer
	var writer *os.File
	if *outputPath == "" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(*outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			if closeErr := writer.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close output file: %w", closeErr)
			}
		}()
	}

	// Export graph in requested format
	switch strings.ToLower(*outputFormat) {
	case "dot":
		return exportDOT(depGraph, writer, export.DOTOptions{
			Title:               fmt.Sprintf("Dependency Graph: %s", filepath.Base(path)),
			ShowLayers:          *showLayers,
			HighlightViolations: *highlightViolations,
			FilterLayer:         *filterLayer,
			MaxDepth:            *maxDepth,
			ShowCycles:          *showCycles,
			SimplifyPaths:       *simplifyPaths,
		})

	case "mermaid":
		return exportMermaid(depGraph, writer, false, export.MermaidOptions{
			Title:               fmt.Sprintf("Dependency Graph: %s", filepath.Base(path)),
			ShowLayers:          *showLayers,
			HighlightViolations: *highlightViolations,
			FilterLayer:         *filterLayer,
			MaxDepth:            *maxDepth,
			ShowCycles:          *showCycles,
			SimplifyPaths:       *simplifyPaths,
			Direction:           "LR",
		})

	case "mermaid-html", "html":
		return exportMermaid(depGraph, writer, true, export.MermaidOptions{
			Title:               fmt.Sprintf("Dependency Graph: %s", filepath.Base(path)),
			ShowLayers:          *showLayers,
			HighlightViolations: *highlightViolations,
			FilterLayer:         *filterLayer,
			MaxDepth:            *maxDepth,
			ShowCycles:          *showCycles,
			SimplifyPaths:       *simplifyPaths,
			Direction:           "LR",
		})

	default:
		return fmt.Errorf("unsupported format: %s (supported: dot, mermaid, mermaid-html)", *outputFormat)
	}
}

func exportDOT(g *graph.ImportGraph, writer *os.File, options export.DOTOptions) error {
	exporter := export.NewDOTExporter(g, options)
	if err := exporter.Export(writer); err != nil {
		return fmt.Errorf("failed to export DOT: %w", err)
	}

	// If writing to file, print success message
	if writer != os.Stdout {
		fmt.Fprintf(os.Stderr, "✓ Graph exported to %s\n", writer.Name())
		fmt.Fprintf(os.Stderr, "  To visualize: dot -Tsvg %s -o graph.svg\n", writer.Name())
	}

	return nil
}

func exportMermaid(g *graph.ImportGraph, writer *os.File, asHTML bool, options export.MermaidOptions) error {
	exporter := export.NewMermaidExporter(g, options)

	var err error
	if asHTML {
		err = exporter.ExportHTML(writer)
	} else {
		err = exporter.ExportWithWrapper(writer)
	}

	if err != nil {
		return fmt.Errorf("failed to export Mermaid: %w", err)
	}

	// If writing to file, print success message
	if writer != os.Stdout {
		if asHTML {
			fmt.Fprintf(os.Stderr, "✓ Interactive HTML graph exported to %s\n", writer.Name())
			fmt.Fprintf(os.Stderr, "  Open in browser to view\n")
		} else {
			fmt.Fprintf(os.Stderr, "✓ Mermaid graph exported to %s\n", writer.Name())
			fmt.Fprintf(os.Stderr, "  Render in GitHub README or use mermaid-cli\n")
		}
	}

	return nil
}

func reportCycles(g *graph.ImportGraph) error {
	detector := analysis.NewCycleDetector(g)
	cycles := detector.FindAllCycles()

	if len(cycles) == 0 {
		fmt.Println("✓ No circular dependencies found")
		return nil
	}

	fmt.Printf("✗ Found %d circular dependencies:\n\n", len(cycles))

	for i, cycle := range cycles {
		fmt.Printf("%d. Cycle of length %d:\n", i+1, cycle.Length)
		fmt.Printf("   %s\n\n", cycle.String())
	}

	// Also report strongly connected components
	sccs := detector.GetStronglyConnectedComponents()
	if len(sccs) > 0 {
		fmt.Printf("Strongly Connected Components: %d\n", len(sccs))
		for i, scc := range sccs {
			if len(scc) > 1 {
				fmt.Printf("%d. %d files in cycle: %s\n", i+1, len(scc), strings.Join(scc, ", "))
			}
		}
	}

	// Exit with error code if cycles found
	os.Exit(1)
	return nil
}

func printGraphHelp() {
	fmt.Println(`structurelint graph - Visualize and analyze dependency graphs

Usage:
  structurelint graph [options] [path]

Description:
  Generate visual dependency graphs in various formats (DOT, Mermaid, HTML).
  Analyze circular dependencies and layer violations.

Output Options:
  --output <path>            Write to file instead of stdout
  --format <format>          Output format: dot, mermaid, mermaid-html
                             (default: dot)

Filtering Options:
  --layer <name>             Show only files in this layer
  --depth <n>                Limit dependency depth (0 = unlimited)

Analysis Options:
  --cycles                   Highlight circular dependencies (orange edges)
  --cycles-only              Only detect cycles, don't generate graph
  --violations               Highlight layer violations (red edges) (default: true)
  --show-layers              Color nodes by their layer (default: true)
  --simplify                 Shorten file paths for readability (default: true)

Examples:
  # Generate DOT file for GraphViz
  structurelint graph --output graph.dot

  # Generate SVG using GraphViz
  structurelint graph --output graph.dot && dot -Tsvg graph.dot -o graph.svg

  # Generate interactive HTML
  structurelint graph --format mermaid-html --output graph.html

  # Generate Mermaid for GitHub
  structurelint graph --format mermaid --output graph.md

  # Show only domain layer dependencies
  structurelint graph --layer domain --output domain.dot

  # Detect circular dependencies
  structurelint graph --cycles-only

  # Limit depth to 3 levels
  structurelint graph --depth 3 --output shallow.dot

Output Formats:
  dot           - GraphViz DOT format (use with 'dot' command)
  mermaid       - Mermaid markdown format (for GitHub/docs)
  mermaid-html  - Interactive HTML with embedded Mermaid

Visualization Tips:
  - Use DOT format with GraphViz for high-quality diagrams:
      dot -Tsvg graph.dot -o graph.svg
      dot -Tpng graph.dot -o graph.png

  - Use Mermaid HTML for interactive exploration:
      structurelint graph --format mermaid-html -o graph.html
      open graph.html

  - Use Mermaid markdown in GitHub READMEs:
      structurelint graph --format mermaid >> README.md

  - Use --depth to focus on specific areas:
      structurelint graph --depth 2 --layer api

Layer Violations:
  - Red edges: Dependencies that violate layer rules
  - Orange edges: Circular dependencies (with --cycles)
  - Colored nodes: Files grouped by architectural layer

Cycle Detection:
  - --cycles-only reports all circular dependencies
  - Useful in CI/CD to enforce acyclic architecture
  - Exit code 1 if cycles found, 0 otherwise

Documentation:
  https://github.com/Jonathangadeaharder/structurelint`)
}
