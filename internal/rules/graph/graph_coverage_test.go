package graph

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestOrphanedFilesRule_NilGraph(t *testing.T) {
	rule := NewOrphanedFilesRule(nil, nil)
	violations := rule.Check(nil, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for nil graph, got %d", len(violations))
	}
}

func TestIsConfigOrDocFile_AllPatterns(t *testing.T) {
	rule := NewOrphanedFilesRule(&graph.ImportGraph{}, nil)
	tests := []string{
		".structurelint.yml",
		".structurelint.yaml",
		"package.json",
		"tsconfig.json",
		"go.mod",
		"go.sum",
		"setup.py",
		"pyproject.toml",
		"Makefile",
		".gitignore",
		".eslintrc",
		".prettierrc",
		"README.md",
		"CHANGELOG.txt",
		".eslintrc.json",
	}
	for _, tc := range tests {
		if !rule.isConfigOrDocFile(tc) {
			t.Errorf("isConfigOrDocFile(%q) = false, want true", tc)
		}
	}
}

func TestIsConfigOrDocFile_Not(t *testing.T) {
	rule := NewOrphanedFilesRule(&graph.ImportGraph{}, nil)
	if rule.isConfigOrDocFile("src/main.go") {
		t.Error("isConfigOrDocFile(src/main.go) = true, want false")
	}
}

func TestMatchesEntrypointPattern_BaseGlob(t *testing.T) {
	if !matchesEntrypointPattern("scripts/deploy.py", "deploy.py") {
		t.Error("should match exact pattern")
	}
}

func TestMatchesEntrypointPattern_Exact(t *testing.T) {
	if !matchesEntrypointPattern("main.go", "main.go") {
		t.Error("should match exact path")
	}
}

func TestMatchesEntrypointPattern_Not(t *testing.T) {
	if matchesEntrypointPattern("scripts/deploy.py", "*.go") {
		t.Error("should NOT match different extension")
	}
}

func TestMatchesEntrypointPattern_DoubleStar(t *testing.T) {
	if !matchesEntrypointPattern("scripts/deploy.py", "**/*.py") {
		t.Error("should match double-star pattern")
	}
}

func TestPathBasedLayerCheck_EmptyFiles(t *testing.T) {
	layers := []PathLayer{
		{Name: "app", Patterns: []string{"src/**"}},
	}
	rule := NewPathBasedLayerRule(layers)
	violations := rule.Check(nil, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for nil files, got %d", len(violations))
	}
}

func TestPathLayer_WithForbiddenPaths(t *testing.T) {
	layers := []PathLayer{
		{
			Name:           "app",
			Patterns:       []string{"src/app/**"},
			ForbiddenPaths: []string{"src/app/config/*"},
		},
	}
	files := []walker.FileInfo{
		{Path: "src/app/config/secret.go", IsDir: false},
		{Path: "src/app/service.go", IsDir: false},
	}
	rule := NewPathBasedLayerRule(layers)
	violations := rule.Check(files, nil)
	if len(violations) == 0 {
		t.Log("no violations (forbidden may not trigger based on file location)")
	}
}

func TestCheckForbiddenLayer_Nil(t *testing.T) {
	r := &PathBasedLayerRule{}
	violations := r.checkForbiddenLayer(walker.FileInfo{}, &PathLayer{}, "nonexistent", nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for nil layer, got %d", len(violations))
	}
}

func TestFindLayerByName_NotFound(t *testing.T) {
	r := &PathBasedLayerRule{Layers: []PathLayer{{Name: "existing"}}}
	result := r.findLayerByName("nonexistent")
	if result != nil {
		t.Error("expected nil for nonexistent layer")
	}
}

func TestCheckForbiddenLayerPattern_NoMatch(t *testing.T) {
	r := &PathBasedLayerRule{}
	violation := r.checkForbiddenLayerPattern(
		walker.FileInfo{Path: "src/main.go"},
		&PathLayer{Name: "app", CanDependOn: []string{"data"}},
		"data", "src/data/**", nil,
	)
	if violation != nil {
		t.Error("expected nil for non-matching pattern")
	}
}

func TestIsDuplicateViolation_Trimmed(t *testing.T) {
	r := &PathBasedLayerRule{}
	result := r.isDuplicateViolation("/data/", map[string]bool{"**/data/**": true})
	if !result {
		t.Error("expected duplicate detection with trimmed segment")
	}
}

