package test

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestTestAdjacencyRule_Check_Adjacent(t *testing.T) {
	tests := []struct {
		name          string
		filePatterns  []string
		exemptions    []string
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:         "Go file with test - no violation",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "main.go", ParentPath: "", IsDir: false},
				{Path: "main_test.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "Go file without test - violation",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "service.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 1,
		},
		{
			name:         "main.go is framework-exempt entrypoint",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "main.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "SvelteKit +page.svelte is framework-exempt",
			filePatterns: []string{"**/*.svelte"},
			files: []walker.FileInfo{
				{Path: "src/routes/+page.svelte", ParentPath: "src/routes", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "TS declaration files are framework-exempt",
			filePatterns: []string{"**/*.ts"},
			files: []walker.FileInfo{
				{Path: "types/global.d.ts", ParentPath: "types", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "TypeScript file with test - no violation",
			filePatterns: []string{"**/*.ts"},
			files: []walker.FileInfo{
				{Path: "utils.ts", ParentPath: "src", IsDir: false},
				{Path: "utils.test.ts", ParentPath: "src", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "exempted file - no violation",
			filePatterns: []string{"**/*.go"},
			exemptions:   []string{"cmd/**/*.go"},
			files: []walker.FileInfo{
				{Path: "cmd/main.go", ParentPath: "cmd", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "test file itself - no violation",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "main_test.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:         "multiple files in same dir - some missing tests",
			filePatterns: []string{"**/*.go"},
			files: []walker.FileInfo{
				{Path: "file1.go", ParentPath: "", IsDir: false},
				{Path: "file1_test.go", ParentPath: "", IsDir: false},
				{Path: "file2.go", ParentPath: "", IsDir: false},
			},
			wantViolCount: 1, // file2.go missing test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewTestAdjacencyRule("adjacent", "", tt.filePatterns, tt.exemptions)

			// Act
			violations := rule.Check(tt.files, nil)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestTestAdjacencyRule_getTestFileName(t *testing.T) {
	rule := &TestAdjacencyRule{}

	tests := []struct {
		sourcePath   string
		wantTestName string
	}{
		{"main.go", "main_test.go"},
		{"utils.ts", "utils.test.ts"},
		{"component.tsx", "component.test.tsx"},
		{"helper.js", "helper.spec.js"},
		{"module.py", "test_module.py"},
		{"src/file.go", "file_test.go"},
	}

	for _, tt := range tests {
		t.Run(tt.sourcePath, func(t *testing.T) {
			got := rule.getTestFileName(tt.sourcePath)
			if got != tt.wantTestName {
				t.Errorf("getTestFileName(%q) = %q, want %q", tt.sourcePath, got, tt.wantTestName)
			}
		})
	}
}

func TestTestAdjacencyRule_isTestFile(t *testing.T) {
	rule := &TestAdjacencyRule{}

	tests := []struct {
		path string
		want bool
	}{
		{"main_test.go", true},
		{"utils.test.ts", true},
		{"component.spec.js", true},
		{"test_module.py", false}, // Python test prefix not detected by basename check
		{"FileTest.java", true},
		{"file_spec.rb", true},
		{"main.go", false},
		{"utils.ts", false},
		{"module.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := rule.isTestFile(tt.path)
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestTestAdjacencyRule_Name(t *testing.T) {
	rule := NewTestAdjacencyRule("adjacent", "", []string{"**/*.go"}, nil)
	if got := rule.Name(); got != "test-adjacency" {
		t.Errorf("Name() = %v, want test-adjacency", got)
	}
}

func TestTestAdjacencyRule_SeparatePattern_Violation(t *testing.T) {
	rule := NewTestAdjacencyRule("separate", "tests", []string{"**/*.go"}, nil)
	files := []walker.FileInfo{
		{Path: "service.go", ParentPath: "", IsDir: false},
	}
	v := rule.Check(files, nil)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Path != "service.go" {
		t.Errorf("violation path = %s, want service.go", v[0].Path)
	}
}

func TestTestAdjacencyRule_SeparatePattern_NoViolation(t *testing.T) {
	rule := NewTestAdjacencyRule("separate", "tests", []string{"**/*.go"}, nil)
	files := []walker.FileInfo{
		{Path: "service.go", ParentPath: "", IsDir: false},
		{Path: "tests/service_test.go", ParentPath: "tests", IsDir: false},
	}
	v := rule.Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(v), v)
	}
}

func TestTestAdjacencyRule_SeparatePattern_ExemptedFile(t *testing.T) {
	rule := NewTestAdjacencyRule("separate", "tests", []string{"**/*.go"}, []string{"cmd/**"})
	files := []walker.FileInfo{
		{Path: "cmd/main.go", ParentPath: "cmd", IsDir: false},
	}
	v := rule.Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for exempted file, got %d", len(v))
	}
}

func TestTestAdjacencyRule_SeparatePattern_EmptyFilePatterns(t *testing.T) {
	rule := NewTestAdjacencyRule("separate", "tests", []string{}, nil)
	files := []walker.FileInfo{
		{Path: "service.go", ParentPath: "", IsDir: false},
	}
	v := rule.Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations with empty patterns, got %d", len(v))
	}
}

func TestTestAdjacencyRule_getExpectedTestPath(t *testing.T) {
	rule := NewTestAdjacencyRule("separate", "tests", nil, nil)
	tests := []struct {
		source string
		want   string
	}{
		{"service.go", "tests/service_test.go"},
		{"src/service.go", "tests/src/service_test.go"},
		{"sub/service.go", "tests/sub/service_test.go"},
	}
	for _, c := range tests {
		got := rule.getExpectedTestPath(c.source)
		if got != c.want {
			t.Errorf("getExpectedTestPath(%q) = %q, want %q", c.source, got, c.want)
		}
	}
}
