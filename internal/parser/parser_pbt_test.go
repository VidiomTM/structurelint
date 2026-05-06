package parser

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

func arbIdent(t *rapid.T, label string) string {
	for {
		name := rapid.StringMatching(`[a-z][a-z0-9_]{0,10}`).Draw(t, label)
		if !goKeywords[name] {
			return name
		}
	}
}

func arbGoSource(t *rapid.T) string {
	pkg := arbIdent(t, "pkg")
	numDecl := rapid.IntRange(0, 5).Draw(t, "numDecl")
	var decls []string
	for i := 0; i < numDecl; i++ {
		name := arbIdent(t, "declName")
		val := rapid.IntRange(-100, 100).Draw(t, "val")
		decls = append(decls, fmt.Sprintf("var %s = %d", name, val))
	}
	return fmt.Sprintf("package %s\n\n%s\n", pkg, strings.Join(decls, "\n"))
}

func arbGoSourceWithComments(t *rapid.T) string {
	pkg := arbIdent(t, "pkg")
	numDecl := rapid.IntRange(1, 5).Draw(t, "numDecl")
	var decls []string
	for i := 0; i < numDecl; i++ {
		name := arbIdent(t, "declName")
		val := rapid.IntRange(-100, 100).Draw(t, "val")
		comment := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9 ]{0,20}`).Draw(t, "comment")
		decls = append(decls, fmt.Sprintf("// %s\nvar %s = %d", comment, name, val))
	}
	return fmt.Sprintf("package %s // pkgdoc\n\n%s\n", pkg, strings.Join(decls, "\n"))
}

func arbGoSourceUnicode(t *rapid.T) string {
	pkg := rapid.OneOf(
		rapid.Just("main"),
		rapid.Just("api"),
		rapid.Just("util"),
	).Draw(t, "pkg")
	names := []string{"日本語", "Ñoño", "Über", "Δelta", "Café"}
	name := names[rapid.IntRange(0, len(names)-1).Draw(t, "nameIdx")]
	val := rapid.IntRange(-100, 100).Draw(t, "val")
	return fmt.Sprintf("package %s\n\nvar %s = %d\n", pkg, name, val)
}

func countComments(f *ast.File) int {
	n := 0
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if c != nil {
				n++
			}
		}
	}
	return n
}

func TestParseSafety(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		src := arbGoSource(t)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse error: %v\nsource:\n%s", err, src)
		}
		if f == nil {
			t.Fatal("nil AST")
		}
	})
}

func TestParseSafetyEmpty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		pkg := arbIdent(t, "pkg")
		src := fmt.Sprintf("package %s\n", pkg)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse error: %v\nsource:\n%s", err, src)
		}
		if f == nil {
			t.Fatal("nil AST")
		}
	})
}

func TestParseSafetyUnicode(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		src := arbGoSourceUnicode(t)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse error: %v\nsource:\n%s", err, src)
		}
		if f == nil {
			t.Fatal("nil AST")
		}
	})
}

func TestRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		src := arbGoSource(t)
		fset := token.NewFileSet()
		_, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("initial parse error: %v\nsource:\n%s", err, src)
		}

		formatted, err := format.Source([]byte(src))
		if err != nil {
			t.Fatalf("format error: %v\nsource:\n%s", err, src)
		}

		fset2 := token.NewFileSet()
		f2, err := parser.ParseFile(fset2, "", formatted, parser.ParseComments)
		if err != nil {
			t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
		}
		if f2 == nil {
			t.Fatal("nil AST after round-trip")
		}
	})
}

func TestCommentPreservation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		src := arbGoSourceWithComments(t)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse error: %v\nsource:\n%s", err, src)
		}
		if f == nil {
			t.Fatal("nil AST")
		}
		if countComments(f) == 0 {
			t.Fatalf("expected comments but found none\nsource:\n%s", src)
		}
	})
}
