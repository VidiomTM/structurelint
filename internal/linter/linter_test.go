package linter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Act
	linter := New()

	// Assert
	require.NotNil(t, linter)
	assert.Nil(t, linter.config)
}

func TestGetRuleConfig_Exists(t *testing.T) {
	// Arrange
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
			},
		},
	}

	// Act
	cfg, ok := linter.getRuleConfig("max-depth")

	// Assert
	assert.True(t, ok)
	assert.NotNil(t, cfg)
}

func TestGetRuleConfig_NotExists(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{},
		},
	}

	_, ok := linter.getRuleConfig("nonexistent-rule")
	assert.False(t, ok)
}

func TestGetRuleConfig_NilConfig(t *testing.T) {
	linter := &Linter{
		config: nil,
	}

	_, ok := linter.getRuleConfig("max-depth")
	assert.False(t, ok)
}

func TestGetRuleConfig_NilRules(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: nil,
		},
	}

	_, ok := linter.getRuleConfig("max-depth")
	assert.False(t, ok)
}

func TestGetRuleConfig_DisabledByZero(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": 0,
			},
		},
	}

	_, ok := linter.getRuleConfig("max-depth")
	assert.False(t, ok)
}

func TestGetRuleConfig_DisabledByFalse(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": false,
			},
		},
	}

	_, ok := linter.getRuleConfig("max-depth")
	assert.False(t, ok)
}

func TestGetRuleConfig_EnabledByTrue(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"enforce-layer-boundaries": true,
			},
		},
	}

	_, ok := linter.getRuleConfig("enforce-layer-boundaries")
	assert.True(t, ok)
}

func TestIsRuleEnabled_Enabled(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
			},
		},
	}

	assert.True(t, linter.isRuleEnabled("max-depth"))
}

func TestIsRuleEnabled_Disabled(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": 0,
			},
		},
	}

	assert.False(t, linter.isRuleEnabled("max-depth"))
}

func TestIsRuleEnabled_NotExists(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{},
		},
	}

	assert.False(t, linter.isRuleEnabled("nonexistent-rule"))
}

func TestLint_BasicRules(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "linter-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a simple project structure
	srcDir := filepath.Join(tmpDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))

	// Create a config file with max-depth rule
	configContent := `
root: true
rules:
  max-depth:
    max: 2
`
	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Create a deeply nested file (should violate max-depth: 2)
	deepDir := filepath.Join(srcDir, "level1", "level2", "level3")
	require.NoError(t, os.MkdirAll(deepDir, 0755))

	deepFile := filepath.Join(deepDir, "file.ts")
	require.NoError(t, os.WriteFile(deepFile, []byte("// test"), 0644))

	// Run linter
	linter := New()
	violations, err := linter.Lint(tmpDir)

	require.NoError(t, err)

	// Should have violations for exceeding max depth
	assert.NotEmpty(t, violations)

	// Verify at least one violation is about depth
	hasDepthViolation := false
	for _, v := range violations {
		if v.Rule == "max-depth" {
			hasDepthViolation = true
			break
		}
	}

	assert.True(t, hasDepthViolation)
}

func TestLint_NoConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "linter-test-noconfig")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.ts")
	require.NoError(t, os.WriteFile(testFile, []byte("// test"), 0644))

	l := New()
	_, err = l.Lint(tmpDir)

	assert.ErrorIs(t, err, ErrNoConfig)
}

func TestLint_WithConfig_NoViolations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "linter-test-withconfig")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configContent := `
root: true
rules:
  max-depth:
    max: 10
`
	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	testFile := filepath.Join(tmpDir, "test.ts")
	require.NoError(t, os.WriteFile(testFile, []byte("// test"), 0644))

	l := New()
	violations, err := l.Lint(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, violations)
}

func TestLint_WithLayers(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "linter-layers-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create directory structure
	domainDir := filepath.Join(tmpDir, "src", "domain")
	appDir := filepath.Join(tmpDir, "src", "application")

	require.NoError(t, os.MkdirAll(domainDir, 0755))
	require.NoError(t, os.MkdirAll(appDir, 0755))

	// Create domain file that violates layer boundaries by importing from application
	domainFile := filepath.Join(domainDir, "user.ts")
	domainContent := `
import { UserService } from '../application/userService';
export class User {}
`
	require.NoError(t, os.WriteFile(domainFile, []byte(domainContent), 0644))

	// Create application file
	appFile := filepath.Join(appDir, "userService.ts")
	appContent := `
import { User } from '../domain/user';
export class UserService {}
`
	require.NoError(t, os.WriteFile(appFile, []byte(appContent), 0644))

	// Create config with layers
	configContent := `
root: true
rules:
  enforce-layer-boundaries:
    enabled: true
layers:
  - name: domain
    path: src/domain/**
    dependsOn: []
  - name: application
    path: src/application/**
    dependsOn: [domain]
`
	configFile := filepath.Join(tmpDir, ".structurelint.yml")
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	// Run linter
	linter := New()
	violations, err := linter.Lint(tmpDir)

	require.NoError(t, err)

	// Should have violations for layer boundaries
	hasLayerViolation := false
	for _, v := range violations {
		if v.Rule == "enforce-layer-boundaries" {
			hasLayerViolation = true
			break
		}
	}

	assert.True(t, hasLayerViolation)
}

func TestCreateRules_MaxDepth(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
			},
		},
	}

	rules, _ := linter.createRules([]walker.FileInfo{}, nil)

	// Should have created max-depth rule
	hasMaxDepth := false
	for _, rule := range rules {
		if rule.Name() == "max-depth" {
			hasMaxDepth = true
			break
		}
	}

	assert.True(t, hasMaxDepth)
}

func TestCreateRules_NoRules(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{},
		},
	}

	rules, _ := linter.createRules([]walker.FileInfo{}, nil)

	assert.Empty(t, rules)
}

func TestCreateRules_MultipleRules(t *testing.T) {
	linter := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{
					"max": 5,
				},
				"max-files-in-dir": map[string]interface{}{
					"max": 10,
				},
				"max-subdirs": map[string]interface{}{
					"max": 8,
				},
			},
		},
	}

	rules, _ := linter.createRules([]walker.FileInfo{}, nil)

	// Should have created 3 rules
	assert.GreaterOrEqual(t, len(rules), 3)

	// Check that all rules are present
	ruleNames := make(map[string]bool)
	for _, rule := range rules {
		ruleNames[rule.Name()] = true
	}

	assert.True(t, ruleNames["max-depth"])
	assert.True(t, ruleNames["max-files-in-dir"])
	assert.True(t, ruleNames["max-subdirs"])
}
