package parser

import (
	"bufio"
	"os"
	"strings"
)

// DirectiveType represents the type of directive
type DirectiveType string

const (
	// DirectiveIgnore ignores specific rules for a file
	DirectiveIgnore DirectiveType = "ignore"
	// DirectiveNoTest indicates a file doesn't need tests (legacy, kept for backward compatibility)
	DirectiveNoTest DirectiveType = "no-test"
)

// Directive represents a structurelint directive found in a file
type Directive struct {
	Type   DirectiveType // Type of directive (ignore, no-test, etc.)
	Rules  []string      // Specific rules to ignore (empty means all rules)
	Reason string        // Human-readable reason for the directive
	Line   int           // Line number where directive was found
}

// ParseDirectives scans a file and extracts all structurelint directives
// Directives should be placed near the top of the file (scans first 100 lines)
func ParseDirectives(absPath string) []Directive {
	file, err := os.Open(absPath)
	if err != nil {
		return nil
	}
	defer func() { _ = file.Close() }()

	var directives []Directive
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Only scan first 100 lines (directives should be at the top)
	for scanner.Scan() && lineNum < 100 {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Look for @structurelint: prefix
		if !strings.Contains(line, "@structurelint:") {
			continue
		}

		// Parse the directive
		if directive := parseDirectiveLine(line, lineNum); directive != nil {
			directives = append(directives, *directive)
		}
	}

	return directives
}

// parseDirectiveLine parses a single line containing a structurelint directive
func parseDirectiveLine(line string, lineNum int) *Directive {
	// Only parse directives in comments (lines starting with //, #, /*, or *)
	trimmed := strings.TrimSpace(line)
	if !isComment(trimmed) {
		return nil
	}

	// Find the directive marker
	idx := strings.Index(line, "@structurelint:")
	if idx == -1 {
		return nil
	}

	// Extract content after @structurelint:
	content := line[idx+len("@structurelint:"):]

	// Parse different directive types
	if strings.HasPrefix(content, "ignore") {
		return parseIgnoreDirective(content, lineNum)
	} else if strings.HasPrefix(content, "no-test") {
		return parseNoTestDirective(content, lineNum)
	}

	return nil
}

// isComment checks if a line is a comment
func isComment(line string) bool {
	return strings.HasPrefix(line, "//") ||
		strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "/*") ||
		strings.HasPrefix(line, "*")
}

// parseIgnoreDirective parses an @structurelint:ignore directive
// Format: @structurelint:ignore [rule1 rule2 ...] [reason]
// Examples:
//   @structurelint:ignore max-depth max-files-in-dir Legacy code, refactor planned
//   @structurelint:ignore Generated file, do not lint
func parseIgnoreDirective(content string, lineNum int) *Directive {
	// Remove "ignore" prefix
	content = strings.TrimPrefix(content, "ignore")
	content = strings.TrimSpace(content)

	if content == "" {
		// No rules or reason specified - ignore all rules
		return &Directive{
			Type:   DirectiveIgnore,
			Rules:  []string{},
			Reason: "no reason provided",
			Line:   lineNum,
		}
	}

	// Parse rules and reason
	// Rules are dash-separated words, reason is the remaining text
	var rules []string
	var reasonParts []string
	inReason := false

	words := strings.Fields(content)
	for _, word := range words {
		// If we see a word that doesn't look like a rule name, it's the start of the reason
		if !inReason && !isRuleName(word) {
			inReason = true
		}

		if inReason {
			reasonParts = append(reasonParts, word)
		} else {
			rules = append(rules, word)
		}
	}

	reason := strings.Join(reasonParts, " ")
	if reason == "" {
		reason = "no reason provided"
	}

	return &Directive{
		Type:   DirectiveIgnore,
		Rules:  rules,
		Reason: reason,
		Line:   lineNum,
	}
}

// parseNoTestDirective parses an @structurelint:no-test directive (legacy support)
// Format: @structurelint:no-test [reason]
func parseNoTestDirective(content string, lineNum int) *Directive {
	// Remove "no-test" prefix
	content = strings.TrimPrefix(content, "no-test")
	reason := strings.TrimSpace(content)

	if reason == "" {
		reason = "no reason provided"
	}

	return &Directive{
		Type:   DirectiveNoTest,
		Rules:  []string{"test-adjacency", "test-location"}, // no-test applies to test rules
		Reason: reason,
		Line:   lineNum,
	}
}

// isRuleName checks if a word is a valid rule name format.
// Rules are expected to be kebab-case: lowercase letters, numbers, and hyphens.
func isRuleName(word string) bool {
	if len(word) == 0 {
		return false
	}

	// Rule names must not start or end with a hyphen
	if strings.HasPrefix(word, "-") || strings.HasSuffix(word, "-") {
		return false
	}

	// Check each character is lowercase letter, digit, or hyphen
	for _, r := range word {
		isLowercaseLetter := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isHyphen := r == '-'
		if !isLowercaseLetter && !isDigit && !isHyphen {
			return false
		}
	}

	// Rule names should contain at least one hyphen (to distinguish from regular words)
	return strings.Contains(word, "-")
}

// HasDirectiveForRule checks if a file has a directive that applies to a specific rule
func HasDirectiveForRule(directives []Directive, ruleName string) (bool, string) {
	for _, directive := range directives {
		// Check if this directive applies to the rule
		if directive.Type == DirectiveIgnore {
			// Empty rules list means ignore all rules
			if len(directive.Rules) == 0 {
				return true, directive.Reason
			}

			// Check if this specific rule is in the list
			for _, rule := range directive.Rules {
				if rule == ruleName {
					return true, directive.Reason
				}
			}
		} else if directive.Type == DirectiveNoTest && isTestRule(ruleName) {
			// no-test directive applies to test-related rules
			return true, directive.Reason
		}
	}

	return false, ""
}

// isTestRule checks if a rule is test-related
func isTestRule(ruleName string) bool {
	return ruleName == "test-adjacency" || ruleName == "test-location"
}
