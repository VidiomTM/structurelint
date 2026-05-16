package syntactic

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

func TestNewIndex(t *testing.T) {
	idx := NewIndex()
	if idx == nil {
		t.Fatal("NewIndex returned nil")
	}
	if idx.index == nil {
		t.Error("index map should not be nil")
	}
}

func TestIndex_Add(t *testing.T) {
	idx := NewIndex()
	s := types.Shingle{Hash: 42, FilePath: "a.go", StartToken: 0, EndToken: 19}
	idx.Add(s)
	candidates := idx.GetCandidates(42)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].FilePath != "a.go" {
		t.Errorf("FilePath = %q, want %q", candidates[0].FilePath, "a.go")
	}
}

func TestIndex_Add_MultipleSameHash(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 1, FilePath: "b.go"})
	candidates := idx.GetCandidates(1)
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
}

func TestIndex_AddBatch(t *testing.T) {
	idx := NewIndex()
	shingles := []types.Shingle{
		{Hash: 1, FilePath: "a.go"},
		{Hash: 2, FilePath: "b.go"},
		{Hash: 1, FilePath: "c.go"},
	}
	idx.AddBatch(shingles)
	if idx.Size() != 2 {
		t.Errorf("Size = %d, want 2", idx.Size())
	}
}

func TestIndex_GetCandidates_Nonexistent(t *testing.T) {
	idx := NewIndex()
	candidates := idx.GetCandidates(999)
	if candidates != nil {
		t.Error("expected nil for nonexistent hash")
	}
}

func TestIndex_FindCollisions(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 1, FilePath: "b.go"})
	idx.Add(types.Shingle{Hash: 2, FilePath: "c.go"})

	collisions := idx.FindCollisions()
	if len(collisions) != 1 {
		t.Fatalf("expected 1 collision, got %d", len(collisions))
	}
	shingles, ok := collisions[1]
	if !ok {
		t.Fatal("expected collision for hash 1")
	}
	if len(shingles) != 2 {
		t.Errorf("expected 2 shingles, got %d", len(shingles))
	}
}

func TestIndex_FindCollisions_NoCollisions(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 2, FilePath: "b.go"})

	collisions := idx.FindCollisions()
	if len(collisions) != 0 {
		t.Errorf("expected 0 collisions, got %d", len(collisions))
	}
}

func TestIndex_FindCrossFileCollisions(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 1, FilePath: "b.go"})
	idx.Add(types.Shingle{Hash: 2, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 2, FilePath: "a.go"})

	collisions := idx.FindCrossFileCollisions()
	if len(collisions) != 1 {
		t.Fatalf("expected 1 cross-file collision, got %d", len(collisions))
	}
}

func TestIndex_FindCrossFileCollisions_SameFile(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})

	collisions := idx.FindCrossFileCollisions()
	if len(collisions) != 0 {
		t.Errorf("expected 0 cross-file collisions for same file, got %d", len(collisions))
	}
}

func TestIndex_Stats(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Add(types.Shingle{Hash: 1, FilePath: "b.go"})
	idx.Add(types.Shingle{Hash: 2, FilePath: "c.go"})

	stats := idx.Stats()
	if stats.TotalHashes != 2 {
		t.Errorf("TotalHashes = %d, want 2", stats.TotalHashes)
	}
	if stats.TotalShingles != 3 {
		t.Errorf("TotalShingles = %d, want 3", stats.TotalShingles)
	}
	if stats.CollisionCount != 1 {
		t.Errorf("CollisionCount = %d, want 1", stats.CollisionCount)
	}
}

func TestIndex_Clear(t *testing.T) {
	idx := NewIndex()
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	idx.Clear()
	if idx.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", idx.Size())
	}
}

func TestIndex_Size(t *testing.T) {
	idx := NewIndex()
	if idx.Size() != 0 {
		t.Errorf("initial Size = %d, want 0", idx.Size())
	}
	idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
	if idx.Size() != 1 {
		t.Errorf("Size = %d, want 1", idx.Size())
	}
}

func TestIndex_ConcurrentSafe(t *testing.T) {
	idx := NewIndex()
	done := make(chan bool)
	go func() {
		idx.Add(types.Shingle{Hash: 1, FilePath: "a.go"})
		done <- true
	}()
	go func() {
		idx.GetCandidates(1)
		done <- true
	}()
	<-done
	<-done
}
