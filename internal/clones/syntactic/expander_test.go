package syntactic

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

func TestNewExpander(t *testing.T) {
	e := NewExpander()
	if e == nil {
		t.Fatal("NewExpander returned nil")
	}
	if e.tokenCache == nil {
		t.Error("tokenCache should not be nil")
	}
}

func TestExpandClone_Basic(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{
		{Value: "func", Line: 1},
		{Value: "_ID_", Line: 1},
		{Value: "(", Line: 1},
		{Value: "_ID_", Line: 1},
		{Value: "_ID_", Line: 1},
		{Value: ")", Line: 1},
		{Value: "{", Line: 1},
		{Value: "return", Line: 2},
		{Value: "_LIT_", Line: 2},
		{Value: "}", Line: 3},
	}
	e.SetTokenCache(map[string][]types.Token{
		"a.go": tokens,
		"b.go": tokens,
	})

	s1 := types.Shingle{Hash: 123, FilePath: "a.go", StartToken: 0, EndToken: 9}
	s2 := types.Shingle{Hash: 123, FilePath: "b.go", StartToken: 0, EndToken: 9}

	clone := e.ExpandClone(s1, s2)
	if clone == nil {
		t.Fatal("ExpandClone returned nil")
	}
	if clone.TokenCount != 10 {
		t.Errorf("TokenCount = %d, want 10", clone.TokenCount)
	}
	if clone.Type != types.Type2 {
		t.Errorf("Type = %v, want Type2", clone.Type)
	}
	if len(clone.Locations) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(clone.Locations))
	}
}

func TestExpandClone_Mismatch(t *testing.T) {
	e := NewExpander()
	tokens1 := []types.Token{
		{Value: "func", Line: 1},
		{Value: "_ID_", Line: 1},
	}
	tokens2 := []types.Token{
		{Value: "return", Line: 1},
		{Value: "_LIT_", Line: 1},
	}
	e.SetTokenCache(map[string][]types.Token{
		"a.go": tokens1,
		"b.go": tokens2,
	})

	s1 := types.Shingle{Hash: 123, FilePath: "a.go", StartToken: 0, EndToken: 1}
	s2 := types.Shingle{Hash: 123, FilePath: "b.go", StartToken: 0, EndToken: 1}

	clone := e.ExpandClone(s1, s2)
	if clone != nil {
		t.Error("expected nil for mismatched tokens")
	}
}

func TestExpandClone_MissingFile(t *testing.T) {
	e := NewExpander()
	e.SetTokenCache(map[string][]types.Token{
		"a.go": {{Value: "func"}},
	})

	s1 := types.Shingle{Hash: 123, FilePath: "a.go", StartToken: 0, EndToken: 0}
	s2 := types.Shingle{Hash: 123, FilePath: "missing.go", StartToken: 0, EndToken: 0}

	clone := e.ExpandClone(s1, s2)
	if clone != nil {
		t.Error("expected nil for missing file")
	}
}

func TestExpandClone_OutOfBounds(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{{Value: "func", Line: 1}}
	e.SetTokenCache(map[string][]types.Token{
		"a.go": tokens,
		"b.go": tokens,
	})

	s1 := types.Shingle{Hash: 123, FilePath: "a.go", StartToken: 0, EndToken: 100}
	s2 := types.Shingle{Hash: 123, FilePath: "b.go", StartToken: 0, EndToken: 0}

	clone := e.ExpandClone(s1, s2)
	if clone != nil {
		t.Error("expected nil for out-of-bounds")
	}
}

func TestExpandAllCollisions_Empty(t *testing.T) {
	e := NewExpander()
	clones := e.ExpandAllCollisions(nil)
	if len(clones) != 0 {
		t.Errorf("expected 0 clones, got %d", len(clones))
	}
}

func TestExpandAllCollisions_NoMatches(t *testing.T) {
	e := NewExpander()
	collisions := map[uint64][]types.Shingle{
		1: {
			{Hash: 1, FilePath: "a.go", StartToken: 0, EndToken: 0},
			{Hash: 1, FilePath: "b.go", StartToken: 0, EndToken: 0},
		},
	}
	clones := e.ExpandAllCollisions(collisions)
	if len(clones) != 0 {
		t.Errorf("expected 0 clones (no token cache), got %d", len(clones))
	}
}

