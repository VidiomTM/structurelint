package autofix

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine(true)
	require.NotNil(t, e)
	assert.True(t, e.dryRun)

	e2 := NewEngine(false)
	assert.False(t, e2.dryRun)
}

func TestRegisterFixer(t *testing.T) {
	e := NewEngine(true)
	assert.Equal(t, 1, len(e.fixers))
	e.RegisterFixer(&FileLocationFixer{})
	assert.Equal(t, 2, len(e.fixers))
}

func TestGenerateFixes_WithAutoFix(t *testing.T) {
	e := NewEngine(true)
	violations := []rules.Violation{
		{
			Rule:    "test-rule",
			Path:    "test.go",
			Message: "test violation",
			AutoFix: &rules.AutoFix{
				FilePath: "test.go",
				Content:  "fixed content",
			},
		},
	}
	fixes, err := e.GenerateFixes(violations, nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(fixes))
	assert.Equal(t, 0.95, fixes[0].Confidence)
	assert.True(t, fixes[0].Safe)
	assert.Equal(t, "Apply auto-fix for test-rule", fixes[0].Description)
}

func TestGenerateFixes_NoAutoFix(t *testing.T) {
	e := NewEngine(true)
	violations := []rules.Violation{
		{Rule: "test-rule", Path: "test.go", Message: "no autofix"},
	}
	fixes, err := e.GenerateFixes(violations, nil)
	require.NoError(t, err)
	assert.Empty(t, fixes)
}

func TestApplyFixes_DryRun(t *testing.T) {
	e := NewEngine(true)
	fixes := []*Fix{
		{
			Violation:   rules.Violation{Rule: "test", Path: "test.go"},
			Description: "Test fix",
			Actions:     []Action{},
		},
	}
	count, err := e.ApplyFixes(fixes)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMoveFileAction_Apply(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.go")
	dst := filepath.Join(dir, "dst.go")
	require.NoError(t, os.WriteFile(src, []byte("package main"), 0644))

	a := &MoveFileAction{SourcePath: src, TargetPath: dst}
	err := a.Apply()
	require.NoError(t, err)

	_, err = os.Stat(src)
	assert.True(t, os.IsNotExist(err))

	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "package main", string(data))
}

func TestMoveFileAction_Apply_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.go")
	dst := filepath.Join(dir, "sub", "dst.go")
	require.NoError(t, os.WriteFile(src, []byte("package main"), 0644))

	a := &MoveFileAction{SourcePath: src, TargetPath: dst}
	err := a.Apply()
	require.NoError(t, err)

	_, err = os.Stat(src)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(dst)
	assert.NoError(t, err)
}

func TestMoveFileAction_Describe(t *testing.T) {
	a := &MoveFileAction{SourcePath: "a.go", TargetPath: "b.go"}
	assert.Equal(t, "Move a.go → b.go", a.Describe())
}

func TestMoveFileAction_Revert(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.go")
	dst := filepath.Join(dir, "dst.go")
	require.NoError(t, os.WriteFile(src, []byte("package main"), 0644))

	a := &MoveFileAction{SourcePath: src, TargetPath: dst}
	require.NoError(t, a.Apply())
	require.NoError(t, a.Revert())

	data, err := os.ReadFile(src)
	require.NoError(t, err)
	assert.Equal(t, "package main", string(data))
	_, err = os.Stat(dst)
	assert.True(t, os.IsNotExist(err))
}

func TestMoveFileAction_Revert_NoOriginal(t *testing.T) {
	a := &MoveFileAction{SourcePath: "a.go", TargetPath: "b.go"}
	err := a.Revert()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no original content")
}

func TestMoveFileAction_Apply_NonexistentSource(t *testing.T) {
	a := &MoveFileAction{SourcePath: "/nonexistent/file.go", TargetPath: "/tmp/dst.go"}
	err := a.Apply()
	assert.Error(t, err)
}

