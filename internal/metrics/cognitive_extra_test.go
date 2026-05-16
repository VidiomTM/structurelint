package metrics

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCognitiveVisitor_VisitNil(t *testing.T) {
	v := &cognitiveVisitor{}
	v.visit(nil, 0)
	if v.complexity != 0 {
		t.Errorf("Expected complexity 0 for nil visit, got %d", v.complexity)
	}
}

func TestProcessSwitchCases_NilBody(t *testing.T) {
	v := &cognitiveVisitor{}
	v.processSwitchCases(nil, 0)
	if v.complexity != 0 {
		t.Errorf("Expected complexity 0 for nil body, got %d", v.complexity)
	}
}

func TestProcessSelectCases_NilBody(t *testing.T) {
	v := &cognitiveVisitor{}
	v.processSelectCases(nil, 0)
	if v.complexity != 0 {
		t.Errorf("Expected complexity 0 for nil body, got %d", v.complexity)
	}
}

func TestCognitiveComplexityAnalyzer_Name(t *testing.T) {
	a := NewCognitiveComplexityAnalyzer()
	if got := a.Name(); got != "cognitive-complexity" {
		t.Errorf("Name() = %q, want %q", got, "cognitive-complexity")
	}
}

func TestCognitiveComplexityAnalyzer_AnalyzeFunction_NilBody(t *testing.T) {
	code := `package main; func foo()`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			a := NewCognitiveComplexityAnalyzer()
			m := a.AnalyzeFunction(fn)
			if m.Value != 0 {
				t.Errorf("Expected 0 for nil body, got %v", m.Value)
			}
		}
	}
}

func TestCognitiveComplexity_ForLoop_CStyle(t *testing.T) {
	code := `
package main

func cStyleFor(n int) {
	for i := 0; i < n; i++ {
		println(i)
	}
}
`
	complexity := analyzeCode(t, code)
	// +1 for = 1
	if complexity != 1 {
		t.Errorf("Expected complexity 1 for C-style for, got %d", complexity)
	}
}

func TestCognitiveComplexity_TypeSwitch(t *testing.T) {
	code := `
package main

func typeSwitch(x interface{}) {
	switch v := x.(type) {
	case int:
		println(v)
	default:
		println("other")
	}
}
`
	complexity := analyzeCode(t, code)
	// +1 switch +1 case clause = 2
	if complexity != 2 {
		t.Errorf("Expected complexity 2 for type switch, got %d", complexity)
	}
}

func TestCognitiveComplexity_Select(t *testing.T) {
	code := `
package main

func selectStmt(ch chan int) {
	select {
	case msg := <-ch:
		println(msg)
	default:
		println("default")
	}
}
`
	complexity := analyzeCode(t, code)
	// +1 select +1 case clause = 2
	if complexity != 2 {
		t.Errorf("Expected complexity 2 for select, got %d", complexity)
	}
}

func TestCognitiveComplexity_GoStmt(t *testing.T) {
	code := `
package main

func goStmt() {
	go func() {
		println("hello")
	}()
}
`
	complexity := analyzeCode(t, code)
	// +1 go (nesting 0) = 1
	if complexity != 1 {
		t.Errorf("Expected complexity 1 for go, got %d", complexity)
	}
}

func TestCognitiveComplexity_NestedSelectInFor(t *testing.T) {
	code := `
package main

func nestedSelect(ch chan int) {
	for i := 0; i < 10; i++ {
		select {
		case <-ch:
			println("got")
		default:
			println("none")
		}
	}
}
`
	complexity := analyzeCode(t, code)
	// +1 for +2 select(nesting1) +1 case(nesting2) = 4
	if complexity != 4 {
		t.Errorf("Expected complexity 4 for nested select, got %d", complexity)
	}
}

func TestCognitiveComplexity_DeferStmt(t *testing.T) {
	code := `
package main

func withDefer() {
	defer func() { println("cleanup") }()
}
`
	complexity := analyzeCode(t, code)
	if complexity != 0 {
		t.Errorf("Expected complexity 0 for defer, got %d", complexity)
	}
}

