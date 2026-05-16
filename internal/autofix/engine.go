// Package autofix provides automatic fixing capabilities for structurelint violations
package autofix

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// Fix represents a proposed fix for a violation
type Fix struct {
	// Violation that triggered this fix
	Violation rules.Violation

	// Description of what the fix does
	Description string

	// Actions to perform
	Actions []Action

	// Confidence level (0.0-1.0)
	Confidence float64

	// Safe indicates if this fix is safe to apply automatically
	Safe bool
}

// Action represents a single fixable action
type Action interface {
	// Apply executes the action
	Apply() error

	// Describe returns a human-readable description
	Describe() string

	// Revert undoes the action (best-effort)
	Revert() error
}

// MoveFileAction moves a file to a new location
type MoveFileAction struct {
	SourcePath string
	TargetPath string
	UpdateImports bool
	originalContent []byte
}

// Apply moves the file
func (a *MoveFileAction) Apply() error {
	// Read original content for potential revert
	content, err := os.ReadFile(a.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	a.originalContent = content

	// Create target directory if needed
	targetDir := filepath.Dir(a.TargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy file to new location
	if err := os.WriteFile(a.TargetPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write target file: %w", err)
	}

	// Remove original file
	if err := os.Remove(a.SourcePath); err != nil {
		if removeErr := os.Remove(a.TargetPath); removeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clean up target file %s: %v\n", a.TargetPath, removeErr)
		}
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

// Describe returns a description of the action
func (a *MoveFileAction) Describe() string {
	return fmt.Sprintf("Move %s → %s", a.SourcePath, a.TargetPath)
}

// Revert undoes the file move
func (a *MoveFileAction) Revert() error {
	if a.originalContent == nil {
		return fmt.Errorf("no original content to revert to")
	}

	// Restore original file
	if err := os.WriteFile(a.SourcePath, a.originalContent, 0644); err != nil {
		return fmt.Errorf("failed to restore original file: %w", err)
	}

	// Remove moved file
	if err := os.Remove(a.TargetPath); err != nil {
		return fmt.Errorf("failed to remove moved file: %w", err)
	}

	return nil
}

// WriteFileAction writes content to a file (for auto-generated content)
type WriteFileAction struct {
	FilePath       string
	Content        string
	backupPath     string
	originalExists bool
}

// Apply writes content to the file
func (a *WriteFileAction) Apply() error {
	// Check if file already exists
	if _, err := os.Stat(a.FilePath); err == nil {
		a.originalExists = true
		// Create backup
		a.backupPath = a.FilePath + ".backup"
		content, err := os.ReadFile(a.FilePath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
		if err := os.WriteFile(a.backupPath, content, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Create directory if needed
	dir := filepath.Dir(a.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write new content
	if err := os.WriteFile(a.FilePath, []byte(a.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Describe returns a description of the action
func (a *WriteFileAction) Describe() string {
	if a.originalExists {
		return fmt.Sprintf("Update %s", a.FilePath)
	}
	return fmt.Sprintf("Create %s", a.FilePath)
}

// Revert undoes the file write
func (a *WriteFileAction) Revert() error {
	if a.originalExists && a.backupPath != "" {
		// Restore from backup
		content, err := os.ReadFile(a.backupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup: %w", err)
		}
		if err := os.WriteFile(a.FilePath, content, 0644); err != nil {
			return fmt.Errorf("failed to restore file: %w", err)
		}
		if err := os.Remove(a.backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove backup %s: %v\n", a.backupPath, err)
		}
	} else {
		// Remove created file
		if err := os.Remove(a.FilePath); err != nil {
			return fmt.Errorf("failed to remove created file: %w", err)
		}
	}
	return nil
}

// UpdateImportAction updates import statements in a file
type UpdateImportAction struct {
	FilePath   string
	OldImport  string
	NewImport  string
	backupPath string
}

// Apply updates the imports
func (a *UpdateImportAction) Apply() error {
	// Read file content
	content, err := os.ReadFile(a.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create backup
	a.backupPath = a.FilePath + ".backup"
	if err := os.WriteFile(a.backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Update imports (simplified - real implementation would use AST)
		// This is a placeholder for demonstration

	return nil
}

// Describe returns a description of the action
func (a *UpdateImportAction) Describe() string {
	return fmt.Sprintf("Update import in %s: %s → %s", a.FilePath, a.OldImport, a.NewImport)
}

// Revert undoes the import update
func (a *UpdateImportAction) Revert() error {
	if a.backupPath == "" {
		return fmt.Errorf("no backup to revert to")
	}

	// Read backup
	content, err := os.ReadFile(a.backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Restore original content
	if err := os.WriteFile(a.FilePath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	// Remove backup
	if err := os.Remove(a.backupPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove backup %s: %v\n", a.backupPath, err)
	}

	return nil
}

// Fixer generates fixes for violations
type Fixer interface {
	// CanFix determines if this fixer can handle the violation
	CanFix(v rules.Violation) bool

	// GenerateFix creates a fix for the violation
	GenerateFix(v rules.Violation, files []walker.FileInfo) (*Fix, error)
}

// Engine applies fixes
type Engine struct {
	fixers []Fixer
	dryRun bool
}

// NewEngine creates a new fix engine
func NewEngine(dryRun bool) *Engine {
	return &Engine{
		fixers: []Fixer{
			// Register built-in fixers
			NewFileLocationFixer(),
		},
		dryRun: dryRun,
	}
}

// RegisterFixer adds a fixer to the engine
func (e *Engine) RegisterFixer(f Fixer) {
	e.fixers = append(e.fixers, f)
}

// GenerateFixes creates fixes for all violations
func (e *Engine) GenerateFixes(violations []rules.Violation, files []walker.FileInfo) ([]*Fix, error) {
	var fixes []*Fix

	for _, v := range violations {
		fix := e.generateFixForViolation(v, files)
		if fix != nil {
			fixes = append(fixes, fix)
			continue

		}
	}

	return fixes, nil
}

func (e *Engine) generateFixForViolation(v rules.Violation, files []walker.FileInfo) *Fix {
	// Check if violation has built-in AutoFix
	if v.AutoFix != nil {
		return &Fix{
			Violation:   v,
			Description: fmt.Sprintf("Apply auto-fix for %s", v.Rule),
			Actions: []Action{
				&WriteFileAction{
					FilePath: v.AutoFix.FilePath,
					Content:  v.AutoFix.Content,
				},
			},
			Confidence: 0.95, // High confidence for rule-provided fixes
			Safe:       true,  // Assume rule-provided fixes are safe
		}
	}
	return nil
}

// ApplyFixes executes the fixes
func (e *Engine) ApplyFixes(fixes []*Fix) (int, error) {
	applied := 0

	for _, fix := range fixes {
		if e.dryRun {
			e.applyDryRun(fix)
			applied++
			continue
		}
		if err := e.applyFix(fix); err != nil {
			return applied, err
		}
		applied++
	}

	return applied, nil
}

func (e *Engine) applyDryRun(fix *Fix) {
	fmt.Printf("[DRY RUN] Would apply fix: %s\n", fix.Description)
	for _, action := range fix.Actions {
		fmt.Printf("  - %s\n", action.Describe())
	}
}

func (e *Engine) applyFix(fix *Fix) error {
	var appliedActions []Action
	for _, action := range fix.Actions {
		if err := action.Apply(); err != nil {
			e.revertActions(appliedActions)
			return fmt.Errorf("failed to apply fix for %s: %w", fix.Violation.Path, err)
		}
		appliedActions = append(appliedActions, action)
	}
	return nil
}

func (e *Engine) revertActions(actions []Action) {
	for i := len(actions) - 1; i >= 0; i-- {
		if revertErr := actions[i].Revert(); revertErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to revert action: %v\n", revertErr)
		}
	}
}

// FixResult represents the result of applying fixes
type FixResult struct {
	Applied   int
	Failed    int
	Skipped   int
	DryRun    bool
	Fixes     []*Fix
	Errors    []error
}

// String returns a human-readable summary
func (r *FixResult) String() string {
	if r.DryRun {
		return fmt.Sprintf("Dry run: %d fixes would be applied, %d skipped", r.Applied, r.Skipped)
	}
	return fmt.Sprintf("Applied %d fixes, %d failed, %d skipped", r.Applied, r.Failed, r.Skipped)
}
