package rules

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/predicate"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestPredicateRule_Check_ReturnsViolations(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool {
		return file.Path == "bad.go"
	}
	r := NewPredicateRule("test-rule", "test description", pred, "file %s is bad")
	files := []walker.FileInfo{{Path: "good.go"}, {Path: "bad.go"}, {Path: "other.go"}}
	v := r.Check(files, nil)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Path != "bad.go" {
		t.Errorf("violation path = %s, want bad.go", v[0].Path)
	}
	if v[0].Rule != "test-rule" {
		t.Errorf("violation rule = %s, want test-rule", v[0].Rule)
	}
}

func TestPredicateRule_Check_NoViolations(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool { return false }
	r := NewPredicateRule("test-rule", "test", pred, "")
	v := r.Check([]walker.FileInfo{{Path: "good.go"}}, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations, got %d", len(v))
	}
}

func TestPredicateRule_DefaultMessage(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool { return true }
	r := NewPredicateRule("test-rule", "test description", pred, "")
	v := r.Check([]walker.FileInfo{{Path: "f.go"}}, nil)
	if len(v) != 1 || v[0].Message != "test description: f.go" {
		t.Errorf("unexpected message: %s", v[0].Message)
	}
}

func TestPredicateRule_Name(t *testing.T) {
	r := NewPredicateRule("my-rule", "", nil, "")
	if r.Name() != "my-rule" {
		t.Errorf("Name() = %s, want my-rule", r.Name())
	}
}

func TestPredicateRule_WithGraph(t *testing.T) {
	r := NewPredicateRule("r", "", nil, "")
	r2 := r.WithGraph(nil)
	if r != r2 {
		t.Error("WithGraph should return the same pointer")
	}
}

func TestRequireFileRule_Found(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool {
		return file.Path == "main.go"
	}
	r := NewRequireFileRule("require-test", "test", pred, "")
	v := r.Check([]walker.FileInfo{{Path: "main.go"}, {Path: "other.go"}}, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations, got %d", len(v))
	}
}

func TestRequireFileRule_NotFound(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool { return false }
	r := NewRequireFileRule("require-test", "test", pred, "custom message")
	v := r.Check([]walker.FileInfo{{Path: "other.go"}}, nil)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Message != "custom message" {
		t.Errorf("message = %s, want custom message", v[0].Message)
	}
}

func TestRequireFileRule_DefaultMessage(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool { return false }
	r := NewRequireFileRule("require-test", "test description", pred, "")
	v := r.Check([]walker.FileInfo{{Path: "f.go"}}, nil)
	if len(v) != 1 || v[0].Message != "Required file not found: test description" {
		t.Errorf("unexpected message: %s", v[0].Message)
	}
}

func TestRequireFileRule_WithGraph(t *testing.T) {
	r := NewRequireFileRule("r", "", nil, "")
	r2 := r.WithGraph(nil)
	if r != r2 {
		t.Error("WithGraph should return the same pointer")
	}
}

func TestRequireFileRule_Name(t *testing.T) {
	r := NewRequireFileRule("my-require", "", nil, "")
	if r.Name() != "my-require" {
		t.Errorf("Name() = %s, want my-require", r.Name())
	}
}

func TestDisallowFilesWhere(t *testing.T) {
	pred := func(file walker.FileInfo, ctx *predicate.Context) bool { return file.Path == "bad.go" }
	r := DisallowFilesWhere("no-bad", "no bad files", pred)
	v := r.Check([]walker.FileInfo{{Path: "bad.go"}}, nil)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Message != "bad.go violates rule: no bad files" {
		t.Errorf("message = %s", v[0].Message)
	}
}
