package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
rules:
  max-depth:
    max: 5
  naming-convention:
    "*.ts": "camelCase"
`

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	// Act
	config, err := Load(configFile)

	// Assert
	require.NoError(t, err)
	assert.True(t, config.Root)
	assert.Equal(t, 2, len(config.Rules))
	assert.Contains(t, config.Rules, "max-depth")
}

func TestLoadNonExistentFile(t *testing.T) {
	config, err := Load("/nonexistent/file.yml")

	require.NoError(t, err)
	assert.NotNil(t, config.Rules)
}

func TestMerge(t *testing.T) {
	config1 := &Config{
		Root: false,
		Rules: map[string]interface{}{
			"max-depth": map[string]interface{}{"max": 5},
		},
	}

	config2 := &Config{
		Root: true,
		Rules: map[string]interface{}{
			"max-depth": map[string]interface{}{"max": 7},  // Override
			"max-files": map[string]interface{}{"max": 10}, // New rule
		},
	}

	merged := Merge(config1, config2)

	assert.True(t, merged.Root)

	maxDepth := merged.Rules["max-depth"].(map[string]interface{})["max"]
	assert.Equal(t, 7, maxDepth)

	assert.Contains(t, merged.Rules, "max-files")
}

func TestMergeWithLayers(t *testing.T) {
	config1 := &Config{
		Layers: []Layer{
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	config2 := &Config{
		Layers: []Layer{
			{Name: "app", Path: "src/app/**", DependsOn: []string{"domain"}},
		},
	}

	merged := Merge(config1, config2)

	assert.Equal(t, 2, len(merged.Layers))
	assert.Equal(t, "domain", merged.Layers[0].Name)
	assert.Equal(t, "app", merged.Layers[1].Name)
}

func TestMergeWithEntrypoints(t *testing.T) {
	config1 := &Config{
		Entrypoints: []string{"src/index.ts"},
	}

	config2 := &Config{
		Entrypoints: []string{"src/main.ts"},
	}

	merged := Merge(config1, config2)

	assert.Equal(t, 2, len(merged.Entrypoints))
}

func TestMergeWithOverrides(t *testing.T) {
	config1 := &Config{
		Overrides: []Override{
			{
				Files: []string{"src/**"},
				Rules: map[string]interface{}{"max-depth": 5},
			},
		},
	}

	config2 := &Config{
		Overrides: []Override{
			{
				Files: []string{"tests/**"},
				Rules: map[string]interface{}{"max-depth": 0},
			},
		},
	}

	merged := Merge(config1, config2)

	assert.Equal(t, 2, len(merged.Overrides))
}

func TestFindConfigs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "src", "components")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Root config
	rootConfig := filepath.Join(tmpDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(rootConfig, []byte("root: true\nrules:\n  max-depth:\n    max: 5"), 0644))

	// Sub config
	subConfig := filepath.Join(subDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(subConfig, []byte("rules:\n  max-depth:\n    max: 10"), 0644))

	configs, err := FindConfigs(subDir)

	require.NoError(t, err)

	// Should find both configs (root and sub)
	assert.Equal(t, 2, len(configs))

	// First config should be root
	assert.True(t, configs[0].Root)
}

func TestMergeWithExclude(t *testing.T) {
	config1 := &Config{
		Exclude: []string{"node_modules/**", "dist/**"},
	}

	config2 := &Config{
		Exclude: []string{"vendor/**"},
	}

	merged := Merge(config1, config2)

	assert.Equal(t, 3, len(merged.Exclude))

	// Verify all patterns are present
	assert.Contains(t, merged.Exclude, "node_modules/**")
	assert.Contains(t, merged.Exclude, "dist/**")
	assert.Contains(t, merged.Exclude, "vendor/**")
}

func TestLoadWithExclude(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
exclude:
  - testdata/**
  - build/**
rules:
  max-depth:
    max: 5
`

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	config, err := Load(configFile)

	require.NoError(t, err)
	assert.Equal(t, 2, len(config.Exclude))
	assert.Equal(t, "testdata/**", config.Exclude[0])
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	// Use invalid YAML with mismatched indentation and tabs/spaces
	content := "\troot: true\n\t  rules:\n\t\tinvalid:\t  [\n"

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	_, err := Load(configFile)

	assert.Error(t, err)
}

