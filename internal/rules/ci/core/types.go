package core

import "github.com/Jonathangadeaharder/structurelint/internal/rules"

type ProjectType string

const (
	SvelteKit ProjectType = "sveltekit"
	Python    ProjectType = "python"
	Go        ProjectType = "golang"
	Rust      ProjectType = "rust"
)

type CoverageThresholds struct {
	Branches   float64
	Lines      float64
	Functions  float64
	Statements float64
}

type CIGate struct {
	Name     string
	Required bool
	Hint     string
}

type LinterTool struct {
	Name     string
	Required bool
	Hint     string
}

type FileInfo struct {
	Path    string
	AbsPath string
	IsDir   bool
}

type JobInfo struct {
	Name  string
	Steps []StepInfo
}

type StepInfo struct {
	Name            string
	Run             string
	ContinueOnError string
	Uses            string
	Line            int
}

type CheckResult struct {
	Rule    string
	Path    string
	Message string
	Fix     string
}

func (c CheckResult) ToViolation() rules.Violation {
	v := rules.Violation{
		Rule:    "github-workflows",
		Path:    c.Path,
		Message: c.Message,
	}
	if c.Fix != "" {
		v.Suggestions = []string{c.Fix}
	}
	return v
}

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type Strategy interface {
	ProjectType() ProjectType
	RequiredCoverage() CoverageThresholds
	RequiredCIGates() []CIGate
	RequiredLinters() []LinterTool
	CheckProjectConfig(files []FileInfo, reader FileReader) []CheckResult
	CheckWorkflowSteps(jobs map[string]JobInfo) []CheckResult
	CheckSuppressions(files []FileInfo, reader FileReader) []CheckResult
}
