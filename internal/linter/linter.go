package linter

import (
	"errors"
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

var ErrNoConfig = errors.New("no .structurelint.yml configuration file found")

type Linter struct {
	config  *config.Config
	rootDir string
}

type Violation = rules.Violation

func New() *Linter {
	return &Linter{}
}

func (l *Linter) Lint(path string) ([]Violation, error) {
	l.rootDir = path

	configs, found, err := config.FindConfigsWithGitignore(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if !found {
		return nil, ErrNoConfig
	}

	l.config = config.Merge(configs...)

	w := walker.New(path).WithExclude(l.config.Exclude)
	if err := w.Walk(); err != nil {
		return nil, fmt.Errorf("failed to walk filesystem: %w", err)
	}

	files := w.GetFiles()
	dirs := w.GetDirs()

	var importGraph *graph.ImportGraph
	needsGraph := len(l.config.Layers) > 0 ||
		l.isRuleEnabled("enforce-layer-boundaries") ||
		l.isRuleEnabled("disallow-orphaned-files") ||
		l.isRuleEnabled("disallow-import-cycles") ||
		l.isRuleEnabled("path-based-layers")

	if needsGraph {
		builder := graph.NewBuilder(path, l.config.Layers)
		var err error
		importGraph, err = builder.Build(files)
		if err != nil {
			return nil, fmt.Errorf("failed to build import graph: %w", err)
		}
	}

	factory := NewRuleFactory(l.rootDir, l.config, importGraph)
	rulesList, err := factory.CreateRules(files, dirs)
	if err != nil {
		return nil, fmt.Errorf("failed to create rules: %w", err)
	}

	var violations []Violation
	for _, rule := range rulesList {
		ruleViolations := rule.Check(files, dirs)
		violations = append(violations, ruleViolations...)
	}

	return violations, nil
}

func (l *Linter) isRuleEnabled(ruleName string) bool {
	if l.config == nil || l.config.Rules == nil {
		return false
	}

	value, exists := l.config.Rules[ruleName]
	if !exists {
		return false
	}

	switch v := value.(type) {
	case int:
		return v != 0
	case bool:
		return v
	}

	return true
}

func (l *Linter) getRuleConfig(ruleName string) (interface{}, bool) {
	if l.config == nil || l.config.Rules == nil {
		return nil, false
	}

	value, exists := l.config.Rules[ruleName]
	if !exists {
		return nil, false
	}

	switch v := value.(type) {
	case int:
		if v == 0 {
			return nil, false
		}
	case bool:
		if !v {
			return nil, false
		}
	}

	return value, true
}

func (l *Linter) createRules(files []walker.FileInfo, importGraph *graph.ImportGraph) ([]rules.Rule, error) {
	factory := NewRuleFactory(l.rootDir, l.config, importGraph)
	return factory.CreateRules(files, nil)
}
