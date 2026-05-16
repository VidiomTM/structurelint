package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents a .structurelint.yml configuration file
type Config struct {
	Root                   bool                   `yaml:"root"`
	Extends                interface{}            `yaml:"extends"`                // string or []string
	Exclude                []string               `yaml:"exclude"`                // Patterns to exclude from linting
	AutoLoadGitignore      *bool                  `yaml:"autoLoadGitignore"`      // Auto-load .gitignore patterns (default: true)
	AutoLanguageNaming     *bool                  `yaml:"autoLanguageNaming"`     // Auto-apply language-specific naming conventions (default: true)
	InfrastructurePatterns []string               `yaml:"infrastructurePatterns"` // Additional patterns for infrastructure code (Priority 2 feature)
	Rules                  map[string]interface{} `yaml:"rules"`
	Overrides              []Override             `yaml:"overrides"`
	Layers                 []Layer                `yaml:"layers"`      // Phase 1: Layer definitions
	Entrypoints            []string               `yaml:"entrypoints"` // Phase 2: Entry points for orphan detection
}

// Override represents a configuration override for specific file patterns
type Override struct {
	Files []string               `yaml:"files"`
	Rules map[string]interface{} `yaml:"rules"`
}

// Layer represents an architectural layer definition (Phase 1)
type Layer struct {
	Name      string   `yaml:"name"`
	Path      string   `yaml:"path"`
	DependsOn []string `yaml:"dependsOn"`
}

// PathBasedLayer represents a path-based layer (Priority 3 feature)
type PathBasedLayer struct {
	Name           string   `yaml:"name"`
	Patterns       []string `yaml:"patterns"`       // Regex/glob patterns for matching files
	CanDependOn    []string `yaml:"canDependOn"`    // Names of layers this can depend on
	ForbiddenPaths []string `yaml:"forbiddenPaths"` // Path patterns this layer cannot contain
}

// MaxDepthRule represents the max-depth rule configuration
type MaxDepthRule struct {
	Max int `yaml:"max"`
}

// MaxFilesRule represents the max-files-in-dir rule configuration
type MaxFilesRule struct {
	Max int `yaml:"max"`
}

// MaxSubdirsRule represents the max-subdirs rule configuration
type MaxSubdirsRule struct {
	Max int `yaml:"max"`
}

// NamingConventionRule represents naming convention patterns
type NamingConventionRule map[string]string

// FileExistenceRule represents file existence requirements
type FileExistenceRule map[string]string

// AllowedLocationsRule represents allowed file locations
type AllowedLocationsRule struct {
	Files       []string `yaml:"files"`
	Destination []string `yaml:"destination"`
	StartsWith  string   `yaml:"startsWith,omitempty"`
}

// DisallowedPatternsRule represents disallowed file patterns
type DisallowedPatternsRule []string

// Load loads a configuration file from the given path
func Load(path string) (*Config, error) {
	visited := make(map[string]bool)
	return loadWithVisited(path, visited)
}

// LoadWithGitignore loads a configuration and auto-applies .gitignore patterns
// rootDir is the directory to search for .gitignore (typically the project root)
func LoadWithGitignore(path, rootDir string) (*Config, error) {
	config, err := Load(path)
	if err != nil {
		return nil, err
	}

	return ApplyGitignore(config, rootDir)
}

// ApplyGitignore applies .gitignore patterns to a config if autoLoadGitignore is true
func ApplyGitignore(config *Config, rootDir string) (*Config, error) {
	// Check if auto-load is disabled
	if config.AutoLoadGitignore != nil && !*config.AutoLoadGitignore {
		return config, nil
	}

	// Default to true if not specified
	gitignorePatterns, err := LoadGitignorePatterns(rootDir)
	if err != nil {
		// Don't fail if .gitignore can't be read, just skip it
		return config, nil
	}

	// Merge gitignore patterns with existing exclusions
	config.Exclude = MergeWithGitignore(config.Exclude, gitignorePatterns)

	return config, nil
}

// loadWithVisited loads a config and tracks visited paths to detect cycles
func loadWithVisited(path string, visited map[string]bool) (*Config, error) {
	// Normalize path for cycle detection
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path // fallback to original path if Abs fails
	}

	// Check for cycles
	if visited[absPath] {
		return nil, fmt.Errorf("circular dependency detected: config '%s' is already being loaded", path)
	}
	visited[absPath] = true

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Rules: make(map[string]interface{})}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Rules == nil {
		config.Rules = make(map[string]interface{})
	}

	// Resolve extends if present
	if config.Extends != nil {
		extendedConfigs, err := resolveExtendsWithVisited(config.Extends, filepath.Dir(path), visited)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve extends: %w", err)
		}

		// Merge extended configs with this config
		// Extended configs come first, so this config overrides them
		allConfigs := append(extendedConfigs, &config)
		merged := Merge(allConfigs...)

		// Clear the Extends field to avoid infinite recursion
		merged.Extends = nil

		return merged, nil
	}

	return &config, nil
}

