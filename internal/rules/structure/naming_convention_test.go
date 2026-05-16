package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
)

func TestNamingConventionRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		patterns      map[string]string
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:     "camelCase valid",
			patterns: map[string]string{"*.js": "camelCase"},
			files: []walker.FileInfo{
				{Path: "myFile.js"},
				{Path: "anotherFile.js"},
			},
			wantViolCount: 0,
		},
		{
			name:     "camelCase invalid - starts with uppercase",
			patterns: map[string]string{"*.js": "camelCase"},
			files: []walker.FileInfo{
				{Path: "MyFile.js"},
			},
			wantViolCount: 1,
		},
		{
			name:     "PascalCase valid",
			patterns: map[string]string{"*.tsx": "PascalCase"},
			files: []walker.FileInfo{
				{Path: "MyComponent.tsx"},
				{Path: "Button.tsx"},
			},
			wantViolCount: 0,
		},
		{
			name:     "PascalCase invalid - starts with lowercase",
			patterns: map[string]string{"*.tsx": "PascalCase"},
			files: []walker.FileInfo{
				{Path: "myComponent.tsx"},
			},
			wantViolCount: 1,
		},
		{
			name:     "kebab-case valid",
			patterns: map[string]string{"*.css": "kebab-case"},
			files: []walker.FileInfo{
				{Path: "my-styles.css"},
				{Path: "button-component.css"},
			},
			wantViolCount: 0,
		},
		{
			name:     "kebab-case invalid - has uppercase",
			patterns: map[string]string{"*.css": "kebab-case"},
			files: []walker.FileInfo{
				{Path: "MyStyles.css"},
			},
			wantViolCount: 1,
		},
		{
			name:     "snake_case valid",
			patterns: map[string]string{"*.py": "snake_case"},
			files: []walker.FileInfo{
				{Path: "my_module.py"},
				{Path: "test_utils.py"},
			},
			wantViolCount: 0,
		},
		{
			name:     "snake_case invalid - has uppercase",
			patterns: map[string]string{"*.py": "snake_case"},
			files: []walker.FileInfo{
				{Path: "MyModule.py"},
			},
			wantViolCount: 1,
		},
		{
			name:     "lowercase valid",
			patterns: map[string]string{"*.txt": "lowercase"},
			files: []walker.FileInfo{
				{Path: "readme.txt"},
			},
			wantViolCount: 0,
		},
		{
			name:     "multiple patterns",
			patterns: map[string]string{
				"*.js":  "camelCase",
				"*.tsx": "PascalCase",
			},
			files: []walker.FileInfo{
				{Path: "myFile.js"},
				{Path: "MyComponent.tsx"},
			},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNamingConventionRule(tt.patterns)
			violations := rule.Check(tt.files, nil)
			assert.Equal(t, tt.wantViolCount, len(violations), "Check() got %d violations, want %d", len(violations), tt.wantViolCount)
			for _, v := range violations {
				t.Logf("  - %s: %s", v.Path, v.Message)
			}
		})
	}
}

func Test_isCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"myFile", true},
		{"myLongFileName", true},
		{"a", true},
		{"MyFile", false},
		{"my-file", false},
		{"my_file", false},
		{"my file", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, isCamelCase(tt.input), "isCamelCase(%q)", tt.input)
		})
	}
}

func Test_isPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"MyFile", true},
		{"MyLongFileName", true},
		{"A", true},
		{"myFile", false},
		{"My-File", false},
		{"My_File", false},
		{"My File", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, isPascalCase(tt.input), "isPascalCase(%q)", tt.input)
		})
	}
}

func Test_isKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"my-file", true},
		{"my-long-file-name", true},
		{"file", true},
		{"MyFile", false},
		{"my_file", false},
		{"my file", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, isKebabCase(tt.input), "isKebabCase(%q)", tt.input)
		})
	}
}

func Test_isSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"my_file", true},
		{"my_long_file_name", true},
		{"file", true},
		{"MyFile", false},
		{"my-file", false},
		{"my file", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, isSnakeCase(tt.input), "isSnakeCase(%q)", tt.input)
		})
	}
}

func TestNamingConventionRule_Name(t *testing.T) {
	rule := NewNamingConventionRule(map[string]string{"*.js": "camelCase"})
	assert.Equal(t, "naming-convention", rule.Name())
}

func TestIsUpperCase(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"ABC", true},
		{"abc", false},
		{"Abc", false},
		{"ABC123", true},
		{"", true},
	}
	for _, c := range cases {
		got := isUpperCase(c.input)
		if got != c.want {
			t.Errorf("isUpperCase(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
