package rules

import (
	"context"
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// ExternalRuleAdapter adapts a RulePlugin to the Rule interface
type ExternalRuleAdapter struct {
	plugin interface {
		Name() string
		Check(ctx context.Context, files []walker.FileInfo, config map[string]interface{}) ([]Violation, error)
	}
	config map[string]interface{}
}

// NewExternalRuleAdapter creates a new adapter
func NewExternalRuleAdapter(plugin interface {
	Name() string
	Check(ctx context.Context, files []walker.FileInfo, config map[string]interface{}) ([]Violation, error)
}, config map[string]interface{}) *ExternalRuleAdapter {
	return &ExternalRuleAdapter{
		plugin: plugin,
		config: config,
	}
}

func (r *ExternalRuleAdapter) Name() string {
	return r.plugin.Name()
}

func (r *ExternalRuleAdapter) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	ctx := context.Background()
	
	violations, err := r.plugin.Check(ctx, files, r.config)
	if err != nil {
		// Report plugin failure as a violation
		return []Violation{{
			Rule:    r.Name(),
			Path:    ".",
			Message: fmt.Sprintf("Plugin execution failed: %v", err),
		}}
	}
	return violations
}
