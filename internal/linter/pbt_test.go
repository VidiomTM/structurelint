package linter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	rulesgraph "github.com/Jonathangadeaharder/structurelint/internal/rules/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules/structure"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"pgregory.net/rapid"
)

func arbPathComponent() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-z][a-z0-9]{2,7}`)
}

func arbFileExtension() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{
		".ts", ".js", ".go", ".py", ".rs", ".tsx", ".jsx", ".java",
	})
}

func arbWalkerFile(t *rapid.T) walker.FileInfo {
	depth := rapid.IntRange(0, 6).Draw(t, "depth")
	var components []string
	for i := 0; i < depth; i++ {
		components = append(components, arbPathComponent().Draw(t, fmt.Sprintf("dir_%d", i)))
	}
	dir := filepath.Join(components...)
	ext := arbFileExtension().Draw(t, "ext")
	name := arbPathComponent().Draw(t, "filename") + ext
	relPath := name
	if dir != "" && dir != "." {
		relPath = filepath.Join(dir, name)
	}
	parentPath := dir
	if parentPath == "" {
		parentPath = "."
	}
	return walker.FileInfo{
		Path:       relPath,
		AbsPath:    filepath.Join("/root", relPath),
		IsDir:      false,
		Depth:      depth,
		ParentPath: parentPath,
	}
}

func arbWalkerFiles(t *rapid.T) []walker.FileInfo {
	n := rapid.IntRange(0, 30).Draw(t, "fileCount")
	files := make([]walker.FileInfo, n)
	for i := 0; i < n; i++ {
		files[i] = arbWalkerFile(t)
	}
	return files
}

func buildDirs(files []walker.FileInfo) map[string]*walker.DirInfo {
	dirs := make(map[string]*walker.DirInfo)
	seenChildren := make(map[string]map[string]struct{})
	dirs["."] = &walker.DirInfo{Path: ".", Depth: 0}

	for _, f := range files {
		parent := f.ParentPath
		if _, ok := dirs[parent]; !ok {
			dirs[parent] = &walker.DirInfo{Path: parent, Depth: f.Depth}
		}
		if !f.IsDir {
			dirs[parent].FileCount++
		}

		for d := parent; d != "" && d != "."; {
			p := filepath.Dir(d)
			if p == d {
				break
			}
			if _, ok := dirs[p]; !ok {
				dirs[p] = &walker.DirInfo{Path: p, Depth: depthOf(p)}
			}
			if seenChildren[p] == nil {
				seenChildren[p] = map[string]struct{}{}
			}
			if _, ok := seenChildren[p][d]; !ok {
				seenChildren[p][d] = struct{}{}
				dirs[p].SubdirCount++
			}
			d = p
		}
	}
	return dirs
}

func depthOf(path string) int {
	if path == "" || path == "." {
		return 0
	}
	return strings.Count(filepath.ToSlash(path), "/") + 1
}

func arbRule(t *rapid.T) rules.Rule {
	ruleType := rapid.IntRange(0, 5).Draw(t, "ruleType")
	switch ruleType {
	case 0:
		maxDepth := rapid.IntRange(1, 10).Draw(t, "maxDepth")
		return structure.NewMaxDepthRule(maxDepth)
	case 1:
		maxFiles := rapid.IntRange(1, 20).Draw(t, "maxFiles")
		return structure.NewMaxFilesRule(maxFiles)
	case 2:
		maxSubdirs := rapid.IntRange(1, 15).Draw(t, "maxSubdirs")
		return structure.NewMaxSubdirsRule(maxSubdirs)
	case 3:
		patternCount := rapid.IntRange(1, 5).Draw(t, "patternCount")
		patterns := make(map[string]string)
		for i := 0; i < patternCount; i++ {
			pat := rapid.SampledFrom([]string{"*.ts", "*.go", "*.py", "index.*", "*.test.js"}).Draw(t, fmt.Sprintf("pat_%d", i))
			req := rapid.SampledFrom([]string{"exists:0", "exists:1", "exists:0-5"}).Draw(t, fmt.Sprintf("req_%d", i))
			patterns[pat] = req
		}
		return structure.NewFileExistenceRule(patterns)
	case 4:
		patternCount := rapid.IntRange(1, 5).Draw(t, "disallowedCount")
		patterns := make([]string, patternCount)
		for i := 0; i < patternCount; i++ {
			patterns[i] = rapid.SampledFrom([]string{
				"*.bak", "*.tmp", "*.log", "**/generated/**", "*.secret",
			}).Draw(t, fmt.Sprintf("disPat_%d", i))
		}
		return structure.NewDisallowedPatternsRule(patterns)
	default:
		patternCount := rapid.IntRange(1, 4).Draw(t, "namingPatternCount")
		patterns := make(map[string]string)
		for i := 0; i < patternCount; i++ {
			pat := rapid.SampledFrom([]string{"*.ts", "*.go", "*.py", "*.rs"}).Draw(t, fmt.Sprintf("namingPat_%d", i))
			conv := rapid.SampledFrom([]string{"camelCase", "PascalCase", "snake_case", "kebab-case"}).Draw(t, fmt.Sprintf("namingConv_%d", i))
			patterns[pat] = conv
		}
		return structure.NewNamingConventionRule(patterns)
	}
}

func arbPathBasedLayerRule(t *rapid.T) rules.Rule {
	layerCount := rapid.IntRange(2, 4).Draw(t, "layerCount")
	layers := make([]rulesgraph.PathLayer, layerCount)
	for i := 0; i < layerCount; i++ {
		layers[i] = rulesgraph.PathLayer{
			Name:     fmt.Sprintf("layer_%d", i),
			Patterns: []string{fmt.Sprintf("layer%d/**", i)},
		}
		if i > 0 {
			deps := make([]string, i)
			for j := 0; j < i; j++ {
				deps[j] = fmt.Sprintf("layer_%d", j)
			}
			layers[i].CanDependOn = deps
		}
	}
	return rulesgraph.NewPathBasedLayerRule(layers)
}

func violationsKey(v []rules.Violation) string {
	paths := make([]string, len(v))
	for i, vi := range v {
		paths[i] = fmt.Sprintf("%s|%s|%s", vi.Rule, vi.Path, vi.Message)
	}
	sort.Strings(paths)
	var sb strings.Builder
	for _, p := range paths {
		sb.WriteString(p)
		sb.WriteByte(';')
	}
	return sb.String()
}

func filesForDeepPath(base string, depth int, ext string) []walker.FileInfo {
	var files []walker.FileInfo
	components := []string{base}
	for i := 0; i <= depth; i++ {
		components = append(components, fmt.Sprintf("level%d", i))
	}
	dir := filepath.Join(components...)
	files = append(files, walker.FileInfo{
		Path:       filepath.Join(dir, "deep_file"+ext),
		AbsPath:    filepath.Join("/root", dir, "deep_file"+ext),
		IsDir:      false,
		Depth:      depth + 2,
		ParentPath: dir,
	})
	return files
}

func filesForManyFiles(dir string, count int, ext string) []walker.FileInfo {
	files := make([]walker.FileInfo, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("file_%d%s", i, ext)
		files[i] = walker.FileInfo{
			Path:       filepath.Join(dir, name),
			AbsPath:    filepath.Join("/root", dir, name),
			IsDir:      false,
			Depth:      1,
			ParentPath: dir,
		}
	}
	return files
}

func TestLintRuleIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		files := arbWalkerFiles(t)
		dirs := buildDirs(files)
		rule := arbRule(t)

		v1 := rule.Check(files, dirs)
		v2 := rule.Check(files, dirs)

		if len(v1) != len(v2) {
			t.Fatalf("violation count mismatch: %d vs %d for rule %s", len(v1), len(v2), rule.Name())
		}
		if violationsKey(v1) != violationsKey(v2) {
			t.Fatalf("violations differ between runs for rule %s", rule.Name())
		}
	})
}

func TestLintRuleDeterminism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		files := arbWalkerFiles(t)
		dirs := buildDirs(files)
		rule := arbRule(t)

		first := violationsKey(rule.Check(files, dirs))
		for i := 0; i < 99; i++ {
			run := violationsKey(rule.Check(files, dirs))
			if first != run {
				t.Fatalf("non-deterministic rule %s: run %d differs from first", rule.Name(), i+1)
			}
		}
	})
}

func TestMaxDepthNoFalseNegatives(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxDepth := rapid.IntRange(1, 5).Draw(t, "maxDepth")
		extraDepth := rapid.IntRange(1, 5).Draw(t, "extraDepth")
		violatingDepth := maxDepth + extraDepth

		ext := arbFileExtension().Draw(t, "ext")
		files := filesForDeepPath("src", violatingDepth, ext)
		dirs := buildDirs(files)

		rule := structure.NewMaxDepthRule(maxDepth)
		violations := rule.Check(files, dirs)

		if len(violations) == 0 {
			t.Fatalf("false negative: depth %d > max %d not detected", violatingDepth, maxDepth)
		}

		for _, v := range violations {
			if v.Rule != "max-depth" {
				t.Fatalf("expected rule 'max-depth', got '%s'", v.Rule)
			}
		}
	})
}

func TestMaxFilesNoFalseNegatives(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxFiles := rapid.IntRange(1, 5).Draw(t, "maxFiles")
		extraFiles := rapid.IntRange(1, 10).Draw(t, "extraFiles")
		totalFiles := maxFiles + extraFiles

		ext := arbFileExtension().Draw(t, "ext")
		files := filesForManyFiles("src", totalFiles, ext)
		dirs := buildDirs(files)

		rule := structure.NewMaxFilesRule(maxFiles)
		violations := rule.Check(files, dirs)

		if len(violations) == 0 {
			t.Fatalf("false negative: %d files > max %d not detected", totalFiles, maxFiles)
		}
	})
}

func TestMaxSubdirsNoFalseNegatives(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		maxSubdirs := rapid.IntRange(1, 4).Draw(t, "maxSubdirs")
		extraSubdirs := rapid.IntRange(1, 6).Draw(t, "extraSubdirs")
		totalSubdirs := maxSubdirs + extraSubdirs

		ext := arbFileExtension().Draw(t, "ext")
		parentDir := "modules"
		files := make([]walker.FileInfo, totalSubdirs)
		for i := 0; i < totalSubdirs; i++ {
			subdirName := fmt.Sprintf("sub%d", i)
			files[i] = walker.FileInfo{
				Path:       filepath.Join(parentDir, subdirName, "mod"+ext),
				AbsPath:    filepath.Join("/root", parentDir, subdirName, "mod"+ext),
				IsDir:      false,
				Depth:      2,
				ParentPath: filepath.Join(parentDir, subdirName),
			}
		}
		dirs := buildDirs(files)

		rule := structure.NewMaxSubdirsRule(maxSubdirs)
		violations := rule.Check(files, dirs)

		if len(violations) == 0 {
			t.Fatalf("false negative: %d subdirs > max %d not detected", totalSubdirs, maxSubdirs)
		}
	})
}

func TestDisallowedPatternsNoFalseNegatives(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		disallowedExt := rapid.SampledFrom([]string{".bak", ".tmp", ".log"}).Draw(t, "disallowedExt")
		pattern := "*" + disallowedExt

		files := []walker.FileInfo{
			{
				Path:       "src/config" + disallowedExt,
				AbsPath:    "/root/src/config" + disallowedExt,
				IsDir:      false,
				Depth:      1,
				ParentPath: "src",
			},
		}
		dirs := buildDirs(files)

		rule := structure.NewDisallowedPatternsRule([]string{pattern})
		violations := rule.Check(files, dirs)

		if len(violations) == 0 {
			t.Fatalf("false negative: file matching '%s' not detected", pattern)
		}
	})
}

func TestNamingConventionNoFalseNegatives(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		convention := rapid.SampledFrom([]string{"camelCase", "PascalCase", "snake_case", "kebab-case"}).Draw(t, "convention")
		violatingName := ""

		switch convention {
		case "camelCase":
			violatingName = "My_Component.ts"
		case "PascalCase":
			violatingName = "myComponent.ts"
		case "snake_case":
			violatingName = "my-component.ts"
		case "kebab-case":
			violatingName = "my_component.ts"
		}

		files := []walker.FileInfo{
			{
				Path:       "src/" + violatingName,
				AbsPath:    "/root/src/" + violatingName,
				IsDir:      false,
				Depth:      1,
				ParentPath: "src",
			},
		}
		dirs := buildDirs(files)

		rule := structure.NewNamingConventionRule(map[string]string{"*.ts": convention})
		violations := rule.Check(files, dirs)

		if len(violations) == 0 {
			t.Fatalf("false negative: '%s' violates '%s' but not detected", violatingName, convention)
		}
	})
}

func TestPathBasedLayerRuleIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		layerCount := rapid.IntRange(2, 4).Draw(t, "layerCount")
		fileCount := rapid.IntRange(1, 10).Draw(t, "fileCount")

		layers := make([]rulesgraph.PathLayer, layerCount)
		var files []walker.FileInfo
		for i := 0; i < layerCount; i++ {
			layers[i] = rulesgraph.PathLayer{
				Name:     fmt.Sprintf("layer_%d", i),
				Patterns: []string{fmt.Sprintf("layer%d/**", i)},
			}
			if i > 0 {
				deps := make([]string, i)
				for j := 0; j < i; j++ {
					deps[j] = fmt.Sprintf("layer_%d", j)
				}
				layers[i].CanDependOn = deps
			}

			for j := 0; j < fileCount; j++ {
				name := fmt.Sprintf("file_%d.ts", j)
				files = append(files, walker.FileInfo{
					Path:       filepath.Join(fmt.Sprintf("layer%d", i), name),
					AbsPath:    filepath.Join("/root", fmt.Sprintf("layer%d", i), name),
					IsDir:      false,
					Depth:      1,
					ParentPath: fmt.Sprintf("layer%d", i),
				})
			}
		}

		dirs := buildDirs(files)
		rule := rulesgraph.NewPathBasedLayerRule(layers)

		v1 := rule.Check(files, dirs)
		v2 := rule.Check(files, dirs)

		if len(v1) != len(v2) {
			t.Fatalf("path-layer rule violation count mismatch: %d vs %d", len(v1), len(v2))
		}
		if violationsKey(v1) != violationsKey(v2) {
			t.Fatalf("path-layer rule violations differ between runs")
		}
	})
}
