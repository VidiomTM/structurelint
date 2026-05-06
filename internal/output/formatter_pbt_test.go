package output

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func arbViolation(t *rapid.T) rules.Violation {
	return rules.Violation{
		Rule:    rapid.StringMatching(`[a-z][a-z-]*[a-z]`).Draw(t, "rule"),
		Path:    rapid.StringMatching(`[a-z]+/[a-z]+\.go`).Draw(t, "path"),
		Message: rapid.StringOfN(rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.,:;\"\\<>&\t\r")), 1, 80, -1).Draw(t, "message"),
	}
}

func arbViolations(t *rapid.T) []rules.Violation {
	n := rapid.IntRange(0, 20).Draw(t, "n")
	violations := make([]rules.Violation, n)
	for i := range violations {
		violations[i] = arbViolation(t)
	}
	return violations
}

func TestJSONOutputValidity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		f := &JSONFormatter{Version: "test"}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("json format error: %v", err)
		}

		var result JSONOutput
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid json: %v", err)
		}

		if len(result.Results) != len(violations) {
			t.Fatalf("result count mismatch: got %d, want %d", len(result.Results), len(violations))
		}
	})
}

func TestJSONOutputRequiredFields(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		f := &JSONFormatter{Version: "test"}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("json format error: %v", err)
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(out), &raw); err != nil {
			t.Fatalf("invalid json: %v", err)
		}

		for _, field := range []string{"version", "timestamp", "violations", "results"} {
			if _, ok := raw[field]; !ok {
				t.Fatalf("missing required field: %s", field)
			}
		}

		if raw["version"] != "test" {
			t.Fatalf("version mismatch: got %v, want test", raw["version"])
		}

		violationsCount, ok := raw["violations"].(float64)
		if !ok {
			t.Fatalf("violations is not a number")
		}
		if int(violationsCount) != len(violations) {
			t.Fatalf("violations count mismatch: got %d, want %d", int(violationsCount), len(violations))
		}
	})
}

func TestJSONOutputResultFields(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		f := &JSONFormatter{Version: "test"}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("json format error: %v", err)
		}

		var result JSONOutput
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid json: %v", err)
		}

		for i, r := range result.Results {
			if r.Rule == "" {
				t.Fatalf("result[%d]: empty rule", i)
			}
			if r.Path == "" {
				t.Fatalf("result[%d]: empty path", i)
			}
			if r.Message == "" {
				t.Fatalf("result[%d]: empty message", i)
			}
		}
	})
}

func TestTextFormatLineCount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		f := &TextFormatter{}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("text format error: %v", err)
		}

		if len(violations) == 0 {
			if out != "" {
				t.Fatalf("expected empty output for no violations, got: %q", out)
			}
			return
		}

		lines := strings.Count(out, "\n")
		if lines != len(violations) {
			t.Fatalf("line count mismatch: got %d, want %d", lines, len(violations))
		}
	})
}

func TestTextFormatContainsPathAndMessage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)
		if len(violations) == 0 {
			return
		}

		f := &TextFormatter{}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("text format error: %v", err)
		}

		for _, v := range violations {
			if !strings.Contains(out, v.Path) {
				t.Fatalf("output missing path %q", v.Path)
			}
			if !strings.Contains(out, v.Message) {
				t.Fatalf("output missing message %q", v.Message)
			}
		}
	})
}

func TestCountMatchAcrossFormatters(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		jf := &JSONFormatter{Version: "test"}
		jsonOut, err := jf.Format(violations)
		if err != nil {
			t.Fatalf("json format error: %v", err)
		}
		var jsonResult JSONOutput
		if err := json.Unmarshal([]byte(jsonOut), &jsonResult); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if jsonResult.Violations != len(violations) {
			t.Fatalf("json count mismatch: got %d, want %d", jsonResult.Violations, len(violations))
		}

		juf := &JUnitFormatter{}
		junitOut, err := juf.Format(violations)
		if err != nil {
			t.Fatalf("junit format error: %v", err)
		}
		var junitResult JUnitTestSuites
		if err := xml.Unmarshal([]byte(junitOut), &junitResult); err != nil {
			t.Fatalf("invalid junit xml: %v", err)
		}
		if junitResult.Failures != len(violations) {
			t.Fatalf("junit failures mismatch: got %d, want %d", junitResult.Failures, len(violations))
		}
	})
}

func TestJUnitXMLOutputValidity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		f := &JUnitFormatter{}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("junit format error: %v", err)
		}

		if !strings.HasPrefix(out, xml.Header) {
			t.Fatalf("junit output missing xml header")
		}

		var result JUnitTestSuites
		if err := xml.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid junit xml: %v", err)
		}

		if result.Name != "structurelint" {
			t.Fatalf("junit name mismatch: got %q, want %q", result.Name, "structurelint")
		}
	})
}

func TestJUnitTestCaseCount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		violations := arbViolations(t)

		f := &JUnitFormatter{}
		out, err := f.Format(violations)
		if err != nil {
			t.Fatalf("junit format error: %v", err)
		}

		var result JUnitTestSuites
		if err := xml.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("invalid junit xml: %v", err)
		}

		totalCases := 0
		for _, ts := range result.TestSuites {
			totalCases += len(ts.TestCases)
		}

		expected := len(violations)
		if expected == 0 {
			expected = 1
		}
		if totalCases != expected {
			t.Fatalf("test case count mismatch: got %d, want %d", totalCases, expected)
		}
	})
}
