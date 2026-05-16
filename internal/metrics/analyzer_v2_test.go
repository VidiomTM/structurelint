package metrics

import (
	"testing"
)

func TestCognitiveComplexityAnalyzerV2_Constructor_And_Name(t *testing.T) {
	a := NewCognitiveComplexityAnalyzerV2()
	if a == nil {
		t.Fatal("Expected non-nil analyzer")
	}
	if a.Name() != MetricCognitiveComplexity {
		t.Errorf("Name() = %q, want %q", a.Name(), MetricCognitiveComplexity)
	}
}

func TestHalsteadAnalyzerV2_Constructor_And_Name(t *testing.T) {
	a := NewHalsteadAnalyzerV2()
	if a == nil {
		t.Fatal("Expected non-nil analyzer")
	}
	if a.Name() != MetricHalstead {
		t.Errorf("Name() = %q, want %q", a.Name(), MetricHalstead)
	}
}

func TestAnalyzerV2_AnalyzeUnsupportedExtension(t *testing.T) {
	a := NewHalsteadAnalyzerV2()
	metrics, err := a.AnalyzeFileByPath("test.xyz")
	if err != nil {
		t.Fatalf("Expected no error for unsupported extension, got %v", err)
	}
	if metrics.FilePath != "test.xyz" {
		t.Errorf("Expected empty metrics, got filepath %q", metrics.FilePath)
	}
}

func TestCognitiveComplexityAnalyzerV2_AnalyzeUnsupportedExtension(t *testing.T) {
	a := NewCognitiveComplexityAnalyzerV2()
	metrics, err := a.AnalyzeFileByPath("test.xyz")
	if err != nil {
		t.Fatalf("Expected no error for unsupported extension, got %v", err)
	}
	if metrics.FilePath != "test.xyz" {
		t.Errorf("Expected empty metrics, got filepath %q", metrics.FilePath)
	}
}
