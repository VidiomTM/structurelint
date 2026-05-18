package structure

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// LinterConfigRule validates project-level linter and tooling configuration
// files (pyproject.toml, tsconfig.json, .semgrep.yml).
type LinterConfigRule struct {
	rootDir string
}

// NewLinterConfigRule creates a new LinterConfigRule.
func NewLinterConfigRule(rootDir string) *LinterConfigRule {
	return &LinterConfigRule{rootDir: rootDir}
}

func (r *LinterConfigRule) Name() string { return "linter-config" }

func (r *LinterConfigRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	// Index root-level files.
	rootFiles := make(map[string]walker.FileInfo)
	for _, f := range files {
		if !f.IsDir && f.ParentPath == "" {
			rootFiles[f.Path] = f
		}
	}

	// .semgrep.yml is required for ALL projects.
	if _, ok := rootFiles[".semgrep.yml"]; !ok {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    ".semgrep.yml",
			Message: "missing required config file: .semgrep.yml",
			Suggestions: []string{
				"Create a .semgrep.yml file at the project root",
				"See https://semgrep.dev/docs/writing-rules/overview/",
			},
		})
	}

	// Python project: validate pyproject.toml.
	if f, ok := rootFiles["pyproject.toml"]; ok {
		violations = append(violations, r.validatePyproject(f)...)
	}

	// TypeScript project: validate tsconfig.json.
	if f, ok := rootFiles["tsconfig.json"]; ok {
		violations = append(violations, r.validateTsconfig(f)...)
	}

	return violations
}

// validatePyproject checks that [tool.pyright] has typeCheckingMode = "strict"
// and that a [tool.ruff] section exists.
func (r *LinterConfigRule) validatePyproject(f walker.FileInfo) []rules.Violation {
	var violations []rules.Violation

	data, err := os.ReadFile(f.AbsPath)
	if err != nil {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    f.Path,
			Message: fmt.Sprintf("could not read pyproject.toml: %v", err),
		})
		return violations
	}

	var cfg struct {
		Tool struct {
			Pyright *struct {
				TypeCheckingMode string `toml:"typeCheckingMode"`
			} `toml:"pyright"`
			Ruff *struct{} `toml:"ruff"`
		} `toml:"tool"`
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    f.Path,
			Message: fmt.Sprintf("could not parse pyproject.toml: %v", err),
		})
		return violations
	}

	if cfg.Tool.Pyright == nil {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    f.Path,
			Message: "Python project missing [tool.pyright] configuration",
			Suggestions: []string{
				"Add a [tool.pyright] section to pyproject.toml",
				"Example:\n[tool.pyright]\ntypeCheckingMode = \"strict\"",
			},
		})
	} else if cfg.Tool.Pyright.TypeCheckingMode != "strict" {
		violations = append(violations, rules.Violation{
			Rule:     r.Name(),
			Path:     f.Path,
			Message:  "pyright typeCheckingMode should be \"strict\"",
			Expected: "strict",
			Actual:   cfg.Tool.Pyright.TypeCheckingMode,
			Suggestions: []string{
				"Set typeCheckingMode = \"strict\" under [tool.pyright]",
			},
		})
	}

	if cfg.Tool.Ruff == nil {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    f.Path,
			Message: "Python project missing [tool.ruff] configuration",
			Suggestions: []string{
				"Add a [tool.ruff] section to pyproject.toml",
				"Example:\n[tool.ruff]\nline-length = 100",
			},
		})
	}

	return violations
}

// validateTsconfig checks that compilerOptions.strict is true.
func (r *LinterConfigRule) validateTsconfig(f walker.FileInfo) []rules.Violation {
	var violations []rules.Violation

	data, err := os.ReadFile(f.AbsPath)
	if err != nil {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    f.Path,
			Message: fmt.Sprintf("could not read tsconfig.json: %v", err),
		})
		return violations
	}

	var cfg struct {
		CompilerOptions struct {
			Strict bool `json:"strict"`
		} `json:"compilerOptions"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    f.Path,
			Message: fmt.Sprintf("could not parse tsconfig.json: %v", err),
		})
		return violations
	}

	if !cfg.CompilerOptions.Strict {
		violations = append(violations, rules.Violation{
			Rule:     r.Name(),
			Path:     f.Path,
			Message:  "TypeScript project should have strict mode enabled",
			Expected: "compilerOptions.strict = true",
			Actual:   "compilerOptions.strict = false or missing",
			Suggestions: []string{
				"Add \"strict\": true to compilerOptions in tsconfig.json",
			},
		})
	}

	return violations
}
