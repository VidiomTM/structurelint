package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	init_pkg "github.com/Jonathangadeaharder/structurelint/internal/init"
	"github.com/Jonathangadeaharder/structurelint/internal/linter"
	"github.com/Jonathangadeaharder/structurelint/internal/output"
)

// Version is set during build via ldflags
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check for subcommands
	if len(os.Args) > 1 {
		if err := handleSubcommands(os.Args[1:]); err != nil {
			return err
		}
		// If handleSubcommands returns nil but we processed a command, we should exit
		// However, the current logic implies that if it returns nil, we might continue?
		// Looking at the original code, runGraph etc return error.
		// If we matched a subcommand, we should return the result of that subcommand.
		// But wait, the original code had a switch and returned.
		// I need to extract the switch into a function.
		return nil
	}

	return runLinterWithArgs([]string{})
}

func handleSubcommands(args []string) error {
	switch args[0] {
	case "graph":
		return runGraph(args[1:])
	case "clones":
		return runClones(args[1:])
	case "fix":
		return runFix(args[1:])
	case "tui":
		return runTUI(args[1:])
	case "scaffold":
		return runScaffold(args[1:])
	case "help":
		return handleHelpCommand(args)
	}
	// Not a subcommand, continue to linter
	return runLinterWithArgs(args)
}

func handleHelpCommand(args []string) error {
	if len(args) > 1 {
		switch args[1] {
		case "graph":
			printGraphHelp()
			return nil
		case "clones":
			printClonesHelp()
			return nil
		case "fix":
			printFixHelp()
			return nil
		case "tui":
			printTUIHelp()
			return nil
		case "scaffold":
			printScaffoldHelp()
			return nil
		}
	}
	printHelp()
	return nil
}

func runLinterWithArgs(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("structurelint", flag.ContinueOnError)
	formatFlag := fs.String("format", "text", "Output format: text, json, junit")
	versionFlag := fs.Bool("version", false, "Show version information")
	versionFlagShort := fs.Bool("v", false, "Show version information (shorthand)")
	helpFlag := fs.Bool("help", false, "Show help message")
	helpFlagShort := fs.Bool("h", false, "Show help message (shorthand)")
	initFlag := fs.Bool("init", false, "Initialize configuration")
	presetFlag := fs.String("preset", "", "Preset for --init (e.g. sveltekit, python-monorepo, go-stdlayout)")
	listPresetsFlag := fs.Bool("list-presets", false, "List available --init presets and exit")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Extract path argument once
	path := "."
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	// Handle version flag
	if *versionFlag || *versionFlagShort {
		fmt.Printf("structurelint version %s\n", Version)
		return nil
	}

	// Handle help flag
	if *helpFlag || *helpFlagShort {
		printHelp()
		return nil
	}

	if *listPresetsFlag {
		for _, name := range init_pkg.ListPresets() {
			fmt.Println(name)
		}
		return nil
	}

	// Handle init flag
	if *initFlag {
		return runInit(path, *presetFlag)
	}

	return executeLinter(path, *formatFlag)
}

func executeLinter(path, format string) error {
	// Get output formatter
	formatter, err := output.GetFormatter(format, Version)
	if err != nil {
		return err
	}

	// Create and run linter
	l := linter.New()
	violations, err := l.Lint(path)
	if err != nil {
		if errors.Is(err, linter.ErrNoConfig) {
			fmt.Fprintf(os.Stderr, "Error: no .structurelint.yml configuration file found.\n")
			fmt.Fprintf(os.Stderr, "Create one with: structurelint --init\n")
			os.Exit(1)
		}
		return err
	}

	// Format and report violations
	if len(violations) > 0 {
		formatted, err := formatter.Format(violations)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(formatted)

		// For text format, print error message to stderr
		if format == "text" || format == "" {
			fmt.Fprintf(os.Stderr, "Error: found %d violation(s)\n", len(violations))
		}
		os.Exit(1)
	}

	// Only print success message for text format
	if format == "text" || format == "" {
		fmt.Println("✓ All checks passed")
	} else {
		// For JSON/JUnit, output empty success structure
		formatted, err := formatter.Format(violations)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(formatted)
	}

	return nil
}