func TestWriteFileAction_Apply(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.go")

	a := &WriteFileAction{FilePath: path, Content: "package main"}
	err := a.Apply()
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "package main", string(data))
}

func TestWriteFileAction_Apply_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.go")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0644))

	a := &WriteFileAction{FilePath: path, Content: "updated"}
	err := a.Apply()
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "updated", string(data))
	assert.True(t, a.originalExists)
}

func TestWriteFileAction_Describe(t *testing.T) {
	a := &WriteFileAction{FilePath: "new.go", Content: "x"}
	assert.Equal(t, "Create new.go", a.Describe())

	a2 := &WriteFileAction{FilePath: "existing.go", Content: "x", originalExists: true}
	assert.Equal(t, "Update existing.go", a2.Describe())
}

func TestWriteFileAction_Revert_RestoreBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "revert.go")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0644))

	a := &WriteFileAction{FilePath: path, Content: "updated"}
	require.NoError(t, a.Apply())
	require.NoError(t, a.Revert())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "original", string(data))
}

func TestWriteFileAction_Revert_RemoveNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new_revert.go")

	a := &WriteFileAction{FilePath: path, Content: "new"}
	require.NoError(t, a.Apply())
	require.NoError(t, a.Revert())

	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestUpdateImportAction_Apply(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	require.NoError(t, os.WriteFile(path, []byte("package main"), 0644))

	a := &UpdateImportAction{FilePath: path, OldImport: "old/pkg", NewImport: "new/pkg"}
	err := a.Apply()
	require.NoError(t, err)
	assert.NotEmpty(t, a.backupPath)
}

func TestUpdateImportAction_Describe(t *testing.T) {
	a := &UpdateImportAction{FilePath: "main.go", OldImport: "old", NewImport: "new"}
	assert.Equal(t, "Update import in main.go: old → new", a.Describe())
}

func TestUpdateImportAction_Revert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "revert_import.go")
	require.NoError(t, os.WriteFile(path, []byte("package main"), 0644))

	a := &UpdateImportAction{FilePath: path, OldImport: "old", NewImport: "new"}
	require.NoError(t, a.Apply())

	data1, _ := os.ReadFile(path)
	require.NoError(t, a.Revert())

	data2, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, data1, data2)
}

func TestUpdateImportAction_Revert_NoBackup(t *testing.T) {
	a := &UpdateImportAction{FilePath: "main.go"}
	err := a.Revert()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backup")
}

func TestFixResult_String(t *testing.T) {
	r := &FixResult{Applied: 3, Failed: 1, Skipped: 2, DryRun: false}
	assert.Equal(t, "Applied 3 fixes, 1 failed, 2 skipped", r.String())

	r2 := &FixResult{Applied: 3, Skipped: 1, DryRun: true}
	assert.Equal(t, "Dry run: 3 fixes would be applied, 1 skipped", r2.String())
}

func TestFileLocationFixer_CanFix(t *testing.T) {
	f := NewFileLocationFixer()
	assert.True(t, f.CanFix(rules.Violation{Rule: "test-location"}))
	assert.True(t, f.CanFix(rules.Violation{Rule: "file-location"}))
	assert.True(t, f.CanFix(rules.Violation{Rule: "other", Expected: "expected", Message: "should be in src/"}))
	assert.False(t, f.CanFix(rules.Violation{Rule: "naming-convention"}))
}

func TestFileLocationFixer_GenerateFix(t *testing.T) {
	f := NewFileLocationFixer()
	v := rules.Violation{
		Rule:    "test-location",
		Path:    "wrong/path/file.go",
		Message: "should be in src/components/",
	}
	fix, err := f.GenerateFix(v, nil)
	require.NoError(t, err)
	require.NotNil(t, fix)
	assert.Contains(t, fix.Description, "Move")
	assert.Equal(t, 0.85, fix.Confidence)
	assert.False(t, fix.Safe)
}

