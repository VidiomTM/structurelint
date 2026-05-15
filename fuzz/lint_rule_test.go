package fuzz

import (
	"encoding/json"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/output"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func FuzzLintRule(f *testing.F) {
	seeds := []string{
		"max-depth",
		"naming-convention",
		"file-existence",
		"disallowed-patterns",
		"",
		"unknown-rule",
		"a",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, ruleName string) {
		v := rules.Violation{Rule: ruleName, Path: "test.txt", Message: "test"}
		violations := []rules.Violation{v}

		textF := &output.TextFormatter{}
		_, _ = textF.Format(violations)

		jsonF := &output.JSONFormatter{Version: "fuzz"}
		result, err := jsonF.Format(violations)
		if err == nil {
			var parsed output.JSONOutput
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("invalid json output: %v", err)
			}
		}
	})
}
