package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type PythonStrategy struct {
	reader             ci.FileReader
	coverage           ci.CoverageThresholds
	requirePytestLinter bool
}

func NewPythonStrategy(reader ci.FileReader, cfg map[string]interface{}) *PythonStrategy {
	s := &PythonStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{
		Branches:   90,
		Lines:      80,
		Functions:  90,
		Statements: 80,
	}
	if cfg != nil {
		if v, ok := cfg["require-pytest-linter"].(bool); ok {
			s.requirePytestLinter = v
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

func (s *PythonStrategy) ProjectType() ci.ProjectType { return ci.Python }
func (s *PythonStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *PythonStrategy) RequiredCIGates() []ci.CIGate {
	gates := []ci.CIGate{
		{Name: "ruff check", Required: true, Hint: "Add ruff check to CI"},
		{Name: "pyright", Required: true, Hint: "Add pyright type-checking to CI"},
		{Name: "pytest --cov-branch --cov-fail-under", Required: true, Hint: "Ensure pytest uses --cov-branch and --cov-fail-under=90"},
	}
	if s.requirePytestLinter {
		gates = append(gates, ci.CIGate{Name: "pytest-linter", Required: true, Hint: "Add pytest-linter CI gate"})
	}
	return gates
}
func (s *PythonStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "ruff", Required: true, Hint: "Configure ruff in pyproject.toml"},
		{Name: "pyright", Required: true, Hint: "Configure pyright in pyproject.toml or pyrightconfig.json"},
	}
}
func (s *PythonStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
func (s *PythonStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	for _, job := range jobs {
		for _, step := range job.Steps {
			runLower := strings.ToLower(step.Run)
			if !strings.Contains(runLower, "pytest") {
				continue
			}
			if !strings.Contains(runLower, "--cov-branch") {
				results = append(results, ci.CheckResult{
					Message: "pytest command missing --cov-branch",
					Fix:     "Add --cov-branch to pytest command for branch coverage.",
				})
			}
			if !strings.Contains(runLower, "--cov-fail-under") {
				results = append(results, ci.CheckResult{
					Message: "pytest command missing --cov-fail-under",
					Fix:     "Add --cov-fail-under=90 to pytest command.",
				})
			}
		}
	}

	foundRuff := false
	foundPyright := false
	foundPytestLinter := false
	for _, job := range jobs {
		for _, step := range job.Steps {
			combined := strings.ToLower(step.Run + " " + step.Name)
			if strings.Contains(combined, "ruff") {
				foundRuff = true
			}
			if strings.Contains(combined, "pyright") {
				foundPyright = true
			}
			if s.requirePytestLinter && strings.Contains(combined, "pytest-linter") {
				foundPytestLinter = true
			}
		}
	}
	if !foundRuff {
		results = append(results, ci.CheckResult{
			Message: "Missing ruff check in CI",
			Fix:     "Add ruff check to CI workflow.",
		})
	}
	if !foundPyright {
		results = append(results, ci.CheckResult{
			Message: "Missing pyright in CI",
			Fix:     "Add pyright type-checking to CI workflow.",
		})
	}
	if s.requirePytestLinter && !foundPytestLinter {
		results = append(results, ci.CheckResult{
			Message: "Missing pytest-linter CI gate",
			Fix:     "Add pytest-linter CI gate.",
		})
	}

	return results
}
func (s *PythonStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult {
	var results []ci.CheckResult
	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".py") {
			continue
		}
		data, err := reader.ReadFile(f.AbsPath)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		count := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "# noqa") || strings.Contains(trimmed, "# type: ignore") {
				count++
			}
		}
		if count > 0 {
			results = append(results, ci.CheckResult{
				Path:    f.Path,
				Message: "Python suppression comments exceed threshold",
				Fix:     "Reduce # noqa / # type: ignore comments. Address root causes.",
			})
		}
	}
	return results
}
