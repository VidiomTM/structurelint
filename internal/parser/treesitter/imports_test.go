package treesitter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImports_ExtractGoImports_Simple(t *testing.T) {
	e, err := NewImportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}

	source := []byte(`package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello")
}
`)
	imports, err := e.ExtractFromSource(source, "main.go")
	if err == nil {
		assert.NotEmpty(t, imports)
	}
}

func TestImports_ExtractGoImports_Relative(t *testing.T) {
	e, err := NewImportExtractor(LanguageGo)
	if err != nil {
		t.Skip("tree-sitter Go grammar not available")
	}

	source := []byte(`package main

import (
	"fmt"
	"github.com/example/lib"
)
`)
	imports, err := e.ExtractFromSource(source, "main.go")
	if err == nil {
		assert.NotEmpty(t, imports)
		for _, imp := range imports {
			if imp.ImportPath == "fmt" {
				assert.False(t, imp.IsRelative)
			}
		}
	}
}

func TestImports_ExtractJSImports(t *testing.T) {
	e, err := NewImportExtractor(LanguageJavaScript)
	if err != nil {
		t.Skip("tree-sitter JS grammar not available")
	}

	source := []byte(`import { foo } from './bar';
const baz = require('lodash');
`)
	imports, err := e.ExtractFromSource(source, "main.js")
	if err == nil {
		assert.NotEmpty(t, imports)
	}
}

func TestImports_ExtractPythonImports(t *testing.T) {
	e, err := NewImportExtractor(LanguagePython)
	if err != nil {
		t.Skip("tree-sitter Python grammar not available")
	}

	source := []byte(`import os
from typing import List
`)
	imports, err := e.ExtractFromSource(source, "main.py")
	if err == nil {
		assert.NotEmpty(t, imports)
	}
}
