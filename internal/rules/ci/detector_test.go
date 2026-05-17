package ci

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestDetectSvelteKit(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json":     `{"devDependencies": {"svelte": "^5.0.0"}}`,
		"/project/svelte.config.js": `import adapter from '@sveltejs/adapter-auto'`,
	}}
	detector := NewProjectDetector(reader)
	files := []core.FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "svelte.config.js", AbsPath: "/project/svelte.config.js"},
	}
	types := detector.Detect(files)
	if len(types) != 1 || types[0] != core.SvelteKit {
		t.Fatalf("expected SvelteKit, got %v", types)
	}
}

func TestDetectGo(t *testing.T) {
	detector := NewProjectDetector(nil)
	files := []core.FileInfo{
		{Path: "go.mod", AbsPath: "/project/go.mod"},
	}
	types := detector.Detect(files)
	if len(types) != 1 || types[0] != core.Go {
		t.Fatalf("expected Go, got %v", types)
	}
}

func TestDetectMultiple(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json":   `{"dependencies": {"svelte": "^5.0.0"}}`,
		"/project/pyproject.toml": `[project]`,
	}}
	detector := NewProjectDetector(reader)
	files := []core.FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "pyproject.toml", AbsPath: "/project/pyproject.toml"},
		{Path: "svelte.config.ts", AbsPath: "/project/svelte.config.ts"},
	}
	types := detector.Detect(files)
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d: %v", len(types), types)
	}
	hasSK, hasPy := false, false
	for _, t := range types {
		if t == core.SvelteKit {
			hasSK = true
		}
		if t == core.Python {
			hasPy = true
		}
	}
	if !hasSK || !hasPy {
		t.Fatalf("expected SvelteKit and Python, got %v", types)
	}
}
