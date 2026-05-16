package treesitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateCognitiveComplexity_Simple(t *testing.T) {
	p, err := New(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	source := []byte(`package main
func main() {
	if true {
		return
	}
}
`)
	tree, err := p.Parse(source)
	require.NoError(t, err)
	defer tree.Close()

	m := &MetricsCalculator{parser: p, language: LanguageGo}
	c := m.calculateCognitiveComplexity(tree, source)
	assert.Greater(t, c, 0)
}

func TestCalculateCognitiveComplexity_Nested(t *testing.T) {
	p, err := New(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	source := []byte(`package main
func main() {
	if true {
		for i := 0; i < 10; i++ {
			if false {
				return
			}
		}
	}
}
`)
	tree, err := p.Parse(source)
	require.NoError(t, err)
	defer tree.Close()

	m := &MetricsCalculator{parser: p, language: LanguageGo}
	c := m.calculateCognitiveComplexity(tree, source)
	assert.Greater(t, c, 2)
}

func TestCalculateHalsteadMetrics_Basic(t *testing.T) {
	p, err := New(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}
	source := []byte(`package main
func add(a int, b int) int {
	return a + b
}
`)
	tree, err := p.Parse(source)
	require.NoError(t, err)
	defer tree.Close()

	m := &MetricsCalculator{parser: p, language: LanguageGo}
	metrics := m.calculateHalsteadMetrics(tree, source)
	assert.Greater(t, metrics.UniqueOperators, 0)
	assert.Greater(t, metrics.UniqueOperands, 0)
	assert.Greater(t, metrics.Volume, 0.0)
}

func TestIsGoExported(t *testing.T) {
	assert.True(t, isGoExported("Foo"))
	assert.True(t, isGoExported("ExportedFunction"))
	assert.False(t, isGoExported("foo"))
	assert.False(t, isGoExported("_private"))
	assert.False(t, isGoExported(""))
}