func TestFindConfigsWithYamlExtension(t *testing.T) {
	tmpDir := t.TempDir()

	// Use .yaml extension instead of .yml
	configFile := filepath.Join(tmpDir, ".structurelint.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte("root: true\nrules:\n  max-depth:\n    max: 5"), 0644))

	configs, err := FindConfigs(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, 1, len(configs))
	assert.True(t, configs[0].Root)
}

func TestMergeWithNilConfig(t *testing.T) {
	config1 := &Config{
		Rules: map[string]interface{}{
			"max-depth": 5,
		},
	}

	merged := Merge(config1, nil)

	assert.Equal(t, 1, len(merged.Rules))
}

func TestFindConfigsNoConfigFound(t *testing.T) {
	tmpDir := t.TempDir()

	configs, err := FindConfigs(tmpDir)

	require.NoError(t, err)

	// Should return default config
	assert.Equal(t, 1, len(configs))
	assert.NotNil(t, configs[0].Rules)
}

func TestLoadWithGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
rules:
  max-depth:
    max: 5
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	gitignoreContent := "node_modules\n*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	config, err := LoadWithGitignore(configFile, tmpDir)

	require.NoError(t, err)
	assert.True(t, config.Root)
	assert.Contains(t, config.Exclude, "**/node_modules")
}

func TestLoadWithGitignoreDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
autoLoadGitignore: false
rules:
  max-depth:
    max: 5
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	gitignoreContent := "node_modules\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	config, err := LoadWithGitignore(configFile, tmpDir)

	require.NoError(t, err)
	assert.Empty(t, config.Exclude)
}

func TestLoadWithGitignoreLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	content := `root: true
rules:
  max-depth:
    max: 5
`
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	// Don't create .gitignore but don't fail — ApplyGitignore handles missing gracefully
	config, err := LoadWithGitignore(configFile, tmpDir)

	require.NoError(t, err)
	assert.True(t, config.Root)
}

func TestApplyGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	gitignoreContent := "node_modules\n*.log\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	cfg := &Config{
		Rules:    make(map[string]interface{}),
		Exclude:  []string{"dist/**"},
	}

	result, err := ApplyGitignore(cfg, tmpDir)

	require.NoError(t, err)
	assert.Contains(t, result.Exclude, "dist/**")
	assert.Contains(t, result.Exclude, "**/node_modules")
}

func TestApplyGitignoreAutoLoadDisabled(t *testing.T) {
	autoLoadFalse := false
	cfg := &Config{
		Rules:             make(map[string]interface{}),
		AutoLoadGitignore: &autoLoadFalse,
	}

	result, err := ApplyGitignore(cfg, "/tmp")

	require.NoError(t, err)
	assert.Empty(t, result.Exclude)
}

func TestLoadWithExtendsString(t *testing.T) {
	tmpDir := t.TempDir()

	baseConfig := filepath.Join(tmpDir, "base.yml")
	require.NoError(t, os.WriteFile(baseConfig, []byte("rules:\n  max-depth:\n    max: 3\n"), 0644))

	mainConfig := filepath.Join(tmpDir, ".structurelint.yml")
	content := "extends: base.yml\nrules:\n  max-files:\n    max: 10\n"
	require.NoError(t, os.WriteFile(mainConfig, []byte(content), 0644))

	config, err := Load(mainConfig)

	require.NoError(t, err)
	assert.Equal(t, 3, config.Rules["max-depth"].(map[string]interface{})["max"])
	assert.Equal(t, 10, config.Rules["max-files"].(map[string]interface{})["max"])
	assert.Nil(t, config.Extends)
}

func TestLoadWithExtendsArray(t *testing.T) {
	tmpDir := t.TempDir()

	base1 := filepath.Join(tmpDir, "base1.yml")
	require.NoError(t, os.WriteFile(base1, []byte("rules:\n  max-depth:\n    max: 5\n"), 0644))

	base2 := filepath.Join(tmpDir, "base2.yml")
	require.NoError(t, os.WriteFile(base2, []byte("rules:\n  naming-convention:\n    \"*.ts\": \"camelCase\"\n"), 0644))

	mainConfig := filepath.Join(tmpDir, ".structurelint.yml")
	content := "extends:\n  - base1.yml\n  - base2.yml\nrules:\n  max-files:\n    max: 10\n"
	require.NoError(t, os.WriteFile(mainConfig, []byte(content), 0644))

	config, err := Load(mainConfig)

	require.NoError(t, err)
	assert.NotNil(t, config.Rules["max-depth"])
	assert.NotNil(t, config.Rules["naming-convention"])
	assert.Equal(t, 10, config.Rules["max-files"].(map[string]interface{})["max"])
}

