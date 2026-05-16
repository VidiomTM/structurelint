package scaffold

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplate(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)

	tmpl, err := g.GetTemplate(LangGo, TypeService)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.Equal(t, TypeService, tmpl.Type)

	_, err = g.GetTemplate("nonexistent", TypeService)
	assert.Error(t, err)
}

func TestTemplateConstants(t *testing.T) {
	assert.Equal(t, TemplateType("service"), TypeService)
	assert.Equal(t, TemplateType("repository"), TypeRepository)
	assert.Equal(t, TemplateType("controller"), TypeController)
	assert.Equal(t, TemplateType("model"), TypeModel)
	assert.Equal(t, TemplateType("handler"), TypeHandler)
	assert.Equal(t, TemplateType("middleware"), TypeMiddleware)
	assert.Equal(t, TemplateType("test"), TypeTest)

	assert.Equal(t, Language("go"), LangGo)
	assert.Equal(t, Language("typescript"), LangTypeScript)
	assert.Equal(t, Language("python"), LangPython)
	assert.Equal(t, Language("java"), LangJava)
}
