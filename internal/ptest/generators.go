package ptest

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"

	"pgregory.net/rapid"
)

func Path() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		parts := rapid.SliceOfN(rapid.StringMatching(`[a-z][a-z0-9_]{0,15}`), 0, 6).Draw(t, "segments")
		if len(parts) == 0 {
			return ""
		}
		return strings.Join(parts, "/")
	})
}

func FilePath() *rapid.Generator[string] {
	exts := []string{".go", ".ts", ".js", ".py", ".rs", ".java", ".yaml", ".yml", ".json", ".md"}
	return rapid.Custom(func(t *rapid.T) string {
		p := Path().Draw(t, "path")
		ext := rapid.SampledFrom(exts).Draw(t, "ext")
		if p == "" {
			return "file" + ext
		}
		return p + "/file" + ext
	})
}

func RuleName() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-z]{2,8}-[a-z]{2,12}`)
}

func Directive() *rapid.Generator[parser.Directive] {
	return rapid.Custom(func(t *rapid.T) parser.Directive {
		dt := rapid.SampledFrom([]parser.DirectiveType{
			parser.DirectiveIgnore,
			parser.DirectiveNoTest,
		}).Draw(t, "directive_type")

		ruleList := rapid.SliceOfN(RuleName(), 0, 5).Draw(t, "rules")
		reason := rapid.StringN(0, 80, 120).Draw(t, "reason")
		line := rapid.IntRange(1, 200).Draw(t, "line")

		return parser.Directive{
			Type:   dt,
			Rules:  ruleList,
			Reason: reason,
			Line:   line,
		}
	})
}

func FileInfo() *rapid.Generator[walker.FileInfo] {
	return rapid.Custom(func(t *rapid.T) walker.FileInfo {
		isDir := rapid.Bool().Draw(t, "is_dir")

		var path string
		if isDir {
			path = Path().Draw(t, "dir_path")
		} else {
			path = FilePath().Draw(t, "file_path")
		}

		depth := rapid.IntRange(0, 20).Draw(t, "depth")
		parent := Path().Draw(t, "parent")
		dirs := rapid.SliceOfN(Directive(), 0, 3).Draw(t, "directives")

		return walker.FileInfo{
			Path:       path,
			AbsPath:    "/root/" + path,
			IsDir:      isDir,
			Depth:      depth,
			ParentPath: parent,
			Directives: dirs,
		}
	})
}

func DirInfo() *rapid.Generator[walker.DirInfo] {
	return rapid.Custom(func(t *rapid.T) walker.DirInfo {
		return walker.DirInfo{
			Path:        Path().Draw(t, "path"),
			FileCount:   rapid.IntRange(0, 100).Draw(t, "file_count"),
			SubdirCount: rapid.IntRange(0, 20).Draw(t, "subdir_count"),
			Depth:       rapid.IntRange(0, 15).Draw(t, "depth"),
		}
	})
}

func Violation() *rapid.Generator[rules.Violation] {
	return rapid.Custom(func(t *rapid.T) rules.Violation {
		sliceOfStrings := rapid.SliceOfN(rapid.StringN(1, 40, 80), 0, 5)
		return rules.Violation{
			Rule:        RuleName().Draw(t, "rule"),
			Path:        FilePath().Draw(t, "path"),
			Message:     rapid.StringN(5, 100, 200).Draw(t, "message"),
			Expected:    rapid.StringN(0, 30, 60).Draw(t, "expected"),
			Actual:      rapid.StringN(0, 30, 60).Draw(t, "actual"),
			Suggestions: sliceOfStrings.Draw(t, "suggestions"),
			Context:     rapid.StringN(0, 50, 100).Draw(t, "context"),
		}
	})
}

func Layer() *rapid.Generator[config.Layer] {
	return rapid.Custom(func(t *rapid.T) config.Layer {
		return config.Layer{
			Name:      rapid.StringMatching(`[a-z]{3,12}`).Draw(t, "name"),
			Path:      Path().Draw(t, "path"),
			DependsOn: rapid.SliceOf(rapid.StringMatching(`[a-z]{3,12}`)).Draw(t, "depends_on"),
		}
	})
}

func Override() *rapid.Generator[config.Override] {
	return rapid.Custom(func(t *rapid.T) config.Override {
		glob := rapid.StringMatching(`\*\*?\.[a-z]{1,5}`).Draw(t, "glob")
		ruleName := RuleName().Draw(t, "rule_name")
		ruleMap := map[string]interface{}{
			ruleName: rapid.IntRange(0, 100).Draw(t, "rule_value"),
		}
		return config.Override{
			Files: []string{glob},
			Rules: ruleMap,
		}
	})
}

func Config() *rapid.Generator[*config.Config] {
	return rapid.Custom(func(t *rapid.T) *config.Config {
		layers := rapid.SliceOfN(Layer(), 0, 5).Draw(t, "layers")
		overrides := rapid.SliceOfN(Override(), 0, 3).Draw(t, "overrides")
		exclude := rapid.SliceOfN(
			rapid.StringMatching(`[a-z*]{1,10}`), 0, 5,
		).Draw(t, "exclude")
		entrypoints := rapid.SliceOfN(
			rapid.StringMatching(`[a-z_]{2,12}\.(go|ts|js)`), 0, 5,
		).Draw(t, "entrypoints")

		ruleNames := []string{
			"max-depth", "max-files-in-dir", "max-subdirs",
			"naming-convention", "disallowed-patterns",
		}
		ruleMap := make(map[string]interface{})
		for _, name := range ruleNames {
			if rapid.Bool().Draw(t, fmt.Sprintf("has_%s", name)) {
				ruleMap[name] = rapid.IntRange(1, 100).Draw(t, name)
			}
		}

		return &config.Config{
			Root:        rapid.Bool().Draw(t, "root"),
			Exclude:     exclude,
			Rules:       ruleMap,
			Overrides:   overrides,
			Layers:      layers,
			Entrypoints: entrypoints,
		}
	})
}

func FileInfos() *rapid.Generator[[]walker.FileInfo] {
	return rapid.SliceOfN(FileInfo(), 0, 50)
}

func DirInfos() *rapid.Generator[map[string]*walker.DirInfo] {
	return rapid.Custom(func(t *rapid.T) map[string]*walker.DirInfo {
		dirs := rapid.SliceOfN(DirInfo(), 0, 20).Draw(t, "dirs")
		m := make(map[string]*walker.DirInfo, len(dirs))
		for i := range dirs {
			m[dirs[i].Path] = &dirs[i]
		}
		return m
	})
}

func Violations() *rapid.Generator[[]rules.Violation] {
	return rapid.SliceOfN(Violation(), 0, 20)
}
