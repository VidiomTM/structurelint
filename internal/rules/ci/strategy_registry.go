package ci

import "github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"

type StrategyRegistry struct {
	strategies map[core.ProjectType]core.Strategy
}

func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{strategies: make(map[core.ProjectType]core.Strategy)}
}

func (r *StrategyRegistry) Register(s core.Strategy) {
	r.strategies[s.ProjectType()] = s
}

func (r *StrategyRegistry) StrategiesFor(types []core.ProjectType) []core.Strategy {
	var out []core.Strategy
	for _, t := range types {
		if s, ok := r.strategies[t]; ok {
			out = append(out, s)
		}
	}
	return out
}
