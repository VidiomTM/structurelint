package syntactic

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

func TestNewHasher(t *testing.T) {
	h := NewHasher(DefaultKGramSize)
	if h == nil {
		t.Fatal("NewHasher returned nil")
	}
	if h.kGramSize != DefaultKGramSize {
		t.Errorf("kGramSize = %d, want %d", h.kGramSize, DefaultKGramSize)
	}
}

func TestNewHasher_ZeroKGramSize(t *testing.T) {
	h := NewHasher(0)
	if h.kGramSize != DefaultKGramSize {
		t.Errorf("kGramSize = %d, want %d", h.kGramSize, DefaultKGramSize)
	}
}

func TestNewHasher_NegativeKGramSize(t *testing.T) {
	h := NewHasher(-5)
	if h.kGramSize != DefaultKGramSize {
		t.Errorf("kGramSize = %d, want %d", h.kGramSize, DefaultKGramSize)
	}
}

func TestGenerateShingles_SmallFile(t *testing.T) {
	h := NewHasher(20)
	tokens := make([]types.Token, 5)
	for i := range tokens {
		tokens[i] = types.Token{Value: "_ID_"}
	}
	ft := &types.FileTokens{FilePath: "small.go", Tokens: tokens}
	shingles := h.GenerateShingles(ft)
	if shingles != nil {
		t.Error("expected nil for file smaller than kGramSize")
	}
}

func TestGenerateShingles_ExactMatch(t *testing.T) {
	h := NewHasher(3)
	tokens := make([]types.Token, 3)
	for i := range tokens {
		tokens[i] = types.Token{Value: "_ID_", Line: i + 1}
	}
	ft := &types.FileTokens{FilePath: "exact.go", Tokens: tokens}
	shingles := h.GenerateShingles(ft)
	if len(shingles) != 1 {
		t.Fatalf("expected 1 shingle, got %d", len(shingles))
	}
	if shingles[0].StartToken != 0 {
		t.Errorf("StartToken = %d, want 0", shingles[0].StartToken)
	}
}

func TestGenerateShingles_Multiple(t *testing.T) {
	h := NewHasher(3)
	tokens := make([]types.Token, 10)
	for i := range tokens {
		tokens[i] = types.Token{Value: "_ID_", Line: i + 1}
	}
	ft := &types.FileTokens{FilePath: "multi.go", Tokens: tokens}
	shingles := h.GenerateShingles(ft)
	expected := 8
	if len(shingles) != expected {
		t.Errorf("expected %d shingles, got %d", expected, len(shingles))
	}
}

func TestHashTokens_Deterministic(t *testing.T) {
	h := NewHasher(3)
	tokens := []types.Token{
		{Value: "func"},
		{Value: "_ID_"},
		{Value: "return"},
	}
	h1 := h.hashTokens(tokens)
	h2 := h.hashTokens(tokens)
	if h1 != h2 {
		t.Error("hashTokens is not deterministic")
	}
}

func TestHashTokens_DifferentInputs(t *testing.T) {
	h := NewHasher(3)
	tokens1 := []types.Token{{Value: "func"}, {Value: "_ID_"}, {Value: "return"}}
	tokens2 := []types.Token{{Value: "func"}, {Value: "_ID_"}, {Value: "if"}}
	h1 := h.hashTokens(tokens1)
	h2 := h.hashTokens(tokens2)
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestRollingHash(t *testing.T) {
	h := NewHasher(3)
	removed := types.Token{Value: "func"}
	added := types.Token{Value: "if"}
	prevHash := uint64(12345)
	result := h.rollingHash(prevHash, removed, added, 3)
	if result == 0 {
		t.Error("rolling hash should not be zero")
	}
}

func TestHashKGram(t *testing.T) {
	h := NewHasher(3)
	tokens := []types.Token{{Value: "package"}, {Value: "_ID_"}, {Value: "_LIT_"}}
	hash := h.HashKGram(tokens)
	if hash == 0 {
		t.Error("HashKGram should not return 0")
	}
}

func TestVerifyShingle_Valid(t *testing.T) {
	h := NewHasher(3)
	tokens := []types.Token{
		{Value: "func"}, {Value: "_ID_"}, {Value: "return"},
		{Value: "_LIT_"}, {Value: "_ID_"},
	}
	ft := &types.FileTokens{FilePath: "verify.go", Tokens: tokens}
	shingles := h.GenerateShingles(ft)
	if len(shingles) == 0 {
		t.Fatal("expected shingles")
	}
	if !h.VerifyShingle(shingles[0], tokens) {
		t.Error("shingle should verify correctly")
	}
}

func TestVerifyShingle_OutOfBounds(t *testing.T) {
	h := NewHasher(3)
	tokens := []types.Token{{Value: "a"}, {Value: "b"}, {Value: "c"}}
	s := types.Shingle{StartToken: 0, EndToken: 10, Hash: 123}
	if h.VerifyShingle(s, tokens) {
		t.Error("should fail for out-of-bounds shingle")
	}
}

func TestVerifyShingle_Mismatch(t *testing.T) {
	h := NewHasher(3)
	tokens := []types.Token{{Value: "a"}, {Value: "b"}, {Value: "c"}}
	s := types.Shingle{StartToken: 0, EndToken: 2, Hash: 999}
	if h.VerifyShingle(s, tokens) {
		t.Error("should fail for hash mismatch")
	}
}

func TestGenerateShingles_EmptyTokens(t *testing.T) {
	h := NewHasher(5)
	ft := &types.FileTokens{FilePath: "empty.go", Tokens: nil}
	shingles := h.GenerateShingles(ft)
	if shingles != nil {
		t.Error("expected nil for empty tokens")
	}
}

func TestGenerateShingles_ShinglePositions(t *testing.T) {
	h := NewHasher(2)
	tokens := make([]types.Token, 4)
	for i := range tokens {
		tokens[i] = types.Token{Value: "_ID_", Line: i + 1}
	}
	ft := &types.FileTokens{FilePath: "pos.go", Tokens: tokens}
	shingles := h.GenerateShingles(ft)
	if len(shingles) != 3 {
		t.Fatalf("expected 3 shingles, got %d", len(shingles))
	}
	if shingles[0].StartToken != 0 || shingles[0].EndToken != 1 {
		t.Errorf("first shingle: start=%d end=%d", shingles[0].StartToken, shingles[0].EndToken)
	}
	if shingles[1].StartToken != 1 || shingles[1].EndToken != 2 {
		t.Errorf("second shingle: start=%d end=%d", shingles[1].StartToken, shingles[1].EndToken)
	}
	if shingles[2].StartToken != 2 || shingles[2].EndToken != 3 {
		t.Errorf("third shingle: start=%d end=%d", shingles[2].StartToken, shingles[2].EndToken)
	}
}
