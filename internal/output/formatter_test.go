package output

import (
	"testing"
)

func TestGetFormatter_Text(t *testing.T) {
	f, err := GetFormatter("text", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(*TextFormatter); !ok {
		t.Error("expected TextFormatter")
	}
}

func TestGetFormatter_Empty(t *testing.T) {
	f, err := GetFormatter("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(*TextFormatter); !ok {
		t.Error("expected TextFormatter for empty format")
	}
}

func TestGetFormatter_JSON(t *testing.T) {
	f, err := GetFormatter("json", "1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	jf, ok := f.(*JSONFormatter)
	if !ok {
		t.Fatal("expected JSONFormatter")
	}
	if jf.Version != "1.0" {
		t.Errorf("Version = %s, want 1.0", jf.Version)
	}
}

func TestGetFormatter_JUnit(t *testing.T) {
	f, err := GetFormatter("junit", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(*JUnitFormatter); !ok {
		t.Error("expected JUnitFormatter")
	}
}

func TestGetFormatter_JUnitXML(t *testing.T) {
	f, err := GetFormatter("junit-xml", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(*JUnitFormatter); !ok {
		t.Error("expected JUnitFormatter")
	}
}

func TestGetFormatter_Unknown(t *testing.T) {
	_, err := GetFormatter("unknown", "")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestTextFormatter_Empty(t *testing.T) {
	f := &TextFormatter{}
	s, err := f.Format(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "" {
		t.Errorf("expected empty string, got %s", s)
	}
}

func TestJSONFormatter_Empty(t *testing.T) {
	f := &JSONFormatter{Version: "1.0"}
	s, err := f.Format(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == "" {
		t.Error("expected non-empty JSON output")
	}
}
