package config

// RuleConfigs holds all rule configurations with type safety.
// Each field corresponds to a rule name in .structurelint.yml.
type RuleConfigs struct {
	MaxDepth               *MaxDepthConfig               `yaml:"max-depth,omitempty"`
	MaxFilesInDir          *MaxFilesInDirConfig          `yaml:"max-files-in-dir,omitempty"`
	MaxSubdirs             *MaxSubdirsConfig             `yaml:"max-subdirs,omitempty"`
	NamingConvention       NamingConventionConfig        `yaml:"naming-convention,omitempty"`
	FileExistence          FileExistenceConfig           `yaml:"file-existence,omitempty"`
	RegexMatch             RegexMatchConfig              `yaml:"regex-match,omitempty"`
	DisallowedPatterns     DisallowedPatternsConfig      `yaml:"disallowed-patterns,omitempty"`
	TestAdjacency          *TestAdjacencyConfig          `yaml:"test-adjacency,omitempty"`
	TestLocation           *TestLocationConfig           `yaml:"test-location,omitempty"`
	EnforceLayerBoundaries *EnforceLayerBoundariesConfig `yaml:"enforce-layer-boundaries,omitempty"`
	DisallowOrphanedFiles  *DisallowOrphanedFilesConfig  `yaml:"disallow-orphaned-files,omitempty"`
	DisallowImportCycles   *DisallowImportCyclesConfig   `yaml:"disallow-import-cycles,omitempty"`
	PathBasedLayers        *PathBasedLayersConfig        `yaml:"path-based-layers,omitempty"`
}

// MaxDepthConfig configures the max-depth rule.
type MaxDepthConfig struct {
	Max int `json:"max" yaml:"max"`
}

// MaxFilesInDirConfig configures the max-files-in-dir rule.
type MaxFilesInDirConfig struct {
	Max int `json:"max" yaml:"max"`
}

// MaxSubdirsConfig configures the max-subdirs rule.
type MaxSubdirsConfig struct {
	Max int `yaml:"max"`
}

// NamingConventionConfig configures the naming-convention rule.
// Maps file glob patterns to naming conventions (e.g., "*.ts" -> "camelCase").
type NamingConventionConfig map[string]string

// FileExistenceConfig configures the file-existence rule.
// Maps file patterns to existence requirements (e.g., "index.ts" -> "exists:1").
type FileExistenceConfig map[string]string

// RegexMatchConfig configures the regex-match rule.
// Maps file patterns to regex requirements.
type RegexMatchConfig map[string]string

// DisallowedPatternsConfig configures the disallowed-patterns rule.
// A list of glob patterns that are not allowed.
type DisallowedPatternsConfig []string

// TestAdjacencyConfig configures the test-adjacency rule.
type TestAdjacencyConfig struct {
	Pattern      string   `json:"pattern" yaml:"pattern"`
	TestDir      string   `json:"test-dir" yaml:"test-dir,omitempty"`
	FilePatterns []string `json:"file-patterns" yaml:"file-patterns,omitempty"`
	Exemptions   []string `json:"exemptions" yaml:"exemptions,omitempty"`
}

// TestLocationConfig configures the test-location rule.
type TestLocationConfig struct {
	IntegrationTestDir string   `yaml:"integration-test-dir,omitempty"`
	AllowAdjacent      bool     `yaml:"allow-adjacent,omitempty"`
	FilePatterns       []string `yaml:"file-patterns,omitempty"`
	Exemptions         []string `yaml:"exemptions,omitempty"`
}

// EnforceLayerBoundariesConfig configures the enforce-layer-boundaries rule.
// An empty struct means the rule only needs to be enabled (no parameters).
type EnforceLayerBoundariesConfig struct{}

// DisallowOrphanedFilesConfig configures the disallow-orphaned-files rule.
type DisallowOrphanedFilesConfig struct {
	EntryPointPatterns []string `json:"entry-point-patterns" yaml:"entry-point-patterns,omitempty"`
}

// DisallowImportCyclesConfig configures the disallow-import-cycles rule.
// An empty struct means the rule only needs to be enabled (no parameters).
type DisallowImportCyclesConfig struct{}

// PathLayerConfig represents a single layer in the path-based-layers rule.
type PathLayerConfig struct {
	Name           string   `json:"name" yaml:"name"`
	Patterns       []string `json:"patterns" yaml:"patterns,omitempty"`
	CanDependOn    []string `json:"canDependOn" yaml:"canDependOn,omitempty"`
	ForbiddenPaths []string `json:"forbiddenPaths" yaml:"forbiddenPaths,omitempty"`
}

// PathBasedLayersConfig configures the path-based-layers rule.
type PathBasedLayersConfig struct {
	Layers []PathLayerConfig `json:"layers" yaml:"layers,omitempty"`
}
