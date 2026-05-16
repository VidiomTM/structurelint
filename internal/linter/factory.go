package linter

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	rulesgraph "github.com/Jonathangadeaharder/structurelint/internal/rules/graph"
	rulestest "github.com/Jonathangadeaharder/structurelint/internal/rules/test"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"

	// Blank imports to trigger sub-package init() registrations
	_ "github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
	_ "github.com/Jonathangadeaharder/structurelint/internal/rules/structure"
)

type RuleFactory struct {
	rootDir     string
	config      *config.Config
	importGraph *graph.ImportGraph
	files       []walker.FileInfo
	dirs        map[string]*walker.DirInfo
}

func NewRuleFactory(rootDir string, cfg *config.Config, importGraph *graph.ImportGraph) *RuleFactory {
	return &RuleFactory{
		rootDir:     rootDir,
		config:      cfg,
		importGraph: importGraph,
	}
}

func (f *RuleFactory) CreateRules(files []walker.FileInfo, dirs map[string]*walker.DirInfo) ([]rules.Rule, error) {
	f.files = files
	f.dirs = dirs

	if err := f.checkBreakingChanges(); err != nil {
		return nil, err
	}

	var rulesList []rules.Rule
	rulesList = append(rulesList, f.createRegistryRules()...)
	rulesList = append(rulesList, f.createGraphDependentRules()...)
	rulesList = append(rulesList, f.createPathBasedLayerRules()...)
	rulesList = append(rulesList, f.createTestValidationRules()...)

	return rulesList, nil
}

func (f *RuleFactory) checkBreakingChanges() error {
	removed := map[string]string{
		"max-cyclomatic-complexity": "Function-level complexity is out of scope for structurelint. Use a language-specific tool (gocognit, ruff, eslint complexity).",
		"max-cognitive-complexity":  "Function-level complexity is out of scope. Use gocognit / ruff / eslint complexity.",
		"max-halstead-effort":       "Function-level complexity is out of scope. Use a language-specific complexity tool.",

		"linter-config":             "Linter presence checks are out of scope. Use a presence rule via file-existence.",
		"contract-framework":        "Dependency-presence checks are out of scope. Encode requirements in a presence rule via file-existence.",
		"api-spec":                  "Replace with file-existence: `api/openapi.yaml: exists:1`.",
		"spec-adr-enforcement":      "Replace with file-existence + naming-convention.",
		"file-content":              "Template enforcement is out of scope. Use copier / cookiecutter.",
		"disallow-unused-exports":   "Cannot be done correctly without per-language symbol resolution. Use ts-prune / knip / ruff F401 / deadcode.",
		"property-enforcement":      "Replaced by 'disallow-import-cycles' (cycle detection only). max_dependencies_per_file / max_dependency_depth dropped — arbitrary metrics. forbidden_patterns is covered by 'path-based-layers' forbiddenPaths.",
	}
	for ruleName, advice := range removed {
		if _, ok := f.config.Rules[ruleName]; ok {
			return fmt.Errorf("rule '%s' has been removed in this version. %s", ruleName, advice)
		}
	}
	return nil
}

func (f *RuleFactory) createRegistryRules() []rules.Rule {
	var rulesList []rules.Rule

	for ruleName, ruleConfig := range f.config.Rules {
		if !f.isRuleEnabled(ruleName) {
			continue
		}

		factory, ok := rules.GetFactory(ruleName)
		if !ok {
			continue
		}

		ctx := &rules.RuleContext{
			RootDir:     f.rootDir,
			ImportGraph: f.importGraph,
			Config:      f.normalizeConfig(ruleConfig),
		}

		rule, err := factory(ctx)
		if err == nil && rule != nil {
			rulesList = append(rulesList, rule)
		}
	}

	return rulesList
}

func (f *RuleFactory) normalizeConfig(config interface{}) map[string]interface{} {
	if m, ok := config.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{
		"": config,
	}
}

func (f *RuleFactory) createGraphDependentRules() []rules.Rule {
	if f.importGraph == nil {
		return nil
	}

	var rulesList []rules.Rule
	if rule := f.createLayerBoundariesRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createOrphanedFilesRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	return rulesList
}

func (f *RuleFactory) createLayerBoundariesRule() rules.Rule {
	if _, ok := f.config.Rules["enforce-layer-boundaries"]; !ok {
		return nil
	}
	if !f.isRuleEnabled("enforce-layer-boundaries") || len(f.config.Layers) == 0 {
		return nil
	}
	return rulesgraph.NewLayerBoundariesRule(f.importGraph)
}

