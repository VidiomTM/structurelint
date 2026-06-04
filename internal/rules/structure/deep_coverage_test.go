package structure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestFileExistence_checkRequirement_InvalidFormat(t *testing.T) {
	r := &FileExistenceRule{}
	err := r.checkRequirement("", "README.md", "noexists:1", nil)
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestFileExistence_checkRequirement_RangeViolation(t *testing.T) {
	r := NewFileExistenceRule(map[string]string{"*": "exists:1-2"})
	files := []walker.FileInfo{
		{Path: "a.go", ParentPath: "", IsDir: false},
		{Path: "b.go", ParentPath: "", IsDir: false},
		{Path: "c.go", ParentPath: "", IsDir: false},
	}
	dirs := map[string]*walker.DirInfo{"": {}}
	v := r.Check(files, dirs)
	if len(v) != 1 {
		t.Errorf("expected 1 violation for range violation, got %d", len(v))
	}
}

func TestFileExistence_checkRequirement_MaxRange(t *testing.T) {
	r := NewFileExistenceRule(map[string]string{"*.go": "exists:0"})
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: "", IsDir: false},
	}
	dirs := map[string]*walker.DirInfo{"": {}}
	v := r.Check(files, dirs)
	if len(v) != 1 {
		t.Errorf("expected 1 violation for max 0 range, got %d", len(v))
	}
}

func TestFileExistence_parseCountSpec_RangeInvalidMin(t *testing.T) {
	r := &FileExistenceRule{}
	_, _, err := r.parseCountSpec("abc-5")
	if err == nil {
		t.Error("expected error for invalid min in range")
	}
}

func TestFileExistence_parseCountSpec_RangeInvalidMax(t *testing.T) {
	r := &FileExistenceRule{}
	_, _, err := r.parseCountSpec("1-xyz")
	if err == nil {
		t.Error("expected error for invalid max in range")
	}
}

func TestFileExistence_fileMatchesPattern_Exact(t *testing.T) {
	r := &FileExistenceRule{}
	file := walker.FileInfo{Path: "README.md", ParentPath: ""}
	if !r.fileMatchesPattern(file, "README.md") {
		t.Error("expected exact match")
	}
}

func TestFileExistence_fileMatchesPattern_NoMatch(t *testing.T) {
	r := &FileExistenceRule{}
	file := walker.FileInfo{Path: "main.go", ParentPath: ""}
	if r.fileMatchesPattern(file, "*.py") {
		t.Error("expected no match for different ext")
	}
}

func TestMaxFiles_Check_RootDir(t *testing.T) {
	r := NewMaxFilesRule(2)
	files := []walker.FileInfo{
		{Path: "f1.go", ParentPath: "", IsDir: false},
		{Path: "f2.go", ParentPath: "", IsDir: false},
		{Path: "f3.go", ParentPath: "", IsDir: false},
	}
	dirs := map[string]*walker.DirInfo{"": {}}
	v := r.Check(files, dirs)
	if len(v) != 1 {
		t.Errorf("expected 1 violation for root dir, got %d", len(v))
	}
}

func TestMatchesConvention_DirPattern(t *testing.T) {
	r := NewNamingConventionRule(map[string]string{"components/": "PascalCase"})
	v := r.Check([]walker.FileInfo{
		{Path: "components/myComponent", IsDir: true},
	}, nil)
	if len(v) != 1 {
		t.Errorf("expected 1 violation for dir pattern, got %d", len(v))
	}
}

func TestMatchesConvention_UnknownConvention(t *testing.T) {
	r := NewNamingConventionRule(map[string]string{"*.xyz": "unknown-convention"})
	v := r.Check([]walker.FileInfo{
		{Path: "test.xyz"},
	}, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for unknown convention, got %d", len(v))
	}
}

func TestMatchesPattern_WithPathSeparator(t *testing.T) {
	if !matchesPattern("src/components/index.ts", "src/**/*.ts") {
		t.Error("matchesPattern should match path with globstar")
	}
}

func TestDetectConvention_KebabCase(t *testing.T) {
	r := NewNamingConventionRule(nil)
	got := r.detectConvention("my-component")
	if got != "kebab-case" {
		t.Errorf("detectConvention(my-component) = %q, want kebab-case", got)
	}
}

func TestDetectConvention_Lowercase(t *testing.T) {
	r := NewNamingConventionRule(nil)
	got := r.detectConvention("hello world")
	if got != "lowercase" {
		t.Errorf("detectConvention(hello world) = %q, want lowercase", got)
	}
}

func TestDetectConvention_Unknown(t *testing.T) {
	r := NewNamingConventionRule(nil)
	got := r.detectConvention("Mixed_Case-Name")
	if got != "unknown/mixed" {
		t.Errorf("detectConvention(Mixed_Case-Name) = %q, want unknown/mixed", got)
	}
}

