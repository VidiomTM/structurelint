package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

type dummyRule struct {
	name string
}

func (r *dummyRule) Name() string { return r.name }

func (r *dummyRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	return nil
}

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
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, path)
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			return
		}

		file := walker.FileInfo{
			Path:    path,
			AbsPath: testFile,
			IsDir:   false,
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