func TestLoadCircularExtends(t *testing.T) {
	tmpDir := t.TempDir()

	cfgA := filepath.Join(tmpDir, "a.yml")
	cfgB := filepath.Join(tmpDir, "b.yml")

	require.NoError(t, os.WriteFile(cfgA, []byte("extends: b.yml\nrules:\n  max-depth:\n    max: 5\n"), 0644))
	require.NoError(t, os.WriteFile(cfgB, []byte("extends: a.yml\nrules:\n  max-files:\n    max: 10\n"), 0644))

	_, err := Load(cfgA)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestLoadWithInvalidExtendsType(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	content := "extends: 42\nrules:\n  max-depth:\n    max: 5\n"
	require.NoError(t, os.WriteFile(configFile, []byte(content), 0644))

	_, err := Load(configFile)
	assert.Error(t, err)
}

func TestResolveExtendPathAbsolute(t *testing.T) {
	tmpDir := t.TempDir()
	extPath := filepath.Join(tmpDir, "ext.yml")
	require.NoError(t, os.WriteFile(extPath, []byte("rules:\n  max-depth:\n    max: 5\n"), 0644))

	result, err := resolveExtendPath(extPath, "/some/base")
	require.NoError(t, err)
	assert.Equal(t, extPath, result)
}

func TestResolveExtendPathRelative(t *testing.T) {
	tmpDir := t.TempDir()
	extPath := filepath.Join(tmpDir, "ext.yml")
	require.NoError(t, os.WriteFile(extPath, []byte("rules:\n  max-depth:\n    max: 5\n"), 0644))

	result, err := resolveExtendPath("ext.yml", tmpDir)
	require.NoError(t, err)
	assert.Equal(t, extPath, result)
}

func TestResolveExtendPathNotFound(t *testing.T) {
	_, err := resolveExtendPath("/nonexistent/path.yml", "/tmp")
	assert.Error(t, err)
}

func TestFindConfigsWithGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("root: true\nrules:\n  max-depth:\n    max: 5\n"), 0644))

	gitignoreContent := "node_modules\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644))

	configs, found, err := FindConfigsWithGitignore(tmpDir)

	require.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, configs)
	assert.Contains(t, configs[0].Exclude, "**/node_modules")
}

func TestFindConfigsWithGitignoreNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	configs, found, err := FindConfigsWithGitignore(tmpDir)

	require.NoError(t, err)
	assert.False(t, found)
	assert.NotEmpty(t, configs) // default config
}

func TestGetRuleConfigInt(t *testing.T) {
	cfg := &Config{
		Rules: map[string]interface{}{
			"max-depth": map[string]interface{}{"max": 5},
		},
	}

	type MaxDepth struct {
		Max int
	}
	result, ok := GetRuleConfig[*MaxDepth](cfg, "max-depth")

	assert.True(t, ok)
	assert.Equal(t, 5, result.Max)
}

func TestGetRuleConfigNotFound(t *testing.T) {
	cfg := &Config{
		Rules: map[string]interface{}{},
	}

	type MaxDepth struct {
		Max int
	}
	_, ok := GetRuleConfig[*MaxDepth](cfg, "nonexistent")
	assert.False(t, ok)
}

func TestGetRuleConfigDisabledByZero(t *testing.T) {
	cfg := &Config{
		Rules: map[string]interface{}{
			"max-depth": 0,
		},
	}

	type MaxDepth struct {
		Max int
	}
	_, ok := GetRuleConfig[*MaxDepth](cfg, "max-depth")
	assert.False(t, ok)
}

func TestGetRuleConfigDisabledByFalse(t *testing.T) {
	cfg := &Config{
		Rules: map[string]interface{}{
			"max-depth": false,
		},
	}

	type MaxDepth struct {
		Max int
	}
	_, ok := GetRuleConfig[*MaxDepth](cfg, "max-depth")
	assert.False(t, ok)
}

