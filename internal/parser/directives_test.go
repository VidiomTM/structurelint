package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDirectives(t *testing.T) {
	t.Run("no directives", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.go")
		os.WriteFile(f, []byte("package main\n"), 0644)
		directives := ParseDirectives(f)
		if len(directives) != 0 {
			t.Errorf("expected 0 directives, got %d", len(directives))
		}
	})

	t.Run("ignore directive - all rules", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.go")
		os.WriteFile(f, []byte("// @structurelint:ignore Legacy code\npackage main\n"), 0644)
		directives := ParseDirectives(f)
		if len(directives) != 1 {
			t.Fatalf("expected 1 directive, got %d", len(directives))
		}
		if directives[0].Type != DirectiveIgnore {
			t.Errorf("expected ignore, got %s", directives[0].Type)
		}
		if len(directives[0].Rules) != 0 {
			t.Errorf("expected empty rules, got %v", directives[0].Rules)
		}
	})

	t.Run("ignore directive - specific rules", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.go")
		os.WriteFile(f, []byte("// @structurelint:ignore max-depth max-files-in-dir Legacy code\npackage main\n"), 0644)
		directives := ParseDirectives(f)
		if len(directives) != 1 {
			t.Fatalf("expected 1 directive, got %d", len(directives))
		}
		if len(directives[0].Rules) != 2 {
			t.Errorf("expected 2 rules, got %d", len(directives[0].Rules))
		}
	})

	t.Run("no-test directive", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.go")
		os.WriteFile(f, []byte("// @structurelint:no-test Interface definitions\npackage main\n"), 0644)
		directives := ParseDirectives(f)
		if len(directives) != 1 {
			t.Fatalf("expected 1 directive, got %d", len(directives))
		}
		if directives[0].Type != DirectiveNoTest {
			t.Errorf("expected no-test, got %s", directives[0].Type)
		}
	})

	t.Run("invalid file returns nil", func(t *testing.T) {
		directives := ParseDirectives("/nonexistent/file.go")
		if directives != nil {
			t.Errorf("expected nil, got %v", directives)
		}
	})

	t.Run("only scans first 100 lines", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.go")
		content := make([]byte, 0, 5000)
		for i := 0; i < 150; i++ {
			content = append(content, []byte("// line\n")...)
		}
		content = append(content, []byte("// @structurelint:ignore test-rule\n")...)
		os.WriteFile(f, content, 0644)
		directives := ParseDirectives(f)
		if len(directives) != 0 {
			t.Errorf("expected 0 directives (past 100 lines), got %d", len(directives))
		}
	})
}

func TestParseDirectiveLine(t *testing.T) {
	t.Run("not a comment", func(t *testing.T) {
		d := parseDirectiveLine("@structurelint:ignore", 1)
		if d != nil {
			t.Error("expected nil for non-comment line")
		}
	})

	t.Run("no marker", func(t *testing.T) {
		d := parseDirectiveLine("// nothing here", 1)
		if d != nil {
			t.Error("expected nil")
		}
	})

	t.Run("ignore all rules no reason", func(t *testing.T) {
		d := parseDirectiveLine("// @structurelint:ignore", 5)
		if d == nil {
			t.Fatal("expected directive")
		}
		if d.Type != DirectiveIgnore {
			t.Errorf("expected ignore, got %s", d.Type)
		}
		if len(d.Rules) != 0 {
			t.Errorf("expected empty rules, got %v", d.Rules)
		}
	})

	t.Run("ignore with rules and reason", func(t *testing.T) {
		d := parseDirectiveLine("// @structurelint:ignore max-depth max-files-in-dir Legacy code", 5)
		if d == nil {
			t.Fatal("expected directive")
		}
		if d.Type != DirectiveIgnore {
			t.Errorf("expected ignore, got %s", d.Type)
		}
		if len(d.Rules) != 2 {
			t.Errorf("expected 2 rules, got %d", len(d.Rules))
		}
		expectedReason := "Legacy code"
		if d.Reason != expectedReason {
			t.Errorf("expected reason %q, got %q", expectedReason, d.Reason)
		}
	})

	t.Run("no-test with reason", func(t *testing.T) {
		d := parseDirectiveLine("// @structurelint:no-test Interface only", 5)
		if d == nil {
			t.Fatal("expected directive")
		}
		if d.Type != DirectiveNoTest {
			t.Errorf("expected no-test, got %s", d.Type)
		}
		expectedReason := "Interface only"
		if d.Reason != expectedReason {
			t.Errorf("expected reason %q, got %q", expectedReason, d.Reason)
		}
	})

	t.Run("no-test without reason", func(t *testing.T) {
		d := parseDirectiveLine("// @structurelint:no-test", 5)
		if d == nil {
			t.Fatal("expected directive")
		}
		if d.Type != DirectiveNoTest {
			t.Errorf("expected no-test, got %s", d.Type)
		}
	})

	t.Run("unknown directive", func(t *testing.T) {
		d := parseDirectiveLine("// @structurelint:unknown", 1)
		if d != nil {
			t.Error("expected nil for unknown directive")
		}
	})
}

func TestIsComment(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"// comment", true},
		{"# comment", true},
		{"/* comment */", true},
		{"* comment", true},
		{"not a comment", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			got := isComment(tc.line)
			if got != tc.want {
				t.Errorf("isComment(%q) = %v, want %v", tc.line, got, tc.want)
			}
		})
	}
}

