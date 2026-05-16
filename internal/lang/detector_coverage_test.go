package lang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectInDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644))

	languages, err := DetectInDirectory(tmpDir)
	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, Go, languages[0].Language)
}

func TestDetectInDirectory_Error(t *testing.T) {
	_, err := DetectInDirectory("/nonexistent")
	assert.Error(t, err)
}

func TestDetector_DetectRubyProject(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte("source 'https://rubygems.org'\n"), 0644))

	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, Ruby, languages[0].Language)
}

func TestDetector_DetectJavaScriptProject_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := filepath.Join(tmpDir, "package.json")
	// Create empty dir instead of file — NewDetector uses filepath.Walk which uses os.Stat
	// Actually, we need an unreadable file or a broken read
	require.NoError(t, os.WriteFile(packageJSON, []byte(""), 0000))
	defer os.Chmod(packageJSON, 0644)

	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()

	require.NoError(t, err)
	// Should find it as JavaScript (default when can't read)
	require.Len(t, languages, 1)
}

func TestDetector_DetectJavaScriptProject_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("not json"), 0644))

	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, JavaScript, languages[0].Language)
}

func TestDetector_DetectJavaScriptProject_DevDepsTypeScript(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := `{"devDependencies": {"typescript": "^5.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644))

	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, TypeScript, languages[0].Language)
}

func TestDetector_DetectJavaScriptProject_ReactWithoutTypeScript(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := `{"dependencies": {"react": "^18.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644))
	// Also create tsconfig.json in the same dir to trigger TypeScript detection
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte("{}"), 0644))

	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, TypeScript, languages[0].Language)
	require.Len(t, languages[0].SubLanguages, 1)
	assert.Equal(t, React, languages[0].SubLanguages[0])
}

func TestDetector_DetectJavaScriptProject_ReactNoTypeScript(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSON := `{"dependencies": {"react": "^18.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644))

	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, JavaScript, languages[0].Language)
	require.Len(t, languages[0].SubLanguages, 1)
	assert.Equal(t, React, languages[0].SubLanguages[0])
}

func TestLanguage_String_Unknown(t *testing.T) {
	assert.Equal(t, "Unknown", Language(999).String())
}

func TestLanguage_DefaultNamingConvention_Unknown(t *testing.T) {
	assert.Equal(t, "snake_case", Language(999).DefaultNamingConvention())
}

func TestDetector_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	detector := NewDetector(tmpDir)
	languages, err := detector.Detect()
	require.NoError(t, err)
	assert.Empty(t, languages)
}
