package detector

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/clones/types"
)

// Reporter formats and outputs clone detection results
type Reporter struct {
	format string // "console", "json", "sarif"
}

// NewReporter creates a new reporter with the specified format
func NewReporter(format string) *Reporter {
	return &Reporter{format: format}
}

// Report outputs the clone detection results
func (r *Reporter) Report(clones []*types.Clone) string {
	switch r.format {
	case "json":
		return r.reportJSON(clones)
	case "sarif":
		return r.reportSARIF(clones)
	default:
		return r.reportConsole(clones)
	}
}

// reportConsole formats clones for console output
func (r *Reporter) reportConsole(clones []*types.Clone) string {
	if len(clones) == 0 {
		return "✓ No code clones detected!\n"
	}

	var output strings.Builder

	fmt.Fprintf(&output, "\n🔍 Found %d code clone pairs:\n", len(clones))
	output.WriteString(strings.Repeat("=", 80) + "\n\n")

	for i, clone := range clones {
		fmt.Fprintf(&output, "Clone Pair #%d (%d tokens, ~%d lines) [%s]\n",
			i+1, clone.TokenCount, clone.LineCount, clone.Type.String())
		output.WriteString(strings.Repeat("-", 80) + "\n")

		for j, loc := range clone.Locations {
			fmt.Fprintf(&output, "  Location %c: %s:%d-%d\n",
				'A'+j, loc.FilePath, loc.StartLine, loc.EndLine)
		}

		fmt.Fprintf(&output, "  Similarity: %.1f%%\n", clone.Similarity*100)
		output.WriteString("\n")
	}

	fmt.Fprintf(&output, "Total: %d clone pairs detected\n", len(clones))

	return output.String()
}

// reportJSON formats clones as JSON
func (r *Reporter) reportJSON(clones []*types.Clone) string {
	type JSONClone struct {
		Type       string          `json:"type"`
		TokenCount int             `json:"token_count"`
		LineCount  int             `json:"line_count"`
		Similarity float64         `json:"similarity"`
		Locations  []types.Location `json:"locations"`
	}

	type JSONReport struct {
		TotalClones int         `json:"total_clones"`
		Clones      []JSONClone `json:"clones"`
	}

	report := JSONReport{
		TotalClones: len(clones),
		Clones:      make([]JSONClone, len(clones)),
	}

	for i, clone := range clones {
		report.Clones[i] = JSONClone{
			Type:       clone.Type.String(),
			TokenCount: clone.TokenCount,
			LineCount:  clone.LineCount,
			Similarity: clone.Similarity,
			Locations:  clone.Locations,
		}
	}

	jsonBytes, _ := json.MarshalIndent(report, "", "  ")
	return string(jsonBytes)
}

// reportSARIF formats clones in SARIF format for IDE integration
func (r *Reporter) reportSARIF(clones []*types.Clone) string {
	// SARIF (Static Analysis Results Interchange Format)
	// Simplified version for POC
	type SARIFResult struct {
		RuleID  string `json:"ruleId"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
		Locations []struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
				Region struct {
					StartLine int `json:"startLine"`
					EndLine   int `json:"endLine"`
				} `json:"region"`
			} `json:"physicalLocation"`
		} `json:"locations"`
	}

	type SARIFReport struct {
		Version string `json:"version"`
		Runs    []struct {
			Tool struct {
				Driver struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				} `json:"driver"`
			} `json:"tool"`
			Results []SARIFResult `json:"results"`
		} `json:"runs"`
	}

	report := SARIFReport{Version: "2.1.0"}
	report.Runs = make([]struct {
		Tool struct {
			Driver struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"driver"`
		} `json:"tool"`
		Results []SARIFResult `json:"results"`
	}, 1)

	report.Runs[0].Tool.Driver.Name = "structurelint-clones"
	report.Runs[0].Tool.Driver.Version = "1.0.0"
	report.Runs[0].Results = make([]SARIFResult, len(clones))

	for i, clone := range clones {
		result := SARIFResult{
			RuleID: "code-clone-" + strings.ToLower(clone.Type.String()),
		}
		result.Message.Text = fmt.Sprintf("Code clone detected: %d tokens, %s",
			clone.TokenCount, clone.Type.String())

		result.Locations = make([]struct {
			PhysicalLocation struct {
				ArtifactLocation struct {
					URI string `json:"uri"`
				} `json:"artifactLocation"`
				Region struct {
					StartLine int `json:"startLine"`
					EndLine   int `json:"endLine"`
				} `json:"region"`
			} `json:"physicalLocation"`
		}, len(clone.Locations))

		for j, loc := range clone.Locations {
			result.Locations[j].PhysicalLocation.ArtifactLocation.URI = loc.FilePath
			result.Locations[j].PhysicalLocation.Region.StartLine = loc.StartLine
			result.Locations[j].PhysicalLocation.Region.EndLine = loc.EndLine
		}

		report.Runs[0].Results[i] = result
	}

	jsonBytes, _ := json.MarshalIndent(report, "", "  ")
	return string(jsonBytes)
}

// Summary provides a quick summary of clone detection results
func (r *Reporter) Summary(clones []*types.Clone) string {
	if len(clones) == 0 {
		return "No clones detected"
	}

	type1Count := 0
	type2Count := 0
	type3Count := 0

	totalTokens := 0
	totalLines := 0

	for _, clone := range clones {
		switch clone.Type {
		case types.Type1:
			type1Count++
		case types.Type2:
			type2Count++
		case types.Type3:
			type3Count++
		}
		totalTokens += clone.TokenCount
		totalLines += clone.LineCount
	}

	return fmt.Sprintf("Clones: %d total (Type-1: %d, Type-2: %d, Type-3: %d) | Tokens: %d | Lines: ~%d",
		len(clones), type1Count, type2Count, type3Count, totalTokens, totalLines)
}
