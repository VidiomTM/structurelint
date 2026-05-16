package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTypeScriptImports(t *testing.T) {
	// Arrange
	// Create temp file
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `import { foo } from './foo';
import bar from '../bar';
import * as baz from './baz';
const qux = require('./qux');
`

	require.NoError(t, os.WriteFile(tsFile, []byte(content), 0644))

	parser := New(tmpDir)

	// Act
	imports, err := parser.ParseFile(tsFile)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 4, len(imports))

	// Check first import
	assert.Equal(t, "./foo", imports[0].ImportPath)
	assert.True(t, imports[0].IsRelative)
}

func TestParseGoImports(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

import (
	"fmt"
	"github.com/user/repo"
)

import "strings"
`

	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(goFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))

	expectedPaths := map[string]bool{
		"fmt":                  true,
		"github.com/user/repo": true,
		"strings":              true,
	}

	for _, imp := range imports {
		assert.True(t, expectedPaths[imp.ImportPath], "Unexpected import path: %s", imp.ImportPath)
	}
}

func TestParsePythonImports(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `import os
from typing import List
import json.decoder
`

	require.NoError(t, os.WriteFile(pyFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(pyFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))
}

func TestParseTypeScriptExports(t *testing.T) {
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `export const foo = 'bar';
export function hello() {}
export class MyClass {}
export { one, two };
export default something;
`

	require.NoError(t, os.WriteFile(tsFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(tsFile)

	require.NoError(t, err)
	assert.Equal(t, 5, len(exports))

	// Check for default export
	hasDefault := false
	for _, exp := range exports {
		if exp.IsDefault {
			hasDefault = true
			break
		}
	}

	assert.True(t, hasDefault)
}

func TestParseGoExports(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func HelloWorld() {}
func privateFunc() {}
type PublicType struct{}
type privateType struct{}
const PublicConst = 42
var PrivateVar = "test"
`

	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(goFile)

	require.NoError(t, err)

	// Should only find exported (uppercase) symbols
	expectedExports := map[string]bool{
		"HelloWorld":  true,
		"PublicType":  true,
		"PublicConst": true,
		"PrivateVar":  true,
	}

	exportCount := 0
	for _, exp := range exports {
		for _, name := range exp.Names {
			if expectedExports[name] {
				exportCount++
			}
		}
	}

	assert.Equal(t, 4, exportCount)
}

