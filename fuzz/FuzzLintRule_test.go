package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func FuzzRuleViolationFormat(f *testing.F) {
	f.Add("test-rule", "/some/path.go", "something went wrong", "PascalCase", "camelCase")
	f.Add("", "", "", "", "")
	f.Add("rule-with-quotes", `path\with\backslash`, "msg with \"quotes\"", "", "")
	f.Fuzz(func(t *testing.T, rule, path, message, expected, actual string) {
		v := rules.Violation{
			Rule:     rule,
			Path:     path,
			Message:  message,
			Expected: expected,
			Actual:   actual,
		}
		_ = v.FormatDetailed()
	})
}

func FuzzRuleShouldIgnoreFile(f *testing.F) {
	seeds := []struct {
		path      string
		content   string
		ruleName  string
	}{
		{"test.go", "// @structurelint:ignore test-rule\npackage main\n", "test-rule"},
		{"test.go", "package main\n", "test-rule"},
		{"test.go", "// @structurelint:ignore other-rule\npackage main\n", "test-rule"},
		{"test.go", "// @structurelint:ignore test-rule other-rule\npackage main\n", "test-rule"},
	}
	for _, s := range seeds {
		f.Add(s.path, s.content, s.ruleName)
	}
	f.Fuzz(func(t *testing.T, path, content, ruleName string) {
		if strings.Contains(path, "\x00") || strings.Contains(content, "\x00") {
			t.Skip()
		}

		tmpDir := t.TempDir()
		cleanPath := filepath.Clean(path)
		if filepath.IsAbs(cleanPath) || cleanPath == ".." ||
			strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
			t.Skip()
		}
		testFile := filepath.Join(tmpDir, cleanPath)
		if rel, err := filepath.Rel(tmpDir, testFile); err != nil || strings.HasPrefix(rel, "..") {
			t.Skip()
		}
		if err := os.MkdirAll(filepath.Dir(testFile), 0o755); err != nil {
			t.Skip()
		}
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			return
		}

		file := walker.FileInfo{
			Path:       path,
			AbsPath:    testFile,
			IsDir:      false,
			Directives: parser.ParseDirectives(testFile),
		}

		shouldIgnore, _ := rules.ShouldIgnoreFile(file, ruleName)
		if shouldIgnore && content == "" {
			t.Error("empty content should not be ignored")
		}
	})
}

func FuzzFilterIgnoredFiles(f *testing.F) {
	f.Add("test-rule", 3)
	f.Add("", 0)
	f.Add("some-very-long-rule-name-with-special-chars-!@#", 100)
	f.Fuzz(func(t *testing.T, ruleName string, fileCount int) {
		if fileCount < 0 {
			fileCount = 0
		}
		if fileCount > 1000 {
			fileCount = 1000
		}

		files := make([]walker.FileInfo, fileCount)
		for i := range files {
			files[i] = walker.FileInfo{
				Path:    "file.go",
				AbsPath: "/tmp/file.go",
				IsDir:   false,
			}
		}

		filtered := rules.FilterIgnoredFiles(files, ruleName)
		if len(filtered) > len(files) {
			t.Errorf("filtered %d > original %d", len(filtered), len(files))
		}
	})
}