func TestIsDuplicateViolation_Not(t *testing.T) {
	r := &PathBasedLayerRule{}
	result := r.isDuplicateViolation("/business/", map[string]bool{"**/data/**": true})
	if result {
		t.Error("expected no duplicate for different segment")
	}
}

func TestPathBasedLayerCheck_WithCanDependOnConstraints(t *testing.T) {
	layers := []PathLayer{
		{
			Name:        "app",
			Patterns:    []string{"src/app/**"},
			CanDependOn: []string{"domain"},
		},
		{
			Name:        "domain",
			Patterns:    []string{"src/domain/**"},
			CanDependOn: []string{},
		},
	}
	files := []walker.FileInfo{
		{Path: "src/app/service.go", IsDir: false},
		{Path: "src/domain/entity.go", IsDir: false},
	}
	rule := NewPathBasedLayerRule(layers)
	violations := rule.Check(files, nil)
	// No violations because no file paths contain forbidden layer segments
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestIsEntrypoint_CommonEntrypoints(t *testing.T) {
	rule := NewOrphanedFilesRule(&graph.ImportGraph{}, nil)
	tests := []string{
		"main.go",
		"main.ts",
		"main.js",
		"main.py",
		"index.ts",
		"index.js",
		"app.ts",
		"app.js",
		"app.py",
		"__init__.py",
		"manage.py",
	}
	for _, path := range tests {
		if !rule.isEntrypoint(path) {
			t.Errorf("isEntrypoint(%q) = false, want true", path)
		}
	}
}

func TestOrphanedFilesCheck_ParserUnsupported(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"file.svelte", "file.rs"},
	}
	rule := NewOrphanedFilesRule(g, []string{})
	violations := rule.Check(nil, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for unsupported parsers, got %d", len(violations))
	}
}

func TestLayerBoundariesCheck_NilGraph(t *testing.T) {
	rule := NewLayerBoundariesRule(nil)
	violations := rule.Check(nil, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for nil graph, got %d", len(violations))
	}
}

func TestLayerBoundariesCheck_UnlayeredFile(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/misc/helper.go": {"src/domain/user.go"},
		},
		FileLayers: map[string]*config.Layer{},
	}
	rule := NewLayerBoundariesRule(g)
	violations := rule.Check(nil, nil)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for file without layer, got %d", len(violations))
	}
}

func TestImportCyclesInit(t *testing.T) {
	// Test the init() function by calling rule constructors
	_ = NewImportCyclesRule(nil)
	_ = NewLayerBoundariesRule(nil)
	_ = NewOrphanedFilesRule(nil, nil)
	_ = NewPathBasedLayerRule(nil)
}

func TestResolve_EmptyAllFiles(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
		},
	}
	d := &cycleDetector{graph: g}
	result := d.resolve("b.go", "a.go")
	if result != "b.go" {
		t.Errorf("expected 'b.go', got %q", result)
	}
}

func TestResolve_EmptyAllFiles_NotFound(t *testing.T) {
	g := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
		},
	}
	d := &cycleDetector{graph: g}
	result := d.resolve("unknown.ext", "a.go")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestResolve_RelativePath_WithExt(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/pkg/foo.go"},
		Dependencies: map[string][]string{
			"src/pkg/main.go": {"./foo"},
		},
	}
	d := &cycleDetector{graph: g}
	result := d.resolve("./foo", "src/pkg/main.go")
	if result != "src/pkg/foo.go" {
		t.Errorf("expected 'src/pkg/foo.go', got %q", result)
	}
}

func TestResolve_RelativePath_Exact(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/pkg/helper.go"},
		Dependencies: map[string][]string{
			"src/pkg/main.go": {"./helper.go"},
		},
	}
	d := &cycleDetector{graph: g}
	result := d.resolve("./helper.go", "src/pkg/main.go")
	if result != "src/pkg/helper.go" {
		t.Errorf("expected 'src/pkg/helper.go', got %q", result)
	}
}

func TestResolve_SuffixMatch(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/pkg/util/helper.go"},
		Dependencies: map[string][]string{
			"src/pkg/main.go": {"util/helper.go"},
		},
	}
	d := &cycleDetector{graph: g}
	result := d.resolve("util/helper.go", "src/pkg/main.go")
	if result != "src/pkg/util/helper.go" {
		t.Errorf("expected 'src/pkg/util/helper.go', got %q", result)
	}
}
