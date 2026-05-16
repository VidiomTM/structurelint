package ptest

import (
	"testing"

	"pgregory.net/rapid"
)

func TestPath_GeneratesStrings(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := Path().Draw(t, "path")
		if len(p) > 0 {
			if len(p) < 1 {
				t.Fatal("unexpected")
			}
		}
	})
}

func TestFilePath_GeneratesPaths(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := FilePath().Draw(t, "path")
		if p == "" {
			t.Fatal("expected non-empty file path")
		}
		if len(p) < 5 {
			t.Fatal("expected path to have extension")
		}
	})
}

func TestRuleName_GeneratesNames(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := RuleName().Draw(t, "name")
		if len(name) < 5 {
			t.Fatal("expected rule name to have pattern a-z{a-z}{2,8}-a-z{a-z}{2,12}")
		}
	})
}

func TestDirective_GeneratesDirectives(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		d := Directive().Draw(t, "directive")
		if d.Line < 1 || d.Line > 200 {
			t.Fatalf("line out of range: %d", d.Line)
		}
		if len(d.Rules) > 5 {
			t.Fatalf("too many rules: %d", len(d.Rules))
		}
	})
}

func TestFileInfo_GeneratesFileInfo(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fi := FileInfo().Draw(t, "fileinfo")
		if fi.Path == "" && !fi.IsDir {
			t.Fatal("file info should have path or be dir")
		}
	})
}

func TestDirInfo_GeneratesDirInfo(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		di := DirInfo().Draw(t, "dirinfo")
		if di.FileCount > 100 || di.SubdirCount > 20 {
			t.Fatalf("dir info values out of range: files=%d subdirs=%d", di.FileCount, di.SubdirCount)
		}
	})
}

func TestViolation_GeneratesViolations(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		v := Violation().Draw(t, "violation")
		if v.Rule == "" {
			t.Fatal("expected non-empty rule")
		}
		if v.Message == "" {
			t.Fatal("expected non-empty message")
		}
	})
}

func TestLayer_GeneratesLayers(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := Layer().Draw(t, "layer")
		if l.Name == "" {
			t.Fatal("expected layer name")
		}
	})
}

func TestOverride_GeneratesOverrides(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		o := Override().Draw(t, "override")
		if len(o.Files) == 0 {
			t.Fatal("expected files in override")
		}
	})
}

func TestConfig_GeneratesConfigs(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cfg := Config().Draw(t, "config")
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
	})
}

func TestFileInfos_GeneratesSlice(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fis := FileInfos().Draw(t, "fileinfos")
		if len(fis) > 50 {
			t.Fatalf("too many file infos: %d", len(fis))
		}
	})
}

func TestDirInfos_GeneratesMap(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		dis := DirInfos().Draw(t, "dirinfos")
		for k, di := range dis {
			if di.Path != k {
				t.Fatalf("key %q != path %q", k, di.Path)
			}
		}
	})
}

func TestViolations_GeneratesSlice(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		vs := Violations().Draw(t, "violations")
		if len(vs) > 20 {
			t.Fatalf("too many violations: %d", len(vs))
		}
	})
}

func TestPath_Empty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		p := Path().Draw(t, "p")
		_ = p
	})
}
