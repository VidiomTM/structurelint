package detector

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

func TestNewReporter(t *testing.T) {
	r := NewReporter("console")
	if r == nil {
		t.Fatal("NewReporter returned nil")
	}
	if r.format != "console" {
		t.Errorf("format = %q, want %q", r.format, "console")
	}
}

func TestReporter_Report_Console_NoClones(t *testing.T) {
	r := NewReporter("console")
	output := r.Report(nil)
	if !strings.Contains(output, "No code clones") {
		t.Errorf("expected 'No code clones', got %q", output)
	}
}

func TestReporter_Report_Console_WithClones(t *testing.T) {
	r := NewReporter("console")
	clones := []*types.Clone{
		{
			Type:       types.Type1,
			TokenCount: 50,
			LineCount:  10,
			Similarity: 1.0,
			Locations: []types.Location{
				{FilePath: "a.go", StartLine: 1, EndLine: 10},
				{FilePath: "b.go", StartLine: 20, EndLine: 30},
			},
		},
	}
	output := r.Report(clones)
	if !strings.Contains(output, "Clone Pair #1") {
		t.Errorf("expected 'Clone Pair #1', got %q", output)
	}
	if !strings.Contains(output, "a.go") {
		t.Errorf("expected 'a.go', got %q", output)
	}
}

func TestReporter_Report_JSON(t *testing.T) {
	r := NewReporter("json")
	clones := []*types.Clone{
		{
			Type:       types.Type2,
			TokenCount: 30,
			LineCount:  5,
			Similarity: 0.95,
			Locations: []types.Location{
				{FilePath: "x.go", StartLine: 1, EndLine: 5},
			},
		},
	}
	output := r.Report(clones)
	if !strings.Contains(output, `"total_clones": 1`) {
		t.Errorf("expected JSON with 1 clone, got %q", output)
	}
	if !strings.Contains(output, "Type-2") {
		t.Errorf("expected Type-2 in JSON, got %q", output)
	}
}

func TestReporter_Report_JSON_Empty(t *testing.T) {
	r := NewReporter("json")
	output := r.Report(nil)
	if !strings.Contains(output, `"total_clones": 0`) {
		t.Errorf("expected JSON with 0 clones, got %q", output)
	}
}

func TestReporter_Report_SARIF(t *testing.T) {
	r := NewReporter("sarif")
	clones := []*types.Clone{
		{
			Type:       types.Type1,
			TokenCount: 20,
			LineCount:  5,
			Similarity: 1.0,
			Locations: []types.Location{
				{FilePath: "a.go", StartLine: 1, EndLine: 5},
				{FilePath: "b.go", StartLine: 10, EndLine: 15},
			},
		},
	}
	output := r.Report(clones)
	if !strings.Contains(output, `"version": "2.1.0"`) {
		t.Errorf("expected SARIF version, got %q", output)
	}
	if !strings.Contains(output, `"ruleId": "code-clone-type-1`) {
		t.Errorf("expected ruleId in SARIF, got %q", output)
	}
}

func TestReporter_Report_SARIF_Empty(t *testing.T) {
	r := NewReporter("sarif")
	output := r.Report(nil)
	if !strings.Contains(output, `"results": []`) {
		t.Errorf("expected empty results in SARIF, got %q", output)
	}
}

func TestReporter_Report_DefaultFormat(t *testing.T) {
	r := NewReporter("unknown")
	clones := []*types.Clone{
		{
			Type: types.Type3, TokenCount: 10, LineCount: 3, Similarity: 0.8,
			Locations: []types.Location{
				{FilePath: "a.go", StartLine: 1, EndLine: 3},
			},
		},
	}
	output := r.Report(clones)
	if !strings.Contains(output, "Clone Pair #1") {
		t.Errorf("expected console output, got %q", output)
	}
}

func TestReporter_Summary_NoClones(t *testing.T) {
	r := NewReporter("console")
	s := r.Summary(nil)
	if !strings.Contains(s, "No clones") {
		t.Errorf("expected 'No clones', got %q", s)
	}
}

func TestReporter_Summary_WithClones(t *testing.T) {
	r := NewReporter("console")
	clones := []*types.Clone{
		{Type: types.Type1, TokenCount: 50, LineCount: 10},
		{Type: types.Type2, TokenCount: 30, LineCount: 5},
		{Type: types.Type3, TokenCount: 20, LineCount: 3},
		{Type: types.Type4, TokenCount: 15, LineCount: 2},
	}
	s := r.Summary(clones)
	if !strings.Contains(s, "4 total") {
		t.Errorf("expected '4 total', got %q", s)
	}
	if !strings.Contains(s, "Type-1: 1") {
		t.Errorf("expected Type-1: 1, got %q", s)
	}
	if !strings.Contains(s, "Type-2: 1") {
		t.Errorf("expected Type-2: 1, got %q", s)
	}
	if !strings.Contains(s, "Type-3: 1") {
		t.Errorf("expected Type-3: 1, got %q", s)
	}
}
