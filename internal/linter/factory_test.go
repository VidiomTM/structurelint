package linter

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/stretchr/testify/assert"
)

func TestNewRuleFactory(t *testing.T) {
	f := NewRuleFactory("/test", &config.Config{}, nil)
	assert.NotNil(t, f)
	assert.Equal(t, "/test", f.rootDir)
}

func TestCheckBreakingChanges_NoRemoved(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{"max-depth": 5}}}
	assert.NoError(t, f.checkBreakingChanges())
}

func TestCheckBreakingChanges_Removed(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{"max-cyclomatic-complexity": 5}}}
	assert.Error(t, f.checkBreakingChanges())
}

func TestNormalizeConfig_Map(t *testing.T) {
	f := &RuleFactory{}
	m := map[string]interface{}{"key": "val"}
	assert.Equal(t, m, f.normalizeConfig(m))
}

func TestNormalizeConfig_NonMap(t *testing.T) {
	f := &RuleFactory{}
	m := f.normalizeConfig(5)
	assert.Equal(t, map[string]interface{}{"": 5}, m)
}

func TestExtractStringSlice(t *testing.T) {
	f := &RuleFactory{}
	r := f.extractStringSlice([]interface{}{"a", "b"})
	assert.Equal(t, []string{"a", "b"}, r)
}

func TestGetStringFromMap(t *testing.T) {
	f := &RuleFactory{}
	r := f.getStringFromMap(map[string]interface{}{"key": "val"}, "key")
	assert.Equal(t, "val", r)
}

func TestGetStringFromMap_Missing(t *testing.T) {
	f := &RuleFactory{}
	r := f.getStringFromMap(map[string]interface{}{}, "key")
	assert.Equal(t, "", r)
}

func TestGetBoolFromMap(t *testing.T) {
	f := &RuleFactory{}
	r := f.getBoolFromMap(map[string]interface{}{"key": true}, "key")
	assert.True(t, r)
}

func TestGetBoolFromMap_Missing(t *testing.T) {
	f := &RuleFactory{}
	r := f.getBoolFromMap(map[string]interface{}{}, "key")
	assert.False(t, r)
}

func TestGetIntFromMap(t *testing.T) {
	f := &RuleFactory{}
	r := f.getIntFromMap(map[string]interface{}{"key": 42}, "key")
	assert.Equal(t, 42, r)
}

func TestGetIntFromMap_Float(t *testing.T) {
	f := &RuleFactory{}
	r := f.getIntFromMap(map[string]interface{}{"key": float64(42)}, "key")
	assert.Equal(t, 42, r)
}

func TestGetIntFromMap_Missing(t *testing.T) {
	f := &RuleFactory{}
	r := f.getIntFromMap(map[string]interface{}{}, "key")
	assert.Equal(t, 0, r)
}

func TestGetStringSliceFromMap(t *testing.T) {
	f := &RuleFactory{}
	r := f.getStringSliceFromMap(map[string]interface{}{"key": []interface{}{"a", "b"}}, "key")
	assert.Equal(t, []string{"a", "b"}, r)
}

func TestGetStringSliceFromMap_Missing(t *testing.T) {
	f := &RuleFactory{}
	r := f.getStringSliceFromMap(map[string]interface{}{}, "key")
	assert.Nil(t, r)
}

func TestParsePathLayers(t *testing.T) {
	f := &RuleFactory{}
	layers := f.parsePathLayers([]interface{}{
		map[string]interface{}{
			"name":     "app",
			"patterns": []interface{}{"src/**"},
		},
	})
	assert.Len(t, layers, 1)
	assert.Equal(t, "app", layers[0].Name)
}

func TestParsePathLayers_BadEntry(t *testing.T) {
	f := &RuleFactory{}
	layers := f.parsePathLayers([]interface{}{"not-a-map"})
	assert.Len(t, layers, 0)
}

func TestCreateGraphDependentRules_NilGraph(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{}}, importGraph: nil}
	rules := f.createGraphDependentRules()
	assert.Nil(t, rules)
}

func TestCreateLayerBoundariesRule_NoConfig(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{}}, importGraph: &graph.ImportGraph{}}
	assert.Nil(t, f.createLayerBoundariesRule())
}

func TestCreateOrphanedFilesRule_NoConfig(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{}}}
	assert.Nil(t, f.createOrphanedFilesRule())
}

func TestFactoryIsRuleEnabled_NilConfig(t *testing.T) {
	f := &RuleFactory{}
	assert.False(t, f.isRuleEnabled("any"))
}

func TestFactoryIsRuleEnabled_NotExists(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{}}}
	assert.False(t, f.isRuleEnabled("any"))
}

func TestFactoryIsRuleEnabled_Zero(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{"any": 0}}}
	assert.False(t, f.isRuleEnabled("any"))
}

func TestFactoryIsRuleEnabled_False(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{"any": false}}}
	assert.False(t, f.isRuleEnabled("any"))
}

func TestFactoryIsRuleEnabled_Enabled(t *testing.T) {
	f := &RuleFactory{config: &config.Config{Rules: map[string]interface{}{"any": true}}}
	assert.True(t, f.isRuleEnabled("any"))
}
