package rules

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

type passRule struct{ name string }

func (r *passRule) Name() string { return r.name }
func (r *passRule) Check([]walker.FileInfo, map[string]*walker.DirInfo) []Violation { return nil }

type failRule struct{ name string }

func (r *failRule) Name() string { return r.name }
func (r *failRule) Check([]walker.FileInfo, map[string]*walker.DirInfo) []Violation {
	return []Violation{{Rule: r.name, Path: ".", Message: "failed"}}
}

func TestCompositeRule_AND(t *testing.T) {
	r := NewCompositeRule("test-and", "all must pass", OperatorAND, &passRule{"a"}, &passRule{"b"})
	assertNoViolations(t, r.Check(nil, nil))

	r = NewCompositeRule("test-and", "all must pass", OperatorAND, &passRule{"a"}, &failRule{"b"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestCompositeRule_OR(t *testing.T) {
	r := NewCompositeRule("test-or", "at least one", OperatorOR, &passRule{"a"}, &failRule{"b"})
	assertNoViolations(t, r.Check(nil, nil))

	r = NewCompositeRule("test-or", "at least one", OperatorOR, &failRule{"a"}, &failRule{"b"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestCompositeRule_NOT(t *testing.T) {
	r := NewCompositeRule("test-not", "invert", OperatorNOT, &failRule{"a"})
	assertNoViolations(t, r.Check(nil, nil))

	r = NewCompositeRule("test-not", "invert", OperatorNOT, &passRule{"a"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestCompositeRule_NOT_NoRules(t *testing.T) {
	r := NewCompositeRule("test-not", "invert", OperatorNOT)
	v := r.Check(nil, nil)
	assertHasViolations(t, v)
	if len(v) > 0 && v[0].Message != "NOT operator requires at least one rule" {
		t.Errorf("unexpected message: %s", v[0].Message)
	}
}

func TestCompositeRule_XOR(t *testing.T) {
	r := NewCompositeRule("test-xor", "exactly one", OperatorXOR, &passRule{"a"}, &failRule{"b"})
	assertNoViolations(t, r.Check(nil, nil))

	r = NewCompositeRule("test-xor", "exactly one", OperatorXOR, &passRule{"a"}, &passRule{"b"})
	assertHasViolations(t, r.Check(nil, nil))

	r = NewCompositeRule("test-xor", "exactly one", OperatorXOR, &failRule{"a"}, &failRule{"b"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestCompositeRule_UnknownOperator(t *testing.T) {
	r := NewCompositeRule("test-unknown", "unknown", CompositeOperator(99))
	v := r.Check(nil, nil)
	assertHasViolations(t, v)
	if len(v) > 0 && v[0].Message != "Unknown composite operator: 99" {
		t.Errorf("unexpected message: %s", v[0].Message)
	}
}

func TestCompositeRule_Name(t *testing.T) {
	r := NewCompositeRule("my-rule", "desc", OperatorAND)
	if r.Name() != "my-rule" {
		t.Errorf("Name() = %s, want my-rule", r.Name())
	}
}

func TestAllOf(t *testing.T) {
	r := AllOf("all-of", "all must pass", &passRule{"a"}, &passRule{"b"})
	assertNoViolations(t, r.Check(nil, nil))

	r = AllOf("all-of", "all must pass", &passRule{"a"}, &failRule{"b"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestAnyOf(t *testing.T) {
	r := AnyOf("any-of", "at least one", &passRule{"a"}, &failRule{"b"})
	assertNoViolations(t, r.Check(nil, nil))

	r = AnyOf("any-of", "at least one", &failRule{"a"}, &failRule{"b"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestNotRule(t *testing.T) {
	r := NotRule("not-rule", "invert", &failRule{"a"})
	assertNoViolations(t, r.Check(nil, nil))

	r = NotRule("not-rule", "invert", &passRule{"a"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestExactlyOneOf(t *testing.T) {
	r := ExactlyOneOf("xor-rule", "exactly one", &passRule{"a"}, &failRule{"b"})
	assertNoViolations(t, r.Check(nil, nil))
}

func TestConditionalRule_ConditionMet(t *testing.T) {
	cond := func(files []walker.FileInfo, dirs map[string]*walker.DirInfo) bool { return true }
	r := NewConditionalRule("cond", cond, &failRule{"inner"})
	assertHasViolations(t, r.Check(nil, nil))
}

func TestConditionalRule_ConditionNotMet(t *testing.T) {
	cond := func(files []walker.FileInfo, dirs map[string]*walker.DirInfo) bool { return false }
	r := NewConditionalRule("cond", cond, &failRule{"inner"})
	assertNoViolations(t, r.Check(nil, nil))
}

func TestConditionalRule_Name(t *testing.T) {
	r := NewConditionalRule("cond-rule", nil, nil)
	if r.Name() != "cond-rule" {
		t.Errorf("Name() = %s, want cond-rule", r.Name())
	}
}

func TestIfProjectHas(t *testing.T) {
	files := []walker.FileInfo{{Path: "src/main.go"}}
	inner := &failRule{"inner"}
	r := IfProjectHas("main.go", inner)
	assertHasViolations(t, r.Check(files, nil))

	r = IfProjectHas("nonexistent.go", inner)
	assertNoViolations(t, r.Check(files, nil))
}

func TestIfProjectLanguage(t *testing.T) {
	files := []walker.FileInfo{{Path: "main.go"}}
	inner := &failRule{"inner"}
	r := IfProjectLanguage(".go", inner)
	assertHasViolations(t, r.Check(files, nil))

	r = IfProjectLanguage(".py", inner)
	assertNoViolations(t, r.Check(files, nil))
}

func assertNoViolations(t *testing.T, v []Violation) {
	t.Helper()
	if len(v) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(v), v)
	}
}

func assertHasViolations(t *testing.T, v []Violation) {
	t.Helper()
	if len(v) == 0 {
		t.Error("expected violations, got none")
	}
}
