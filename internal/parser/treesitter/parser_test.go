package treesitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectLanguageFromExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected Language
		hasErr   bool
	}{
		{".go", LanguageGo, false},
		{".py", LanguagePython, false},
		{".js", LanguageJavaScript, false},
		{".jsx", LanguageJavaScript, false},
		{".mjs", LanguageJavaScript, false},
		{".ts", LanguageTypeScript, false},
		{".tsx", LanguageTypeScript, false},
		{".java", LanguageJava, false},
		{".cpp", LanguageCpp, false},
		{".cc", LanguageCpp, false},
		{".cxx", LanguageCpp, false},
		{".c", LanguageCpp, false},
		{".h", LanguageCpp, false},
		{".hpp", LanguageCpp, false},
		{".cs", LanguageCSharp, false},
		{".rs", "", true},
		{".unknown", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.ext, func(t *testing.T) {
			lang, err := DetectLanguageFromExtension(tc.ext)
			if tc.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, lang)
			}
		})
	}
}

func TestNewParser_UnsupportedLanguage(t *testing.T) {
	p, err := New("unsupported")
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestNewParser_SupportedLanguages(t *testing.T) {
	languages := []Language{LanguageGo, LanguagePython, LanguageJavaScript, LanguageTypeScript, LanguageJava, LanguageCpp, LanguageCSharp}
	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			p, err := New(lang)
			if err != nil {
				t.Skipf("tree-sitter grammar not available for %s: %v", lang, err)
				return
			}
			assert.NotNil(t, p)
			assert.Equal(t, lang, p.language)
		})
	}
}

func TestParse_SimpleGo(t *testing.T) {
	p, err := New(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	source := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	tree, err := p.Parse(source)
	require.NoError(t, err)
	require.NotNil(t, tree)
	defer tree.Close()

	root := tree.RootNode()
	assert.Equal(t, "source_file", root.Type())
}

func TestParse_EmptyInput(t *testing.T) {
	p, err := New(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	tree, err := p.Parse([]byte{})
	require.NoError(t, err)
	require.NotNil(t, tree)
	defer tree.Close()
}

func TestParse_NilError(t *testing.T) {
	p, err := New(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	tree, err := p.Parse([]byte("package main"))
	require.NoError(t, err)
	defer tree.Close()
	assert.NotNil(t, tree)
}

func TestNewMetricsCalculator(t *testing.T) {
	m, err := NewMetricsCalculator(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	require.NotNil(t, m)
	assert.Equal(t, LanguageGo, m.language)
}

func TestNewMetricsCalculator_Unsupported(t *testing.T) {
	m, err := NewMetricsCalculator("unsupported")
	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestIsNestingStructure(t *testing.T) {
	m := &MetricsCalculator{}
	assert.True(t, m.isNestingStructure("if_statement"))
	assert.True(t, m.isNestingStructure("for_statement"))
	assert.True(t, m.isNestingStructure("function_declaration"))
	assert.True(t, m.isNestingStructure("with_statement"))
	assert.True(t, m.isNestingStructure("match_statement"))
	assert.False(t, m.isNestingStructure("identifier"))
	assert.False(t, m.isNestingStructure("binary_expression"))
}

func TestIsComplexityNode(t *testing.T) {
	m := &MetricsCalculator{}
	assert.True(t, m.isComplexityNode("if_statement"))
	assert.True(t, m.isComplexityNode("else_clause"))
	assert.True(t, m.isComplexityNode("for_statement"))
	assert.True(t, m.isComplexityNode("try_statement"))
	assert.True(t, m.isComplexityNode("return_statement"))
	assert.True(t, m.isComplexityNode("binary_expression"))
	assert.False(t, m.isComplexityNode("identifier"))
}

func TestIsOperator(t *testing.T) {
	m := &MetricsCalculator{}
	assert.True(t, m.isOperator("+"))
	assert.True(t, m.isOperator("=="))
	assert.True(t, m.isOperator("&&"))
	assert.True(t, m.isOperator("binary_expression"))
	assert.True(t, m.isOperator("assignment_expression"))
	assert.False(t, m.isOperator("identifier"))
}

func TestIsOperand(t *testing.T) {
	m := &MetricsCalculator{}
	assert.True(t, m.isOperand("identifier"))
	assert.True(t, m.isOperand("number"))
	assert.True(t, m.isOperand("string"))
	assert.True(t, m.isOperand("true"))
	assert.True(t, m.isOperand("type_identifier"))
	assert.False(t, m.isOperand("binary_expression"))
}

func TestNewImportExtractor(t *testing.T) {
	e, err := NewImportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	require.NotNil(t, e)
	assert.Equal(t, LanguageGo, e.language)
}

func TestNewExportExtractor(t *testing.T) {
	e, err := NewExportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	require.NotNil(t, e)
	assert.Equal(t, LanguageGo, e.language)
}

func TestResolveImportPath(t *testing.T) {
	result := ResolveImportPath("src/main.go", "./util")
	assert.Equal(t, "src/util", result)

	result2 := ResolveImportPath("src/main.go", "../lib/helper")
	assert.Equal(t, "lib/helper", result2)
}
