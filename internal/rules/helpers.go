// Package rules provides shared helper functions for rule implementations.
package rules

import (
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// MatchesGlobPattern checks if a path matches a glob pattern.
// Supports `*` (no slash crossing), `?`, `[...]`, and `**` (any path including empty).
// Negation patterns must be handled by the caller (look for `!` prefix).
func MatchesGlobPattern(path, pattern string) bool {
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	if path == pattern {
		return true
	}

	re := globToRegexp(pattern)
	if re == nil {
		return false
	}
	if re.MatchString(path) {
		return true
	}

	// Common ergonomic case: pattern without `**` should also match by basename
	// (e.g., `*.pyc` matches `foo/bar.pyc`). Avoid for `**` patterns where the
	// caller intended a full-path match.
	if !strings.Contains(pattern, "**") && !strings.Contains(pattern, "/") {
		base := filepath.Base(path)
		if re.MatchString(base) {
			return true
		}
	}
	return false
}

var globRegexpCache sync.Map // pattern → *regexp.Regexp

// globToRegexp converts a glob pattern to an anchored regexp.
// Returns nil if the regexp could not be built (caller falls back to no match).
func globToRegexp(pattern string) *regexp.Regexp {
	if v, ok := globRegexpCache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}
	var b strings.Builder
	b.WriteString("^")
	i := 0
	for i < len(pattern) {
		switch c := pattern[i]; c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// `**` — match anything including slashes.
				// Special-case `**/` and `/**` to make leading/trailing
				// segments optional (so `**/foo` matches `foo` and `bar/foo`).
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					b.WriteString("(?:.*/)?")
					i += 3
					continue
				}
				if i > 0 && pattern[i-1] == '/' {
					// Already wrote a `/`, replace by `(?:.*)?`.
					s := b.String()
					b.Reset()
					b.WriteString(strings.TrimSuffix(s, "/"))
					b.WriteString("(?:/.*)?")
					i += 2
					continue
				}
				b.WriteString(".*")
				i += 2
			} else {
				b.WriteString("[^/]*")
				i++
			}
		case '?':
			b.WriteString("[^/]")
			i++
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
			i++
		case '[':
			// Pass through character class as-is until the closing `]`.
			j := i + 1
			for j < len(pattern) && pattern[j] != ']' {
				j++
			}
			if j < len(pattern) {
				b.WriteString(pattern[i : j+1])
				i = j + 1
			} else {
				b.WriteString("\\[")
				i++
			}
		default:
			b.WriteByte(c)
			i++
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	globRegexpCache.Store(pattern, re)
	return re
}
