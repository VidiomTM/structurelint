package fuzz

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	lintparser "github.com/Jonathangadeaharder/structurelint/internal/parser"
)

func FuzzParseGoAST(f *testing.F) {
	f.Add([]byte("package main\n"))
	f.Add([]byte(""))
	f.Add([]byte("package main\nfunc main() {}\n"))
	f.Add([]byte("package foo\nimport \"fmt\"\nfunc bar() { fmt.Println(\"hi\") }\n"))
	f.Add([]byte("\x00\x01\x02"))
	f.Add([]byte("package main\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "", data, 0)
		if err != nil {
			return
		}
		if file == nil {
			t.Fatal("nil AST without error")
		}
	})
}

func FuzzParseGoImports(f *testing.F) {
	seeds := []string{
		"package main\n",
		"package main\nimport \"fmt\"\n",
		"package main\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n",
		"package main\nfunc main() {}\n",
		"import \"github.com/foo/bar\"\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "fuzz.go")
		if err := os.WriteFile(testFile, []byte(src), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		p := lintparser.New(tmpDir)
		imports, err := p.ParseFile(testFile)
		if err != nil {
			return
		}
		for _, imp := range imports {
			if imp.SourceFile == "" {
				t.Error("empty SourceFile in import")
			}
		}
	})
}

func FuzzParseTypeScriptJavaScript(f *testing.F) {
	seeds := []string{
		"import foo from 'bar';",
		"import { x, y } from 'module';",
		"import * as lib from 'lib';",
		"import 'side-effects';",
		"const x = require('pkg');",
		"export default function() {}",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "fuzz.ts")
		if err := os.WriteFile(testFile, []byte(src), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		p := lintparser.New(tmpDir)
		imports, err := p.ParseFile(testFile)
		if err != nil {
			return
		}
		for _, imp := range imports {
			if imp.SourceFile == "" {
				t.Error("empty SourceFile in import")
			}
		}
	})
}

func FuzzParsePython(f *testing.F) {
	seeds := []string{
		"import os",
		"from sys import argv",
		"import os.path",
		"from . import sibling",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "fuzz.py")
		if err := os.WriteFile(testFile, []byte(src), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		p := lintparser.New(tmpDir)
		imports, err := p.ParseFile(testFile)
		if err != nil {
			return
		}
		for _, imp := range imports {
			if imp.SourceFile == "" {
				t.Error("empty SourceFile in import")
			}
		}
	})
}
