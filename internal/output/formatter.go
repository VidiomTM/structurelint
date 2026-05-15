// @structurelint:ignore test-adjacency Output formatters are integration-tested through CLI
package output

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

// Formatter defines the interface for output formatters
type Formatter interface {
	Format(violations []rules.Violation) (string, error)
}

// TextFormatter formats violations as plain text
type TextFormatter struct{}

// Format formats violations as plain text
func (f *TextFormatter) Format(violations []rules.Violation) (string, error) {
	if len(violations) == 0 {
		return "", nil
	}

	var sb strings.Builder
	for _, v := range violations {
		fmt.Fprintf(&sb, "%s: %s\n", v.Path, v.Message)
	}
	return sb.String(), nil
}

// JSONFormatter formats violations as JSON
type JSONFormatter struct {
	Version string
}

// JSONOutput represents the JSON output structure
type JSONOutput struct {
	Version    string          `json:"version"`
	Timestamp  string          `json:"timestamp"`
	Violations int             `json:"violations"`
	Results    []JSONViolation `json:"results"`
}

// JSONViolation represents a single violation in JSON format
type JSONViolation struct {
	Rule    string `json:"rule"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// Format formats violations as JSON
func (f *JSONFormatter) Format(violations []rules.Violation) (string, error) {
	output := JSONOutput{
		Version:    f.Version,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Violations: len(violations),
		Results:    make([]JSONViolation, 0, len(violations)),
	}

	for _, v := range violations {
		output.Results = append(output.Results, JSONViolation{
			Rule:    v.Rule,
			Path:    v.Path,
			Message: v.Message,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// JUnitFormatter formats violations as JUnit XML
type JUnitFormatter struct{}

// JUnitTestSuites represents the root element of JUnit XML
type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	XMLNS      string           `xml:"xmlns,attr,omitempty"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Time       string           `xml:"time,attr"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite represents a test suite
type JUnitTestSuite struct {
	Name      string            `xml:"name,attr"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      string            `xml:"time,attr"`
	Timestamp string            `xml:"timestamp,attr"`
	TestCases []JUnitTestCase   `xml:"testcase"`
}

// JUnitTestCase represents a test case
type JUnitTestCase struct {
	Name      string         `xml:"name,attr"`
	Classname string         `xml:"classname,attr"`
	Time      string         `xml:"time,attr"`
	Failure   *JUnitFailure  `xml:"failure,omitempty"`
}

// JUnitFailure represents a test failure
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// Format formats violations as JUnit XML
func (f *JUnitFormatter) Format(violations []rules.Violation) (string, error) {
	// Capture timestamp once for all test suites
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Group violations by rule
	ruleViolations := make(map[string][]rules.Violation)
	for _, v := range violations {
		ruleViolations[v.Rule] = append(ruleViolations[v.Rule], v)
	}

	// Sort rule names for deterministic output
	var ruleNames []string
	for ruleName := range ruleViolations {
		ruleNames = append(ruleNames, ruleName)
	}
	sort.Strings(ruleNames)

	var testSuites []JUnitTestSuite
	totalTests := 0
	totalFailures := len(violations)

	// Create a test suite for each rule in sorted order
	for _, ruleName := range ruleNames {
		ruleViols := ruleViolations[ruleName]
		var testCases []JUnitTestCase

		for _, v := range ruleViols {
			testCases = append(testCases, JUnitTestCase{
				Name:      v.Path,
				Classname: ruleName,
				Time:      "0",
				Failure: &JUnitFailure{
					Message: v.Message,
					Type:    "StructureLintViolation",
					Content: fmt.Sprintf("%s: %s", v.Path, v.Message),
				},
			})
		}

		testSuites = append(testSuites, JUnitTestSuite{
			Name:      ruleName,
			Tests:     len(testCases),
			Failures:  len(testCases),
			Errors:    0,
			Time:      "0",
			Timestamp: timestamp,
			TestCases: testCases,
		})

		totalTests += len(testCases)
	}

	// If no violations, create a passing test suite
	if len(violations) == 0 {
		testSuites = append(testSuites, JUnitTestSuite{
			Name:      "structurelint",
			Tests:     1,
			Failures:  0,
			Errors:    0,
			Time:      "0",
			Timestamp: timestamp,
			TestCases: []JUnitTestCase{
				{
					Name:      "all-rules",
					Classname: "structurelint",
					Time:      "0",
				},
			},
		})
		totalTests = 1
	}

	output := JUnitTestSuites{
		XMLNS:      "https://github.com/Jonathangadeaharder/structurelint",
		Name:       "structurelint",
		Tests:      totalTests,
		Failures:   totalFailures,
		Errors:     0,
		Time:       "0",
		TestSuites: testSuites,
	}

	data, err := xml.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	return xml.Header + string(data), nil
}

// GetFormatter returns a formatter based on the format name
func GetFormatter(format string, version string) (Formatter, error) {
	switch strings.ToLower(format) {
	case "text", "":
		return &TextFormatter{}, nil
	case "json":
		return &JSONFormatter{Version: version}, nil
	case "junit", "junit-xml":
		return &JUnitFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s (supported: text, json, junit)", format)
	}
}
