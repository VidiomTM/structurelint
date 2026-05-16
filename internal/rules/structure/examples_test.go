package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestExampleTestingStrategyRule_Name(t *testing.T) {
	rule := ExampleTestingStrategyRule()
	if rule.Name() != "testing-strategy" {
		t.Fatalf("Name() = %q", rule.Name())
	}
}

func TestExampleDocumentationCompletenessRule_Name(t *testing.T) {
	rule := ExampleDocumentationCompletenessRule()
	if rule.Name() != "documentation-completeness" {
		t.Fatalf("Name() = %q", rule.Name())
	}
}

func TestExampleArchitectureConsistencyRule_Name(t *testing.T) {
	rule := ExampleArchitectureConsistencyRule()
	if rule.Name() != "architecture-consistency" {
		t.Fatalf("Name() = %q", rule.Name())
	}
}

func TestExampleDocumentationCompletenessRule_AlwaysViolates(t *testing.T) {
	rule := ExampleDocumentationCompletenessRule()
	v := rule.Check(nil, map[string]*walker.DirInfo{"": {}})
	// Example functions use description strings instead of valid "exists:N" format
	if len(v) == 0 {
		t.Error("expected violations (example uses broken requirement format)")
	}
}