func TestCognitiveComplexity_GoInNestedIf(t *testing.T) {
	code := `
package main

func goInIf(x int) {
	if x > 0 {
		go func() { println("hi") }()
	}
}
`
	complexity := analyzeCode(t, code)
	// +1 if +2 go(nesting1) = 3
	if complexity != 3 {
		t.Errorf("Expected complexity 3 for go in nested if, got %d", complexity)
	}
}

func TestCognitiveComplexity_MethodReceiver(t *testing.T) {
	code := `
package main

type S struct{}

func (s S) method() {
	if true {
		println("yes")
	}
}
`
	complexity := analyzeCode(t, code)
	if complexity != 1 {
		t.Errorf("Expected complexity 1 for method with if, got %d", complexity)
	}
}

func TestCognitiveComplexity_DeclStmt(t *testing.T) {
	code := `
package main

func withVarDecl() {
	var x = 42
	println(x)
}
`
	complexity := analyzeCode(t, code)
	if complexity != 0 {
		t.Errorf("Expected complexity 0 for var decl, got %d", complexity)
	}
}

func TestCognitiveComplexity_AssignExprReturnDeclStmt(t *testing.T) {
	code := `
package main

func assignExpr() int {
	x := func() int { return 1 }()
	return x
}
`
	complexity := analyzeCode(t, code)
	// Func literal body walks inner return: no complexity increment
	if complexity != 0 {
		t.Errorf("Expected complexity 0 for func literal, got %d", complexity)
	}
}

func TestHalsteadAnalyzer_Name(t *testing.T) {
	a := NewHalsteadAnalyzer()
	if got := a.Name(); got != "halstead" {
		t.Errorf("Name() = %q, want %q", got, "halstead")
	}
}

func TestHalsteadAnalyzer_AnalyzeFunction_NilBody(t *testing.T) {
	code := `package main; func foo()`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			a := NewHalsteadAnalyzer()
			m := a.AnalyzeFunction(fn)
			if m.Value != 0 {
				t.Errorf("Expected 0 for nil body, got %v", m.Value)
			}
		}
	}
}

func TestHalstead_TypeSwitch(t *testing.T) {
	code := `
package main

func handle(v interface{}) {
	switch t := v.(type) {
	case int:
		println(t)
	case string:
		println(t)
	}
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
	if metrics.FileLevel["function_count"] != 1 {
		t.Errorf("Expected 1 function, got %v", metrics.FileLevel["function_count"])
	}
}

func TestHalstead_Select(t *testing.T) {
	code := `
package main

func selectOp(ch chan int) {
	select {
	case <-ch:
		println("received")
	}
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_ForLoopCStyle(t *testing.T) {
	code := `
package main

func loop() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_SendStmt(t *testing.T) {
	code := `
package main

func send(ch chan int) {
	ch <- 42
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_StarExpr(t *testing.T) {
	code := `
package main

func deref(p *int) int {
	return *p
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_TypeAssertExpr(t *testing.T) {
	code := `
package main

func assert(x interface{}) int {
	return x.(int)
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_GoDeferBranch(t *testing.T) {
	code := `
package main

func goDefer() {
	defer println("done")
	go println("async")
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_IndexSliceExpr(t *testing.T) {
	code := `
package main

func sliceIdx(s []int) int {
	return s[1:3][0]
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestHalstead_DefaultCase(t *testing.T) {
	code := `
package main

func classify(x int) string {
	switch x {
	default:
		return "unknown"
	}
}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	analyzer := NewHalsteadAnalyzer()
	metrics := analyzer.AnalyzeFile(node)
	if len(metrics.Functions) == 0 {
		t.Fatal("Expected at least one function")
	}
}

func TestCognitiveComplexity_MethodReceiverName(t *testing.T) {
	code := `
package main

type T struct{}

func (t T) m() {}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			name := getFunctionName(fn)
			if name != "(t).m" {
				t.Errorf("getFunctionName = %q, want %q", name, "(t).m")
			}
		}
	}
}

func TestHalstead_MethodReceiverName(t *testing.T) {
	code := `
package main

type T struct{}

func (t T) m() {}
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			name := getFunctionNameHalstead(fn)
			if name != "(t).m" {
				t.Errorf("getFunctionNameHalstead = %q, want %q", name, "(t).m")
			}
		}
	}
}
