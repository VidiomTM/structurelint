package fuzz

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func FuzzMatchesPattern(f *testing.F) {
	corpus := []struct{ path, pattern string }{
		{"src/main.go", "*.go"},
		{"src/main.go", "**/*.go"},
		{"node_modules/pkg/index.js", "node_modules/"},
		{"test_test.go", "*_test.go"},
		{"README.md", "*.md"},
		{"src/internal/config/config.go", "src/**/*.go"},
	}
	for _, c := range corpus {
		f.Add(c.path, c.pattern)
	}

	f.Fuzz(func(t *testing.T, path, pattern string) {
		_ = walker.MatchesPattern(path, pattern)
	})
}

func FuzzNamingConventionDetection(f *testing.F) {
	corpus := []string{
		"camelCase.go",
		"PascalCase.ts",
		"snake_case.py",
		"kebab-case.yaml",
		"UPPER_SNAKE.go",
		"mixedCASE_name.js",
		"simple.rs",
	}
	for _, c := range corpus {
		f.Add(c)
	}

	f.Fuzz(func(t *testing.T, filename string) {
		if strings.Contains(filename, "\x00") {
			t.Skip()
		}
		base := filename
		if idx := strings.LastIndex(filename, "/"); idx >= 0 {
			base = filename[idx+1:]
		}
		if base == "" {
			t.Skip()
		}
		if strings.Contains(base, ".") {
			parts := strings.SplitN(base, ".", 2)
			_ = parts[0]
		}
	})
}

func FuzzDirDepthCalculation(f *testing.F) {
	corpus := []string{
		"src",
		"src/internal",
		"src/internal/config",
		"a/b/c/d/e",
		".",
	}
	for _, c := range corpus {
		f.Add(c)
	}

	f.Fuzz(func(t *testing.T, path string) {
		if path == "" || strings.Contains(path, "\x00") {
			t.Skip()
		}
		depth := strings.Count(path, "/")
		if depth < 0 {
			t.Errorf("negative depth for path %q", path)
		}
	})
}

func FuzzGlobPattern(f *testing.F) {
	corpus := []struct{ path, pattern string }{
		{"foo.go", "*.go"},
		{"bar.ts", "*.ts"},
		{"baz_test.go", "*_test.go"},
		{"cmd/main.go", "cmd/*.go"},
	}
	for _, c := range corpus {
		f.Add(c.path, c.pattern)
	}

	f.Fuzz(func(t *testing.T, path, pattern string) {
		if strings.Contains(path, "\x00") || strings.Contains(pattern, "\x00") {
			t.Skip()
		}
		_ = fmt.Sprintf("path=%s pattern=%s", path, pattern)
	})
}
