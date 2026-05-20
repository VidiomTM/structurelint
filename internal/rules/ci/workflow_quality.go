package ci

import (
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"gopkg.in/yaml.v3"
)

type WorkflowQualityRule struct {
	registry *StrategyRegistry
	reader   core.FileReader
	detector *ProjectDetector
	config   map[string]interface{}
}

func NewWorkflowQualityRule(registry *StrategyRegistry, reader core.FileReader, detector *ProjectDetector, config map[string]interface{}) *WorkflowQualityRule {
	return &WorkflowQualityRule{
		registry: registry,
		reader:   reader,
		detector: detector,
		config:   config,
	}
}

func (r *WorkflowQualityRule) Name() string {
	return "github-workflows"
}

func (r *WorkflowQualityRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	internalFiles := toInternalFiles(files)

	projectTypes := r.detector.Detect(internalFiles)
	if len(projectTypes) == 0 {
		return nil
	}

	strats := r.registry.StrategiesFor(projectTypes)

	workflowFiles := filterWorkflowFiles(files)
	workflowJobs := r.parseAllWorkflows(workflowFiles)

	var results []core.CheckResult

	globalCfg := extractConfig("global", r.config)
	if v, ok := globalCfg["disallow-command-masking"].(bool); ok && v {
		results = append(results, checkCommandMasking(workflowJobs)...)
	}
	if v, ok := globalCfg["disallow-continue-on-error-on-quality"].(bool); ok && v {
		results = append(results, checkContinueOnError(workflowJobs)...)
	}
	if v, ok := globalCfg["require-required-checks-aggregator"].(bool); ok && v {
		results = append(results, checkRequiredChecksAggregator(workflowJobs)...)
	}

	allFiles := toInternalFiles(files)
	for _, s := range strats {
		results = append(results, s.CheckWorkflowSteps(workflowJobs)...)
		results = append(results, s.CheckProjectConfig(allFiles, r.reader)...)
		results = append(results, s.CheckSuppressions(allFiles, r.reader)...)
	}

	var violations []rules.Violation
	for _, cr := range results {
		violations = append(violations, cr.ToViolation())
	}
	return violations
}

func (r *WorkflowQualityRule) parseAllWorkflows(files []walker.FileInfo) map[string]core.JobInfo {
	jobs := make(map[string]core.JobInfo)
	for _, f := range files {
		for name, job := range parseWorkflowJobs(f, r.reader) {
			jobs[name] = job
		}
	}
	return jobs
}

func parseWorkflowJobs(f walker.FileInfo, reader core.FileReader) map[string]core.JobInfo {
	data, err := reader.ReadFile(f.AbsPath)
	if err != nil {
		return nil
	}
	var workflow yaml.Node
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil
	}
	rawJobs := findJobs(&workflow)
	jobs := make(map[string]core.JobInfo, len(rawJobs))
	for name, jobNode := range rawJobs {
		ji := core.JobInfo{Name: name}
		if jobNode.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i < len(jobNode.Content)-1; i += 2 {
			if jobNode.Content[i].Value == "steps" && jobNode.Content[i+1].Kind == yaml.SequenceNode {
				for _, stepNode := range jobNode.Content[i+1].Content {
					if stepNode.Kind != yaml.MappingNode {
						continue
					}
					si := core.StepInfo{Line: stepNode.Line}
					for j := 0; j < len(stepNode.Content)-1; j += 2 {
						key := stepNode.Content[j].Value
						val := stepNode.Content[j+1].Value
						switch key {
						case "name":
							si.Name = val
						case "run":
							si.Run = val
						case "continue-on-error":
							si.ContinueOnError = val
						case "uses":
							si.Uses = val
						}
					}
					ji.Steps = append(ji.Steps, si)
				}
			}
		}
		jobs[name] = ji
	}
	return jobs
}

func findJobs(workflow *yaml.Node) map[string]*yaml.Node {
	jobs := make(map[string]*yaml.Node)
	if workflow.Kind != yaml.DocumentNode || len(workflow.Content) == 0 {
		return jobs
	}
	root := workflow.Content[0]
	if root.Kind != yaml.MappingNode {
		return jobs
	}
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "jobs" && root.Content[i+1].Kind == yaml.MappingNode {
			jobsNode := root.Content[i+1]
			for j := 0; j < len(jobsNode.Content)-1; j += 2 {
				if jobsNode.Content[j+1].Kind == yaml.MappingNode {
					jobs[jobsNode.Content[j].Value] = jobsNode.Content[j+1]
				} else if jobsNode.Content[j+1].Kind == yaml.SequenceNode {
					jobs[jobsNode.Content[j].Value] = jobsNode.Content[j+1]
				}
			}
		}
	}
	return jobs
}

func toInternalFiles(files []walker.FileInfo) []core.FileInfo {
	var out []core.FileInfo
	for _, f := range files {
		out = append(out, core.FileInfo{
			Path:    f.Path,
			AbsPath: f.AbsPath,
			IsDir:   f.IsDir,
		})
	}
	return out
}

func filterWorkflowFiles(files []walker.FileInfo) []walker.FileInfo {
	var out []walker.FileInfo
	for _, f := range files {
		if f.IsDir {
			continue
		}
		path := filepath.ToSlash(f.Path)
		if strings.Contains(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			out = append(out, f)
		}
	}
	return out
}
