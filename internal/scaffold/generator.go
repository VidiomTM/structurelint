package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// TemplateType represents the type of component to scaffold
type TemplateType string

const (
	TypeService    TemplateType = "service"
	TypeRepository TemplateType = "repository"
	TypeController TemplateType = "controller"
	TypeModel      TemplateType = "model"
	TypeHandler    TemplateType = "handler"
	TypeMiddleware TemplateType = "middleware"
	TypeTest       TemplateType = "test"
)

// Language represents the target programming language
type Language string

const (
	LangGo         Language = "go"
	LangTypeScript Language = "typescript"
	LangPython     Language = "python"
	LangJava       Language = "java"
)

// Template represents a code template
type Template struct {
	Type        TemplateType
	Language    Language
	Name        string
	Description string
	Files       []TemplateFile
}

// TemplateFile represents a single file in a template
type TemplateFile struct {
	Path     string // Relative path with template vars (e.g., "services/{{.Name}}.go")
	Content  string // File content with template vars
	IsTest   bool   // Whether this is a test file
	Optional bool   // Whether this file is optional
}

// Variables holds template variable values
type Variables struct {
	Name           string // Component name (e.g., "UserService")
	NameLower      string // lowercase name (e.g., "userservice")
	NameSnake      string // snake_case name (e.g., "user_service")
	NameKebab      string // kebab-case name (e.g., "user-service")
	NameCamel      string // camelCase name (e.g., "userService")
	Package        string // Package/module name
	Description    string // Component description
	Author         string // Author name
	IncludeTests   bool   // Whether to include test files
	CustomVars     map[string]string
}

// Generator handles code generation from templates
type Generator struct {
	rootDir   string
	templates map[string]*Template
}

// NewGenerator creates a new scaffold generator
func NewGenerator(rootDir string) *Generator {
	g := &Generator{
		rootDir:   rootDir,
		templates: make(map[string]*Template),
	}

	// Register built-in templates
	g.registerBuiltInTemplates()

	return g
}

// Generate generates code from a template
func (g *Generator) Generate(templateName, componentName string, vars Variables) error {
	// Find template
	tmpl, ok := g.templates[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Complete variables
	vars = g.completeVariables(componentName, vars)

	// Generate each file
	for _, file := range tmpl.Files {
		if file.Optional && !vars.IncludeTests && file.IsTest {
			continue
		}

		// Render file path
		filePath, err := g.renderTemplate(file.Path, vars)
		if err != nil {
			return fmt.Errorf("failed to render file path: %w", err)
		}

		// Render file content
		content, err := g.renderTemplate(file.Content, vars)
		if err != nil {
			return fmt.Errorf("failed to render file content: %w", err)
		}

		// Create file
		fullPath := filepath.Join(g.rootDir, filePath)
		if err := g.writeFile(fullPath, content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		fmt.Printf("✓ Created %s\n", filePath)
	}

	return nil
}

// completeVariables fills in derived variable values
func (g *Generator) completeVariables(name string, vars Variables) Variables {
	vars.Name = name
	vars.NameLower = strings.ToLower(name)
	vars.NameSnake = toSnakeCase(name)
	vars.NameKebab = toKebabCase(name)
	vars.NameCamel = toCamelCase(name)

	vars = g.fillDefaultVars(vars)

	return vars
}

func (g *Generator) fillDefaultVars(vars Variables) Variables {
	if vars.Package == "" {
		vars.Package = g.detectPackage()
	}
	if vars.Description == "" {
		vars.Description = fmt.Sprintf("%s component", vars.Name)
	}
	if vars.Author == "" {
		vars.Author = os.Getenv("USER")
		if vars.Author == "" {
			vars.Author = "Developer"
		}
	}
	if vars.CustomVars == nil {
		vars.CustomVars = make(map[string]string)
	}
	return vars
}

// renderTemplate renders a template string with variables
func (g *Generator) renderTemplate(tmplStr string, vars Variables) (string, error) {
	tmpl, err := template.New("scaffold").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// writeFile writes content to a file, creating directories as needed
func (g *Generator) writeFile(path, content string) error {
	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	// Write file
	return os.WriteFile(path, []byte(content), 0644)
}

// detectPackage detects the package/module name from the project
func (g *Generator) detectPackage() string {
	// Try to read go.mod for Go projects
	goModPath := filepath.Join(g.rootDir, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				return strings.TrimPrefix(line, "module ")
			}
		}
	}

	// Try to read package.json for TypeScript/JavaScript projects
	pkgJSONPath := filepath.Join(g.rootDir, "package.json")
	if data, err := os.ReadFile(pkgJSONPath); err == nil {
		// Simple extraction without JSON parsing
		content := string(data)
		if idx := strings.Index(content, `"name"`); idx != -1 {
			rest := content[idx+6:]
			if idx2 := strings.Index(rest, `"`); idx2 != -1 {
				rest = rest[idx2+1:]
				if idx3 := strings.Index(rest, `"`); idx3 != -1 {
					return rest[:idx3]
				}
			}
		}
	}

	// Default to directory name
	return filepath.Base(g.rootDir)
}

// ListTemplates returns all available templates
func (g *Generator) ListTemplates() []*Template {
	var templates []*Template
	for _, tmpl := range g.templates {
		templates = append(templates, tmpl)
	}
	return templates
}

// Helper functions for name case conversions

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
