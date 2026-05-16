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

func TestMoveFileAction_Apply_MkdirFail(t *testing.T) {
	a := &MoveFileAction{
		SourcePath: "/nonexistent/src.go",
		TargetPath: "/dev/null/sub/dst.go",
	}
	err := a.Apply()
	assert.Error(t, err)
}

func TestMoveFileAction_WriteTargetFail(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.go")
	require.NoError(t, os.WriteFile(src, []byte("content"), 0644))

	targetDir := filepath.Join(dir, "sub")
	os.MkdirAll(targetDir, 0755)
	os.Chmod(targetDir, 0444)
	defer os.Chmod(targetDir, 0755)

	a := &MoveFileAction{
		SourcePath: src,
		TargetPath: filepath.Join(targetDir, "dst.go"),
	}
	err := a.Apply()
	assert.Error(t, err)
	os.Chmod(targetDir, 0755)
}

func TestWriteFileAction_BackupWriteFail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0644))

	a := &WriteFileAction{
		FilePath:   path,
		Content:    "updated",
		backupPath: "/nonexistent/backup",
	}
	err := a.Apply()
	// Should still work - backup error is not fatal
	assert.NoError(t, err)
}

func TestWriteFileAction_Revert_BackupReadFail(t *testing.T) {
	a := &WriteFileAction{
		FilePath:       "/tmp/nonexistent.go",
		originalExists: true,
		backupPath:     "/nonexistent/backup",
	}
	err := a.Revert()
	assert.Error(t, err)
}

func TestWriteFileAction_Revert_RemoveCreatedFail(t *testing.T) {
	a := &WriteFileAction{
		FilePath: "/nonexistent/newfile.go",
	}
	err := a.Revert()
	assert.Error(t, err)
}

func TestUpdateImportAction_Revert_BackupReadFail(t *testing.T) {
	a := &UpdateImportAction{
		FilePath:   "/tmp/f.go",
		backupPath: "/nonexistent/backup",
	}
	err := a.Revert()
	assert.Error(t, err)
}

func TestUpdateImportAction_Revert_RemoveBackupFail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.go")
	require.NoError(t, os.WriteFile(path, []byte("content"), 0644))

	a := &UpdateImportAction{
		FilePath:   path,
		backupPath: filepath.Join(dir, "backup"),
	}
	require.NoError(t, os.WriteFile(a.backupPath, []byte("original"), 0644))

	err := a.Revert()
	assert.NoError(t, err)
}

func TestExtractTargetPath_LowercaseSuggestion(t *testing.T) {
	f := &FileLocationFixer{}
	path := f.extractTargetPath(rules.Violation{
		Suggestions: []string{"move to correct/dir/"},
	})
	assert.Equal(t, "correct/dir/", path)
}

func TestExtractTargetPath_FromMessage(t *testing.T) {
	f := &FileLocationFixer{}
	path := f.extractTargetPath(rules.Violation{
		Path:    "src/wrong/test.go",
		Message: "file should be in 'src/components/'",
	})
	assert.Equal(t, "src/components/test.go", path)
}

func TestImportRewriter_GenerateUpdateActions_WithMightImport(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src/pkg/importer.go", AbsPath: "/root/src/pkg/importer.go"},
		{Path: "src/pkg/dep.go", AbsPath: "/root/src/pkg/dep.go"},
	}
	r := NewImportRewriter("/root", files)
	r.RegisterMove("src/pkg/dep.go", "src/lib/dep.go")
	actions := r.GenerateUpdateActions()
	require.NotEmpty(t, actions)
}

func TestPathToImport_TSRelError(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	// Invalid Rel causes fallback to filePath
	result := r.pathToImport("/different/root/helper.ts", "main.ts", "typescript")
	assert.Contains(t, result, "helper")
}

func TestApplyFixes_DryRunActions(t *testing.T) {
	e := NewEngine(true)
	fixes := []*Fix{
		{
			Violation:   rules.Violation{Rule: "test", Path: "test.go"},
			Description: "Test fix",
			Actions: []Action{
				&WriteFileAction{FilePath: filepath.Join(t.TempDir(), "f.go"), Content: "x"},
			},
		},
	}
	count, err := e.ApplyFixes(fixes)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGenerateUpdateActions_MultipleLangs(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "main.go", AbsPath: "/root/main.go"},
		{Path: "main.py", AbsPath: "/root/main.py"},
		{Path: "main.rs", AbsPath: "/root/main.rs"},
	}
	r := NewImportRewriter("/root", files)
	assert.NotNil(t, r)
}

func TestPathToImport_RelDirSame(t *testing.T) {
	r := NewImportRewriter("/root", nil)
	result := r.pathToImport("helper.ts", "helper.ts", "typescript")
	assert.Contains(t, result, "./")
}
