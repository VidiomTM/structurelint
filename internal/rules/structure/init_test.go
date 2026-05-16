package structure

import "testing"

func TestParseMaxDepthOverrides_Nil(t *testing.T) {
	overrides := parseMaxDepthOverrides(nil)
	if len(overrides) != 0 {
		t.Errorf("expected nil returns empty, got %d", len(overrides))
	}
}

func TestParseMaxDepthOverrides_Map(t *testing.T) {
	overrides := parseMaxDepthOverrides(map[string]interface{}{
		"src/**": 8,
		"tests/**": 6,
	})
	if len(overrides) != 2 {
		t.Fatalf("expected 2 overrides, got %d", len(overrides))
	}
	// order is non-deterministic, check by pattern
	m := map[string]int{}
	for _, o := range overrides {
		m[o.Pattern] = o.Max
	}
	if m["src/**"] != 8 {
		t.Errorf("src/** max = %d, want 8", m["src/**"])
	}
	if m["tests/**"] != 6 {
		t.Errorf("tests/** max = %d, want 6", m["tests/**"])
	}
}

func TestParseMaxDepthOverrides_MapFloat(t *testing.T) {
	overrides := parseMaxDepthOverrides(map[string]interface{}{
		"src/**": float64(8),
	})
	if len(overrides) != 1 || overrides[0].Max != 8 {
		t.Errorf("expected 1 override with max=8, got %v", overrides)
	}
}

func TestParseMaxDepthOverrides_MapZeroValue(t *testing.T) {
	overrides := parseMaxDepthOverrides(map[string]interface{}{
		"src/**": 0,
	})
	if len(overrides) != 0 {
		t.Errorf("expected 0 overrides for zero value, got %d", len(overrides))
	}
}

func TestParseMaxDepthOverrides_List(t *testing.T) {
	overrides := parseMaxDepthOverrides([]interface{}{
		map[string]interface{}{"pattern": "src/**", "max": 8},
		map[string]interface{}{"pattern": "tests/**", "max": 6},
	})
	if len(overrides) != 2 {
		t.Fatalf("expected 2 overrides, got %d", len(overrides))
	}
	m := map[string]int{}
	for _, o := range overrides {
		m[o.Pattern] = o.Max
	}
	if m["src/**"] != 8 || m["tests/**"] != 6 {
		t.Errorf("unexpected overrides: %v", overrides)
	}
}

func TestParseMaxDepthOverrides_ListInvalidItem(t *testing.T) {
	overrides := parseMaxDepthOverrides([]interface{}{
		"not a map",
	})
	if len(overrides) != 0 {
		t.Errorf("expected 0 overrides for invalid item, got %d", len(overrides))
	}
}

func TestParseMaxDepthOverrides_ListMissingFields(t *testing.T) {
	overrides := parseMaxDepthOverrides([]interface{}{
		map[string]interface{}{"pattern": "src/**"}, // missing max
	})
	if len(overrides) != 0 {
		t.Errorf("expected 0 overrides for missing max, got %d", len(overrides))
	}
}

func TestParseMaxDepthOverrides_UnsupportedType(t *testing.T) {
	overrides := parseMaxDepthOverrides("string")
	if len(overrides) != 0 {
		t.Errorf("expected 0 overrides for string, got %d", len(overrides))
	}
}

func TestToInt(t *testing.T) {
	if got := toInt(42); got != 42 {
		t.Errorf("toInt(42) = %d", got)
	}
	if got := toInt(float64(42)); got != 42 {
		t.Errorf("toInt(float64(42)) = %d", got)
	}
	if got := toInt("hello"); got != 0 {
		t.Errorf("toInt('hello') = %d, want 0", got)
	}
	if got := toInt(nil); got != 0 {
		t.Errorf("toInt(nil) = %d, want 0", got)
	}
}
