package fuzz

import (
	"encoding/json"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/output"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func FuzzJSONFormatter(f *testing.F) {
	f.Add("test-rule", "/path/to/file.go", "violation message", "1.0.0")
	f.Add("", "", "", "")
	f.Add("rule", "path", "msg with \"quotes\" and \n newlines", "2.0.0")
	f.Fuzz(func(t *testing.T, rule, path, message, version string) {
		violations := []rules.Violation{
			{
				Rule:    rule,
				Path:    path,
				Message: message,
			},
		}

		formatter := &output.JSONFormatter{Version: version}
		result, err := formatter.Format(violations)
		if err != nil {
			t.Fatalf("JSONFormatter.Format returned error: %v", err)
		}

		if !json.Valid([]byte(result)) {
			t.Error("JSONFormatter produced invalid JSON")
		}

		var parsed output.JSONOutput
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON output: %v", err)
		}

		if parsed.Violations != 1 {
			t.Errorf("violations count: got %d, want 1", parsed.Violations)
		}
	})
}

func FuzzTextFormatter(f *testing.F) {
	f.Add("rule", "/path", "message")
	f.Add("", "", "")
	f.Add("r", "p", "msg with \x00 null bytes")
	f.Fuzz(func(t *testing.T, rule, path, message string) {
		violations := []rules.Violation{
			{
				Rule:    rule,
				Path:    path,
				Message: message,
			},
		}

		formatter := &output.TextFormatter{}
		result, err := formatter.Format(violations)
		if err != nil {
			t.Fatalf("TextFormatter.Format returned error: %v", err)
		}

		if len(violations) > 0 && result == "" {
			t.Error("non-empty violations produced empty text output")
		}
	})
}

func FuzzJUnitFormatter(f *testing.F) {
	f.Add("rule-a", "/file1.go", "msg1", "rule-b", "/file2.go", "msg2")
	f.Add("", "", "", "", "", "")
	f.Fuzz(func(t *testing.T, rule1, path1, msg1, rule2, path2, msg2 string) {
		violations := []rules.Violation{
			{Rule: rule1, Path: path1, Message: msg1},
			{Rule: rule2, Path: path2, Message: msg2},
		}

		formatter := &output.JUnitFormatter{}
		result, err := formatter.Format(violations)
		if err != nil {
			t.Fatalf("JUnitFormatter.Format returned error: %v", err)
		}

		if len(result) == 0 && len(violations) > 0 {
			t.Error("non-empty violations produced empty JUnit output")
		}
	})
}

func FuzzJSONFormatterEmptyViolations(f *testing.F) {
	f.Add("1.0.0")
	f.Add("")
	f.Add("v2.0.0-beta.1+build")
	f.Fuzz(func(t *testing.T, version string) {
		formatter := &output.JSONFormatter{Version: version}
		result, err := formatter.Format(nil)
		if err != nil {
			t.Fatalf("JSONFormatter.Format(nil) returned error: %v", err)
		}

		var parsed output.JSONOutput
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if parsed.Violations != 0 {
			t.Errorf("expected 0 violations, got %d", parsed.Violations)
		}

		if len(parsed.Results) != 0 {
			t.Errorf("expected 0 results, got %d", len(parsed.Results))
		}
	})
}

func FuzzJSONFormatterManyViolations(f *testing.F) {
	f.Add(10)
	f.Add(0)
	f.Add(1)
	f.Add(50)
	f.Fuzz(func(t *testing.T, count int) {
		if count < 0 {
			count = 0
		}
		if count > 500 {
			count = 500
		}

		violations := make([]rules.Violation, count)
		for i := range violations {
			violations[i] = rules.Violation{
				Rule:    "test-rule",
				Path:    "/path/to/file.go",
				Message: "test message",
			}
		}

		formatter := &output.JSONFormatter{Version: "1.0.0"}
		result, err := formatter.Format(violations)
		if err != nil {
			t.Fatalf("JSONFormatter.Format returned error: %v", err)
		}

		if !json.Valid([]byte(result)) {
			t.Error("produced invalid JSON")
		}
	})
}
