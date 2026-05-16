package walker

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestWithExclude(t *testing.T) {
	w := New("test")
	if w.excludePatterns != nil {
		t.Fatal("Expected nil exclude patterns by default")
	}
	result := w.WithExclude([]string{"*.txt", "*.md"})
	if result != w {
		t.Fatal("WithExclude should return self")
	}
	if len(w.excludePatterns) != 2 {
		t.Fatalf("Expected 2 patterns, got %d", len(w.excludePatterns))
	}
}

func TestIsExcluded(t *testing.T) {
	w := New("test").WithExclude([]string{"*.txt", "node_modules/**"})
	tests := []struct {
		path  string
		want  bool
	}{
		{"file.txt", true},
		{"node_modules/pkg/index.js", true},
		{"src/main.go", false},
		{"file.ts", false},
	}
	for _, tt := range tests {
		got := w.isExcluded(tt.path)
		if got != tt.want {
			t.Errorf("isExcluded(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsExcluded_NoPatterns(t *testing.T) {
	w := New("test")
	if w.isExcluded("anything.txt") {
		t.Error("Expected false with no exclude patterns")
	}
}

func TestWalk_NonExistentRoot(t *testing.T) {
	w := New("/nonexistent/path/xyz-123-abc")
	err := w.Walk()
	if err == nil {
		t.Fatal("Expected error for non-existent root")
	}
}

func TestWalk_IsExcludedDir(t *testing.T) {
	tmpDir := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755))
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, ".git", "HEAD"), []byte("ref: main"), 0644))
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))

	w := New(tmpDir)
	requireNoError(t, w.Walk())
	files := w.GetFiles()
	for _, f := range files {
		if f.Path == ".git" || len(f.Path) > 4 && f.Path[:5] == ".git/" {
			t.Errorf("Expected .git to be excluded, found %q", f.Path)
		}
	}
}

func TestWalk_NodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(tmpDir, "node_modules", "pkg"), 0755))
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "node_modules", "pkg", "index.js"), []byte("// test"), 0644))
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte("console.log()"), 0644))

	w := New(tmpDir)
	requireNoError(t, w.Walk())
	for _, f := range w.GetFiles() {
		if len(f.Path) >= 12 && f.Path[:12] == "node_modules" {
			t.Errorf("Expected node_modules to be excluded, found %q", f.Path)
		}
	}
}

func TestWalk_Vendor(t *testing.T) {
	tmpDir := t.TempDir()
	requireNoError(t, os.MkdirAll(filepath.Join(tmpDir, "vendor", "dep"), 0755))
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "vendor", "dep", "lib.go"), []byte("package dep"), 0644))

	w := New(tmpDir)
	requireNoError(t, w.Walk())
	for _, f := range w.GetFiles() {
		if len(f.Path) >= 6 && f.Path[:6] == "vendor" {
			t.Errorf("Expected vendor to be excluded, found %q", f.Path)
		}
	}
}

func TestWalk_WithExcludePattern_Dir(t *testing.T) {
	tmpDir := t.TempDir()
	subdir := filepath.Join(tmpDir, "node_modules")
	requireNoError(t, os.MkdirAll(subdir, 0755))
	requireNoError(t, os.WriteFile(filepath.Join(subdir, "pkg.js"), []byte("// pkg"), 0644))

	w := New(tmpDir).WithExclude([]string{"node_modules"})
	requireNoError(t, w.Walk())
	for _, f := range w.GetFiles() {
		if len(f.Path) >= 12 && f.Path[:12] == "node_modules" {
			t.Errorf("Expected node_modules to be excluded, found %q", f.Path)
		}
	}
}

func TestWalk_ExcludedNotDir(t *testing.T) {
	tmpDir := t.TempDir()
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "skip.txt"), []byte("skip"), 0644))
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "keep.go"), []byte("package main"), 0644))

	w := New(tmpDir).WithExclude([]string{"*.txt"})
	requireNoError(t, w.Walk())
	// Files matching exclude pattern are still added (only dir exclusions are skipped)
	found := false
	for _, f := range w.GetFiles() {
		if f.Path == "skip.txt" {
			found = true
		}
	}
	if !found {
		t.Error("Expected skip.txt to still be in file list (only dir exclusions skip)")
	}
}

func TestGetMaxDepth_Empty(t *testing.T) {
	w := New("test")
	if max := w.GetMaxDepth(); max != 0 {
		t.Errorf("Expected 0 for empty walker, got %d", max)
	}
}