func TestParsePythonExports(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `def public_function():
    pass

def _private_function():
    pass

class PublicClass:
    pass

__all__ = ['public_function', 'PublicClass']
`

	require.NoError(t, os.WriteFile(pyFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(pyFile)

	require.NoError(t, err)
	require.NotEmpty(t, exports, "Expected to find exports")

	// Should prefer __all__ definition
	hasPublicFunction := false
	hasPublicClass := false

	for _, exp := range exports {
		for _, name := range exp.Names {
			if name == "public_function" {
				hasPublicFunction = true
			}
			if name == "PublicClass" {
				hasPublicClass = true
			}
		}
	}

	assert.True(t, hasPublicFunction)
	assert.True(t, hasPublicClass)
}

func TestResolveImportPath(t *testing.T) {
	parser := New("/project")

	tests := []struct {
		sourceFile string
		importPath string
		expected   string
	}{
		{"src/app.ts", "./utils", filepath.Join("src", "utils")},
		{"src/components/Button.tsx", "../hooks/useButton", filepath.Join("src", "hooks", "useButton")},
		{"src/app.ts", "react", "react"}, // External import
	}

	for _, tt := range tests {
		result := parser.ResolveImportPath(tt.sourceFile, tt.importPath)
		assert.Equal(t, tt.expected, result, "ResolveImportPath(%s, %s)", tt.sourceFile, tt.importPath)
	}
}

func TestParseJavaImports(t *testing.T) {
	tmpDir := t.TempDir()
	javaFile := filepath.Join(tmpDir, "Test.java")

	content := `package com.example.app;

import com.example.app.model.User;
import com.example.app.service.*;
import java.util.List;
`

	require.NoError(t, os.WriteFile(javaFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(javaFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))
	assert.Equal(t, "com.example.app.model.User", imports[0].ImportPath)
	assert.True(t, imports[0].IsRelative) // same package prefix
	assert.False(t, imports[2].IsRelative) // java.util is external
}

func TestParseJavaImportsNoPackage(t *testing.T) {
	tmpDir := t.TempDir()
	javaFile := filepath.Join(tmpDir, "Test.java")

	content := `import java.util.List;
`

	require.NoError(t, os.WriteFile(javaFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(javaFile)

	require.NoError(t, err)
	assert.Equal(t, 1, len(imports))
	assert.False(t, imports[0].IsRelative)
}

func TestParseJavaImportsFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/Test.java")
	assert.Error(t, err)
}

func TestParseJavaExports(t *testing.T) {
	tmpDir := t.TempDir()
	javaFile := filepath.Join(tmpDir, "Test.java")

	content := `package com.example;

public class TestClass {
    public void method() {}
}

public interface TestInterface {
    void contract();
}
`

	require.NoError(t, os.WriteFile(javaFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(javaFile)

	require.NoError(t, err)
	require.NotEmpty(t, exports)
	assert.Contains(t, exports[0].Names, "TestClass")
	assert.Contains(t, exports[0].Names, "TestInterface")
}

func TestParseJavaExportsNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	javaFile := filepath.Join(tmpDir, "Test.java")

	content := `package com.example;
class TestClass {}
`

	require.NoError(t, os.WriteFile(javaFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(javaFile)

	require.NoError(t, err)
	assert.Empty(t, exports)
}

func TestParseCppImports(t *testing.T) {
	tmpDir := t.TempDir()
	cppFile := filepath.Join(tmpDir, "test.cpp")

	content := `#include <iostream>
#include "myheader.h"
#include <vector>
// #include <commented_out.hpp>
`

	require.NoError(t, os.WriteFile(cppFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(cppFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))
	assert.False(t, imports[0].IsRelative) // system include
	assert.True(t, imports[1].IsRelative)  // local include
}

func TestParseCppImportsFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/test.cpp")
	assert.Error(t, err)
}

func TestParseCppExports(t *testing.T) {
	tmpDir := t.TempDir()
	hppFile := filepath.Join(tmpDir, "test.hpp")

	content := `#pragma once

namespace myapp {
    class MyClass {
    public:
        void method();
    };

    struct MyStruct {
        int x;
    };
}
`

	require.NoError(t, os.WriteFile(hppFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(hppFile)

	require.NoError(t, err)
	require.NotEmpty(t, exports)
	assert.Contains(t, exports[0].Names, "MyClass")
	assert.Contains(t, exports[0].Names, "MyStruct")
}

func TestParseCppExportsSourceFile(t *testing.T) {
	tmpDir := t.TempDir()
	cppFile := filepath.Join(tmpDir, "test.cpp")

	content := `class Foo {};`
	require.NoError(t, os.WriteFile(cppFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(cppFile)

	require.NoError(t, err)
	assert.Empty(t, exports) // .cpp files don't export
}

func TestParseCppExportsWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	hppFile := filepath.Join(tmpDir, "test.hpp")

	content := `/* block comment */
// line comment
class VisibleClass {};
`

	require.NoError(t, os.WriteFile(hppFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(hppFile)

	require.NoError(t, err)
	require.NotEmpty(t, exports)
	assert.Contains(t, exports[0].Names, "VisibleClass")
}

func TestParseCSharpImports(t *testing.T) {
	tmpDir := t.TempDir()
	csFile := filepath.Join(tmpDir, "Test.cs")

	content := `using System;
using System.Collections.Generic;
using MyApp.Models;
using static System.Math;

namespace MyApp {
    class Test {}
}
`

	require.NoError(t, os.WriteFile(csFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(csFile)

	require.NoError(t, err)
	assert.Equal(t, 4, len(imports))
}

func TestParseCSharpImportsRelative(t *testing.T) {
	tmpDir := t.TempDir()
	csFile := filepath.Join(tmpDir, "Test.cs")

	content := `namespace MyApp.Services {
    using MyApp.Models;
    using MyApp.Utils;
    using System;
    using Newtonsoft.Json;
    class Test {}
}
`

	require.NoError(t, os.WriteFile(csFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(csFile)

	require.NoError(t, err)
	require.Equal(t, 4, len(imports))

	for _, imp := range imports {
		if imp.ImportPath == "MyApp.Models" || imp.ImportPath == "MyApp.Utils" {
			assert.True(t, imp.IsRelative, "expected %s to be relative", imp.ImportPath)
		}
		if imp.ImportPath == "System" || imp.ImportPath == "Newtonsoft.Json" {
			assert.False(t, imp.IsRelative, "expected %s to not be relative", imp.ImportPath)
		}
	}
}

func TestParseCSharpExports(t *testing.T) {
	tmpDir := t.TempDir()
	csFile := filepath.Join(tmpDir, "Test.cs")

	content := `namespace MyApp;

public class TestClass {
    public void Method() {}
}

public interface ITestInterface {
    void Contract();
}

public enum TestEnum { One, Two }

internal class InternalClass {}
`

	require.NoError(t, os.WriteFile(csFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(csFile)

	require.NoError(t, err)
	require.NotEmpty(t, exports)
	assert.Contains(t, exports[0].Names, "TestClass")
	assert.Contains(t, exports[0].Names, "ITestInterface")
	assert.NotContains(t, exports[0].Names, "InternalClass")
}

func TestParseCSharpExportsNone(t *testing.T) {
	tmpDir := t.TempDir()
	csFile := filepath.Join(tmpDir, "Test.cs")

	content := `namespace MyApp;
class InternalClass {}
`

	require.NoError(t, os.WriteFile(csFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(csFile)

	require.NoError(t, err)
	assert.Empty(t, exports)
}

func TestParseUnsupportedFileType(t *testing.T) {
	tmpDir := t.TempDir()
	rbFile := filepath.Join(tmpDir, "test.rb")

	content := `require 'json'`
	require.NoError(t, os.WriteFile(rbFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(rbFile)

	require.NoError(t, err)
	assert.Empty(t, imports)
}

func TestParseExportsUnsupportedFileType(t *testing.T) {
	tmpDir := t.TempDir()
	rbFile := filepath.Join(tmpDir, "test.rb")

	content := `require 'json'`
	require.NoError(t, os.WriteFile(rbFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(rbFile)

	require.NoError(t, err)
	assert.Empty(t, exports)
}

func TestParseGoSingleImport(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

import "fmt"
import "strings"
import "github.com/user/repo"
`

	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(goFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))
}

func TestParseGoRelativeImport(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

import "./mypackage"
`

	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(goFile)

	require.NoError(t, err)
	require.Equal(t, 1, len(imports))
	assert.True(t, imports[0].IsRelative)
}

func TestParseGoFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/test.go")
	assert.Error(t, err)
}

func TestParseGoExportsNone(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := "package main\n"
	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(goFile)

	require.NoError(t, err)
	assert.Empty(t, exports)
}

func TestProcessCppComment(t *testing.T) {
	p := &Parser{}
	tests := []struct {
		name          string
		line          string
		inComment     bool
		wantInComment bool
	}{
		{"single line comment", "// comment", false, false},
		{"already in comment", "some code", true, true},
		{"inline block comment", "/* comment */", false, false},
		{"inline block comment in comment", "/* comment */", true, true},
		{"start multi-line comment", "/* starts", false, true},
		{"end multi-line comment", "ends */", true, false},
		{"not a comment", "int x = 5;", false, false},
		{"empty line", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.processCppComment(tt.line, tt.inComment)
			if got != tt.wantInComment {
				t.Errorf("processCppComment(%q, %v) = %v, want %v", tt.line, tt.inComment, got, tt.wantInComment)
			}
		})
	}
}

func TestExtractCppDeclarations(t *testing.T) {
	p := &Parser{}
	tests := []struct {
		name      string
		line      string
		wantNames []string
	}{
		{"class declaration", "class MyClass {}", []string{"MyClass"}},
		{"struct declaration", "struct MyStruct {}", []string{"MyStruct"}},
		{"namespace declaration", "namespace myapp {}", []string{"myapp"}},
		{"not a declaration", "int x = 5;", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classRegex := regexp.MustCompile(`^\s*(?:class|struct)\s+(\w+)`)
			namespaceRegex := regexp.MustCompile(`^\s*namespace\s+(\w+)`)
			var names []string
			p.extractCppDeclarations(tt.line, classRegex, namespaceRegex, &names)
			if len(names) != len(tt.wantNames) {
				t.Errorf("got %v, want %v", names, tt.wantNames)
				return
			}
			for i, name := range names {
				if name != tt.wantNames[i] {
					t.Errorf("names[%d] = %q, want %q", i, name, tt.wantNames[i])
				}
			}
		})
	}
}

func TestParsePythonRelativeImport(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `from .module import func
from ..parent import ParentClass
`

	require.NoError(t, os.WriteFile(pyFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(pyFile)

	require.NoError(t, err)
	assert.Equal(t, 2, len(imports))
	assert.True(t, imports[0].IsRelative)
}

func TestParseGoImportScannerError(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/file.go")
	assert.Error(t, err)
}

func TestParseTypeScriptImportOnly(t *testing.T) {
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `import 'polyfill';
`

	require.NoError(t, os.WriteFile(tsFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(tsFile)

	require.NoError(t, err)
	assert.Equal(t, 1, len(imports))
	assert.False(t, imports[0].IsRelative)
}

func TestParseCppExportsFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseExports("/nonexistent/test.hpp")
	assert.Error(t, err)
}

func TestParseTSFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/file.ts")
	assert.Error(t, err)
}

func TestParsePyFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/file.py")
	assert.Error(t, err)
}

func TestParseCsFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseFile("/nonexistent/file.cs")
	assert.Error(t, err)
}

func TestParseExportsTSFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseExports("/nonexistent/file.ts")
	assert.Error(t, err)
}

func TestParseExportsPyFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseExports("/nonexistent/file.py")
	assert.Error(t, err)
}

func TestParseExportsCsFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseExports("/nonexistent/file.cs")
	assert.Error(t, err)
}

func TestParseExportsJavaFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseExports("/nonexistent/Test.java")
	assert.Error(t, err)
}

func TestParseExportsGoFileNotFound(t *testing.T) {
	parser := New("/tmp")
	_, err := parser.ParseExports("/nonexistent/file.go")
	assert.Error(t, err)
}

func TestParseIgnoreDirectiveWithRulesNoReason(t *testing.T) {
	d := parseIgnoreDirective("ignore max-depth", 5)
	if d == nil {
		t.Fatal("expected directive")
	}
	if len(d.Rules) != 1 || d.Rules[0] != "max-depth" {
		t.Errorf("expected rules [max-depth], got %v", d.Rules)
	}
	if d.Reason != "no reason provided" {
		t.Errorf("expected default reason, got %q", d.Reason)
	}
}
