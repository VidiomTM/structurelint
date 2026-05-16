package structure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestDeepRelativeImportsRule_Name(t *testing.T) {
	r := NewDeepRelativeImportsRule(3)
	if r.Name() != "disallow-deep-relative-imports" {
		t.Fatalf("Name() = %q", r.Name())
	}
}

func TestDeepRelativeImportsRule_New(t *testing.T) {
	r := NewDeepRelativeImportsRule(5)
	if r.MaxParents != 5 {
		t.Errorf("MaxParents = %d, want 5", r.MaxParents)
	}
}

func TestDeepRelativeImportsRule_Check_Violation(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "service.go")
	code := `package service

import "../../../deep/foo"
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewDeepRelativeImportsRule(3)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "service.go"},
	}
	v := r.Check(files, nil)
	if len(v) == 0 {
		t.Fatal("expected violation for deep relative import")
	}
	if v[0].Rule != "disallow-deep-relative-imports" {
		t.Errorf("Rule = %q", v[0].Rule)
	}
	if v[0].Path != "service.go" {
		t.Errorf("Path = %q", v[0].Path)
	}
}

func TestDeepRelativeImportsRule_Check_NoViolation(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "service.go")
	code := `package service

import "./local"
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewDeepRelativeImportsRule(3)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "service.go"},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for shallow import, got %d", len(v))
	}
}

func TestDeepRelativeImportsRule_Check_DefaultLimit(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "service.go")
	code := `package service

import "../../../deep/foo"
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewDeepRelativeImportsRule(0)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "service.go"},
	}
	v := r.Check(files, nil)
	if len(v) == 0 {
		t.Error("expected violation with default 3 limit")
	}
}

func TestDeepRelativeImportsRule_Check_SkipsDirs(t *testing.T) {
	r := NewDeepRelativeImportsRule(3)
	files := []walker.FileInfo{
		{Path: "dir", IsDir: true},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for dirs, got %d", len(v))
	}
}

func TestDeepRelativeImportsRule_Check_UnreadableFile(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "nonexistent.go")

	r := NewDeepRelativeImportsRule(3)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "nonexistent.go"},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for unreadable file, got %d", len(v))
	}
}
