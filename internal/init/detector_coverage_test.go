package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestDetectLanguages(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "main.go"},
		{Path: "main_test.go"},
		{Path: "lib.py"},
		{Path: "test_lib.py"},
		{Path: "README.md"},
	}
	result := detectLanguages(files)
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	found := false
	for _, l := range result {
		if l.Language == "go" {
			found = true
			if l.TestPattern == "" {
				t.Error("go should have a test pattern")
			}
			break
		}
	}
	if !found {
		t.Error("expected go to be detected")
	}
}
func TestDetectLanguages_EmptyFiles(t *testing.T) {
	result := detectLanguages(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestDetectLanguages_OnlyDirs(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src", IsDir: true},
	}
	result := detectLanguages(files)
	if len(result) != 0 {
		t.Errorf("expected 0 languages for only dirs, got %d", len(result))
	}
}

func TestSortLanguages(t *testing.T) {
	result := sortLanguages(map[string]*LanguageInfo{
		"python":     {Language: "python", FileCount: 10},
		"go":         {Language: "go", FileCount: 50},
		"javascript": {Language: "javascript", FileCount: 20},
	})
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0].Language != "go" {
		t.Errorf("expected go first, got %s", result[0].Language)
	}
	if result[1].Language != "javascript" {
		t.Errorf("expected javascript second, got %s", result[1].Language)
	}
	if result[2].Language != "python" {
		t.Errorf("expected python third, got %s", result[2].Language)
	}
}

func TestSortLanguages_Empty(t *testing.T) {
	result := sortLanguages(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestAnalyzeTestFile_Adjacent(t *testing.T) {
	info := &LanguageInfo{Language: "go"}
	files := []walker.FileInfo{
		{Path: "main.go"},
		{Path: "main_test.go"},
	}
	analyzeTestFile(files[1], files, "go", info)
	if info.TestPattern != "adjacent" {
		t.Errorf("TestPattern = %q, want adjacent", info.TestPattern)
	}
}

func TestAnalyzeTestFile_Separate(t *testing.T) {
	info := &LanguageInfo{Language: "go"}
	files := []walker.FileInfo{
		{Path: "tests/service_test.go"},
	}
	analyzeTestFile(files[0], files, "go", info)
	if info.TestPattern != "separate" {
		t.Errorf("TestPattern = %q, want separate", info.TestPattern)
	}
	if info.TestDir != "tests" {
		t.Errorf("TestDir = %q, want tests", info.TestDir)
	}
}

func TestAnalyzeTestFile_NotTestFile(t *testing.T) {
	info := &LanguageInfo{Language: "go"}
	files := []walker.FileInfo{
		{Path: "main.go"},
	}
	analyzeTestFile(files[0], files, "go", info)
	if info.TestPattern != "" {
		t.Errorf("expected no test pattern for non-test file, got %q", info.TestPattern)
	}
}

func TestAnalyzeTestFile_IntegrationDir(t *testing.T) {
	info := &LanguageInfo{Language: "go"}
	files := []walker.FileInfo{
		{Path: "tests/integration/api_test.go"},
	}
	analyzeTestFile(files[0], files, "go", info)
	if !info.HasIntegrationDir {
		t.Error("expected HasIntegrationDir")
	}
}

func TestIsAdjacentTest(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "main.go"},
		{Path: "main_test.go"},
	}
	if !isAdjacentTest("main_test.go", files, "go") {
		t.Error("expected main_test.go to be adjacent test")
	}
	if isAdjacentTest("no_source_test.go", files, "go") {
		t.Error("expected no_source_test.go to NOT be adjacent")
	}
}

func TestIsAdjacentTest_InvalidLang(t *testing.T) {
	if isAdjacentTest("test.py", nil, "") {
		t.Error("expected false for empty lang")
	}
}

func TestTestToSourceFilename(t *testing.T) {
	tests := []struct {
		testFile, lang, want string
	}{
		{"main_test.go", "go", "main.go"},
		{"utils.test.ts", "typescript", "utils.ts"},
		{"component.spec.js", "javascript", "component.js"},
		{"test_module.py", "python", "module.py"},
		{"nopattern.xyz", "go", ""},
	}
	for _, c := range tests {
		got := testToSourceFilename(c.testFile, c.lang)
		if got != c.want {
			t.Errorf("testToSourceFilename(%q, %q) = %q, want %q", c.testFile, c.lang, got, c.want)
		}
	}
}

func TestFindTestDirectory(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"tests/service_test.go", "tests"},
		{"test/service_test.go", "test"},
		{"__tests__/service.test.ts", "__tests__"},
		{"spec/service_spec.rb", "spec"},
		{"src/main.go", ""},
		{"src/test", "test"},
	}
	for _, c := range tests {
		got := findTestDirectory(c.path)
		if got != c.want {
			t.Errorf("findTestDirectory(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestIsIntegrationTestDir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"tests/integration/api_test.go", true},
		{"e2e/flow_test.go", true},
		{"functional/test_suite.go", true},
		{"unit/math_test.go", false},
		{"src/main.go", false},
	}
	for _, c := range tests {
		got := isIntegrationTestDir(c.path)
		if got != c.want {
			t.Errorf("isIntegrationTestDir(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestExtractIntegrationDir(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"tests/integration/api_test.go", "tests/integration"},
		{"e2e/flow_test.go", "e2e"},
		{"src/main.go", ""},
	}
	for _, c := range tests {
		got := extractIntegrationDir(c.path)
		if got != c.want {
			t.Errorf("extractIntegrationDir(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestCalculateMaxSubdirs(t *testing.T) {
	tests := []struct {
		name string
		dirs map[string]*walker.DirInfo
		want int
	}{
		{"few subdirs", map[string]*walker.DirInfo{
			"src": {SubdirCount: 2},
		}, 10},
		{"many subdirs", map[string]*walker.DirInfo{
			"src": {SubdirCount: 25},
		}, 28},
		{"above cap", map[string]*walker.DirInfo{
			"src": {SubdirCount: 50},
		}, 30},
		{"nil dirs", nil, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMaxSubdirs(tt.dirs)
			if got != tt.want {
				t.Errorf("calculateMaxSubdirs() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDetectProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# readme"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}
	if info == nil {
		t.Fatal("expected non-nil ProjectInfo")
	}
	if info.PrimaryLanguage == nil || info.PrimaryLanguage.Language != "go" {
		t.Errorf("PrimaryLanguage = %v, want go", info.PrimaryLanguage)
	}
	if info.MaxDepth < 4 {
		t.Errorf("MaxDepth = %d, want >= 4", info.MaxDepth)
	}
	if info.MaxFilesInDir < 20 {
		t.Errorf("MaxFilesInDir = %d, want >= 20", info.MaxFilesInDir)
	}
	if info.MaxSubdirs < 10 {
		t.Errorf("MaxSubdirs = %d, want >= 10", info.MaxSubdirs)
	}
	if info.DocumentationStyle != "comprehensive" {
		t.Errorf("DocumentationStyle = %q, want comprehensive", info.DocumentationStyle)
	}
}

func TestDetectProject_NonExistentDir(t *testing.T) {
	_, err := DetectProject("/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent dir")
	}
}
