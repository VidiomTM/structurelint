package walker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func arbDirTree(t *rapid.T) (string, int, int) {
	tmpDir, err := os.MkdirTemp("", "walker-pbt-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	numDirs := rapid.IntRange(0, 10).Draw(t, "numDirs")
	numFiles := rapid.IntRange(0, 10).Draw(t, "numFiles")

	createdDirs := []string{tmpDir}
	for i := 0; i < numDirs; i++ {
		parentIdx := rapid.IntRange(0, len(createdDirs)-1).Draw(t, "parentIdx")
		name := rapid.StringMatching(`[a-z][a-z0-9]{0,7}`).Draw(t, fmt.Sprintf("dirName%d", i))
		parent := createdDirs[parentIdx]
		dirPath := filepath.Join(parent, name)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dirPath, err)
		}
		createdDirs = append(createdDirs, dirPath)
	}

	seen := make(map[string]bool)
	actualFiles := 0
	for i := 0; i < numFiles; i++ {
		parentIdx := rapid.IntRange(0, len(createdDirs)-1).Draw(t, "fileParentIdx")
		name := rapid.StringMatching(`[a-z][a-z0-9]{0,7}\.go`).Draw(t, fmt.Sprintf("fileName%d", i))
		parent := createdDirs[parentIdx]
		filePath := filepath.Join(parent, name)
		if seen[filePath] {
			continue
		}
		seen[filePath] = true
		if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", filePath, err)
		}
		actualFiles++
	}

	uniqueDirs := make(map[string]bool)
	for _, d := range createdDirs[1:] {
		uniqueDirs[d] = true
	}

	return tmpDir, len(uniqueDirs), actualFiles
}

func TestPBT_TraversalCompleteness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		root, expectedDirs, expectedFiles := arbDirTree(t)

		w := New(root)
		if err := w.Walk(); err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		files := w.GetFiles()
		fileCount := 0
		dirCount := 0
		for _, f := range files {
			if f.IsDir {
				dirCount++
			} else {
				fileCount++
			}
		}

		if fileCount != expectedFiles {
			t.Fatalf("expected %d files, visited %d", expectedFiles, fileCount)
		}
		if dirCount != expectedDirs {
			t.Fatalf("expected %d dirs, visited %d", expectedDirs, dirCount)
		}
	})
}

func TestPBT_NodeTypeSafety(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		root, _, _ := arbDirTree(t)

		w := New(root)
		if err := w.Walk(); err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		for _, f := range w.GetFiles() {
			if f.Path == "" {
				t.Fatalf("empty path in results")
			}
			if f.AbsPath == "" {
				t.Fatalf("empty absPath for %s", f.Path)
			}
			if f.Depth < 0 {
				t.Fatalf("negative depth for %s", f.Path)
			}
		}
	})
}

func TestPBT_DepthTracking(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		root, _, _ := arbDirTree(t)

		w := New(root)
		if err := w.Walk(); err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		for _, f := range w.GetFiles() {
			expectedDepth := strings.Count(f.Path, string(filepath.Separator))
			if f.IsDir {
				expectedDepth++
			}
			if f.Depth != expectedDepth {
				t.Fatalf("path=%q isDir=%v depth=%d expected=%d", f.Path, f.IsDir, f.Depth, expectedDepth)
			}
		}
	})
}
