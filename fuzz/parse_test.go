package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/parser"
)

func FuzzParse(f *testing.F) {
	corpus := []string{
		"import { x } from './utils';\n",
		"package main\n\nimport \"fmt\"\n",
		"import os\nfrom typing import List\n",
		"import com.example.Foo;\n",
		"#include <stdio.h>\n#include \"myheader.h\"\n",
		"using System;\nnamespace MyApp {\n}\n",
		"",
	}
	for _, c := range corpus {
		f.Add(c)
	}

	f.Fuzz(func(t *testing.T, content string) {
		exts := []string{".ts", ".go", ".py", ".java", ".cpp", ".cs"}
		for _, ext := range exts {
			tmpDir := t.TempDir()
			fpath := filepath.Join(tmpDir, "test"+ext)
			os.WriteFile(fpath, []byte(content), 0644)

			p := parser.New(tmpDir)
			p.ParseFile(fpath)
		}
	})
}
