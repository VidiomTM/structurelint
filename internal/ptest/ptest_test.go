package ptest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/output"
	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"

	"pgregory.net/rapid"
)

func tempDir(t *rapid.T) string {
	dir, err := os.MkdirTemp("", "ptest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir
}

func TestProperty_ParseFile_NeverPanics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		content := rapid.String().Draw(t, "content")
		ext := rapid.SampledFrom([]string{".ts", ".tsx", ".js", ".jsx", ".go", ".py", ".java", ".cpp", ".cs"}).Draw(t, "ext")

		tmpDir := tempDir(t)
		defer os.RemoveAll(tmpDir)
		fpath := filepath.Join(tmpDir, "test"+ext)
		os.WriteFile(fpath, []byte(content), 0644)

		p := parser.New(tmpDir)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ParseFile panicked: %v", r)
				}
			}()
			p.ParseFile(fpath)
		}()
	})
}

func TestProperty_ParseExports_NeverPanics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		content := rapid.String().Draw(t, "content")
		ext := rapid.SampledFrom([]string{".ts", ".tsx", ".js", ".jsx", ".go", ".py", ".java", ".cpp", ".h", ".cs"}).Draw(t, "ext")

		tmpDir := tempDir(t)
		defer os.RemoveAll(tmpDir)
		fpath := filepath.Join(tmpDir, "test"+ext)
		os.WriteFile(fpath, []byte(content), 0644)

		p := parser.New(tmpDir)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ParseExports panicked: %v", r)
				}
			}()
			p.ParseExports(fpath)
		}()
	})
}

func TestProperty_ResolveImportPath_Idempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sourceFile := rapid.StringMatching(`^[a-zA-Z0-9_./-]+$`).Draw(t, "sourceFile")
		importPath := rapid.OneOf(
			rapid.StringMatching(`^\./[a-zA-Z0-9_./-]+$`),
			rapid.StringMatching(`^\.\./[a-zA-Z0-9_./-]+$`),
			rapid.StringMatching(`^[a-zA-Z0-9_.-]+$`),
		).Draw(t, "importPath")

		p := parser.New("/project")
		r1 := p.ResolveImportPath(sourceFile, importPath)
		r2 := p.ResolveImportPath(sourceFile, importPath)
		if r1 != r2 {
			t.Fatalf("not idempotent: %q != %q", r1, r2)
		}
	})
}

