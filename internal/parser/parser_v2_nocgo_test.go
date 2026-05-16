//go:build !cgo

package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParserV2_NoCGO_Fallback(t *testing.T) {
	p := NewV2(".")

	if p == nil {
		t.Fatal("NewV2 returned nil")
	}

	if p.v1 == nil {
		t.Error("NewV2 did not initialize fallback parser")
	}

	res := p.ResolveImportPath("src/main.go", "./utils")
	if res != "src\\utils" && res != "src/utils" {
		t.Errorf("ResolveImportPath failed: %s", res)
	}
}

func TestParserV2_NoCGO_ParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `import { foo } from './foo';`
	if err := os.WriteFile(tsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewV2(tmpDir)
	imports, err := p.ParseFile(tsFile)

	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if len(imports) != 1 {
		t.Fatalf("got %d imports, want 1", len(imports))
	}

	if imports[0].ImportPath != "./foo" {
		t.Errorf("import path = %q, want %q", imports[0].ImportPath, "./foo")
	}
}

func TestParserV2_NoCGO_ParseExports(t *testing.T) {
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `export const foo = 'bar';`
	if err := os.WriteFile(tsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewV2(tmpDir)
	exports, err := p.ParseExports(tsFile)

	if err != nil {
		t.Fatalf("ParseExports() error = %v", err)
	}

	if len(exports) == 0 {
		t.Fatal("expected at least 1 export")
	}
}
