package ci

type Strategy interface {
	ProjectType() ProjectType
	RequiredCoverage() CoverageThresholds
	RequiredCIGates() []CIGate
	RequiredLinters() []LinterTool
	CheckProjectConfig(files []FileInfo, reader FileReader) []CheckResult
	CheckWorkflowSteps(jobs map[string]JobInfo) []CheckResult
	CheckSuppressions(files []FileInfo, reader FileReader) []CheckResult
}
