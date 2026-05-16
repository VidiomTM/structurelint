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

type graphOptions struct {
	outputPath   string
	outputFormat string
	filterLayer  string
	maxDepth     int
	showCycles   bool
	cyclesOnly   bool
	showLayers   bool
	highlight    bool
	simplify     bool
}

func runGraph(args []string) (err error) {
	opts, err := parseGraphArgs(args)
	if err != nil {
		return err
	}
	if opts == nil {
		return nil
	}

	path := "."
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		path = args[0]
	}

	cfg, depGraph, err := buildGraphForPath(path)
	if err != nil {
		return err
	}
	_ = cfg

	if opts.cyclesOnly {
		return reportCycles(depGraph)
	}

	writer, err := createGraphWriter(opts.outputPath)
	if err != nil {
		return err
	}
	if writer == os.Stdout {
		defer func() { _ = writer.Close() }()
	} else {
		defer func() {
			if closeErr := writer.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close output file: %w", closeErr)
			}
		}()
	}

	return exportGraph(depGraph, writer, opts, path)
}

func parseGraphArgs(args []string) (*graphOptions, error) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	outputPath := fs.String("output", "", "Output file path (default: stdout)")
	outputFormat := fs.String("format", "dot", "Output format: dot, mermaid, mermaid-html (default: dot)")
	filterLayer := fs.String("layer", "", "Show only files in this layer")
	maxDepth := fs.Int("depth", 0, "Limit dependency depth (0 = unlimited)")
	showCycles := fs.Bool("cycles", false, "Highlight circular dependencies")
	cyclesOnly := fs.Bool("cycles-only", false, "Only detect and report cycles (no graph output)")
	showLayers := fs.Bool("show-layers", true, "Color nodes by their layer")
	highlightViolations := fs.Bool("violations", true, "Highlight layer violations in red")
	simplifyPaths := fs.Bool("simplify", true, "Shorten file paths for readability")
	helpFlag := fs.Bool("help", false, "Show help message")
	helpFlagShort := fs.Bool("h", false, "Show help message (shorthand)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if *helpFlag || *helpFlagShort {
		printGraphHelp()
		return nil, nil
	}

	return &graphOptions{
		outputPath:   *outputPath,
		outputFormat: *outputFormat,
		filterLayer:  *filterLayer,
		maxDepth:     *maxDepth,
		showCycles:   *showCycles,
		cyclesOnly:   *cyclesOnly,
		showLayers:   *showLayers,
		highlight:    *highlightViolations,
		simplify:     *simplifyPaths,
	}, nil
}

func buildGraphForPath(path string) (*config.Config, *graph.ImportGraph, error) {
	configs, _, err := config.FindConfigsWithGitignore(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}
	cfg := config.Merge(configs...)

	w := walker.New(path).WithExclude(cfg.Exclude)
	if err := w.Walk(); err != nil {
		return nil, nil, fmt.Errorf("failed to walk project: %w", err)
	}

	builder := graph.NewBuilder(path, cfg.Layers)
	depGraph, err := builder.Build(w.GetFiles())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build graph: %w", err)
	}

	return cfg, depGraph, nil
}

func createGraphWriter(outputPath string) (*os.File, error) {
	if outputPath == "" {
		return os.Stdout, nil
	}
	return os.Create(outputPath)
}

func exportGraph(depGraph *graph.ImportGraph, writer *os.File, opts *graphOptions, path string) error {
	graphTitle := fmt.Sprintf("Dependency Graph: %s", filepath.Base(path))

	switch strings.ToLower(opts.outputFormat) {
	case "dot":
		return exportDOT(depGraph, writer, export.DOTOptions{
			Title:               graphTitle,
			ShowLayers:          opts.showLayers,
			HighlightViolations: opts.highlight,
			FilterLayer:         opts.filterLayer,
			MaxDepth:            opts.maxDepth,
			ShowCycles:          opts.showCycles,
			SimplifyPaths:       opts.simplify,
		})
	case "mermaid":
		return exportMermaid(depGraph, writer, false, export.MermaidOptions{
			Title:               graphTitle,
			ShowLayers:          opts.showLayers,
			HighlightViolations: opts.highlight,
			FilterLayer:         opts.filterLayer,
			MaxDepth:            opts.maxDepth,
			ShowCycles:          opts.showCycles,
			SimplifyPaths:       opts.simplify,
			Direction:           "LR",
		})
	case "mermaid-html", "html":
		return exportMermaid(depGraph, writer, true, export.MermaidOptions{
			Title:               graphTitle,
			ShowLayers:          opts.showLayers,
			HighlightViolations: opts.highlight,
			FilterLayer:         opts.filterLayer,
			MaxDepth:            opts.maxDepth,
			ShowCycles:          opts.showCycles,
			SimplifyPaths:       opts.simplify,
			Direction:           "LR",
		})
	default:
		return fmt.Errorf("unsupported format: %s (supported: dot, mermaid, mermaid-html)", opts.outputFormat)
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