func TestProperty_ParseGo_ImportPathPreserved(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		modPath := rapid.StringMatching(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`).Draw(t, "modPath")

		tmpDir := tempDir(t)
		defer os.RemoveAll(tmpDir)
		content := "package main\n\nimport (\n\t\"" + modPath + "\"\n)\n"
		fpath := filepath.Join(tmpDir, "test.go")
		os.WriteFile(fpath, []byte(content), 0644)

		p := parser.New(tmpDir)
		imports, err := p.ParseFile(fpath)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}
		found := false
		for _, imp := range imports {
			if imp.ImportPath == modPath {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("import %q not found", modPath)
		}
	})
}

func TestProperty_ParseTS_RelativeFlag(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		relPath := rapid.OneOf(
			rapid.StringMatching(`^\./[a-zA-Z0-9_./-]+$`),
			rapid.StringMatching(`^\.\./[a-zA-Z0-9_./-]+$`),
		).Draw(t, "relPath")

		tmpDir := tempDir(t)
		defer os.RemoveAll(tmpDir)
		content := "import { x } from '" + relPath + "';\n"
		fpath := filepath.Join(tmpDir, "test.ts")
		os.WriteFile(fpath, []byte(content), 0644)

		p := parser.New(tmpDir)
		imports, err := p.ParseFile(fpath)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}
		for _, imp := range imports {
			if imp.ImportPath == relPath && !imp.IsRelative {
				t.Fatalf("expected IsRelative=true for %q", relPath)
			}
		}
	})
}

func TestProperty_Load_NeverPanics(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		content := rapid.String().Draw(t, "content")

		tmpDir := tempDir(t)
		defer os.RemoveAll(tmpDir)
		fpath := filepath.Join(tmpDir, ".structurelint.yml")
		os.WriteFile(fpath, []byte(content), 0644)

		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Load panicked: %v", r)
				}
			}()
			config.Load(fpath)
		}()
	})
}

func TestProperty_Load_NonexistentPath_NoError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		path := rapid.StringMatching(`^/[a-zA-Z0-9_][a-zA-Z0-9_./-]*$`).Draw(t, "path")

		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("nil config")
		}
	})
}

func TestProperty_Merge_LastWins(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ruleName := rapid.StringMatching(`^[a-z][a-z0-9-]*$`).Draw(t, "ruleName")
		val1 := rapid.Int().Draw(t, "val1")
		val2 := rapid.Int().Draw(t, "val2")

		c1 := &config.Config{Rules: map[string]interface{}{ruleName: val1}}
		c2 := &config.Config{Rules: map[string]interface{}{ruleName: val2}}

		merged := config.Merge(c1, c2)
		if merged.Rules[ruleName] != val2 {
			t.Fatalf("expected %v, got %v", val2, merged.Rules[ruleName])
		}
	})
}

func TestProperty_Merge_ExcludeAppend(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		pat1 := rapid.StringMatching(`^[a-zA-Z*._-]+$`).Draw(t, "pat1")
		pat2 := rapid.StringMatching(`^[a-zA-Z*._-]+$`).Draw(t, "pat2")

		c1 := &config.Config{Exclude: []string{pat1}}
		c2 := &config.Config{Exclude: []string{pat2}}

		merged := config.Merge(c1, c2)
		if len(merged.Exclude) != 2 {
			t.Fatalf("expected 2 excludes, got %d", len(merged.Exclude))
		}
	})
}

func TestProperty_JSONFormatter_ValidJSON(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(t, "n")
		rule := rapid.StringMatching(`^[a-z][a-z0-9-]*$`).Draw(t, "rule")
		path := rapid.StringMatching(`^[a-zA-Z0-9_./-]+$`).Draw(t, "path")
		msg := rapid.StringN(1, 50, 100).Draw(t, "msg")

		violations := make([]rules.Violation, n)
		for i := range violations {
			violations[i] = rules.Violation{Rule: rule, Path: path, Message: msg}
		}

		f := &output.JSONFormatter{Version: "test"}
		result, err := f.Format(violations)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}

		var parsed output.JSONOutput
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if parsed.Violations != n {
			t.Fatalf("expected %d violations, got %d", n, parsed.Violations)
		}
	})
}

func TestProperty_TextFormatter_NeverFails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 50).Draw(t, "n")
		msg := rapid.StringN(0, 50, 200).Draw(t, "msg")

		violations := make([]rules.Violation, n)
		for i := range violations {
			violations[i] = rules.Violation{Rule: "test", Path: "path", Message: msg}
		}

		f := &output.TextFormatter{}
		result, err := f.Format(violations)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		if n == 0 && result != "" {
			t.Fatal("expected empty for no violations")
		}
	})
}

func TestProperty_JUnitFormatter_NeverFails(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(t, "n")
		rule := rapid.StringMatching(`^[a-z][a-z0-9-]*$`).Draw(t, "rule")

		violations := make([]rules.Violation, n)
		for i := range violations {
			violations[i] = rules.Violation{Rule: rule, Path: "path", Message: "msg"}
		}

		f := &output.JUnitFormatter{}
		result, err := f.Format(violations)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("empty result")
		}
	})
}

func TestProperty_GetFormatter_KnownFormats(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fmtName := rapid.SampledFrom([]string{"text", "json", "junit", "junit-xml", "TEXT", "JSON", ""}).Draw(t, "format")

		f, err := output.GetFormatter(fmtName, "1.0")
		if err != nil {
			t.Fatalf("GetFormatter(%q) error: %v", fmtName, err)
		}
		if f == nil {
			t.Fatalf("GetFormatter(%q) nil", fmtName)
		}
	})
}
