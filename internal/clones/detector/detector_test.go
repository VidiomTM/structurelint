package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MinTokens != 20 {
		t.Errorf("MinTokens = %d, want 20", cfg.MinTokens)
	}
	if cfg.MinLines != 3 {
		t.Errorf("MinLines = %d, want 3", cfg.MinLines)
	}
	if cfg.KGramSize != 20 {
		t.Errorf("KGramSize = %d, want 20", cfg.KGramSize)
	}
	if cfg.NumWorkers != 4 {
		t.Errorf("NumWorkers = %d, want 4", cfg.NumWorkers)
	}
	if !cfg.CrossFileOnly {
		t.Error("CrossFileOnly should be true")
	}
}

func TestNewDetector(t *testing.T) {
	cfg := DefaultConfig()
	d := NewDetector(cfg)
	if d == nil {
		t.Fatal("NewDetector returned nil")
	}
	if d.normalizer == nil {
		t.Error("normalizer should not be nil")
	}
	if d.hasher == nil {
		t.Error("hasher should not be nil")
	}
	if d.index == nil {
		t.Error("index should not be nil")
	}
	if d.expander == nil {
		t.Error("expander should not be nil")
	}
}

func TestNewDetector_ZeroWorkers(t *testing.T) {
	d := NewDetector(Config{NumWorkers: 0})
	if d.config.NumWorkers != 4 {
		t.Errorf("NumWorkers = %d, want 4", d.config.NumWorkers)
	}
}

func TestNewDetector_NegativeMinTokens(t *testing.T) {
	d := NewDetector(Config{MinTokens: -1, NumWorkers: 2})
	if d.config.MinTokens != 20 {
		t.Errorf("MinTokens = %d, want 20", d.config.MinTokens)
	}
}

func TestNewDetector_NegativeMinLines(t *testing.T) {
	d := NewDetector(Config{MinLines: -1, NumWorkers: 2})
	if d.config.MinLines != 3 {
		t.Errorf("MinLines = %d, want 3", d.config.MinLines)
	}
}

func TestFindGoFiles_Basic(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "util.go"), []byte("package util"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "helper.go"), []byte("package helper"), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDetector(DefaultConfig())
	files, err := d.findGoFiles(dir)
	if err != nil {
		t.Fatalf("findGoFiles failed: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 Go files, got %d", len(files))
	}
}

func TestFindGoFiles_WithExclude(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main"), 0644)

	d := NewDetector(DefaultConfig())
	files, err := d.findGoFiles(dir)
	if err != nil {
		t.Fatalf("findGoFiles failed: %v", err)
	}
	// Default config excludes *_test.go
	if len(files) != 1 {
		t.Errorf("expected 1 Go file (test excluded), got %d", len(files))
	}
}

func TestFindGoFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	d := NewDetector(Config{NumWorkers: 2})
	files, err := d.findGoFiles(dir)
	if err != nil {
		t.Fatalf("findGoFiles failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestFindGoFiles_NonexistentDir(t *testing.T) {
	d := NewDetector(Config{NumWorkers: 2})
	_, err := d.findGoFiles("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestMatchesPattern_Simple(t *testing.T) {
	if !matchesPattern("foo.go", "*.go") {
		t.Error("foo.go should match *.go")
	}
	if matchesPattern("foo.txt", "*.go") {
		t.Error("foo.txt should not match *.go")
	}
}

func TestMatchesPattern_DoubleStar(t *testing.T) {
	if !matchesPattern("project/vendor/pkg/file.go", "**/vendor/**") {
		t.Error("should match **/vendor/** for project/vendor/pkg/file.go")
	}
	if !matchesPattern("vendor/foo.go", "**/vendor/**") {
		t.Error("should match vendor in path")
	}
	if matchesPattern("src/foo.go", "**/vendor/**") {
		t.Error("should NOT match path without vendor")
	}
}

func TestMatchesPattern_NoWildcards(t *testing.T) {
	if !matchesPattern("exact.go", "exact.go") {
		t.Error("exact match should work")
	}
}

func TestCheckPatternParts(t *testing.T) {
	if !checkPatternParts("a/b/c/vendor/d/e.go", []string{"vendor"}) {
		t.Error("should find vendor in path")
	}
}

func TestCheckPatternPart_First(t *testing.T) {
	if !checkPatternPart("vendor/x.go", "vendor", 0, 2) {
		t.Error("first part should match")
	}
}

func TestCheckPatternPart_Middle(t *testing.T) {
	if !checkPatternPart("a/vendor/x.go", "vendor", 1, 3) {
		t.Error("middle part should match")
	}
}

func TestCheckPatternPart_Last(t *testing.T) {
	if !checkPatternPart("a/vendor", "vendor", 1, 2) {
		t.Error("last part should match")
	}
}

func TestFilterClones(t *testing.T) {
	d := NewDetector(Config{MinTokens: 10, MinLines: 3, NumWorkers: 2})
	clones := []*types.Clone{
		{TokenCount: 20, LineCount: 5},
		{TokenCount: 5, LineCount: 1},
		{TokenCount: 15, LineCount: 2},
		{TokenCount: 10, LineCount: 3},
	}
	filtered := d.filterClones(clones)
	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered clones, got %d", len(filtered))
	}
}

func TestFilterClones_Empty(t *testing.T) {
	d := NewDetector(Config{NumWorkers: 2})
	filtered := d.filterClones(nil)
	if len(filtered) != 0 {
		t.Errorf("expected 0 filtered clones, got %d", len(filtered))
	}
}

func TestGetIndexStats(t *testing.T) {
	d := NewDetector(DefaultConfig())
	stats := d.GetIndexStats()
	if stats.TotalHashes != 0 {
		t.Errorf("expected 0 total hashes, got %d", stats.TotalHashes)
	}
}

func TestDetectClones_NoGoFiles(t *testing.T) {
	dir := t.TempDir()
	d := NewDetector(DefaultConfig())
	_, err := d.DetectClones(dir)
	if err == nil {
		t.Error("expected error when no Go files present")
	}
}

func TestBuildIndex_EmptyTokenCache(t *testing.T) {
	d := NewDetector(DefaultConfig())
	err := d.buildIndex(make(map[string][]types.Token))
	if err != nil {
		t.Fatalf("buildIndex failed: %v", err)
	}
	stats := d.GetIndexStats()
	if stats.TotalHashes != 0 {
		t.Errorf("expected 0 hashes, got %d", stats.TotalHashes)
	}
}

func TestNormalizeFiles_Empty(t *testing.T) {
	d := NewDetector(Config{NumWorkers: 2})
	cache, err := d.normalizeFiles(nil)
	if err != nil {
		t.Fatalf("normalizeFiles failed: %v", err)
	}
	if len(cache) != 0 {
		t.Errorf("expected empty cache, got %d", len(cache))
	}
}

func TestNormalizeFiles_Nonexistent(t *testing.T) {
	d := NewDetector(Config{NumWorkers: 2})
	_, err := d.normalizeFiles([]string{"/nonexistent.go"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMatchesPattern_EmptyPattern(t *testing.T) {
	if matchesPattern("foo.go", "") {
		t.Error("empty pattern should not match")
	}
}

func TestFindGoFiles_WalkError(t *testing.T) {
	d := NewDetector(DefaultConfig())
	_, err := d.findGoFiles("/dev/null/nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCheckPatternPart_DefaultCase(t *testing.T) {
	// tests the default branch (index != 0 and index != totalParts-1)
	if !checkPatternPart("a/b/c/d/e.go", "c", 2, 5) {
		t.Error("middle part should match via contains")
	}
}

func TestCheckPatternParts_EmptyParts(t *testing.T) {
	if !checkPatternParts("a.go", []string{""}) {
		t.Error("empty parts should not break matching")
	}
}

func TestDetectClones_FullPipeline(t *testing.T) {
	dir := t.TempDir()
	code := `package main

func add(a, b int) int {
	result := a + b
	return result
}

func multiply(x, y int) int {
	result := x * y
	return result
}
`
	code2 := `package utils

func subtract(a, b int) int {
	result := a - b
	return result
}

func divide(x, y int) int {
	result := x / y
	return result
}
`
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)
	_ = os.WriteFile(filepath.Join(dir, "utils.go"), []byte(code2), 0644)

	d := NewDetector(Config{MinTokens: 5, MinLines: 2, KGramSize: 12, NumWorkers: 2, CrossFileOnly: false})
	clones, err := d.DetectClones(dir)
	if err != nil {
		t.Fatalf("DetectClones failed: %v", err)
	}
	// At minimum, should not error and return clones list
	_ = clones
}

func TestDetectClones_CrossFileOnly(t *testing.T) {
	dir := t.TempDir()
	code := `package main
func add(a, b int) int { return a + b }
func sub(a, b int) int { return a - b }
`
	code2 := `package utils
func add(a, b int) int { return a + b }
func mul(a, b int) int { return a * b }
`
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte(code), 0644)
	_ = os.WriteFile(filepath.Join(dir, "utils.go"), []byte(code2), 0644)

	d := NewDetector(Config{MinTokens: 3, MinLines: 1, KGramSize: 8, NumWorkers: 2, CrossFileOnly: true})
	clones, err := d.DetectClones(dir)
	if err != nil {
		t.Fatalf("DetectClones failed: %v", err)
	}
	t.Logf("Found %d clones (cross-file only)", len(clones))
}

func TestNormalizeFiles_WithRealFiles(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "f1.go"), []byte(`package main
func hello() string { return "hello" }
`), 0644)
	_ = os.WriteFile(filepath.Join(dir, "f2.go"), []byte(`package main
func world() string { return "world" }
`), 0644)

	d := NewDetector(Config{NumWorkers: 2})
	cache, err := d.normalizeFiles([]string{filepath.Join(dir, "f1.go"), filepath.Join(dir, "f2.go")})
	if err != nil {
		t.Fatalf("normalizeFiles failed: %v", err)
	}
	if len(cache) != 2 {
		t.Errorf("expected 2 files in cache, got %d", len(cache))
	}
}

func TestBuildIndex_WithShingles(t *testing.T) {
	// Generate enough tokens to produce shingles with k-gram 3
	tokens := make([]types.Token, 10)
	for i := range tokens {
		tokens[i] = types.Token{Type: types.TokenKeyword, Value: "token"}
	}
	d := NewDetector(Config{KGramSize: 3, NumWorkers: 2})
	cache := map[string][]types.Token{
		"file1.go": tokens,
		"file2.go": tokens,
	}
	err := d.buildIndex(cache)
	if err != nil {
		t.Fatalf("buildIndex failed: %v", err)
	}
	stats := d.GetIndexStats()
	if stats.TotalShingles == 0 {
		t.Error("expected some shingles after buildIndex")
	}
}
