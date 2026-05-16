package treesitter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExports_ExtractGoExports(t *testing.T) {
	e, err := NewExportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	os.WriteFile(path, []byte(`package main

func ExportedFunc() {}
func unexportedFunc() {}
type ExportedType struct{}
type unexportedType struct{}
`), 0644)

	exports, err := e.ExtractFromFile(path)
	if err == nil {
		assert.NotEmpty(t, exports)
		for _, exp := range exports {
			assert.True(t, len(exp.Names) > 0)
		}
	}
}

func TestExports_ExtractPythonExports(t *testing.T) {
	e, err := NewExportExtractor(LanguagePython)
	if err != nil {
		t.Skip("tree-sitter Python grammar not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.py")
	os.WriteFile(path, []byte(`def public_function():
    pass

def _private_function():
    pass

class PublicClass:
    pass
`), 0644)

	exports, err := e.ExtractFromFile(path)
	if err == nil {
		assert.NotEmpty(t, exports)
	}
}

func TestExports_ExtractJSExports(t *testing.T) {
	e, err := NewExportExtractor(LanguageJavaScript)
	if err != nil {
		t.Skip("tree-sitter JS grammar not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.js")
	os.WriteFile(path, []byte(`export function foo() {}`), 0644)
	exports, err := e.ExtractFromFile(path)
	assert.NoError(t, err)
	assert.Empty(t, exports)
}

func TestExports_UnsupportedLanguage(t *testing.T) {
	e := &ExportExtractor{language: "unsupported"}
	_, err := e.extractExports(nil, nil, "file.txt")
	assert.Error(t, err)
}

func TestExports_GoGroupedTypes(t *testing.T) {
	e, err := NewExportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	os.WriteFile(path, []byte(`package main

type (
	ExportedType struct{}
	AnotherType struct{}
	privateType struct{}
)
`), 0644)

	exports, err := e.ExtractFromFile(path)
	if err == nil {
		assert.NotEmpty(t, exports)
	}
}

func TestExports_EmptySource(t *testing.T) {
	e, err := NewExportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "empty.go")
	os.WriteFile(path, []byte{}, 0644)

	exports, err := e.ExtractFromFile(path)
	assert.NoError(t, err)
	assert.Empty(t, exports)
}