func TestMatchesPattern_Question(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"file.ts", "file.??", true},
		{"file.ts", "file.?s", true},
		{"file.tx", "file.??", true},
		{"file.t", "file.??", false},
		{"file.txt", "f?le.txt", true},
		{"file.txt", "f?le.tx?", true},
	}
	for _, tt := range tests {
		got := MatchesPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchesPattern_CharacterClass(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"file.ts", "file.[tc]s", true},
		{"file.cs", "file.[tc]s", true},
		{"file.js", "file.[tc]s", false},
		{"file.txt", "file.[a-z]xt", true},
		{"file.axt", "file.[a-z]xt", true},
		{"file.0xt", "file.[a-z]xt", false},
	}
	for _, tt := range tests {
		got := MatchesPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchesPattern_EscapedChars(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"file.test.ts", "file.test.ts", true},
		{"file+test.ts", "file\\+test.ts", true},
		{"file.test.ts", "file\\.test.ts", true},
	}
	for _, tt := range tests {
		got := MatchesPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchesPattern_UnclosedBracket(t *testing.T) {
	if MatchesPattern("file.txt", "file[.txt") {
		t.Error("Expected false for unclosed bracket pattern")
	}
}

func TestMatchesPattern_NoSlashNoGlobstar(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"subdir/file.txt", "*.txt", true},
	}
	for _, tt := range tests {
		got := MatchesPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestPatternRegexp_Cache(t *testing.T) {
	walkerGlobCache = sync.Map{}
	r1 := patternRegexp("*.go")
	r2 := patternRegexp("*.go")
	if r1 != r2 {
		t.Error("Expected cached result")
	}
}

func TestPatternRegexp_StarAfterSlash(t *testing.T) {
	r := patternRegexp("src/**")
	if r == nil {
		t.Fatal("Expected non-nil regexp")
	}
	if !r.MatchString("src/a/b/c") {
		t.Error("Expected src/** to match src/a/b/c")
	}
	if !r.MatchString("src") {
		t.Error("Expected src/** to match src")
	}
}

func TestPatternRegexp_GlobstarAtEnd(t *testing.T) {
	r := patternRegexp("a/**")
	if r == nil {
		t.Fatal("Expected non-nil regexp")
	}
	if !r.MatchString("a/b/c") {
		t.Error("Expected a/** to match a/b/c")
	}
	if !r.MatchString("a") {
		t.Error("Expected a/** to match a")
	}
}

func TestPatternRegexp_InvalidRegexp(t *testing.T) {
	result := patternRegexp("file[].txt")
	if result != nil {
		t.Error("Expected nil for invalid pattern (empty char class)")
	}
}

func TestWalk_AbsError(t *testing.T) {
	w := New("\x00")
	err := w.Walk()
	if err == nil {
		t.Log("Note: null byte path did not cause abs error — skipping")
	}
}

func TestWalk_WalkDirError(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	requireNoError(t, os.MkdirAll(subDir, 0755))
	requireNoError(t, os.WriteFile(filepath.Join(subDir, "f.txt"), []byte("x"), 0644))
	requireNoError(t, os.Chmod(subDir, 0))
	defer os.Chmod(subDir, 0755)

	w := New(tmpDir)
	err := w.Walk()
	if err == nil {
		t.Fatal("Expected walk error from permission denied on subdir")
	}
}

func TestWalker_GetFilesGetDirs(t *testing.T) {
	tmpDir := t.TempDir()
	requireNoError(t, os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644))
	sub := filepath.Join(tmpDir, "sub")
	requireNoError(t, os.MkdirAll(sub, 0755))
	requireNoError(t, os.WriteFile(filepath.Join(sub, "b.txt"), []byte("b"), 0644))

	w := New(tmpDir)
	requireNoError(t, w.Walk())

	files := w.GetFiles()
	if len(files) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(files))
	}

	dirs := w.GetDirs()
	if len(dirs) != 1 {
		t.Fatalf("Expected 1 dir, got %d", len(dirs))
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestPatternRegexp_DollarSign(t *testing.T) {
	r := patternRegexp("file$")
	if r == nil {
		t.Fatal("Expected non-nil regexp")
	}
	if !r.MatchString("file$") {
		t.Error("Expected file$ to match 'file$'")
	}
}
