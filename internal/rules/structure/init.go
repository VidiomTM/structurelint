package structure

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func init() {
	rules.Register("max-depth", func(ctx *rules.RuleContext) (rules.Rule, error) {
		max, ok := ctx.GetInt("max")
		if !ok {
			return nil, fmt.Errorf("missing 'max' parameter")
		}
		overrides := parseMaxDepthOverrides(ctx.Config["overrides"])
		if len(overrides) == 0 {
			return NewMaxDepthRule(max), nil
		}
		return NewMaxDepthRuleWithOverrides(max, overrides), nil
	})

	rules.Register("max-files-in-dir", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxFilesRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	rules.Register("max-subdirs", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxSubdirsRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	rules.Register("file-existence", func(ctx *rules.RuleContext) (rules.Rule, error) {
		reqs, ok := ctx.GetStringMap("")
		if !ok {
			return nil, fmt.Errorf("invalid configuration: file-existence expects map of pattern -> 'exists:N' / 'exists:N-M'")
		}
		if errs := ValidateFileExistenceConfig(reqs); len(errs) > 0 {
			return nil, fmt.Errorf("file-existence config errors: %s", strings.Join(errs, "; "))
		}
		return NewFileExistenceRule(reqs), nil
	})

	rules.Register("regex-match", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewRegexMatchRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("disallowed-patterns", func(ctx *rules.RuleContext) (rules.Rule, error) {
		var patterns []string
		if list, ok := ctx.Config["patterns"].([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
		} else if list, ok := ctx.Config[""].([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
		}
		if len(patterns) > 0 {
			return NewDisallowedPatternsRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("naming-convention", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewNamingConventionRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("uniqueness-constraints", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if constraints, ok := ctx.GetStringMap(""); ok {
			return NewUniquenessConstraintsRule(constraints), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("case-conflicts", func(_ *rules.RuleContext) (rules.Rule, error) {
		return NewCaseConflictsRule(), nil
	})

	rules.Register("disallow-empty-dirs", func(_ *rules.RuleContext) (rules.Rule, error) {
		return NewEmptyDirsRule(), nil
	})

	rules.Register("disallow-symlinks", func(_ *rules.RuleContext) (rules.Rule, error) {
		return NewSymlinksRule(), nil
	})

	rules.Register("disallow-deep-relative-imports", func(ctx *rules.RuleContext) (rules.Rule, error) {
		max := 3
		if v, ok := ctx.GetInt("max-parents"); ok && v > 0 {
			max = v
		}
		return NewDeepRelativeImportsRule(max), nil
	})
}

// parseMaxDepthOverrides accepts:
//
//	overrides:
//	  "src/routes/**": 8
//	  "tests/**": 6
//
// or a list form:
//
//	overrides:
//	  - pattern: "src/routes/**"
//	    max: 8
func parseMaxDepthOverrides(raw interface{}) []MaxDepthOverride {
	if raw == nil {
		return nil
	}
	var overrides []MaxDepthOverride
	switch v := raw.(type) {
	case map[string]interface{}:
		for pattern, val := range v {
			if max := toInt(val); max > 0 {
				overrides = append(overrides, MaxDepthOverride{Pattern: pattern, Max: max})
			}
		}
	case []interface{}:
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			pattern, _ := m["pattern"].(string)
			max := toInt(m["max"])
			if pattern != "" && max > 0 {
				overrides = append(overrides, MaxDepthOverride{Pattern: pattern, Max: max})
			}
		}
	}
	return overrides
}

func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	}
	return 0
}
