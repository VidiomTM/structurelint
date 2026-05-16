package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type SvelteKitStrategy struct {
	reader             ci.FileReader
	coverage           ci.CoverageThresholds
	requireVitestLinter bool
	requireSvelteuml    bool
}

func NewSvelteKitStrategy(reader ci.FileReader, cfg map[string]interface{}) *SvelteKitStrategy {
	s := &SvelteKitStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{
		Branches:   90,
		Lines:      80,
		Functions:  90,
		Statements: 90,
	}
	if cfg != nil {
		if v, ok := cfg["require-vitest-linter"].(bool); ok {
			s.requireVitestLinter = v
		}
		if v, ok := cfg["require-svelteuml"].(bool); ok {
			s.requireSvelteuml = v
		}
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if b, ok := cv["branches"].(float64); ok { s.coverage.Branches = b }
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
			if f, ok := cv["functions"].(float64); ok { s.coverage.Functions = f }
			if st, ok := cv["statements"].(float64); ok { s.coverage.Statements = st }
		}
	}
	return s
}

func (s *SvelteKitStrategy) ProjectType() ci.ProjectType { return ci.SvelteKit }
func (s *SvelteKitStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *SvelteKitStrategy) RequiredCIGates() []ci.CIGate {
	gates := []ci.CIGate{
		{Name: "svelte-check --fail-on-warnings", Required: true, Hint: "Add svelte-check with --fail-on-warnings"},
		{Name: "biome check", Required: true, Hint: "Add biome check to CI"},
		{Name: "vitest coverage", Required: true, Hint: "Add vitest run --coverage"},
		{Name: "build", Required: true, Hint: "Add pnpm build"},
	}
	if s.requireVitestLinter {
		gates = append(gates, ci.CIGate{Name: "vitest-linter", Required: true, Hint: "Add vitest-linter CI gate"})
	}
	if s.requireSvelteuml {
		gates = append(gates, ci.CIGate{Name: "svelteuml", Required: true, Hint: "Add svelteuml diagram generation to CI"})
	}
	return gates
}

func (s *SvelteKitStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "biome", Required: true, Hint: "Configure biome.json"},
	}
}

func (s *SvelteKitStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }

func (s *SvelteKitStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	gates := s.RequiredCIGates()
	for _, gate := range gates {
		found := false
		for _, job := range jobs {
			for _, step := range job.Steps {
				runLower := strings.ToLower(step.Run)
				nameLower := strings.ToLower(step.Name)
				combined := runLower + " " + nameLower
				switch {
				case strings.Contains(gate.Name, "svelte-check"):
					if strings.Contains(combined, "svelte-check") {
						found = true
						if !strings.Contains(combined, "--fail-on-warnings") {
							results = append(results, ci.CheckResult{
								Message: "svelte-check without --fail-on-warnings",
								Fix:     "Add --fail-on-warnings to svelte-check command.",
							})
						}
					}
				case strings.Contains(gate.Name, "vitest-linter"):
					if strings.Contains(combined, "vitest-linter") {
						found = true
					}
				case strings.Contains(gate.Name, "svelteuml"):
					if strings.Contains(combined, "svelteuml") {
						found = true
					}
				case strings.Contains(gate.Name, "biome"):
					if strings.Contains(combined, "biome") {
						found = true
					}
				case strings.Contains(gate.Name, "vitest"):
					if strings.Contains(combined, "vitest") && strings.Contains(combined, "coverage") {
						found = true
					}
				case strings.Contains(gate.Name, "build"):
					if (strings.Contains(runLower, "pnpm") || strings.Contains(runLower, "npm")) && strings.Contains(runLower, "build") {
						found = true
					}
				}
			}
		}
		if !found && gate.Required {
			results = append(results, ci.CheckResult{
				Message: "Missing required CI gate: " + gate.Name,
				Fix:     gate.Hint,
			})
		}
	}
	return results
}

func (s *SvelteKitStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
