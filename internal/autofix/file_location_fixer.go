package autofix

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// FileLocationFixer generates fixes for file location violations
type FileLocationFixer struct{}

// NewFileLocationFixer creates a new file location fixer
func NewFileLocationFixer() *FileLocationFixer {
	return &FileLocationFixer{}
}

// CanFix determines if this fixer can handle the violation
func (f *FileLocationFixer) CanFix(v rules.Violation) bool {
	// Handle violations from location-related rules
	return strings.Contains(v.Rule, "location") ||
		strings.Contains(v.Rule, "test-location") ||
		(v.Expected != "" && strings.Contains(v.Message, "should be in"))
}

// GenerateFix creates a fix for the violation
func (f *FileLocationFixer) GenerateFix(v rules.Violation, files []walker.FileInfo) (*Fix, error) {
	// Parse expected location from violation
	targetPath := f.extractTargetPath(v)
	if targetPath == "" {
		return nil, fmt.Errorf("could not determine target path from violation")
	}

	// Get current path
	sourcePath := v.Path

	// Create fix with file move and import update actions
	fix := &Fix{
		Violation:   v,
		Description: fmt.Sprintf("Move %s to %s", sourcePath, targetPath),
		Actions: []Action{
			&MoveFileAction{
				SourcePath:    sourcePath,
				TargetPath:    targetPath,
				UpdateImports: true,
			},
		},
		Confidence: 0.85, // High confidence but not 100% due to import complexity
		Safe:       false, // File moves require caution
	}

	return fix, nil
}

// extractTargetPath extracts the target path from a violation
func (f *FileLocationFixer) extractTargetPath(v rules.Violation) string {
	// If Expected field is set, use it
	if v.Expected != "" {
		return v.Expected
	}

	// Try to parse from suggestions
	for _, suggestion := range v.Suggestions {
		if strings.HasPrefix(suggestion, "Move to ") {
			return strings.TrimPrefix(suggestion, "Move to ")
		}
		if strings.HasPrefix(suggestion, "move to ") {
			return strings.TrimPrefix(suggestion, "move to ")
		}
	}

	// Try to parse from message
	if strings.Contains(v.Message, "should be in ") {
		parts := strings.Split(v.Message, "should be in ")
		if len(parts) > 1 {
			// Extract path from message (e.g., "'src/components/'")
			path := strings.Trim(parts[1], "' ")
			if idx := strings.Index(path, " "); idx != -1 {
				path = path[:idx]
			}
			// Construct full target path
			base := filepath.Base(v.Path)
			return filepath.Join(strings.Trim(path, "/"), base)
		}
	}

	return ""
}

// ImportRewriter handles updating import statements when files move
type ImportRewriter struct {
	rootPath    string
	movedFiles  map[string]string // old path -> new path
	filesByLang map[string][]walker.FileInfo
}

// NewImportRewriter creates a new import rewriter
func NewImportRewriter(rootPath string, files []walker.FileInfo) *ImportRewriter {
	filesByLang := make(map[string][]walker.FileInfo)

	for _, file := range files {
		if file.IsDir {
			continue
		}

		lang := detectLanguage(file.Path)
		filesByLang[lang] = append(filesByLang[lang], file)
	}

	return &ImportRewriter{
		rootPath:    rootPath,
		movedFiles:  make(map[string]string),
		filesByLang: filesByLang,
	}
}

// RegisterMove registers a file move for import rewriting
func (r *ImportRewriter) RegisterMove(oldPath, newPath string) {
	r.movedFiles[oldPath] = newPath
}

// GenerateUpdateActions generates import update actions for all affected files
func (r *ImportRewriter) GenerateUpdateActions() []Action {
	var actions []Action

	// For each moved file, find all files that import it
	for oldPath, newPath := range r.movedFiles {
		lang := detectLanguage(oldPath)

		// Get files in the same language
		files := r.filesByLang[lang]

		for _, file := range files {
			// Skip the moved file itself
			if file.Path == oldPath {
				continue
			}

			// Check if this file might import the moved file
			// This is a simplified version - full implementation would use AST
			if r.mightImport(file.Path, oldPath, lang) {
				oldImport := r.pathToImport(oldPath, file.Path, lang)
				newImport := r.pathToImport(newPath, file.Path, lang)

				actions = append(actions, &UpdateImportAction{
					FilePath:  file.Path,
					OldImport: oldImport,
					NewImport: newImport,
				})
			}
		}
	}

	return actions
}

// mightImport checks if a file might import another (simplified heuristic)
func (r *ImportRewriter) mightImport(importerPath, importedPath, lang string) bool {
	// Simplified: check if files are in related directories
	// Full implementation would parse imports from the file
	importerDir := filepath.Dir(importerPath)
	importedDir := filepath.Dir(importedPath)

	// Files in the same directory tree are more likely to import each other
	return strings.HasPrefix(importerDir, importedDir) ||
		   strings.HasPrefix(importedDir, importerDir)
}

// pathToImport converts a file path to an import path based on language
func (r *ImportRewriter) pathToImport(filePath, fromPath, lang string) string {
	switch lang {
	case "go":
		// Go uses module-relative imports
		// Remove file extension and convert to package path
		pkg := strings.TrimSuffix(filePath, filepath.Ext(filePath))
		return filepath.Dir(pkg)

	case "typescript", "javascript":
		// TypeScript/JS uses relative imports
		rel, err := filepath.Rel(filepath.Dir(fromPath), filePath)
		if err != nil {
			return filePath
		}
		// Convert to POSIX path and remove extension
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, filepath.Ext(rel))

		// Add ./ prefix for relative imports
		if !strings.HasPrefix(rel, ".") {
			rel = "./" + rel
		}
		return rel

	case "python":
		// Python uses dot-separated module paths
		pkg := strings.TrimSuffix(filePath, ".py")
		return strings.ReplaceAll(pkg, "/", ".")

	default:
		return filePath
	}
}

// detectLanguage detects the programming language from file extension
func detectLanguage(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".hh":
		return "cpp"
	default:
		return "unknown"
	}
}
