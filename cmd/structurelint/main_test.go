package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/autofix"
	"github.com/Jonathangadeaharder/structurelint/internal/linter"
	"github.com/Jonathangadeaharder/structurelint/internal/plugin"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/scaffold"
)

func TestRunLinterWithArgs(t *testing.T) {
	t.Run("version flag", func(t *testing.T) {
		err := runLinterWithArgs([]string{"--version"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("version short flag", func(t *testing.T) {
		err := runLinterWithArgs([]string{"-v"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help flag", func(t *testing.T) {
		err := runLinterWithArgs([]string{"--help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help short flag", func(t *testing.T) {
		err := runLinterWithArgs([]string{"-h"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("list presets flag", func(t *testing.T) {
		err := runLinterWithArgs([]string{"--list-presets"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("invalid flag", func(t *testing.T) {
		err := runLinterWithArgs([]string{"--invalid-flag"})
		if err == nil {
			t.Error("expected error for invalid flag")
		}
	})
}

func TestHandleSubcommands(t *testing.T) {
	t.Run("graph subcommand", func(t *testing.T) {
		err := handleSubcommands([]string{"graph", "--help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("clones subcommand", func(t *testing.T) {
		r := runClones([]string{"--help"})
		if r != nil {
			t.Errorf("expected nil, got %v", r)
		}
	})

	t.Run("fix subcommand", func(t *testing.T) {
		err := handleSubcommands([]string{"fix", "--help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("tui subcommand", func(t *testing.T) {
		err := handleSubcommands([]string{"tui", "--help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("scaffold subcommand", func(t *testing.T) {
		err := handleSubcommands([]string{"scaffold", "--help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help subcommand", func(t *testing.T) {
		err := handleSubcommands([]string{"help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help with topic", func(t *testing.T) {
		err := handleSubcommands([]string{"help", "graph"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestHandleHelpCommand(t *testing.T) {
	t.Run("help graph", func(t *testing.T) {
		err := handleHelpCommand([]string{"help", "graph"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help clones", func(t *testing.T) {
		err := handleHelpCommand([]string{"help", "clones"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help fix", func(t *testing.T) {
		err := handleHelpCommand([]string{"help", "fix"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help tui", func(t *testing.T) {
		err := handleHelpCommand([]string{"help", "tui"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help scaffold", func(t *testing.T) {
		err := handleHelpCommand([]string{"help", "scaffold"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("help default", func(t *testing.T) {
		err := handleHelpCommand([]string{"help"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestRunInit(t *testing.T) {
	t.Run("with unknown preset returns error", func(t *testing.T) {
		err := runInit(t.TempDir(), "nonexistent-preset")
		if err == nil {
			t.Error("expected error for unknown preset")
		}
	})
}

func TestRunLinterWithArgs_init(t *testing.T) {
	t.Run("init flag with unknown preset", func(t *testing.T) {
		err := runLinterWithArgs([]string{"--init", "--preset", "unknown"})
		if err == nil {
			t.Error("expected error for unknown preset")
		}
	})
}

func TestPrintHelp(t *testing.T) {
	printHelp()
}

func TestPrintGraphHelp(t *testing.T) {
	printGraphHelp()
}

func TestPrintClonesHelp(t *testing.T) {
	printClonesHelp()
}

func TestPrintFixHelp(t *testing.T) {
	printFixHelp()
}

func TestPrintTUIHelp(t *testing.T) {
	printTUIHelp()
}

func TestPrintScaffoldHelp(t *testing.T) {
	printScaffoldHelp()
}

func TestResolveAbsolutePath(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		result := resolveAbsolutePath(".")
		if !filepath.IsAbs(result) {
			t.Errorf("expected absolute path, got %s", result)
		}
	})

	t.Run("invalid path returns original", func(t *testing.T) {
		result := resolveAbsolutePath("")
		if result != "" {
			t.Errorf("expected empty, got %s", result)
		}
	})
}

func TestHandlePluginUnavailable(t *testing.T) {
	t.Run("semantic mode returns error", func(t *testing.T) {
		_, err := handlePluginUnavailable("semantic")
		if err == nil {
			t.Error("expected error for semantic mode")
		}
	})

	t.Run("both mode returns nil", func(t *testing.T) {
		_, err := handlePluginUnavailable("both")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestHandleSemanticDetectionError(t *testing.T) {
	t.Run("semantic mode returns error", func(t *testing.T) {
		_, err := handleSemanticDetectionError(fmt.Errorf("test error"), "semantic")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("non-semantic mode returns nil", func(t *testing.T) {
		_, err := handleSemanticDetectionError(fmt.Errorf("test error"), "syntactic")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestParseClonesFlags(t *testing.T) {
	t.Run("default flags", func(t *testing.T) {
		cfg, err := parseClonesFlags([]string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.mode != "syntactic" {
			t.Errorf("expected syntactic, got %s", cfg.mode)
		}
		if !cfg.runSyntactic {
			t.Error("expected syntactic to be true")
		}
		if cfg.runSemantic {
			t.Error("expected semantic to be false")
		}
	})

	t.Run("semantic mode", func(t *testing.T) {
		cfg, err := parseClonesFlags([]string{"--mode", "semantic"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.mode != "semantic" {
			t.Errorf("expected semantic, got %s", cfg.mode)
		}
		if cfg.runSyntactic {
			t.Error("expected syntactic to be false")
		}
		if !cfg.runSemantic {
			t.Error("expected semantic to be true")
		}
	})

	t.Run("both mode", func(t *testing.T) {
		cfg, err := parseClonesFlags([]string{"--mode", "both"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.runSyntactic {
			t.Error("expected syntactic to be true")
		}
		if !cfg.runSemantic {
			t.Error("expected semantic to be true")
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		_, err := parseClonesFlags([]string{"--mode", "invalid"})
		if err == nil {
			t.Error("expected error for invalid mode")
		}
	})

	t.Run("custom flags", func(t *testing.T) {
		cfg, err := parseClonesFlags([]string{
			"--min-tokens", "30",
			"--min-lines", "5",
			"--k-gram", "15",
			"--workers", "8",
			"--similarity", "0.9",
			"--plugin-url", "http://example.com",
			"--format", "json",
			"/path/to/project",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.minTokens != 30 {
			t.Errorf("expected 30, got %d", cfg.minTokens)
		}
		if cfg.minLines != 5 {
			t.Errorf("expected 5, got %d", cfg.minLines)
		}
		if cfg.kGramSize != 15 {
			t.Errorf("expected 15, got %d", cfg.kGramSize)
		}
		if cfg.workers != 8 {
			t.Errorf("expected 8, got %d", cfg.workers)
		}
		if cfg.similarityThreshold != 0.9 {
			t.Errorf("expected 0.9, got %f", cfg.similarityThreshold)
		}
		if cfg.pluginURL != "http://example.com" {
			t.Errorf("expected http://example.com, got %s", cfg.pluginURL)
		}
		if cfg.format != "json" {
			t.Errorf("expected json, got %s", cfg.format)
		}
		if cfg.path != "/path/to/project" {
			t.Errorf("expected /path/to/project, got %s", cfg.path)
		}
	})
}

func TestFilterFixable(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "max-depth", AutoFix: nil},
		{Rule: "naming-convention", AutoFix: &rules.AutoFix{FilePath: "test.go", Content: "fix"}},
		{Rule: "test-adjacency", AutoFix: nil},
		{Rule: "file-existence", AutoFix: &rules.AutoFix{FilePath: "test2.go", Content: "fix2"}},
	}

	fixable := filterFixable(violations, "")
	if len(fixable) != 2 {
		t.Errorf("expected 2 fixable, got %d", len(fixable))
	}
}

func TestFilterFixableWithRuleFilter(t *testing.T) {
	violations := []linter.Violation{
		{Rule: "max-depth", AutoFix: &rules.AutoFix{FilePath: "a.go", Content: "fix"}},
		{Rule: "naming-convention", AutoFix: &rules.AutoFix{FilePath: "b.go", Content: "fix"}},
		{Rule: "file-existence", AutoFix: &rules.AutoFix{FilePath: "c.go", Content: "fix"}},
	}

	t.Run("filter matches", func(t *testing.T) {
		fixable := filterFixable(violations, "naming")
		if len(fixable) != 1 {
			t.Errorf("expected 1, got %d", len(fixable))
		}
	})

	t.Run("filter no match", func(t *testing.T) {
		fixable := filterFixable(violations, "nonexistent")
		if len(fixable) != 0 {
			t.Errorf("expected 0, got %d", len(fixable))
		}
	})
}

func TestRunClones_Help(t *testing.T) {
	err := runClones([]string{"--help"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunFix_Help(t *testing.T) {
	err := runFix([]string{"--help"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunGraph_Help(t *testing.T) {
	err := runGraph([]string{"--help"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunTUI_Help(t *testing.T) {
	err := runTUI([]string{"--help"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunScaffold_Help(t *testing.T) {
	err := runScaffold([]string{"--help"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestExecuteLinter_ErrNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	err := executeLinter(tmpDir, "text")
	if err == nil {
		t.Error("expected error for no config")
	}
}

func TestDecideShouldApply(t *testing.T) {
	t.Run("dry run always applies", func(t *testing.T) {
		ctx := &fixApplyContext{dryRun: true}
		should, quit, err := decideShouldApply(ctx, &autofix.Fix{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if quit {
			t.Error("expected not to quit")
		}
		if !should {
			t.Error("expected shouldApply=true in dry run")
		}
	})

	t.Run("auto mode skips unsafe", func(t *testing.T) {
		ctx := &fixApplyContext{auto: true}
		should, _, _ := decideShouldApply(ctx, &autofix.Fix{Safe: false})
		if should {
			t.Error("expected shouldApply=false for unsafe fix in auto mode")
		}
	})

	t.Run("auto mode applies safe", func(t *testing.T) {
		ctx := &fixApplyContext{auto: true}
		should, _, _ := decideShouldApply(ctx, &autofix.Fix{Safe: true})
		if !should {
			t.Error("expected shouldApply=true for safe fix in auto mode")
		}
	})
}

func TestDecideAutoMode(t *testing.T) {
	if !decideAutoMode(&autofix.Fix{Safe: true}) {
		t.Error("expected true for safe fix")
	}
	if decideAutoMode(&autofix.Fix{Safe: false}) {
		t.Error("expected false for unsafe fix")
	}
}

func TestDisplayFixDetails(t *testing.T) {
	fix := &autofix.Fix{
		Violation:   linter.Violation{Rule: "test-rule", Path: "/test/file.go"},
		Description: "test fix",
		Confidence:  0.95,
		Safe:        true,
	}
	displayFixDetails(fix)
}

func TestReportSemanticClones(t *testing.T) {
	t.Run("no clones", func(t *testing.T) {
		resp := &plugin.SemanticCloneResponse{
			Clones: []plugin.SemanticClone{},
			Stats:  plugin.SemanticCloneStats{},
		}
		count, err := reportSemanticClones(resp)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("with clones", func(t *testing.T) {
		resp := &plugin.SemanticCloneResponse{
			Clones: []plugin.SemanticClone{
				{SourceFile: "a.go", TargetFile: "b.go", Similarity: 0.95, Explanation: "similar code"},
			},
			Stats: plugin.SemanticCloneStats{FilesAnalyzed: 10, FunctionsAnalyzed: 50, DurationMs: 100},
		}
		count, err := reportSemanticClones(resp)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1, got %d", count)
		}
	})
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input string
		want  scaffold.Language
		err   bool
	}{
		{"go", scaffold.LangGo, false},
		{"golang", scaffold.LangGo, false},
		{"typescript", scaffold.LangTypeScript, false},
		{"ts", scaffold.LangTypeScript, false},
		{"python", scaffold.LangPython, false},
		{"py", scaffold.LangPython, false},
		{"java", scaffold.LangJava, false},
		{"rust", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseLanguage(tc.input)
			if tc.err && err == nil {
				t.Error("expected error")
			}
			if !tc.err && got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestParseTemplateType(t *testing.T) {
	tests := []struct {
		input string
		want  scaffold.TemplateType
		err   bool
	}{
		{"service", scaffold.TypeService, false},
		{"repository", scaffold.TypeRepository, false},
		{"repo", scaffold.TypeRepository, false},
		{"controller", scaffold.TypeController, false},
		{"handler", scaffold.TypeHandler, false},
		{"model", scaffold.TypeModel, false},
		{"entity", scaffold.TypeModel, false},
		{"middleware", scaffold.TypeMiddleware, false},
		{"test", scaffold.TypeTest, false},
		{"unknown", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseTemplateType(tc.input)
			if tc.err && err == nil {
				t.Error("expected error")
			}
			if !tc.err && got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	t.Run("empty dir returns empty", func(t *testing.T) {
		got := detectLanguage(t.TempDir())
		if got != "" {
			t.Errorf("expected empty, got %s", got)
		}
	})

	t.Run("go.mod detected", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
		got := detectLanguage(dir)
		if got != "go" {
			t.Errorf("expected go, got %s", got)
		}
	})

	t.Run("tsconfig.json overrides package.json", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
		os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644)
		got := detectLanguage(dir)
		if got != "typescript" {
			t.Errorf("expected typescript, got %s", got)
		}
	})

	t.Run("package.json only defaults to typescript", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
		got := detectLanguage(dir)
		if got != "typescript" {
			t.Errorf("expected typescript, got %s", got)
		}
	})

	t.Run("requirements.txt detected", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(""), 0644)
		got := detectLanguage(dir)
		if got != "python" {
			t.Errorf("expected python, got %s", got)
		}
	})

	t.Run("pom.xml detected", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(""), 0644)
		got := detectLanguage(dir)
		if got != "java" {
			t.Errorf("expected java, got %s", got)
		}
	})

	t.Run("build.gradle detected", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "build.gradle"), []byte(""), 0644)
		got := detectLanguage(dir)
		if got != "java" {
			t.Errorf("expected java, got %s", got)
		}
	})

	t.Run("setup.py detected", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "setup.py"), []byte(""), 0644)
		got := detectLanguage(dir)
		if got != "python" {
			t.Errorf("expected python, got %s", got)
		}
	})
}

func TestExecuteFix(t *testing.T) {
	t.Run("skipped increments skipped", func(t *testing.T) {
		result := &autofix.FixResult{}
		ctx := &fixApplyContext{}
		executeFix(ctx, &autofix.Fix{}, false, result)
		if result.Skipped != 1 {
			t.Errorf("expected 1 skipped, got %d", result.Skipped)
		}
	})

	t.Run("dry run increments applied", func(t *testing.T) {
		result := &autofix.FixResult{}
		ctx := &fixApplyContext{dryRun: true}
		executeFix(ctx, &autofix.Fix{}, true, result)
		if result.Applied != 1 {
			t.Errorf("expected 1 applied, got %d", result.Applied)
		}
	})
}

func TestPrintSemanticStats(t *testing.T) {
	resp := &plugin.SemanticCloneResponse{
		Stats: plugin.SemanticCloneStats{FilesAnalyzed: 5, FunctionsAnalyzed: 20, DurationMs: 50},
	}
	printSemanticStats(resp)
}

func TestListTemplates(t *testing.T) {
	gen := scaffold.NewGenerator(t.TempDir())
	err := listTemplates(gen)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunScaffold_MissingArgs(t *testing.T) {
	err := runScaffold([]string{})
	if err == nil {
		t.Error("expected error for missing args")
	}
}

func TestRunScaffold_UnknownType(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	err := runScaffold([]string{"unknown-type", "TestName"})
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestRunScaffold_UnsupportedLang(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	err := runScaffold([]string{"--lang", "rust", "service", "TestService"})
	if err == nil {
		t.Error("expected error for unsupported lang")
	}
}

func TestHandleSubcommands_NotFound(t *testing.T) {
	dir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origWd)

	err := handleSubcommands([]string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}
