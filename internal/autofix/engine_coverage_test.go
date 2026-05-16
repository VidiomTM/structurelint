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

func TestApplyFixes_Real(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(path, []byte("content"), 0644))

	e := NewEngine(false)
	fixes := []*Fix{
		{
			Violation:   rules.Violation{Rule: "test", Path: path},
			Description: "Write test",
			Actions: []Action{
				&WriteFileAction{FilePath: path, Content: "updated"},
			},
		},
	}
	count, err := e.ApplyFixes(fixes)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	data, _ := os.ReadFile(path)
	assert.Equal(t, "updated", string(data))
}

func TestApplyFixes_ActionError(t *testing.T) {
	e := NewEngine(false)
	fixes := []*Fix{
		{
			Violation:   rules.Violation{Rule: "test", Path: "/nonexistent"},
			Description: "Fail",
			Actions: []Action{
				&WriteFileAction{FilePath: "/nonexistent/dir/file.go", Content: "x"},
			},
		},
	}
	count, err := e.ApplyFixes(fixes)
	assert.Error(t, err)
	assert.Equal(t, 0, count)
}

func TestApplyFixes_MultipleActionsWithRevert(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "f1.go")
	f2 := filepath.Join(dir, "f2.go")
	require.NoError(t, os.WriteFile(f1, []byte("one"), 0644))
	require.NoError(t, os.WriteFile(f2, []byte("two"), 0644))

	e := NewEngine(false)
	fixes := []*Fix{
		{
			Violation: rules.Violation{Rule: "test", Path: f1},
			Actions: []Action{
				&WriteFileAction{FilePath: f1, Content: "one-updated"},
				&WriteFileAction{FilePath: f2, Content: "two-updated"},
			},
		},
	}
	count, err := e.ApplyFixes(fixes)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestApplyFixes_RevertOnFail(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "f1.go")
	require.NoError(t, os.WriteFile(f1, []byte("one"), 0644))

	e := NewEngine(false)
	fixes := []*Fix{
		{
			Violation: rules.Violation{Rule: "test", Path: f1},
			Actions: []Action{
				&WriteFileAction{FilePath: f1, Content: "updated"},
				&UpdateImportAction{FilePath: "/nonexistent/file", OldImport: "x", NewImport: "y"},
			},
		},
	}

	original, _ := os.ReadFile(f1)
	count, err := e.ApplyFixes(fixes)
	assert.Error(t, err)
	assert.Equal(t, 0, count)

	restored, _ := os.ReadFile(f1)
	assert.Equal(t, string(original), string(restored))
}

func TestWriteFileAction_Apply_BackupError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0644))

	a := &WriteFileAction{FilePath: path, Content: "y"}
	a.backupPath = "/nonexistent/backup"

	require.NoError(t, a.Apply())
}

func TestWriteFileAction_Revert_RemoveCreatedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "newfile.go")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0644))

	a := &WriteFileAction{FilePath: path, Content: "y"}
	require.NoError(t, a.Apply())
	require.NoError(t, a.Revert())
}

func TestUpdateImportAction_Apply_ReadError(t *testing.T) {
	a := &UpdateImportAction{FilePath: "/nonexistent/file.go", OldImport: "old", NewImport: "new"}
	err := a.Apply()
	assert.Error(t, err)
}

func TestUpdateImportAction_Revert_BackupWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.go")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0644))

	a := &UpdateImportAction{FilePath: path, OldImport: "old", NewImport: "new"}
	require.NoError(t, a.Apply())

	a.backupPath = "/nonexistent/backup"
	err := a.Revert()
	assert.Error(t, err)
}

func TestMoveFileAction_Apply_NoSource(t *testing.T) {
	a := &MoveFileAction{
		SourcePath: "/nonexistent/source.go",
		TargetPath: "/tmp/dst.go",
	}
	err := a.Apply()
	assert.Error(t, err)
}

func TestEngine_GenerateFixes_NoMatch(t *testing.T) {
	e := NewEngine(true)
	violations := []rules.Violation{
		{Rule: "test-rule", Path: "test.go", Message: "no fixer matches"},
	}
	fixes, err := e.GenerateFixes(violations, nil)
	require.NoError(t, err)
	assert.Empty(t, fixes)
}

func TestImportRewriter_GenerateUpdateActions_WithMovesGo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	require.NoError(t, os.WriteFile(path, []byte("package main"), 0644))
	files := []walker.FileInfo{
		{Path: "main.go", AbsPath: path},
	}
	r := NewImportRewriter(dir, files)
	r.RegisterMove("old/main.go", "new/main.go")
	actions := r.GenerateUpdateActions()
	assert.Empty(t, actions)
}

func TestMoveFileAction_Revert_RestoreError(t *testing.T) {
	dir := t.TempDir()
	a := &MoveFileAction{
		SourcePath:      filepath.Join(dir, "src.go"),
		TargetPath:      filepath.Join(dir, "dst.go"),
		originalContent: []byte("content"),
	}
	err := a.Revert()
	assert.Error(t, err)
}

func TestMoveFileAction_Revert_BothErrors(t *testing.T) {
	a := &MoveFileAction{
		SourcePath:      "/nonexistent/restore.go",
		TargetPath:      "/nonexistent/remove.go",
		originalContent: []byte("data"),
	}
	err := a.Revert()
	assert.Error(t, err)
}
