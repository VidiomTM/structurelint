package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/parser/treesitter"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func generateLargeStruct(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("package domain\n\n")
	b.WriteString("type LargeEntity struct {\n")
	for i := range 100 {
		fmt.Fprintf(&b, "\tField%d string\n", i)
	}
	b.WriteString("}\n")
	return b.String()
}

func generateSmallStruct(t *testing.T) string {
	t.Helper()
	return "package domain\n\ntype SmallEntity struct {\n\tName string\n}\n"
}

func TestNewASTQueryRule(t *testing.T) {
	r := NewASTQueryRule("test-rule", "desc", nil, nil)
	if r.Name() != "test-rule" {
		t.Fatalf("Name() = %q, want %q", r.Name(), "test-rule")
	}
}

func TestExampleRequireInterfaceRule(t *testing.T) {
	dir := t.TempDir()
	domainDir := filepath.Join(dir, "internal", "domain")
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatal(err)
	}
	largeFile := filepath.Join(domainDir, "entity.go")
	if err := os.WriteFile(largeFile, []byte(generateLargeStruct(t)), 0644); err != nil {
		t.Fatal(err)
	}
	smallFile := filepath.Join(domainDir, "small.go")
	if err := os.WriteFile(smallFile, []byte(generateSmallStruct(t)), 0644); err != nil {
		t.Fatal(err)
	}

	rule := ExampleRequireInterfaceRule()
	if rule.Name() != "require-interface" {
		t.Fatalf("Name() = %q", rule.Name())
	}

	files := []walker.FileInfo{
		{AbsPath: largeFile, Path: "internal/domain/entity.go"},
		{AbsPath: smallFile, Path: "internal/domain/small.go"},
	}
	v := rule.Check(files, nil)
	if len(v) == 0 {
		t.Error("expected at least one violation for large struct in domain")
	}
}

func TestExampleRequireInterfaceRule_NonDomain(t *testing.T) {
	dir := t.TempDir()
	largeFile := filepath.Join(dir, "entity.go")
	if err := os.WriteFile(largeFile, []byte(generateLargeStruct(t)), 0644); err != nil {
		t.Fatal(err)
	}

	files := []walker.FileInfo{
		{AbsPath: largeFile, Path: "entity.go"},
	}
	v := ExampleRequireInterfaceRule().Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for non-domain file, got %d", len(v))
	}
}

func TestExampleDisallowDirectDBAccessRule(t *testing.T) {
	dir := t.TempDir()
	domainDir := filepath.Join(dir, "internal", "domain")
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(domainDir, "repo.go")
	code := `package domain

type db interface {
	Query(string, ...interface{}) ([]interface{}, error)
}

func fetchUsers(d db) {
	d.Query("SELECT * FROM users")
}
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	rule := ExampleDisallowDirectDBAccessRule()
	if rule.Name() != "no-direct-db-access" {
		t.Fatalf("Name() = %q", rule.Name())
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "internal/domain/repo.go"},
	}
	v := rule.Check(files, nil)
	if len(v) == 0 {
		t.Error("expected violation for db access in domain")
	}
}

func TestExampleDisallowDirectDBAccessRule_NonDomain(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "infra.go")
	code := `package infra
type db struct{}
func fetchUsers(d *db) { d.Query("SELECT 1") }
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "infrastructure/repo.go"},
	}
	v := ExampleDisallowDirectDBAccessRule().Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for non-domain, got %d", len(v))
	}
}

func TestExampleRequireDeprecationCommentRule(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "service.go")
	code := `package test

func OldService() {}
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	rule := ExampleRequireDeprecationCommentRule()
	if rule.Name() != "require-deprecation-comment" {
		t.Fatalf("Name() = %q", rule.Name())
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "service.go"},
	}
	v := rule.Check(files, nil)
	if len(v) == 0 {
		t.Error("expected violation for Old-prefixed function")
	}
}

func TestExampleRequireDeprecationCommentRule_NoViolation(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "service.go")
	code := `package test

func NewService() {}
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "service.go"},
	}
	v := ExampleRequireDeprecationCommentRule().Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for non-Old function, got %d", len(v))
	}
}

func TestExampleDisallowGlobalVariablesRule(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "config.go")
	code := `package test

import "fmt"

var debugMode = true

func Run() {
	var local = "hello"
	_ = local
	fmt.Println(debugMode)
}
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	rule := ExampleDisallowGlobalVariablesRule()
	if rule.Name() != "no-global-variables" {
		t.Fatalf("Name() = %q", rule.Name())
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "config.go"},
	}
	v := rule.Check(files, nil)
	if len(v) == 0 {
		t.Error("expected violation for global variable")
	}
}

func TestExampleDisallowGlobalVariablesRule_SkipsTestFile(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "config_test.go")
	code := `package test

var debugMode = true
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "config_test.go"},
	}
	v := ExampleDisallowGlobalVariablesRule().Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for test file, got %d", len(v))
	}
}

func TestExampleDisallowGlobalVariablesRule_UpperCaseAllowed(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "constants.go")
	code := `package test

const MaxRetries = 3
`
	if err := os.WriteFile(fixture, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "constants.go"},
	}
	v := ExampleDisallowGlobalVariablesRule().Check(files, nil)
	if len(v) != 0 {
		t.Errorf("expected 0 violations for upper-case, got %d", len(v))
	}
}

func TestASTQueryRule_EmptyCheck(t *testing.T) {
	r := NewASTQueryRule("empty", "desc", nil, nil)
	v := r.Check(nil, nil)
	if len(v) != 0 {
		t.Error("expected no violations for nil files")
	}
}

func TestASTQueryRule_CheckWithDirs(t *testing.T) {
	r := NewASTQueryRule("dirs", "desc", nil, nil)
	files := []walker.FileInfo{
		{IsDir: true, Path: "dir"},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Error("expected no violations for directories")
	}
}

func TestASTQueryRule_CheckUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "data.custom")
	if err := os.WriteFile(fixture, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewASTQueryRule("noext", "desc", nil, nil)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "data.custom"},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Error("expected no violations for unsupported extension")
	}
}

func TestASTQueryRule_CheckNoQueryForLang(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "main.go")
	if err := os.WriteFile(fixture, []byte("package p"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewASTQueryRule("noquery", "desc", nil, nil)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "main.go"},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Error("expected no violations when no queries defined")
	}
}

func TestASTQueryRule_CheckParseError(t *testing.T) {
	dir := t.TempDir()
	fixture := filepath.Join(dir, "broken.go")
	if err := os.WriteFile(fixture, []byte("this is not valid go !!!"), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewASTQueryRule("broken", "desc", map[treesitter.Language]string{treesitter.LanguageGo: ". . ."}, nil)
	files := []walker.FileInfo{
		{AbsPath: fixture, Path: "broken.go"},
	}
	v := r.Check(files, nil)
	if len(v) != 0 {
		t.Error("expected no violations for unparseable file")
	}
}
