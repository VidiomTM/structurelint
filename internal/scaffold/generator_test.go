package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserService", "user_service"},
		{"HTTPServer", "h_t_t_p_server"},
		{"simple", "simple"},
		{"AlreadySnake", "already_snake"},
		{"ABC", "a_b_c"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, toSnakeCase(tc.input))
		})
	}
}

func TestToKebabCase(t *testing.T) {
	assert.Equal(t, "user-service", toKebabCase("UserService"))
	assert.Equal(t, "simple", toKebabCase("Simple"))
}

func TestToCamelCase(t *testing.T) {
	assert.Equal(t, "userService", toCamelCase("UserService"))
	assert.Equal(t, "", toCamelCase(""))
}

func TestNewGenerator(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)
	require.NotNil(t, g)
	assert.Equal(t, dir, g.rootDir)
	assert.NotEmpty(t, g.templates)
}

func TestListTemplates(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)
	templates := g.ListTemplates()
	assert.NotEmpty(t, templates)
}

func TestGenerate_UnknownTemplate(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)
	err := g.Generate("nonexistent", "MyComponent", Variables{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestCompleteVariables(t *testing.T) {
	g := NewGenerator("/test")
	vars := g.completeVariables("UserService", Variables{})
	assert.Equal(t, "UserService", vars.Name)
	assert.Equal(t, "userservice", vars.NameLower)
	assert.Equal(t, "user_service", vars.NameSnake)
	assert.Equal(t, "user-service", vars.NameKebab)
	assert.Equal(t, "userService", vars.NameCamel)
	assert.Equal(t, "UserService component", vars.Description)
	assert.NotEmpty(t, vars.Author)
	assert.NotNil(t, vars.CustomVars)
}

func TestCompleteVariables_CustomValues(t *testing.T) {
	g := NewGenerator("/test")
	vars := g.completeVariables("MyService", Variables{
		Package:     "mypkg",
		Description: "custom desc",
		Author:      "me",
	})
	assert.Equal(t, "mypkg", vars.Package)
	assert.Equal(t, "custom desc", vars.Description)
	assert.Equal(t, "me", vars.Author)
}

func TestDetectPackage_GoMod(t *testing.T) {
	dir := t.TempDir()
	goMod := filepath.Join(dir, "go.mod")
	os.WriteFile(goMod, []byte("module github.com/test/project\n\ngo 1.21"), 0644)

	g := NewGenerator(dir)
	pkg := g.detectPackage()
	assert.Equal(t, "github.com/test/project", pkg)
}

func TestDetectPackage_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)
	pkg := g.detectPackage()
	assert.Equal(t, filepath.Base(dir), pkg)
}

func TestDetectPackage_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	pkgPath := filepath.Join(dir, "package.json")
	os.WriteFile(pkgPath, []byte(`{"name": "my-package", "version": "1.0.0"}`), 0644)

	g := NewGenerator(dir)
	pkg := g.detectPackage()
	assert.Equal(t, "my-package", pkg)
}

func TestRenderTemplate(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)
	result, err := g.renderTemplate("package {{.NameSnake}}", Variables{Name: "MyService", NameSnake: "my_service"})
	require.NoError(t, err)
	assert.Equal(t, "package my_service", result)
}

func TestRenderTemplate_Error(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)
	_, err := g.renderTemplate("{{.nonexistent.field}}", Variables{})
	assert.Error(t, err)
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.go")
	g := NewGenerator(dir)
	err := g.writeFile(path, "content")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "content", string(data))
}

func TestWriteFile_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.go")
	os.WriteFile(path, []byte("original"), 0644)

	g := NewGenerator(dir)
	err := g.writeFile(path, "new content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGenerate_WithTemplate(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)

	tmpl := &Template{
		Type: TypeService,
		Files: []TemplateFile{
			{Path: "services/{{.NameSnake}}.go", Content: "package services"},
		},
	}
	g.templates["custom-service"] = tmpl

	err := g.Generate("custom-service", "UserService", Variables{})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "services/user_service.go"))
	assert.NoError(t, err)
}

func TestGenerate_SkipOptional(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)

	tmpl := &Template{
		Type: TypeService,
		Files: []TemplateFile{
			{Path: "main.go", Content: "package main"},
			{Path: "main_test.go", Content: "package main", IsTest: true, Optional: true},
		},
	}
	g.templates["no-tests"] = tmpl

	err := g.Generate("no-tests", "MyService", Variables{IncludeTests: false})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "main_test.go"))
	assert.True(t, os.IsNotExist(err))
}

func TestTemplateTypes(t *testing.T) {
	assert.Equal(t, TemplateType("service"), TypeService)
	assert.Equal(t, TemplateType("repository"), TypeRepository)
	assert.Equal(t, TemplateType("controller"), TypeController)
	assert.Equal(t, TemplateType("model"), TypeModel)
	assert.Equal(t, TemplateType("handler"), TypeHandler)
	assert.Equal(t, TemplateType("middleware"), TypeMiddleware)
	assert.Equal(t, TemplateType("test"), TypeTest)
}

func TestLanguages(t *testing.T) {
	assert.Equal(t, Language("go"), LangGo)
	assert.Equal(t, Language("typescript"), LangTypeScript)
	assert.Equal(t, Language("python"), LangPython)
	assert.Equal(t, Language("java"), LangJava)
}
