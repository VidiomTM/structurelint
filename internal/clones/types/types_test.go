package types

import (
	"testing"
)

func TestCloneType_String(t *testing.T) {
	tests := []struct {
		ct       CloneType
		expected string
	}{
		{Type1, "Type-1 (exact copy)"},
		{Type2, "Type-2 (renamed)"},
		{Type3, "Type-3 (modified)"},
		{Type4, "Type-4 (semantic)"},
		{CloneType(99), "Unknown"},
	}
	for _, tc := range tests {
		if got := tc.ct.String(); got != tc.expected {
			t.Errorf("CloneType(%d).String() = %q, want %q", tc.ct, got, tc.expected)
		}
	}
}

func TestTokenType_Constants(t *testing.T) {
	if TokenKeyword != 0 {
		t.Errorf("TokenKeyword = %d, want 0", TokenKeyword)
	}
	if TokenIdentifier != 1 {
		t.Errorf("TokenIdentifier = %d, want 1", TokenIdentifier)
	}
	if TokenLiteral != 2 {
		t.Errorf("TokenLiteral = %d, want 2", TokenLiteral)
	}
	if TokenOperator != 3 {
		t.Errorf("TokenOperator = %d, want 3", TokenOperator)
	}
	if TokenPunctuation != 4 {
		t.Errorf("TokenPunctuation = %d, want 4", TokenPunctuation)
	}
}

func TestClone_Fields(t *testing.T) {
	c := &Clone{
		Type:       Type2,
		TokenCount: 42,
		LineCount:  10,
		Hash:       12345,
		Similarity: 0.95,
		Locations: []Location{
			{FilePath: "a.go", StartLine: 1, EndLine: 10, StartToken: 0, EndToken: 41},
		},
	}
	if c.Type != Type2 || c.TokenCount != 42 || c.LineCount != 10 {
		t.Error("Clone fields not set correctly")
	}
	if c.Hash != 12345 || c.Similarity != 0.95 {
		t.Error("Clone fields not set correctly")
	}
	if len(c.Locations) != 1 {
		t.Error("expected 1 location")
	}
}

func TestFileTokens(t *testing.T) {
	ft := &FileTokens{
		FilePath: "test.go",
		Tokens: []Token{
			{Type: TokenKeyword, Value: "package", Line: 1, Position: 0},
			{Type: TokenIdentifier, Value: "_ID_", Line: 1, Position: 1},
		},
	}
	if ft.FilePath != "test.go" || len(ft.Tokens) != 2 {
		t.Error("FileTokens fields not set correctly")
	}
	if ft.Tokens[0].Value != "package" {
		t.Error("Token value incorrect")
	}
}

func TestToken_Position(t *testing.T) {
	tok := Token{
		Type:     TokenOperator,
		Value:    "+",
		Line:     5,
		Column:   10,
		Position: 3,
	}
	if tok.Line != 5 || tok.Column != 10 || tok.Position != 3 {
		t.Error("Token position fields incorrect")
	}
}

func TestLocation_Fields(t *testing.T) {
	loc := Location{
		FilePath:   "src/main.go",
		StartLine:  10,
		EndLine:    20,
		StartToken: 5,
		EndToken:   15,
	}
	if loc.FilePath != "src/main.go" || loc.StartLine != 10 || loc.EndLine != 20 {
		t.Error("Location fields incorrect")
	}
	if loc.StartToken != 5 || loc.EndToken != 15 {
		t.Error("Location token fields incorrect")
	}
}

func TestShingle_Fields(t *testing.T) {
	s := Shingle{
		Hash:       98765,
		StartToken: 0,
		EndToken:   19,
		FilePath:   "a.go",
		StartLine:  1,
		EndLine:    5,
	}
	if s.Hash != 98765 || s.StartToken != 0 || s.EndToken != 19 {
		t.Error("Shingle fields incorrect")
	}
	if s.FilePath != "a.go" || s.StartLine != 1 || s.EndLine != 5 {
		t.Error("Shingle fields incorrect")
	}
}

func TestClonePair_Fields(t *testing.T) {
	locA := Location{FilePath: "a.go", StartLine: 1, EndLine: 5}
	locB := Location{FilePath: "b.go", StartLine: 10, EndLine: 15}
	clone := &Clone{
		Type: Type1, TokenCount: 10, LineCount: 5,
		Locations: []Location{locA, locB},
	}
	cp := ClonePair{
		LocationA:  locA,
		LocationB:  locB,
		Clone:      clone,
		Confidence: 1.0,
	}
	if cp.LocationA.FilePath != "a.go" || cp.Confidence != 1.0 {
		t.Error("ClonePair fields incorrect")
	}
	if cp.Clone.TokenCount != 10 {
		t.Error("ClonePair Clone field incorrect")
	}
}

func TestTokenType_RoundTrip(t *testing.T) {
	types := []TokenType{TokenKeyword, TokenIdentifier, TokenLiteral, TokenOperator, TokenPunctuation}
	for _, tt := range types {
		switch tt {
		case TokenKeyword, TokenIdentifier, TokenLiteral, TokenOperator, TokenPunctuation:
		default:
			t.Errorf("unexpected token type %d", tt)
		}
	}
}