func (f *RuleFactory) createOrphanedFilesRule() rules.Rule {
	config, ok := f.config.Rules["disallow-orphaned-files"]
	if !ok || !f.isRuleEnabled("disallow-orphaned-files") {
		return nil
	}
	rule := rulesgraph.NewOrphanedFilesRule(f.importGraph, f.config.Entrypoints)
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return rule
	}
	patterns, ok := configMap["entry-point-patterns"].([]interface{})
	if !ok {
		return rule
	}
	entryPointPatterns := f.extractStringSlice(patterns)
	if len(entryPointPatterns) > 0 {
		rule = rule.WithEntryPointPatterns(entryPointPatterns)
	}
	return rule
}

func (f *RuleFactory) createPathBasedLayerRules() []rules.Rule {
	ruleConfig, ok := f.config.Rules["path-based-layers"]
	if !ok || !f.isRuleEnabled("path-based-layers") {
		return nil
	}

	configMap, ok := ruleConfig.(map[string]interface{})
	if !ok {
		return nil
	}

	layersConfig, ok := configMap["layers"].([]interface{})
	if !ok {
		return nil
	}

	pathLayers := f.parsePathLayers(layersConfig)
	if len(pathLayers) == 0 {
		return nil
	}

	return []rules.Rule{rulesgraph.NewPathBasedLayerRule(pathLayers)}
}

func (f *RuleFactory) parsePathLayers(layersConfig []interface{}) []rulesgraph.PathLayer {
	var pathLayers []rulesgraph.PathLayer

	for _, layerInterface := range layersConfig {
		layerMap, ok := layerInterface.(map[string]interface{})
		if !ok {
			continue
		}

		layer := rulesgraph.PathLayer{
			Name:           f.getStringFromMap(layerMap, "name"),
			Patterns:       f.getStringSliceFromMap(layerMap, "patterns"),
			CanDependOn:    f.getStringSliceFromMap(layerMap, "canDependOn"),
			ForbiddenPaths: f.getStringSliceFromMap(layerMap, "forbiddenPaths"),
		}
		pathLayers = append(pathLayers, layer)
	}

	return pathLayers
}

func (f *RuleFactory) createTestValidationRules() []rules.Rule {
	var rulesList []rules.Rule
	if rule := f.createTestAdjacencyRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createTestLocationRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	return rulesList
}

func (f *RuleFactory) createTestAdjacencyRule() rules.Rule {
	if !f.isRuleEnabled("test-adjacency") {
		return nil
	}
	testAdj, ok := f.config.Rules["test-adjacency"]
	if !ok {
		return nil
	}
	adjMap, ok := testAdj.(map[string]interface{})
	if !ok {
		return nil
	}
	pattern := f.getStringFromMap(adjMap, "pattern")
	testDir := f.getStringFromMap(adjMap, "test-dir")
	filePatterns := f.getStringSliceFromMap(adjMap, "file-patterns")
	exemptions := f.getStringSliceFromMap(adjMap, "exemptions")
	if pattern != "" && len(filePatterns) > 0 {
		return rulestest.NewTestAdjacencyRule(pattern, testDir, filePatterns, exemptions)
	}
	return nil
}

func (f *RuleFactory) createTestLocationRule() rules.Rule {
	if !f.isRuleEnabled("test-location") {
		return nil
	}
	testLoc, ok := f.config.Rules["test-location"]
	if !ok {
		return nil
	}
	locMap, ok := testLoc.(map[string]interface{})
	if !ok {
		return nil
	}
	integrationDir := f.getStringFromMap(locMap, "integration-test-dir")
	allowAdjacent := f.getBoolFromMap(locMap, "allow-adjacent")
	filePatterns := f.getStringSliceFromMap(locMap, "file-patterns")
	exemptions := f.getStringSliceFromMap(locMap, "exemptions")
	return rulestest.NewTestLocationRule(integrationDir, allowAdjacent, filePatterns, exemptions)
}

func (f *RuleFactory) isRuleEnabled(ruleName string) bool {
	if f.config == nil || f.config.Rules == nil {
		return false
	}

	value, exists := f.config.Rules[ruleName]
	if !exists {
		return false
	}

	switch v := value.(type) {
	case nil:
		return false
	case int:
		return v != 0
	case bool:
		return v
	}

	return true
}

func (f *RuleFactory) extractStringSlice(patterns []interface{}) []string {
	var result []string
	for _, p := range patterns {
		if str, ok := p.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func (f *RuleFactory) getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func (f *RuleFactory) getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func (f *RuleFactory) getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func (f *RuleFactory) getStringSliceFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if strVal, ok := v.(string); ok {
				result = append(result, strVal)
			}
		}
		return result
	}
	return nil
}
