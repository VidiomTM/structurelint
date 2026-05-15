package fuzz

import (
	"encoding/json"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/output"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func FuzzJSONOutput(f *testing.F) {
	seeds := []string{
		`{"rule":"test","path":"a.ts","message":"violation"}`,
		`{"rule":"","path":"","message":""}`,
		`null`,
		`{}`,
		`[]`,
		"invalid",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var violations []rules.Violation
		_ = json.Unmarshal([]byte(input), &violations)

		if violations == nil {
			violations = []rules.Violation{}
		}

		fmtr := &output.JSONFormatter{Version: "fuzz"}
		result, err := fmtr.Format(violations)
		if err != nil {
			return
		}

		var parsed output.JSONOutput
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Fatalf("formatter produced invalid JSON: %v", err)
		}
	})
}
