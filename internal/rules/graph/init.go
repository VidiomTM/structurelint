package graph

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

var errImportGraphRequired = fmt.Errorf("import graph required")

func requireImportGraph(ctx *rules.RuleContext, constructor func(*graph.ImportGraph) rules.Rule) (rules.Rule, error) {
	if ctx.ImportGraph == nil {
		return nil, errImportGraphRequired
	}
	return constructor(ctx.ImportGraph), nil
}

func init() {
	rules.Register("enforce-layer-boundaries", func(ctx *rules.RuleContext) (rules.Rule, error) {
		return requireImportGraph(ctx, func(g *graph.ImportGraph) rules.Rule { return NewLayerBoundariesRule(g) })
	})

	rules.Register("disallow-orphaned-files", func(ctx *rules.RuleContext) (rules.Rule, error) {
		return requireImportGraph(ctx, func(g *graph.ImportGraph) rules.Rule {
			return NewOrphanedFilesRule(g, []string{})
		})
	})

	rules.Register("disallow-import-cycles", func(ctx *rules.RuleContext) (rules.Rule, error) {
		return requireImportGraph(ctx, func(g *graph.ImportGraph) rules.Rule { return NewImportCyclesRule(g) })
	})
}
