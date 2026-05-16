package structure

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

const (
	errMissingMax    = "missing 'max' parameter"
	errInvalidConfig = "invalid configuration"
)

func init() {
	rules.Register("max-depth", newMaxDepthRule)
	rules.Register("max-files-in-dir", newMaxFilesRule)
	rules.Register("max-subdirs", newMaxSubdirsRule)
	rules.Register("file-existence", newFileExistenceRule)
	rules.Register("regex-match", newRegexMatchRule)
	rules.Register("disallowed-patterns", newDisallowedPatternsRule)
	rules.Register("naming-convention", newNamingConventionRule)
	rules.Register("uniqueness-constraints", newUniquenessConstraintsRule)
	rules.Register("case-conflicts", newSimpleRule(func() rules.Rule { return NewCaseConflictsRule() }))
	rules.Register("disallow-empty-dirs", newSimpleRule(func() rules.Rule { return NewEmptyDirsRule() }))
	rules.Register("disallow-symlinks", newSimpleRule(func() rules.Rule { return NewSymlinksRule() }))
	rules.Register("disallow-deep-relative-imports", newDeepRelativeImportsRule)
}

func newSimpleRule(fn func() rules.Rule) func(*rules.RuleContext) (rules.Rule, error) {
	return func(_ *rules.RuleContext) (rules.Rule, error) {
		return fn(), nil
	}
}

func newMaxDepthRule(ctx *rules.RuleContext) (rules.Rule, error) {
	max, ok := ctx.GetInt("max")
	if !ok {
		return nil, fmt.Errorf(errMissingMax)
	}
	overrides := parseMaxDepthOverrides(ctx.Config["overrides"])
	if len(overrides) == 0 {
		return NewMaxDepthRule(max), nil
	}
	return NewMaxDepthRuleWithOverrides(max, overrides), nil
}

func newMaxFilesRule(ctx *rules.RuleContext) (rules.Rule, error) {
	if max, ok := ctx.GetInt("max"); ok {
		return NewMaxFilesRule(max), nil
	}
	return nil, fmt.Errorf(errMissingMax)
}

func newMaxSubdirsRule(ctx *rules.RuleContext) (rules.Rule, error) {
	if max, ok := ctx.GetInt("max"); ok {
		return NewMaxSubdirsRule(max), nil
	}
	return nil, fmt.Errorf(errMissingMax)
}

func newFileExistenceRule(ctx *rules.RuleContext) (rules.Rule, error) {
	reqs, ok := ctx.GetStringMap("")
	if !ok {
		return nil, fmt.Errorf("invalid configuration: file-existence expects map of pattern -> 'exists:N' / 'exists:N-M'")
	}
	if errs := ValidateFileExistenceConfig(reqs); len(errs) > 0 {
		return nil, fmt.Errorf("file-existence config errors: %s", strings.Join(errs, "; "))
	}
	return NewFileExistenceRule(reqs), nil
}

func newRegexMatchRule(ctx *rules.RuleContext) (rules.Rule, error) {
	if patterns, ok := ctx.GetStringMap(""); ok {
		return NewRegexMatchRule(patterns), nil
	}
	return nil, fmt.Errorf(errInvalidConfig)
}

func newDisallowedPatternsRule(ctx *rules.RuleContext) (rules.Rule, error) {
	patterns := extractPatterns(ctx.Config["patterns"], ctx.Config[""])
	if len(patterns) > 0 {
		return NewDisallowedPatternsRule(patterns), nil
	}
	return nil, fmt.Errorf(errInvalidConfig)
}

func newNamingConventionRule(ctx *rules.RuleContext) (rules.Rule, error) {
	if patterns, ok := ctx.GetStringMap(""); ok {
		return NewNamingConventionRule(patterns), nil
	}
	return nil, fmt.Errorf(errInvalidConfig)
}

func newUniquenessConstraintsRule(ctx *rules.RuleContext) (rules.Rule, error) {
	if constraints, ok := ctx.GetStringMap(""); ok {
		return NewUniquenessConstraintsRule(constraints), nil
	}
	return nil, fmt.Errorf(errInvalidConfig)
}

func newDeepRelativeImportsRule(ctx *rules.RuleContext) (rules.Rule, error) {
	max := 3
	if v, ok := ctx.GetInt("max-parents"); ok && v > 0 {
		max = v
	}
	return NewDeepRelativeImportsRule(max), nil
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

func extractPatterns(configPatterns, emptyPatterns interface{}) []string {
	var patterns []string
	if list, ok := configPatterns.([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				patterns = append(patterns, s)
			}
		}
	} else if list, ok := emptyPatterns.([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				patterns = append(patterns, s)
			}
		}
	}
	return patterns
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
