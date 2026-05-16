package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestFileExistenceRule_WhenChecking(t *testing.T) {
	tests := []struct {
		name          string
		requirements  map[string]string
		files         []walker.FileInfo
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:         "GivenMissingFile_WhenChecking_ThenReportsViolation",
			requirements: map[string]string{"README.md": "exists:1"},
			files:        []walker.FileInfo{},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 1,
		},
		{
			name:         "GivenExistingFile_WhenChecking_ThenNoViolation",
			requirements: map[string]string{"README.md": "exists:1"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: "", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 0,
		},
		{
			name:         "GivenMultipleDirs_WhenAllHaveFile_ThenNoViolation",
			requirements: map[string]string{"README.md": "exists:1"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: "", IsDir: false},
				{Path: "src/README.md", ParentPath: "src", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"":    {},
				"src": {},
			},
			wantViolCount: 0,
		},
		{
			name:         "GivenMultipleDirs_WhenOneMissing_ThenReportsViolation",
			requirements: map[string]string{"README.md": "exists:1"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: "", IsDir: false},
			},
			dirs: map[string]*walker.DirInfo{
				"":    {},
				"src": {},
			},
			wantViolCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewFileExistenceRule(tt.requirements)
			violations := rule.Check(tt.files, tt.dirs)

			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestFileExistenceRule_WhenParsingCountSpec(t *testing.T) {
	tests := []struct {
		spec    string
		wantMin int
		wantMax int
		wantErr bool
	}{
		{"1", 1, 1, false},
		{"0", 0, 0, false},
		{"2-5", 2, 5, false},
		{"invalid", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			rule := &FileExistenceRule{}
			min, max, err := rule.parseCountSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCountSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
				return
			}
			if err == nil {
				if min != tt.wantMin {
					t.Errorf("parseCountSpec(%q) min = %d, want %d", tt.spec, min, tt.wantMin)
				}
				if max != tt.wantMax {
					t.Errorf("parseCountSpec(%q) max = %d, want %d", tt.spec, max, tt.wantMax)
				}
			}
		})
	}
}

func TestFileExistenceRule_WhenGettingName(t *testing.T) {
	rule := NewFileExistenceRule(map[string]string{"README.md": "exists:1"})
	got := rule.Name()
	if got != "file-existence" {
		t.Errorf("Name() = %v, want file-existence", got)
	}
}

func TestValidateFileExistenceConfig_Valid(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "exists:1"})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_ValidRange(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "exists:1-5"})
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_InvalidFormat(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "optional"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_InvalidRange(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "exists:1-2-3"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for invalid range, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_InvalidRangeMin(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "exists:abc-5"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for invalid range min, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_InvalidRangeMax(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "exists:1-xyz"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for invalid range max, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_InvalidCount(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{"README.md": "exists:abc"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for invalid count, got %v", errs)
	}
}

func TestValidateFileExistenceConfig_MultipleErrors(t *testing.T) {
	errs := ValidateFileExistenceConfig(map[string]string{
		"a": "bad",
		"b": "exists:abc",
	})
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %v", errs)
	}
}