func TestExpandAllCollisions_WithDuplicates(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{{Value: "func", Line: 1}}
	e.SetTokenCache(map[string][]types.Token{
		"a.go": tokens,
		"b.go": tokens,
		"c.go": tokens,
	})

	collisions := map[uint64][]types.Shingle{
		1: {
			{Hash: 1, FilePath: "a.go", StartToken: 0, EndToken: 0},
			{Hash: 1, FilePath: "b.go", StartToken: 0, EndToken: 0},
			{Hash: 1, FilePath: "c.go", StartToken: 0, EndToken: 0},
			// Duplicate of a.go — should be deduped
			{Hash: 1, FilePath: "a.go", StartToken: 0, EndToken: 0},
		},
	}
	clones := e.ExpandAllCollisions(collisions)
	if len(clones) != 4 {
		t.Errorf("expected 4 clones (a-b, a-c, a-a, b-c), got %d", len(clones))
	}
}

func TestClonePairKey_Ordering(t *testing.T) {
	e := NewExpander()
	k1 := e.clonePairKey(
		types.Shingle{FilePath: "a.go", StartToken: 0},
		types.Shingle{FilePath: "b.go", StartToken: 0},
	)
	k2 := e.clonePairKey(
		types.Shingle{FilePath: "b.go", StartToken: 0},
		types.Shingle{FilePath: "a.go", StartToken: 0},
	)
	if k1 != k2 {
		t.Error("clonePairKey should be symmetric")
	}
}

func TestExpandBackward(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{
		{Value: "package", Line: 1},
		{Value: "_ID_", Line: 1},
		{Value: "func", Line: 3},
		{Value: "_ID_", Line: 3},
		{Value: "return", Line: 4},
		{Value: "_LIT_", Line: 4},
	}
	start1, start2 := e.expandBackward(tokens, tokens, 5, 5)
	if start1 != 0 || start2 != 0 {
		t.Errorf("expandBackward(5,5) = (%d,%d), want (0,0)", start1, start2)
	}
}

func TestExpandForward(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{
		{Value: "func", Line: 1},
		{Value: "_ID_", Line: 1},
		{Value: "return", Line: 2},
		{Value: "_ID_", Line: 2},
	}
	end1, end2 := e.expandForward(tokens, tokens, 0, 0)
	if end1 != 3 || end2 != 3 {
		t.Errorf("expandForward(0,0) = (%d,%d), want (3,3)", end1, end2)
	}
}

func TestExpandBackward_StopsOnMismatch(t *testing.T) {
	e := NewExpander()
	tokens1 := []types.Token{{Value: "a", Line: 1}, {Value: "b", Line: 1}, {Value: "c", Line: 1}}
	tokens2 := []types.Token{{Value: "x", Line: 1}, {Value: "b", Line: 1}, {Value: "c", Line: 1}}
	start1, start2 := e.expandBackward(tokens1, tokens2, 2, 2)
	if start1 != 1 || start2 != 1 {
		t.Errorf("expandBackward(2,2) = (%d,%d), want (1,1)", start1, start2)
	}
}

func TestDetermineCloneType(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{{Value: "func"}, {Value: "_ID_"}, {Value: "return"}}
	ct := e.determineCloneType(tokens, tokens)
	if ct != types.Type2 {
		t.Errorf("determineCloneType = %v, want Type2", ct)
	}
}

func TestSetTokenCache(t *testing.T) {
	e := NewExpander()
	cache := map[string][]types.Token{"x.go": {{Value: "test"}}}
	e.SetTokenCache(cache)
	if len(e.tokenCache) != 1 {
		t.Error("tokenCache not set correctly")
	}
}

func TestExpandBackward_AtStart(t *testing.T) {
	e := NewExpander()
	tokens := []types.Token{{Value: "a", Line: 1}, {Value: "b", Line: 1}}
	start1, start2 := e.expandBackward(tokens, tokens, 0, 0)
	if start1 != 0 || start2 != 0 {
		t.Errorf("expandBackward at start = (%d,%d), want (0,0)", start1, start2)
	}
}