// resolveExtendsWithVisited resolves extends with cycle detection
func resolveExtendsWithVisited(extends interface{}, baseDir string, visited map[string]bool) ([]*Config, error) {
	var extendPaths []string

	// Handle both string and []string
	switch v := extends.(type) {
	case string:
		extendPaths = []string{v}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				extendPaths = append(extendPaths, s)
			}
		}
	case []string:
		extendPaths = v
	default:
		return nil, fmt.Errorf("extends must be a string or array of strings, got %T", extends)
	}

	var configs []*Config
	for _, extendPath := range extendPaths {
		resolvedPath, err := resolveExtendPath(extendPath, baseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve extend path '%s': %w", extendPath, err)
		}

		config, err := loadWithVisited(resolvedPath, visited)
		if err != nil {
			return nil, fmt.Errorf("failed to load extended config '%s': %w", resolvedPath, err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// resolveExtendPath resolves an extend path to an absolute file path
func resolveExtendPath(extendPath, baseDir string) (string, error) {
	// If it's an absolute path, use it as-is
	if filepath.IsAbs(extendPath) {
		if _, err := os.Stat(extendPath); err != nil {
			return "", fmt.Errorf("extended config not found: %w", err)
		}
		return extendPath, nil
	}

	// If it starts with ./ or ../, or is not an absolute path, treat as relative
	if strings.HasPrefix(extendPath, "./") ||
		strings.HasPrefix(extendPath, "../") ||
		(!filepath.IsAbs(extendPath) && filepath.VolumeName(extendPath) == "") {
		absPath := filepath.Join(baseDir, extendPath)
		if _, err := os.Stat(absPath); err != nil {
			return "", fmt.Errorf("extended config not found: %w", err)
		}
		return absPath, nil
	}

	// Future: Handle package names (e.g., @structurelint/preset-go)
	// For now, treat as a relative path
	absPath := filepath.Join(baseDir, extendPath)
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("extended config not found (package resolution not yet implemented): %w", err)
	}
	return absPath, nil
}

// FindConfigs finds all .structurelint.yml files from the given path up to the root
func FindConfigs(startPath string) ([]*Config, error) {
	configs, _, _, err := findConfigsWithRoot(startPath)
	return configs, err
}

// FindConfigsWithGitignore finds configs and applies .gitignore patterns.
// Returns the configs and whether at least one config file was found on disk.
func FindConfigsWithGitignore(startPath string) ([]*Config, bool, error) {
	configs, rootDir, found, err := findConfigsWithRoot(startPath)
	if err != nil {
		return nil, false, err
	}

	// If we have configs, apply gitignore to the merged config
	if len(configs) > 0 {
		// Apply gitignore to each config before merging
		for i, config := range configs {
			merged, err := ApplyGitignore(config, rootDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to apply gitignore: %v\n", err)
				continue
			}
			configs[i] = merged
		}
	}

	return configs, found, nil
}

// findConfigsWithRoot finds configs and returns the root directory and whether any config was found on disk.
func findConfigsWithRoot(startPath string) ([]*Config, string, bool, error) {
	var configs []*Config
	var rootDir string

	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to get absolute path: %w", err)
	}

	currentPath := absPath
	rootDir = absPath // Default root to the start path
	found := false

	for {
		var configFound bool
		rootDir, configs, configFound, err = checkAndLoadConfig(currentPath, configs)
		if err != nil {
			return nil, "", false, err
		}
		if configFound {
			found = true
		}

		// If we found a root config, stop searching
		if rootDir != "" {
			break
		}

		// Check if we've reached the root directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}

	// If no config found, return a default config
	if len(configs) == 0 {
		return []*Config{{Rules: make(map[string]interface{})}}, absPath, false, nil
	}

	return configs, rootDir, found, nil
}

func checkAndLoadConfig(currentPath string, configs []*Config) (string, []*Config, bool, error) {
	rootDir := ""
	found := false

	configPath := filepath.Join(currentPath, ".structurelint.yml")
	if _, err := os.Stat(configPath); err == nil {
		config, err := Load(configPath)
		if err != nil {
			return "", nil, false, err
		}
		configs = append([]*Config{config}, configs...) // Prepend to maintain order
		found = true

		// Check if this config has Root: true, which means stop searching upwards
		if config.Root {
			return currentPath, configs, true, nil
		}
	}

	// Try .yaml extension
	configPath = filepath.Join(currentPath, ".structurelint.yaml")
	if _, err := os.Stat(configPath); err == nil && len(configs) == 0 {
		config, err := Load(configPath)
		if err != nil {
			return "", nil, false, err
		}
		configs = append([]*Config{config}, configs...)
		found = true
		if config.Root {
			return currentPath, configs, true, nil
		}
	}

	return rootDir, configs, found, nil
}

// Merge merges multiple configs into a single config
// Later configs override earlier ones
func Merge(configs ...*Config) *Config {
	result := &Config{
		Rules:     make(map[string]interface{}),
		Overrides: []Override{},
	}

	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge rules
		for key, value := range config.Rules {
			result.Rules[key] = value
		}

		// Append overrides (they are processed in order)
		result.Overrides = append(result.Overrides, config.Overrides...)

		// Append exclude patterns
		if len(config.Exclude) > 0 {
			result.Exclude = append(result.Exclude, config.Exclude...)
		}

		// Append layers (Phase 1)
		if len(config.Layers) > 0 {
			result.Layers = append(result.Layers, config.Layers...)
		}

		// Append entrypoints (Phase 2)
		if len(config.Entrypoints) > 0 {
			result.Entrypoints = append(result.Entrypoints, config.Entrypoints...)
		}

		// Root flag is taken from the last config that sets it
		if config.Root {
			result.Root = config.Root
		}
	}

	return result
}

// GetRuleConfig extracts a typed rule configuration from the config
func GetRuleConfig[T any](config *Config, ruleName string) (T, bool) {
	var result T
	value, exists := config.Rules[ruleName]
	if !exists {
		return result, false
	}

	// Handle disabled rules (value is 0 or false)
	switch v := value.(type) {
	case int:
		if v == 0 {
			return result, false
		}
	case bool:
		if !v {
			return result, false
		}
	}

	// Convert via JSON marshaling (handles YAML -> Go type conversion)
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return result, false
	}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return result, false
	}
	return result, true
}

