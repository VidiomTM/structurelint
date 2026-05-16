package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

func TestNewNormalizer(t *testing.T) {
	n := NewNormalizer()
	if n == nil {
		t.Fatal("NewNormalizer() returned nil")
	}
	if n.fset == nil {
		t.Error("fset should not be nil")
	}
}

func TestNormalizeFile_Basic(t *testing.T) {
	dir := t.TempDir()
	src := `package main

func add(x int, y int) int {
	return x + y
}
`
	fpath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(fpath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	n := NewNormalizer()
	result, err := n.NormalizeFile(fpath)
	if err != nil {
		t.Fatalf("NormalizeFile failed: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.FilePath != fpath {
		t.Errorf("FilePath = %q, want %q", result.FilePath, fpath)
	}
	if len(result.Tokens) == 0 {
		t.Error("expected non-empty tokens")
	}
}

func TestNormalizeFile_NonexistentFile(t *testing.T) {
	n := NewNormalizer()
	_, err := n.NormalizeFile("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestNormalizeFile_InvalidSyntax(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "bad.go")
	if err := os.WriteFile(fpath, []byte("package main\n\nfunc !!!"), 0644); err != nil {
		t.Fatal(err)
	}

	n := NewNormalizer()
	_, err := n.NormalizeFile(fpath)
	if err == nil {
		t.Error("expected error for invalid syntax")
	}
}

func TestExtractTokens_Identifiers(t *testing.T) {
	n := NewNormalizer()
	dir := t.TempDir()
	src := `package main

func foo() {
	bar := baz + qux
}
`
	fpath := filepath.Join(dir, "id.go")
	if err := os.WriteFile(fpath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := n.NormalizeFile(fpath)
	if err != nil {
		t.Fatalf("NormalizeFile failed: %v", err)
	}

	var idTokens []types.Token
	for _, tok := range result.Tokens {
		if tok.Type == types.TokenIdentifier {
			idTokens = append(idTokens, tok)
		}
	}
	if len(idTokens) == 0 {
		t.Error("expected identifier tokens")
	}
	for _, tok := range idTokens {
		if tok.Value != "_ID_" {
			t.Errorf("expected normalized '_ID_', got %q", tok.Value)
		}
	}
}

func TestExtractTokens_Literals(t *testing.T) {
	n := NewNormalizer()
	dir := t.TempDir()
	src := `package main

var x = 42
var y = "hello"
`
	fpath := filepath.Join(dir, "lit.go")
	if err := os.WriteFile(fpath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := n.NormalizeFile(fpath)
	if err != nil {
		t.Fatalf("NormalizeFile failed: %v", err)
	}

	var litTokens []types.Token
	for _, tok := range result.Tokens {
		if tok.Type == types.TokenLiteral {
			litTokens = append(litTokens, tok)
		}
	}
	if len(litTokens) == 0 {
		t.Error("expected literal tokens")
	}
	for _, tok := range litTokens {
		if tok.Value != "_LIT_" {
			t.Errorf("expected normalized '_LIT_', got %q", tok.Value)
		}
	}
}

func TestProcessKeywords_Nil(t *testing.T) {
	n := NewNormalizer()
	_ = n
	result := TokenStreamToString(nil)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestTokenStreamToString(t *testing.T) {
	tokens := []types.Token{
		{Value: "package"},
		{Value: "_ID_"},
		{Value: "_LIT_"},
	}
	result := TokenStreamToString(tokens)
	if result != "package _ID_ _LIT_ " {
		t.Errorf("got %q, want %q", result, "package _ID_ _LIT_ ")
	}
}

func TestTokenStreamToString_Empty(t *testing.T) {
	result := TokenStreamToString(nil)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestNormalizeFile_IfStmt(t *testing.T) {
	dir := t.TempDir()
	src := `package main

func check(x int) bool {
	if x > 0 {
		return true
	}
	return false
}
`
	fpath := filepath.Join(dir, "if.go")
	if err := os.WriteFile(fpath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	n := NewNormalizer()
	result, err := n.NormalizeFile(fpath)
	if err != nil {
		t.Fatalf("NormalizeFile failed: %v", err)
	}

	foundIf := false
	for _, tok := range result.Tokens {
		if tok.Value == "if" && tok.Type == types.TokenKeyword {
			foundIf = true
			break
		}
	}
	if !foundIf {
		t.Error("expected 'if' keyword token")
	}
}

func TestNormalizeFile_RangeStmt(t *testing.T) {
	dir := t.TempDir()
	src := `package main

func sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}
`
	fpath := filepath.Join(dir, "range.go")
	if err := os.WriteFile(fpath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	n := NewNormalizer()
	result, err := n.NormalizeFile(fpath)
	if err != nil {
		t.Fatalf("NormalizeFile failed: %v", err)
	}

	foundFor := false
	for _, tok := range result.Tokens {
		if tok.Value == "for" && tok.Type == types.TokenKeyword {
			foundFor = true
			break
		}
	}
	if !foundFor {
		t.Error("expected 'for' keyword token")
	}
}
