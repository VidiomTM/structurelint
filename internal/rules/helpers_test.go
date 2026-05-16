package rules

import "testing"

func TestMatchesGlobPattern_EdgeCases(t *testing.T) {
	cases := []struct {
		path, pattern string
		want          bool
	}{
		{"foo/bar/baz.go", "**/baz.go", true},
		{"foo/bar/baz.go", "**/bar/*.go", true},
		{"foo/bar/baz.go", "**/bar/baz.*", true},
		{"foo/bar/baz.go", "**/foo/**", true},
		{"a/b/c/d.go", "**/c/*.go", true},
		{"x", "?", true},
		{"xy", "?", false},
		{"x/y/z", "?/?/*", true},
		{"foo.go", "[a-z]*.go", true},
		{"Foo.go", "[a-z]*.go", false},
		{"path/file.test.js", "*.test.*", true},
		{"path/file.test.js", "*.spec.*", false},
	}
	for _, c := range cases {
		got := MatchesGlobPattern(c.path, c.pattern)
		if got != c.want {
			t.Errorf("MatchesGlobPattern(%q, %q) = %v, want %v", c.path, c.pattern, got, c.want)
		}
	}
}

func TestMatchesGlobPattern_ExactMatch(t *testing.T) {
	if !MatchesGlobPattern("foo.go", "foo.go") {
		t.Error("expected exact match")
	}
}

func TestGlobToRegexp_InvalidPattern(t *testing.T) {
	re := globToRegexp("[z-a]")
	if re != nil {
		t.Error("expected nil for invalid range pattern")
	}
}

func TestGlobToRegexp_NegateCharClass(t *testing.T) {
	re := globToRegexp("[!a-z]*.go")
	if re == nil {
		t.Fatal("expected non-nil regexp")
	}
	if !re.MatchString("Foo.go") {
		t.Error("[!a-z] should match uppercase start")
	}
	if re.MatchString("foo.go") {
		t.Error("[!a-z] should NOT match lowercase start")
	}
}

func TestMatchesGlobPattern_NegationPrefix(t *testing.T) {
	if MatchesGlobPattern("foo.go", "!foo.go") {
		t.Error("negation prefix should not match as glob")
	}
}

func TestMatchesGlobPattern_DoubleStarDir(t *testing.T) {
	if !MatchesGlobPattern("anything", "**") {
		t.Error("** should match everything")
	}
}