// TypedRules converts the raw map[string]interface{} Rules to a type-safe RuleConfigs.
// Returns nil if no rules are configured.
func (c *Config) TypedRules() *RuleConfigs {
	if c == nil || c.Rules == nil || len(c.Rules) == 0 {
		return nil
	}

	result := &RuleConfigs{}

	if v, ok := GetRuleConfig[*MaxDepthConfig](c, "max-depth"); ok {
		result.MaxDepth = v
	}
	if v, ok := GetRuleConfig[*MaxFilesInDirConfig](c, "max-files-in-dir"); ok {
		result.MaxFilesInDir = v
	}
	if v, ok := GetRuleConfig[*MaxSubdirsConfig](c, "max-subdirs"); ok {
		result.MaxSubdirs = v
	}
	if v, ok := GetRuleConfig[NamingConventionConfig](c, "naming-convention"); ok {
		result.NamingConvention = v
	}
	if v, ok := GetRuleConfig[FileExistenceConfig](c, "file-existence"); ok {
		result.FileExistence = v
	}
	if v, ok := GetRuleConfig[RegexMatchConfig](c, "regex-match"); ok {
		result.RegexMatch = v
	}
	if v, ok := GetRuleConfig[DisallowedPatternsConfig](c, "disallowed-patterns"); ok {
		result.DisallowedPatterns = v
	}
	if v, ok := GetRuleConfig[*TestAdjacencyConfig](c, "test-adjacency"); ok {
		result.TestAdjacency = v
	}
	if v, ok := GetRuleConfig[*TestLocationConfig](c, "test-location"); ok {
		result.TestLocation = v
	}
	if v, ok := GetRuleConfig[*EnforceLayerBoundariesConfig](c, "enforce-layer-boundaries"); ok {
		result.EnforceLayerBoundaries = v
	}
	if v, ok := GetRuleConfig[*DisallowOrphanedFilesConfig](c, "disallow-orphaned-files"); ok {
		result.DisallowOrphanedFiles = v
	}
	if v, ok := GetRuleConfig[*DisallowImportCyclesConfig](c, "disallow-import-cycles"); ok {
		result.DisallowImportCycles = v
	}
	if v, ok := GetRuleConfig[*PathBasedLayersConfig](c, "path-based-layers"); ok {
		result.PathBasedLayers = v
	}

	return result
}
