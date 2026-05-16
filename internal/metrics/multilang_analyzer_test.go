package metrics

import (
	"os"
	"testing"
)

func TestDetectLanguageFromPath(t *testing.T) {
	// Arrange
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()
	tests := []struct {
		filePath string
		expected string
	}{
		{"script.py", "python"},
		{"app.js", "javascript"},
		{"component.jsx", "javascript"},
		{"app.ts", "typescript"},
		{"component.tsx", "typescript"},
		{"Main.java", "java"},
		{"example.cpp", "cpp"},
		{"example.cc", "cpp"},
		{"example.cs", "csharp"},
		{"/path/to/script.PY", "python"}, // Case insensitive
		{"/path/to/app.JS", "javascript"},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			// Act
			result, err := analyzer.detectLanguageFromPath(tt.filePath)

			// Assert
			if err != nil {
				t.Errorf("detectLanguageFromPath(%q) returned unexpected error: %v", tt.filePath, err)
			}
			if result != tt.expected {
				t.Errorf("detectLanguageFromPath(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestDetectLanguageFromPath_Unsupported(t *testing.T) {
	// Arrange
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()
	tests := []string{
		"main.go",
		"README.md",
		"data.json",
	}

	for _, filePath := range tests {
		t.Run(filePath, func(t *testing.T) {
			// Act
			_, err := analyzer.detectLanguageFromPath(filePath)

			// Assert
			if err == nil {
				t.Errorf("detectLanguageFromPath(%q) expected error for unsupported extension, got nil", filePath)
			}
		})
	}
}

func TestNewMultiLanguageCognitiveComplexityAnalyzer(t *testing.T) {
	// Act
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	// Assert
	if analyzer == nil {
		t.Fatal("NewMultiLanguageCognitiveComplexityAnalyzer() returned nil")
	}

	if analyzer.metricType != MetricCognitiveComplexity {
		t.Errorf("metricType = %q, want %q", analyzer.metricType, MetricCognitiveComplexity)
	}

	if analyzer.Name() != MetricCognitiveComplexity {
		t.Errorf("Name() = %q, want %q", analyzer.Name(), MetricCognitiveComplexity)
	}
}

func TestNewMultiLanguageHalsteadAnalyzer(t *testing.T) {
	// Act
	analyzer := NewMultiLanguageHalsteadAnalyzer()

	// Assert
	if analyzer == nil {
		t.Fatal("NewMultiLanguageHalsteadAnalyzer() returned nil")
	}

	if analyzer.metricType != MetricHalstead {
		t.Errorf("metricType = %q, want %q", analyzer.metricType, MetricHalstead)
	}

	if analyzer.Name() != MetricHalstead {
		t.Errorf("Name() = %q, want %q", analyzer.Name(), MetricHalstead)
	}
}

func TestAnalyzeFileByPath_UnsupportedLanguage(t *testing.T) {
	// Arrange
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	tmpFile, err := os.CreateTemp("", "test*.go")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	// Act
	_, err = analyzer.AnalyzeFileByPath(tmpFile.Name())

	// Assert
	if err == nil {
		t.Error("AnalyzeFileByPath() with .go file should return error, got nil")
	}

	expectedErrMsg := "unsupported"
	if err != nil && !containsString(err.Error(), expectedErrMsg) {
		t.Errorf("AnalyzeFileByPath() error = %q, want error containing %q", err.Error(), expectedErrMsg)
	}
}

func TestAnalyzePythonFile_ValidPython(t *testing.T) {
	// Arrange
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	tmpFile, err := os.CreateTemp("", "test*.py")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	pythonCode := `def simple_function(x):
    return x + 1

def complex_function(x):
    if x > 0:
        if x > 10:
            return x * 2
    return x
`
	if _, err := tmpFile.WriteString(pythonCode); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Act
	metrics, err := analyzer.AnalyzeFileByPath(tmpFile.Name())

	// Assert
	if err != nil {
		t.Fatalf("AnalyzeFileByPath() error = %v", err)
	}

	if metrics.FilePath != tmpFile.Name() {
		t.Errorf("FilePath = %q, want %q", metrics.FilePath, tmpFile.Name())
	}

	// New implementation provides file-level metrics
	if metrics.FileLevel == nil {
		t.Fatal("Expected FileLevel metrics, got nil")
	}

	if _, ok := metrics.FileLevel["cognitive_complexity"]; !ok {
		t.Error("Expected cognitive_complexity in FileLevel metrics")
	}
}

func TestAnalyzeJavaScriptFile_ValidJS(t *testing.T) {
	// Arrange
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()

	tmpFile, err := os.CreateTemp("", "test*.js")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	jsCode := `function simpleFunction(x) {
    return x + 1;
}

function complexFunction(x) {
    if (x > 0) {
        if (x > 10) {
            return x * 2;
        }
    }
    return x;
}
`
	if _, err := tmpFile.WriteString(jsCode); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Act
	metrics, err := analyzer.AnalyzeFileByPath(tmpFile.Name())

	// Assert
	if err != nil {
		t.Fatalf("AnalyzeFileByPath() error = %v", err)
	}

	if metrics.FilePath != tmpFile.Name() {
		t.Errorf("FilePath = %q, want %q", metrics.FilePath, tmpFile.Name())
	}

	// New implementation provides file-level metrics
	if metrics.FileLevel == nil {
		t.Fatal("Expected FileLevel metrics, got nil")
	}

	if _, ok := metrics.FileLevel["cognitive_complexity"]; !ok {
		t.Error("Expected cognitive_complexity in FileLevel metrics")
	}
}

func TestAnalyzeHalsteadMetrics(t *testing.T) {
	// Arrange
	analyzer := NewMultiLanguageHalsteadAnalyzer()

	tmpFile, err := os.CreateTemp("", "test*.py")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	pythonCode := `def add(a, b):
    return a + b
`
	if _, err := tmpFile.WriteString(pythonCode); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Act
	metrics, err := analyzer.AnalyzeFileByPath(tmpFile.Name())

	// Assert
	if err != nil {
		t.Fatalf("AnalyzeFileByPath() error = %v", err)
	}

	if metrics.FileLevel == nil {
		t.Fatal("Expected FileLevel metrics, got nil")
	}

	expectedMetrics := []string{"halstead_effort", "halstead_volume", "halstead_difficulty"}
	for _, metric := range expectedMetrics {
		if _, ok := metrics.FileLevel[metric]; !ok {
			t.Errorf("Expected %s in FileLevel metrics", metric)
		}
	}
}

func TestConvertToTreeSitterLanguage(t *testing.T) {
	analyzer := NewMultiLanguageCognitiveComplexityAnalyzer()
	tests := []struct {
		lang     string
		want     string
		wantErr  bool
		errMatch string
	}{
		{"python", "python", false, ""},
		{"javascript", "javascript", false, ""},
		{"typescript", "typescript", false, ""},
		{"java", "java", false, ""},
		{"cpp", "cpp", false, ""},
		{"csharp", "csharp", false, ""},
		{"ruby", "", true, "unsupported language"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got, err := analyzer.convertToTreeSitterLanguage(tt.lang)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errMatch != "" && !containsString(err.Error(), tt.errMatch) {
					t.Errorf("Error = %q, want containing %q", err.Error(), tt.errMatch)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("convertToTreeSitterLanguage(%q) = %q, want %q", tt.lang, string(got), tt.want)
			}
		})
	}
}

// Helper functions for tests

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
