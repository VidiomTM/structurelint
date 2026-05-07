package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListPresets_StableNonEmpty(t *testing.T) {
	got := ListPresets()
	if len(got) == 0 {
		t.Fatal("expected at least one preset")
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Errorf("ListPresets must be sorted ascending, got %v", got)
			break
		}
	}
}

func TestPresetConfig_Known(t *testing.T) {
	for _, name := range ListPresets() {
		got, err := PresetConfig(name)
		if err != nil {
			t.Errorf("preset %q returned error: %v", name, err)
			continue
		}
		if !strings.Contains(got, "root: true") {
			t.Errorf("preset %q missing root marker", name)
		}
		if !strings.Contains(got, "rules:") {
			t.Errorf("preset %q missing rules block", name)
		}
	}
}

func TestPresetConfig_Unknown(t *testing.T) {
	if _, err := PresetConfig("nope"); err == nil {
		t.Error("expected error for unknown preset")
	}
}

func TestDetectPreset_SvelteKit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "svelte.config.js"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if got := DetectPreset(dir); got != "sveltekit" {
		t.Errorf("want sveltekit, got %q", got)
	}
}

func TestDetectPreset_NextJS(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "next.config.mjs"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if got := DetectPreset(dir); got != "nextjs-app-router" {
		t.Errorf("want nextjs-app-router, got %q", got)
	}
}

func TestDetectPreset_GoStdLayout(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"go.mod"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"cmd", "internal"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if got := DetectPreset(dir); got != "go-stdlayout" {
		t.Errorf("want go-stdlayout, got %q", got)
	}
}

func TestDetectPreset_PythonMonorepo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"apps", "packages"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if got := DetectPreset(dir); got != "python-monorepo" {
		t.Errorf("want python-monorepo, got %q", got)
	}
}

func TestDetectPreset_None(t *testing.T) {
	dir := t.TempDir()
	if got := DetectPreset(dir); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}