func runInit(path, preset string) error {
	configPath := filepath.Join(path, ".structurelint.yml")

	if preset == "" {
		if detected := init_pkg.DetectPreset(path); detected != "" {
			fmt.Printf("Detected preset: %s (override with --preset=<name>)\n", detected)
			preset = detected
		}
	}

	var config string
	if preset != "" {
		c, err := init_pkg.PresetConfig(preset)
		if err != nil {
			return err
		}
		config = c
		fmt.Printf("Using preset: %s\n", preset)
	} else {
		fmt.Println("Analyzing project structure...")
		info, err := init_pkg.DetectProject(path)
		if err != nil {
			return fmt.Errorf("failed to analyze project: %w", err)
		}
		fmt.Println()
		fmt.Print(init_pkg.GenerateSummary(info))
		fmt.Println()
		config = init_pkg.GenerateConfig(info)
	}

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠ Warning: %s already exists\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted. No changes made.")
			return nil
		}
	}

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("✓ Created %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review and customize .structurelint.yml")
	fmt.Println("  2. Run 'structurelint .' to validate your project")
	fmt.Println("  3. See docs/ for detailed rule documentation")

	return nil
}

func printHelp() {
	fmt.Println(`structurelint - Project structure and architecture linter

Usage:
  structurelint [options] [path]   Lint the project at path (default: current directory)
  structurelint graph [options]    Visualize dependency graphs
  structurelint clones [options]   Detect code clones (duplicated code)
  structurelint fix [options]      Auto-fix detected violations
  structurelint tui [options]      Interactive terminal UI for fixing violations
  structurelint scaffold [options] <type> <name>  Generate code from templates
  structurelint --init [path]      Generate configuration by analyzing project
  structurelint --version          Show version information
  structurelint --help             Show this help message

Commands:
  (default)                    Lint project structure and architecture
  graph                        Visualize dependency graphs (see 'structurelint help graph')
  clones                       Detect code clones (see 'structurelint help clones')
  fix                          Auto-fix violations (see 'structurelint help fix')
  tui                          Interactive terminal UI (see 'structurelint help tui')
  scaffold                     Generate code from templates (see 'structurelint help scaffold')

Options:
  -v, --version                Show version information
  -h, --help                   Show help message
      --init                   Initialize configuration for your project
      --preset <name>          Use a known-shape preset for --init
                                 (sveltekit, nextjs-app-router, go-stdlayout, python-monorepo)
      --list-presets           List available --init presets and exit
      --format <format>        Output format: text, json, junit (default: text)

Configuration:
  structurelint looks for .structurelint.yml or .structurelint.yaml files
  in the current directory and parent directories.

Examples:
  structurelint                     Lint current directory
  structurelint --init                            Generate config based on current project
  structurelint --init --preset sveltekit         Use SvelteKit preset directly
  structurelint --list-presets                    List available presets
  structurelint --format json .                   Output violations as JSON
  structurelint --format junit ./src  Output violations as JUnit XML
  structurelint /path/to/project    Lint specific directory

Output Formats:
  text    - Human-readable text output (default)
  json    - JSON format for machine parsing and CI/CD integration
  junit   - JUnit XML format for Jenkins, GitHub Actions, etc.

Initialization:
  The --init command analyzes your project to detect:
  - Programming languages and test patterns
  - Project structure and organization
  - Documentation style

  It then generates an appropriate .structurelint.yml configuration
  with smart defaults based on detected patterns.

  --preset selects a hand-tuned configuration for a known project shape
  (SvelteKit, Next.js App Router, Go standard layout, Python monorepo).
  --init also auto-detects these presets via marker files (svelte.config.js,
  next.config.*, go.mod + cmd/+ internal/, pyproject.toml + apps/+ packages/)
  and asks before overwriting an existing config.

Documentation:
  https://github.com/Jonathangadeaharder/structurelint`)
}
