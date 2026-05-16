package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestNamingConventionRule_matchesPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"src/components/Button.tsx", "src/components/**/", true},
		{"src/components/index.ts", "src/components/**/", true},
		{"other/file.ts", "src/components/**/", false},
		{"app.test.ts", "*.test.ts", true},
	}
	for _, c := range tests {
		got := matchesPattern(c.path, c.pattern)
		if got != c.want {
			t.Errorf("matchesPattern(%q, %q) = %v, want %v", c.path, c.pattern, got, c.want)
		}
	}
}

func TestNamingConventionRule_DetectConvention(t *testing.T) {
	r := NewNamingConventionRule(nil)
	tests := []struct {
		name string
		want string
	}{
		{"myFile", "camelCase"},
		{"MyFile", "PascalCase"},
		{"my_file", "snake_case"},
		{"MY_FILE", "UPPERCASE"},
	}
	for _, c := range tests {
		got := r.detectConvention(c.name)
		if got != c.want {
			t.Errorf("detectConvention(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestNamingConventionRule_ConvertToConvention(t *testing.T) {
	r := NewNamingConventionRule(nil)
	tests := []struct {
		convention, name, want string
	}{
		{"camelCase", "my_file", "myFile"},
		{"PascalCase", "my_file", "MyFile"},
		{"snake_case", "MyFile", "my_file"},
		{"PascalCase", "myFile", "MyFile"},
	}
	for _, c := range tests {
		got := r.convertToConvention(c.name, c.convention) // (name, convention)
		if got != c.want {
			t.Errorf("convertToConvention(%q, %q) = %q, want %q", c.name, c.convention, got, c.want)
		}
	}
}

func TestIsFrameworkConventionFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"src/routes/+page.svelte", true},
		{"src/routes/+layout.svelte", true},
		{"src/routes/+server.ts", true},
		{"src/routes/+page.ts", true},
		{"src/routes/+error.svelte", true},
		{"src/routes/+page.server.ts", true},
		{"src/routes/page.svelte", false},
		{"src/components/Button.svelte", false},
		{"README.md", false},
	}
	for _, c := range tests {
		got := isFrameworkConventionFile(c.path)
		if got != c.want {
			t.Errorf("isFrameworkConventionFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestFileExistenceRule_Check_RequirementVariants(t *testing.T) {
	tests := []struct {
		name    string
		req     map[string]string
		files   []walker.FileInfo
		dirs    map[string]*walker.DirInfo
		wantDev int
	}{
		{
			name:    "empty dirs",
			req:     map[string]string{"README.md": "exists:1"},
			files:   nil,
			dirs:    nil,
			wantDev: 0,
		},
		{
			name:    "exists:0 (disallowed file exists)",
			req:     map[string]string{"secret.env": "exists:0"},
			files:   []walker.FileInfo{{Path: "secret.env", ParentPath: "", IsDir: false}},
			dirs:    map[string]*walker.DirInfo{"": {}},
			wantDev: 1,
		},
		{
			name:    "exists:0 (disallowed file absent)",
			req:     map[string]string{"secret.env": "exists:0"},
			files:   nil,
			dirs:    map[string]*walker.DirInfo{"": {}},
			wantDev: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewFileExistenceRule(tt.req)
			v := r.Check(tt.files, tt.dirs)
			if len(v) != tt.wantDev {
				t.Errorf("got %d violations, want %d: %v", len(v), tt.wantDev, v)
			}
		})
	}
}

func TestFileExistenceRule_SubdirMatching(t *testing.T) {
	r := NewFileExistenceRule(map[string]string{
		"*.go": "exists:1",
	})
	files := []walker.FileInfo{
		{Path: "sub/main.go", ParentPath: "sub", IsDir: false},
	}
	dirs := map[string]*walker.DirInfo{
		"sub": {},
	}
	v := r.Check(files, dirs)
	if len(v) != 0 {
		t.Errorf("expected 0 violations, got %v", v)
	}
}

func TestMaxFilesRule_Check_EdgeCases(t *testing.T) {
	t.Run("no files", func(t *testing.T) {
		r := NewMaxFilesRule(5)
		v := r.Check(nil, nil)
		if len(v) != 0 {
			t.Errorf("expected 0 violations, got %d", len(v))
		}
	})
}

func TestUniquenessConstraints_Check_EdgeCases(t *testing.T) {
	r := NewUniquenessConstraintsRule(map[string]string{})
	v := r.Check(nil, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations, got %d", len(v))
	}
}

func TestMatchesUniquenessPattern(t *testing.T) {
	tests := []struct {
		file, pattern string
		want          bool
	}{
		{"main.go", "*.go", true},
		{"main.go", "*.ts", false},
		{"user_service.py", "*_service.py", true},
		{"main.go", "*.go", true},
	}
	for _, c := range tests {
		got := matchesUniquenessPattern(c.file, c.pattern)
		if got != c.want {
			t.Errorf("matchesUniquenessPattern(%q, %q) = %v, want %v", c.file, c.pattern, got, c.want)
		}
	}
}
