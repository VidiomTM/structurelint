package rules

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestFormatDetailed_Basic(t *testing.T) {
	v := &Violation{Path: "main.go", Message: "naming violation", Expected: "PascalCase", Actual: "snake_case"}
	msg := v.FormatDetailed()
	if msg != "main.go: naming violation\n  Expected: PascalCase\n  Actual: snake_case" {
		t.Errorf("unexpected format: %s", msg)
	}
}

func TestFormatDetailed_WithContext(t *testing.T) {
	v := &Violation{Path: "main.go", Message: "naming violation", Context: "src/components/**"}
	msg := v.FormatDetailed()
	if msg != "main.go: naming violation\n  Context: src/components/**" {
		t.Errorf("unexpected format: %s", msg)
	}
}

func TestFormatDetailed_WithSuggestions(t *testing.T) {
	v := &Violation{Path: "main.go", Message: "naming violation", Suggestions: []string{"Use PascalCase", "Rename file"}}
	msg := v.FormatDetailed()
	expected := "main.go: naming violation\n  Suggestions:\n    - Use PascalCase\n    - Rename file"
	if msg != expected {
		t.Errorf("unexpected format:\nwant: %s\ngot:  %s", expected, msg)
	}
}

func TestFormatDetailed_Empty(t *testing.T) {
	v := &Violation{}
	msg := v.FormatDetailed()
	expected := ": "
	if msg != expected {
		t.Errorf("unexpected format: %s", msg)
	}
}

func TestFormatDetailed_NoExpectedActual(t *testing.T) {
	v := &Violation{Path: "main.go", Message: "simple violation"}
	msg := v.FormatDetailed()
	if msg != "main.go: simple violation" {
		t.Errorf("unexpected format: %s", msg)
	}
}

func TestShouldIgnoreFile(t *testing.T) {
	file := walker.FileInfo{Path: "main.go", Directives: []parser.Directive{
		{Type: parser.DirectiveIgnore, Rules: []string{"naming-rule"}},
	}}
	ignored, _ := ShouldIgnoreFile(file, "naming-rule")
	if !ignored {
		t.Error("expected file to be ignored")
	}
}

func TestShouldIgnoreFile_IgnoreAll(t *testing.T) {
	file := walker.FileInfo{Path: "main.go", Directives: []parser.Directive{
		{Type: parser.DirectiveIgnore, Reason: "test file"},
	}}
	ignored, reason := ShouldIgnoreFile(file, "any-rule")
	if !ignored {
		t.Error("expected file to be ignored")
	}
	if reason != "test file" {
		t.Errorf("reason = %s, want test file", reason)
	}
}

func TestShouldIgnoreFile_NotIgnored(t *testing.T) {
	file := walker.FileInfo{Path: "main.go"}
	ignored, _ := ShouldIgnoreFile(file, "naming-rule")
	if ignored {
		t.Error("expected file not to be ignored")
	}
}

func TestShouldIgnoreFile_Dir(t *testing.T) {
	file := walker.FileInfo{Path: "dir", IsDir: true}
	ignored, _ := ShouldIgnoreFile(file, "any-rule")
	if ignored {
		t.Error("expected dir not to be ignored")
	}
}

func TestFilterIgnoredFiles(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "keep.go"},
		{Path: "ignore.go", Directives: []parser.Directive{
			{Type: parser.DirectiveIgnore, Rules: []string{"test-rule"}},
		}},
		{Path: "also_keep.go"},
	}
	filtered := FilterIgnoredFiles(files, "test-rule")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 files, got %d", len(filtered))
	}
	if filtered[0].Path != "keep.go" {
		t.Errorf("first file = %s, want keep.go", filtered[0].Path)
	}
	if filtered[1].Path != "also_keep.go" {
		t.Errorf("second file = %s, want also_keep.go", filtered[1].Path)
	}
}

func TestFilterIgnoredFiles_NoFiles(t *testing.T) {
	filtered := FilterIgnoredFiles(nil, "test-rule")
	if len(filtered) != 0 {
		t.Errorf("expected 0 files, got %d", len(filtered))
	}
}
