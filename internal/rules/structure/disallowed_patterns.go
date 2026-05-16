// Package structure provides rule implementations for structurelint.
package structure

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// DisallowedPatternsRule prevents specific file or directory patterns
type DisallowedPatternsRule struct {
	Patterns []string
}

// Name returns the rule name
func (r *DisallowedPatternsRule) Name() string {
	return "disallowed-patterns"
}

// Check validates that disallowed patterns are not present
func (r *DisallowedPatternsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	disallowedPatterns, allowedPatterns := r.separatePatterns()
	var violations []rules.Violation

	for _, file := range files {
		for _, pattern := range disallowedPatterns {
			if rules.MatchesGlobPattern(file.Path, pattern) && !r.isAllowed(file.Path, allowedPatterns) {
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("matches disallowed pattern '%s'", pattern),
				})
			}
		}
	}

	return violations
}

func (r *DisallowedPatternsRule) separatePatterns() (disallowed, allowed []string) {
	for _, pattern := range r.Patterns {
		if strings.HasPrefix(pattern, "!") {
			allowed = append(allowed, strings.TrimPrefix(pattern, "!"))
		} else {
			disallowed = append(disallowed, pattern)
		}
	}
	return
}

func (r *DisallowedPatternsRule) isAllowed(path string, allowedPatterns []string) bool {
	for _, allowPattern := range allowedPatterns {
		if rules.MatchesGlobPattern(path, allowPattern) {
			return true
		}
	}
	return false
}

// NewDisallowedPatternsRule creates a new DisallowedPatternsRule
func NewDisallowedPatternsRule(patterns []string) *DisallowedPatternsRule {
	return &DisallowedPatternsRule{
		Patterns: patterns,
	}
}