func TestTypedRules(t *testing.T) {
	cfg := &Config{
		Rules: map[string]interface{}{
			"max-depth":                    map[string]interface{}{"max": 5},
			"max-files-in-dir":            map[string]interface{}{"max": 100},
			"max-subdirs":                 map[string]interface{}{"max": 10},
			"naming-convention":           map[string]interface{}{"*.ts": "camelCase"},
			"file-existence":              map[string]interface{}{"index.ts": "exists:1"},
			"regex-match":                 map[string]interface{}{"*.go": "package"},
			"disallowed-patterns":         []interface{}{"*.exe", "*.dll"},
			"test-adjacency":              map[string]interface{}{"pattern": "_test.go"},
			"test-location":               map[string]interface{}{"integration-test-dir": "tests/"},
			"enforce-layer-boundaries":    map[string]interface{}{},
			"disallow-orphaned-files":     map[string]interface{}{},
			"disallow-import-cycles":      map[string]interface{}{},
			"path-based-layers":           map[string]interface{}{"layers": []interface{}{}},
		},
	}

	rules := cfg.TypedRules()

	require.NotNil(t, rules)
	require.NotNil(t, rules.MaxDepth)
	assert.Equal(t, 5, rules.MaxDepth.Max)
	require.NotNil(t, rules.MaxFilesInDir)
	assert.Equal(t, 100, rules.MaxFilesInDir.Max)
	require.NotNil(t, rules.MaxSubdirs)
	assert.Equal(t, 10, rules.MaxSubdirs.Max)
	assert.Equal(t, "camelCase", rules.NamingConvention["*.ts"])
	assert.Equal(t, "exists:1", rules.FileExistence["index.ts"])
	assert.Equal(t, "package", rules.RegexMatch["*.go"])
	assert.Equal(t, []string{"*.exe", "*.dll"}, []string(rules.DisallowedPatterns))
	require.NotNil(t, rules.TestAdjacency)
	assert.Equal(t, "_test.go", rules.TestAdjacency.Pattern)
	require.NotNil(t, rules.TestLocation)
	assert.Equal(t, "tests/", rules.TestLocation.IntegrationTestDir)
	require.NotNil(t, rules.EnforceLayerBoundaries)
	require.NotNil(t, rules.DisallowOrphanedFiles)
	require.NotNil(t, rules.DisallowImportCycles)
	require.NotNil(t, rules.PathBasedLayers)
}

func TestTypedRulesNilConfig(t *testing.T) {
	var nilCfg *Config
	rules := nilCfg.TypedRules()
	assert.Nil(t, rules)
}

func TestTypedRulesNoRules(t *testing.T) {
	cfg := &Config{
		Rules: map[string]interface{}{},
	}
	rules := cfg.TypedRules()
	assert.Nil(t, rules)
}

func TestLoadWithExtendsWithYamlSubdir(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subconfig")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	subConfig := filepath.Join(subDir, "base.yml")
	require.NoError(t, os.WriteFile(subConfig, []byte("rules:\n  max-depth:\n    max: 7\n"), 0644))

	mainConfig := filepath.Join(tmpDir, ".structurelint.yml")
	content := "extends: subconfig/base.yml\nrules:\n  max-files:\n    max: 10\n"
	require.NoError(t, os.WriteFile(mainConfig, []byte(content), 0644))

	config, err := Load(mainConfig)
	require.NoError(t, err)
	assert.Equal(t, 7, config.Rules["max-depth"].(map[string]interface{})["max"])
}

func TestResolveExtendPathRelativePrefix(t *testing.T) {
	tmpDir := t.TempDir()
	extPath := filepath.Join(tmpDir, "ext.yml")
	require.NoError(t, os.WriteFile(extPath, []byte("rules: {}"), 0644))

	result, err := resolveExtendPath("./ext.yml", tmpDir)
	require.NoError(t, err)
	assert.Equal(t, extPath, result)
}

func TestApplyGitignoreMissingGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Rules:   make(map[string]interface{}),
		Exclude: []string{"dist/**"},
	}

	result, err := ApplyGitignore(cfg, tmpDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"dist/**"}, result.Exclude)
}

func TestLoadWithNonExistentConfigInvalidAbsError(t *testing.T) {
	// Test that reading a file with a read error path goes to failure
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".structurelint.yml")

	// Create a directory with the same name to make os.ReadFile fail
	require.NoError(t, os.Mkdir(configFile, 0644))

	_, err := Load(configFile)
	assert.Error(t, err)
}
