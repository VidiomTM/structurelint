package linter

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrphanedFilesRule_EnabledWithConfig(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"disallow-orphaned-files": map[string]interface{}{
					"entry-point-patterns": []interface{}{"src/main.ts", "src/index.ts"},
				},
			},
		},
		importGraph: &graph.ImportGraph{},
	}
	rule := f.createOrphanedFilesRule()
	assert.NotNil(t, rule)
}

func TestCreateOrphanedFilesRule_EnabledWithNonMapConfig(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"disallow-orphaned-files": true,
			},
		},
		importGraph: &graph.ImportGraph{},
	}
	rule := f.createOrphanedFilesRule()
	assert.NotNil(t, rule)
}

func TestCreateOrphanedFilesRule_DisabledByZero(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"disallow-orphaned-files": 0,
			},
		},
	}
	rule := f.createOrphanedFilesRule()
	assert.Nil(t, rule)
}

func TestCreatePathBasedLayerRules_EnabledWithConfig(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"path-based-layers": map[string]interface{}{
					"layers": []interface{}{
						map[string]interface{}{
							"name":           "app",
							"patterns":       []interface{}{"src/app/**"},
							"canDependOn":    []interface{}{"domain"},
							"forbiddenPaths": []interface{}{"src/app/legacy/**"},
						},
					},
				},
			},
		},
	}
	rules := f.createPathBasedLayerRules()
	assert.Len(t, rules, 1)
}

func TestCreatePathBasedLayerRules_NonMapConfig(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"path-based-layers": true,
			},
		},
	}
	rules := f.createPathBasedLayerRules()
	assert.Nil(t, rules)
}

func TestCreatePathBasedLayerRules_Disabled(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"path-based-layers": false,
			},
		},
	}
	rules := f.createPathBasedLayerRules()
	assert.Nil(t, rules)
}

func TestCreateTestValidationRules_Adjacency(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"test-adjacency": map[string]interface{}{
					"pattern":      "*.test.ts",
					"test-dir":     "__tests__",
					"file-patterns": []interface{}{"*.ts"},
					"exemptions":   []interface{}{"*.d.ts"},
				},
			},
		},
	}
	rules := f.createTestValidationRules()
	assert.Len(t, rules, 1)
}

func TestCreateTestValidationRules_AdjacencyMissingPattern(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"test-adjacency": map[string]interface{}{
					"pattern": "",
				},
			},
		},
	}
	rules := f.createTestValidationRules()
	assert.Empty(t, rules)
}

func TestCreateTestValidationRules_AdjacencyNonMap(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"test-adjacency": true,
			},
		},
	}
	rules := f.createTestValidationRules()
	assert.Empty(t, rules)
}

func TestCreateTestValidationRules_Location(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"test-location": map[string]interface{}{
					"integration-test-dir": "e2e",
					"allow-adjacent":       true,
					"file-patterns":        []interface{}{"*.test.ts"},
					"exemptions":           []interface{}{"*.d.ts"},
				},
			},
		},
	}
	rules := f.createTestValidationRules()
	assert.Len(t, rules, 1)
}

func TestCreateTestValidationRules_LocationNonMap(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"test-location": true,
			},
		},
	}
	rules := f.createTestValidationRules()
	assert.Empty(t, rules)
}

func TestCreateLayerBoundariesRule_EnabledWithLayers(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"enforce-layer-boundaries": true,
			},
			Layers: []config.Layer{
				{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
			},
		},
		importGraph: &graph.ImportGraph{},
	}
	rule := f.createLayerBoundariesRule()
	assert.NotNil(t, rule)
}

func TestCreateLayerBoundariesRule_EnabledNoLayers(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"enforce-layer-boundaries": true,
			},
		},
		importGraph: &graph.ImportGraph{},
	}
	rule := f.createLayerBoundariesRule()
	assert.Nil(t, rule)
}

func TestCreateGraphDependentRules_WithGraph(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"enforce-layer-boundaries": true,
			},
			Layers: []config.Layer{
				{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
			},
		},
		importGraph: &graph.ImportGraph{},
	}
	rules := f.createGraphDependentRules()
	assert.NotEmpty(t, rules)
}

func TestFactoryCreateRegistryRules_FactoryNotFound(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"nonexistent-rule": true,
			},
		},
	}
	rules := f.createRegistryRules()
	assert.Empty(t, rules)
}

func TestFactoryCreateRegistryRules_FactoryError(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": "invalid",
			},
		},
	}
	rules := f.createRegistryRules()
	assert.Empty(t, rules)
}

func TestFactoryIsRuleEnabled_StringType(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"some-rule": "enabled",
			},
		},
	}
	assert.True(t, f.isRuleEnabled("some-rule"))
}

func TestFactoryIsRuleEnabled_NonZeroInt(t *testing.T) {
	f := &RuleFactory{
		config: &config.Config{
			Rules: map[string]interface{}{
				"some-rule": 5,
			},
		},
	}
	assert.True(t, f.isRuleEnabled("some-rule"))
}

func TestLinterIsRuleEnabled_MapTypeConfig(t *testing.T) {
	l := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": map[string]interface{}{"max": 5},
			},
		},
	}
	assert.True(t, l.isRuleEnabled("max-depth"))
}

func TestLinterGetRuleConfig_NonZeroInt(t *testing.T) {
	l := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": 5,
			},
		},
	}
	val, ok := l.getRuleConfig("max-depth")
	assert.True(t, ok)
	assert.Equal(t, 5, val)
}

func TestLinterGetRuleConfig_TrueBool(t *testing.T) {
	l := &Linter{
		config: &config.Config{
			Rules: map[string]interface{}{
				"max-depth": true,
			},
		},
	}
	val, ok := l.getRuleConfig("max-depth")
	assert.True(t, ok)
	assert.Equal(t, true, val)
}

func TestGetStringSliceFromMap_NonStringItems(t *testing.T) {
	f := &RuleFactory{}
	r := f.getStringSliceFromMap(map[string]interface{}{"key": []interface{}{"a", 5, "b"}}, "key")
	assert.Equal(t, []string{"a", "b"}, r)
}

func TestExtractStringSlice_NonStrings(t *testing.T) {
	f := &RuleFactory{}
	r := f.extractStringSlice([]interface{}{"a", 5, "b"})
	assert.Equal(t, []string{"a", "b"}, r)
}
