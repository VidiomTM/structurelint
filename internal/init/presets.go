package init

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Preset names a known project shape that ships with a battle-tested
// `.structurelint.yml`. Listed via ListPresets, materialized via
// PresetConfig.
type Preset string

const (
	PresetSvelteKit       Preset = "sveltekit"
	PresetPythonMonorepo  Preset = "python-monorepo"
	PresetGoStdLayout     Preset = "go-stdlayout"
	PresetNextJSAppRouter Preset = "nextjs-app-router"
)

// ListPresets returns the known preset identifiers in stable order.
func ListPresets() []string {
	names := []string{
		string(PresetSvelteKit),
		string(PresetPythonMonorepo),
		string(PresetGoStdLayout),
		string(PresetNextJSAppRouter),
	}
	sort.Strings(names)
	return names
}

// PresetConfig returns the raw YAML for a preset name. Returns an error if
// the preset is unknown.
func PresetConfig(name string) (string, error) {
	switch Preset(name) {
	case PresetSvelteKit:
		return svelteKitConfig, nil
	case PresetPythonMonorepo:
		return pythonMonorepoConfig, nil
	case PresetGoStdLayout:
		return goStdLayoutConfig, nil
	case PresetNextJSAppRouter:
		return nextJSAppRouterConfig, nil
	}
	return "", fmt.Errorf("unknown preset %q (known: %v)", name, ListPresets())
}

// DetectPreset inspects the project root for marker files and returns the
// most-likely preset name, or "" if nothing matches.
func DetectPreset(rootPath string) string {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(rootPath, rel))
		return err == nil
	}
	switch {
	case exists("svelte.config.js") || exists("svelte.config.ts"):
		return string(PresetSvelteKit)
	case exists("next.config.js") || exists("next.config.mjs") || exists("next.config.ts"):
		return string(PresetNextJSAppRouter)
	case exists("pyproject.toml") && exists("apps") && exists("packages"):
		return string(PresetPythonMonorepo)
	case exists("go.mod") && exists("cmd") && exists("internal"):
		return string(PresetGoStdLayout)
	}
	return ""
}

const svelteKitConfig = `# structurelint configuration — SvelteKit preset
# https://github.com/Jonathangadeaharder/structurelint
root: true

exclude:
  - node_modules/**
  - .svelte-kit/**
  - build/**
  - dist/**
  - .vercel/**
  - coverage/**
  - testdata/**

rules:
  max-depth:
    max: 6
    overrides:
      "src/routes/**": 10

  max-files-in-dir:
    max: 40

  max-subdirs:
    max: 25

  case-conflicts: true

  disallow-empty-dirs: true
  disallow-symlinks: true

  disallowed-patterns:
    - "*.tmp"
    - "*.bak"
    - ".DS_Store"
    - "**/Thumbs.db"

  naming-convention:
    "*.svelte": "PascalCase"
    "src/lib/**/*.ts": "camelCase"
    "src/lib/**/*.js": "camelCase"

  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "src/lib/**/*.ts"
      - "src/lib/**/*.js"
    exemptions:
      - "**/*.d.ts"
      - "**/index.ts"

  disallow-deep-relative-imports:
    max-parents: 3
`

const nextJSAppRouterConfig = `# structurelint configuration — Next.js App Router preset
root: true

exclude:
  - node_modules/**
  - .next/**
  - out/**
  - build/**
  - coverage/**
  - testdata/**

rules:
  max-depth:
    max: 6
    overrides:
      "app/**": 10
      "pages/**": 8

  max-files-in-dir:
    max: 40

  max-subdirs:
    max: 25

  case-conflicts: true
  disallow-empty-dirs: true
  disallow-symlinks: true

  disallowed-patterns:
    - "*.tmp"
    - "*.bak"
    - ".DS_Store"

  naming-convention:
    "components/**/*.tsx": "PascalCase"
    "lib/**/*.ts": "camelCase"
    "hooks/**/*.ts": "camelCase"

  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "lib/**/*.ts"
      - "hooks/**/*.ts"
    exemptions:
      - "**/*.d.ts"

  disallow-deep-relative-imports:
    max-parents: 3
`

const pythonMonorepoConfig = `# structurelint configuration — Python monorepo preset
# Layout assumed: apps/<service>/, packages/<lib>/
root: true

exclude:
  - .venv/**
  - "**/__pycache__/**"
  - "**/.pytest_cache/**"
  - "**/.mypy_cache/**"
  - "**/.ruff_cache/**"
  - "**/dist/**"
  - "**/build/**"
  - "**/*.egg-info/**"

rules:
  max-depth:
    max: 7

  max-files-in-dir:
    max: 50

  max-subdirs:
    max: 30

  case-conflicts: true
  disallow-empty-dirs: true
  disallow-symlinks: true

  file-existence:
    "README.md": "exists:1"

  disallowed-patterns:
    - "*.pyc"
    - "*.pyo"
    - ".DS_Store"

  naming-convention:
    "**/*.py": "snake_case"

  test-adjacency:
    pattern: "separate"
    test-dir: "tests"
    file-patterns:
      - "apps/*/src/**/*.py"
      - "packages/*/src/**/*.py"
    exemptions:
      - "**/__init__.py"
      - "**/conftest.py"
      - "**/migrations/**"

  disallow-import-cycles: true

path-based-layers:
  layers:
    - name: apps
      patterns: ["apps/**"]
      canDependOn: [packages]
    - name: packages
      patterns: ["packages/**"]
      canDependOn: [packages]
`

const goStdLayoutConfig = `# structurelint configuration — Go standard layout preset
# Layout assumed: cmd/<binary>/, internal/<pkg>/, pkg/<pkg>/
root: true

exclude:
  - vendor/**
  - bin/**
  - dist/**
  - testdata/**
  - "**/testdata/**"

rules:
  max-depth:
    max: 6

  max-files-in-dir:
    max: 40

  max-subdirs:
    max: 20

  case-conflicts: true
  disallow-empty-dirs: true
  disallow-symlinks: true

  file-existence:
    "README.md": "exists:1"

  disallowed-patterns:
    - "*.tmp"
    - "*.bak"
    - ".DS_Store"

  naming-convention:
    "*.go": "snake_case"

  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.go"
    exemptions:
      - "cmd/**/*.go"
      - "**/main.go"
      - "**/doc.go"
      - "**/*_gen.go"

  disallow-import-cycles: true

layers:
  - name: cmd
    path: cmd/**
    dependsOn: [internal, pkg]
  - name: internal
    path: internal/**
    dependsOn: []
  - name: pkg
    path: pkg/**
    dependsOn: []

entrypoints:
  - cmd/**/*.go
  - "**/*_test.go"
`
