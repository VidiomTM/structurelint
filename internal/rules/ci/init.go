package ci

import (
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/strategies"
)

func init() {
	rules.Register("github-workflows", func(ctx *rules.RuleContext) (rules.Rule, error) {
		reader := OSFileReader{}
		registry := newDefaultRegistry(reader, ctx.Config)
		detector := NewProjectDetector(reader)
		return NewWorkflowQualityRule(registry, reader, detector, ctx.Config), nil
	})
}

func newDefaultRegistry(reader core.FileReader, cfg map[string]interface{}) *StrategyRegistry {
	registry := NewStrategyRegistry()
	registry.Register(strategies.NewSvelteKitStrategy(reader, extractConfig("sveltekit", cfg)))
	registry.Register(strategies.NewPythonStrategy(reader, extractConfig("python", cfg)))
	registry.Register(strategies.NewGoStrategy(reader, extractConfig("golang", cfg)))
	registry.Register(strategies.NewRustStrategy(reader, extractConfig("rust", cfg)))
	return registry
}

func extractConfig(pt string, fullCfg map[string]interface{}) map[string]interface{} {
	if fullCfg == nil {
		return nil
	}
	if v, ok := fullCfg[pt].(map[string]interface{}); ok {
		return v
	}
	return nil
}
