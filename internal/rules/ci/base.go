package ci

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

var maskingPattern = regexp.MustCompile(`\|\|\s*(true|echo\s+['"]?['"]?)\s*$`)

var qualityStepNamePatterns = []string{
	"lint", "typecheck", "type.check", "type check",
	"test", "pytest", "check", "quality",
	"coverage", "ruff", "pyright", "biome", "svelte-check",
}

func checkCommandMasking(jobs map[string]core.JobInfo) []core.CheckResult {
	var results []core.CheckResult
	for jobName, job := range jobs {
		for _, step := range job.Steps {
			if maskingPattern.MatchString(step.Run) {
				results = append(results, core.CheckResult{
					Path:    fmt.Sprintf(".github/workflows job=%q step=%q", jobName, step.Name),
					Message: fmt.Sprintf("Command masking on %q: %q", step.Name, strings.TrimSpace(step.Run)),
					Fix:     "Remove '|| true' or '|| echo \"\"' to let command failures propagate.",
				})
			}
		}
	}
	return results
}

func checkContinueOnError(jobs map[string]core.JobInfo) []core.CheckResult {
	var results []core.CheckResult
	for jobName, job := range jobs {
		for _, step := range job.Steps {
			if step.ContinueOnError != "true" && step.ContinueOnError != "yes" {
				continue
			}
			lower := strings.ToLower(step.Name)
			for _, p := range qualityStepNamePatterns {
				if strings.Contains(lower, p) {
					results = append(results, core.CheckResult{
						Path:    fmt.Sprintf(".github/workflows job=%q step=%q", jobName, step.Name),
						Message: fmt.Sprintf("continue-on-error on quality step %q", step.Name),
						Fix:     "Remove continue-on-error: true from quality check steps.",
					})
					break
				}
			}
		}
	}
	return results
}

func checkRequiredChecksAggregator(jobs map[string]core.JobInfo) []core.CheckResult {
	for name := range jobs {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "required-checks") || strings.Contains(lower, "required.checks") {
			return nil
		}
	}
	return []core.CheckResult{{
		Message: "Workflow missing a required-checks aggregator job",
		Fix:     `Add a "required-checks" job that depends on all quality jobs and verifies results.`,
	}}
}