func TestIsRuleName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"max-depth", true},
		{"max-files-in-dir", true},
		{"naming-convention", true},
		{"test-adjacency", true},
		{"simple", false},
		{"-leading-hyphen", false},
		{"trailing-hyphen-", false},
		{"MixedCase", false},
		{"", false},
		{"a-b", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isRuleName(tc.name)
			if got != tc.want {
				t.Errorf("isRuleName(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestHasDirectiveForRule(t *testing.T) {
	t.Run("ignore all rules", func(t *testing.T) {
		directives := []Directive{
			{Type: DirectiveIgnore, Rules: []string{}, Reason: "legacy"},
		}
		ignored, reason := HasDirectiveForRule(directives, "max-depth")
		if !ignored {
			t.Error("expected ignored")
		}
		if reason != "legacy" {
			t.Errorf("expected 'legacy', got %q", reason)
		}
	})

	t.Run("ignore specific rule", func(t *testing.T) {
		directives := []Directive{
			{Type: DirectiveIgnore, Rules: []string{"max-depth", "max-files"}, Reason: "legacy"},
		}
		ignored, _ := HasDirectiveForRule(directives, "max-depth")
		if !ignored {
			t.Error("expected ignored")
		}
	})

	t.Run("ignore different rule", func(t *testing.T) {
		directives := []Directive{
			{Type: DirectiveIgnore, Rules: []string{"max-depth"}, Reason: "legacy"},
		}
		ignored, _ := HasDirectiveForRule(directives, "naming-convention")
		if ignored {
			t.Error("expected not ignored")
		}
	})

	t.Run("no-test directive applies to test rules", func(t *testing.T) {
		directives := []Directive{
			{Type: DirectiveNoTest, Rules: []string{"test-adjacency", "test-location"}, Reason: "interface"},
		}
		ignored, _ := HasDirectiveForRule(directives, "test-adjacency")
		if !ignored {
			t.Error("expected test-adjacency to be ignored by no-test")
		}
	})

	t.Run("no-test does not apply to non-test rules", func(t *testing.T) {
		directives := []Directive{
			{Type: DirectiveNoTest, Rules: []string{"test-adjacency", "test-location"}, Reason: "interface"},
		}
		ignored, _ := HasDirectiveForRule(directives, "max-depth")
		if ignored {
			t.Error("expected max-depth not to be ignored by no-test")
		}
	})

	t.Run("no directives returns false", func(t *testing.T) {
		ignored, _ := HasDirectiveForRule([]Directive{}, "max-depth")
		if ignored {
			t.Error("expected false with no directives")
		}
	})
}

func TestIsTestRule(t *testing.T) {
	if !isTestRule("test-adjacency") {
		t.Error("expected test-adjacency to be test rule")
	}
	if !isTestRule("test-location") {
		t.Error("expected test-location to be test rule")
	}
	if isTestRule("max-depth") {
		t.Error("expected max-depth not to be test rule")
	}
}

func TestParseIgnoreDirective(t *testing.T) {
	t.Run("empty content", func(t *testing.T) {
		d := parseIgnoreDirective("ignore", 1)
		if d == nil {
			t.Fatal("expected directive")
		}
		if len(d.Rules) != 0 {
			t.Errorf("expected empty rules, got %v", d.Rules)
		}
	})

	t.Run("with rules and reason", func(t *testing.T) {
		d := parseIgnoreDirective("ignore max-depth max-files-in-dir legacy", 1)
		if d == nil {
			t.Fatal("expected directive")
		}
		if len(d.Rules) != 2 {
			t.Errorf("expected 2 rules, got %d", len(d.Rules))
		}

		d2 := parseIgnoreDirective("ignore", 1)
		if d2 != nil {
			if d2.Reason != "no reason provided" {
				t.Errorf("expected default reason, got %q", d2.Reason)
			}
		}
	})
}

func TestParseNoTestDirective(t *testing.T) {
	t.Run("with reason", func(t *testing.T) {
		d := parseNoTestDirective("no-test Interface only", 1)
		if d == nil {
			t.Fatal("expected directive")
		}
		if d.Type != DirectiveNoTest {
			t.Errorf("expected no-test, got %s", d.Type)
		}
		if d.Reason != "Interface only" {
			t.Errorf("expected 'Interface only', got %q", d.Reason)
		}
	})

	t.Run("without reason", func(t *testing.T) {
		d := parseNoTestDirective("no-test", 1)
		if d == nil {
			t.Fatal("expected directive")
		}
		if d.Reason != "no reason provided" {
			t.Errorf("expected default reason, got %q", d.Reason)
		}
	})
}

func TestParseDirectives_MultipleDirectives(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.go")
	content := "// @structurelint:ignore max-depth Legacy\n// @structurelint:no-test Interface\npackage main\n"
	os.WriteFile(f, []byte(content), 0644)
	directives := ParseDirectives(f)
	if len(directives) != 2 {
		t.Fatalf("expected 2 directives, got %d", len(directives))
	}
	if directives[0].Type != DirectiveIgnore {
		t.Errorf("expected first to be ignore, got %s", directives[0].Type)
	}
	if directives[1].Type != DirectiveNoTest {
		t.Errorf("expected second to be no-test, got %s", directives[1].Type)
	}
}