func TestConvertToConvention_Lowercase(t *testing.T) {
	r := NewNamingConventionRule(nil)
	got := r.convertToConvention("MyFile", "lowercase")
	if got != "myfile" {
		t.Errorf("convertToConvention(lowercase) = %q, want myfile", got)
	}
}

func TestConvertToConvention_Uppercase(t *testing.T) {
	r := NewNamingConventionRule(nil)
	got := r.convertToConvention("my_file", "uppercase")
	if got != "MYFILE" {
		t.Errorf("convertToConvention(uppercase) = %q, want MYFILE", got)
	}
}

func TestConvertToConvention_Default(t *testing.T) {
	r := NewNamingConventionRule(nil)
	got := r.convertToConvention("myFile", "other-convention")
	if got != "myFile" {
		t.Errorf("convertToConvention(default) = %q, want myFile", got)
	}
}

func TestTitleCase_Empty(t *testing.T) {
	got := titleCase("")
	if got != "" {
		t.Errorf("titleCase('') = %q, want ''", got)
	}
}

func TestToCamelCase_Empty(t *testing.T) {
	got := toCamelCase(nil)
	if got != "" {
		t.Errorf("toCamelCase(nil) = %q, want ''", got)
	}
}

func TestToCamelCase_SingleWord(t *testing.T) {
	got := toCamelCase([]string{"hello"})
	if got != "hello" {
		t.Errorf("toCamelCase([hello]) = %q, want hello", got)
	}
}

func TestMatchesUniquenessPattern_SuffixMatchOnly(t *testing.T) {
	if matchesUniquenessPattern("main.go", "_test.go") {
		t.Error("matchesUniquenessPattern should not match without * prefix")
	}
}

func TestUniqueness_SkipsDirs(t *testing.T) {
	r := NewUniquenessConstraintsRule(map[string]string{"*_service.py": "singleton"})
	v := r.Check([]walker.FileInfo{
		{Path: "src", IsDir: true},
		{Path: "src/services", IsDir: true},
	}, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for dirs, got %d", len(v))
	}
}

func TestDeepRelativeImports_AbsoluteImport(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "service.go")
	code := `package main
import "fmt"
`
	_ = os.WriteFile(fixture, []byte(code), 0644)
	r := NewDeepRelativeImportsRule(3)
	v := r.Check([]walker.FileInfo{
		{AbsPath: fixture, Path: "service.go"},
	}, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for absolute import, got %d", len(v))
	}
}

func TestInit_RegisterFactories(t *testing.T) {
	tests := []struct {
		name string
		cfg  map[string]interface{}
	}{
		{"max-depth", map[string]interface{}{"max": 5}},
		{"max-files-in-dir", map[string]interface{}{"max": 10}},
		{"max-subdirs", map[string]interface{}{"max": 3}},
		{"file-existence", map[string]interface{}{"README.md": "exists:1"}},
		{"regex-match", map[string]interface{}{"pattern.*": "error"}},
		{"disallowed-patterns", map[string]interface{}{"patterns": []interface{}{"vendor/", "node_modules/"}}},
		{"disallowed-patterns", map[string]interface{}{"": []interface{}{"vendor/"}}},
		{"naming-convention", map[string]interface{}{"*.go": "PascalCase"}},
		{"uniqueness-constraints", map[string]interface{}{"*_service.py": "singleton"}},
		{"case-conflicts", map[string]interface{}{}},
		{"disallow-empty-dirs", map[string]interface{}{}},
		{"disallow-symlinks", map[string]interface{}{}},
		{"disallow-deep-relative-imports", map[string]interface{}{"max-parents": 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, ok := rules.GetFactory(tt.name)
			if !ok {
				t.Fatalf("GetFactory(%q) not found", tt.name)
			}
			rule, err := factory(&rules.RuleContext{Config: tt.cfg})
			if err != nil {
				t.Fatalf("factory(%q) error: %v", tt.name, err)
			}
			if rule == nil {
				t.Fatalf("factory(%q) returned nil", tt.name)
			}
		})
	}
}

func TestInit_RegisterErrors(t *testing.T) {
	tests := []struct {
		name string
		cfg  map[string]interface{}
	}{
		{"max-depth", map[string]interface{}{}},
		{"max-files-in-dir", map[string]interface{}{}},
		{"max-subdirs", map[string]interface{}{}},
		{"file-existence", map[string]interface{}{}},
		{"regex-match", map[string]interface{}{}},
		{"disallowed-patterns", map[string]interface{}{}},
		{"naming-convention", map[string]interface{}{}},
		{"uniqueness-constraints", map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, ok := rules.GetFactory(tt.name)
			if !ok {
				t.Fatalf("GetFactory(%q) not found", tt.name)
			}
			_, err := factory(&rules.RuleContext{Config: tt.cfg})
			if err == nil {
				t.Errorf("factory(%q) expected error, got nil", tt.name)
			}
		})
	}
}