func TestFileLocationFixer_GenerateFix_NoTargetPath(t *testing.T) {
	f := NewFileLocationFixer()
	v := rules.Violation{Rule: "other", Path: "file.go"}
	fix, err := f.GenerateFix(v, nil)
	assert.Error(t, err)
	assert.Nil(t, fix)
}

func TestFileLocationFixer_ExtractTargetPath_Expected(t *testing.T) {
	f := &FileLocationFixer{}
	path := f.extractTargetPath(rules.Violation{Expected: "expected/path"})
	assert.Equal(t, "expected/path", path)
}

func TestFileLocationFixer_ExtractTargetPath_Suggestion(t *testing.T) {
	f := &FileLocationFixer{}
	path := f.extractTargetPath(rules.Violation{
		Suggestions: []string{"Move to correct/dir/"},
	})
	assert.Equal(t, "correct/dir/", path)
}

func TestFileLocationFixer_ExtractTargetPath_FromMessage(t *testing.T) {
	f := &FileLocationFixer{}
	path := f.extractTargetPath(rules.Violation{
		Path:    "src/wrong/test.go",
		Message: "file should be in 'src/components/'",
	})
	assert.Equal(t, "src/components/test.go", path)
}

func TestFileLocationFixer_ExtractTargetPath_Empty(t *testing.T) {
	f := &FileLocationFixer{}
	path := f.extractTargetPath(rules.Violation{})
	assert.Empty(t, path)
}

func TestNewImportRewriter(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "main.go", AbsPath: "/root/main.go"},
		{Path: "util.go", AbsPath: "/root/util.go"},
		{Path: "dir", AbsPath: "/root/dir", IsDir: true},
	}
	r := NewImportRewriter("/root", files)
	require.NotNil(t, r)
	assert.Equal(t, "/root", r.rootPath)
}

func TestImportRewriter_RegisterMove(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	r.RegisterMove("old/path", "new/path")
	assert.Equal(t, "new/path", r.movedFiles["old/path"])
}

func TestImportRewriter_GenerateUpdateActions_NoMoves(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	actions := r.GenerateUpdateActions()
	assert.Empty(t, actions)
}

func TestImportRewriter_GenerateUpdateActions_WithMoves(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "importer.go", AbsPath: "/root/importer.go"},
	}
	r := NewImportRewriter("/root", files)
	r.RegisterMove("old/main.go", "new/main.go")
	actions := r.GenerateUpdateActions()
	assert.Empty(t, actions)
}

func TestImportRewriter_MightImport(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	assert.True(t, r.mightImport("src/a/b/file.go", "src/a/b/c/dep.go", "go"))
	assert.False(t, r.mightImport("src/x/file.go", "src/y/dep.go", "go"))
}

func TestImportRewriter_PathToImport_Go(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	result := r.pathToImport("pkg/sub/foo.go", "main.go", "go")
	assert.Equal(t, "pkg/sub", result)
}

func TestImportRewriter_PathToImport_TS(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	result := r.pathToImport("src/utils/helper.ts", "src/app/main.ts", "typescript")
	assert.Contains(t, result, "./")
}

func TestImportRewriter_PathToImport_Python(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	result := r.pathToImport("src/utils/helper.py", "main.py", "python")
	assert.Equal(t, "src.utils.helper", result)
}

func TestImportRewriter_PathToImport_Default(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	result := r.pathToImport("file.rs", "main.rs", "rust")
	assert.Equal(t, "file.rs", result)
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"file.go", "go"},
		{"file.py", "python"},
		{"file.ts", "typescript"},
		{"file.tsx", "typescript"},
		{"file.js", "javascript"},
		{"file.jsx", "javascript"},
		{"file.java", "java"},
		{"file.rs", "rust"},
		{"file.c", "c"},
		{"file.h", "c"},
		{"file.cpp", "cpp"},
		{"file.hpp", "cpp"},
		{"file.unknown", "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, detectLanguage(tc.path))
		})
	}
}
