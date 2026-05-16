package ci

type StrategyRegistry struct {
	strategies map[ProjectType]Strategy
}

func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{strategies: make(map[ProjectType]Strategy)}
}

func (r *StrategyRegistry) Register(s Strategy) {
	r.strategies[s.ProjectType()] = s
}

func (r *StrategyRegistry) StrategiesFor(types []ProjectType) []Strategy {
	var out []Strategy
	for _, t := range types {
		if s, ok := r.strategies[t]; ok {
			out = append(out, s)
		}
	}
	return out
}
