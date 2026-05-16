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
	globPatternToRegex(pattern, &b)
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	globRegexpCache.Store(pattern, re)
	return re
}

func globPatternToRegex(pattern string, b *strings.Builder) {
	i := 0
	for i < len(pattern) {
		switch c := pattern[i]; c {
		case '*':
			i = globAppendStar(pattern, i, b)
		case '?':
			b.WriteString("[^/]")
			i++
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
			i++
		case '[':
			i = globAppendCharClass(pattern, i, b)
		default:
			b.WriteByte(c)
			i++
		}
	}
}

func globAppendStar(pattern string, i int, b *strings.Builder) int {
	if i+1 < len(pattern) && pattern[i+1] == '*' {
		if i+2 < len(pattern) && pattern[i+2] == '/' {
			b.WriteString("(?:.*/)?")
			return i + 3
		}
		if i > 0 && pattern[i-1] == '/' {
			s := b.String()
			b.Reset()
			b.WriteString(strings.TrimSuffix(s, "/"))
			b.WriteString("(?:/.*)?")
			return i + 2
		}
		b.WriteString(".*")
		return i + 2
	}
	b.WriteString("[^/]*")
	return i + 1
}

func globAppendCharClass(pattern string, i int, b *strings.Builder) int {
	j := i + 1
	for j < len(pattern) && pattern[j] != ']' {
		j++
	}
	if j < len(pattern) {
		cls := pattern[i : j+1]
		if len(cls) > 2 && cls[1] == '!' {
			cls = "[" + "^" + cls[2:]
		}
		b.WriteString(cls)
		return j + 1
	}
	b.WriteString("\\[")
	return i + 1
}
