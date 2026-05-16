// Package structure provides rule implementations for structurelint.
package structure

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// FileExistenceRule validates file existence requirements
type FileExistenceRule struct {
	Requirements map[string]string // pattern -> requirement (e.g., "index.ts" -> "exists:1")
}

// Name returns the rule name
func (r *FileExistenceRule) Name() string {
	return "file-existence"
}

// Check validates file existence requirements
func (r *FileExistenceRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	// Group files by directory
	filesByDir := make(map[string][]walker.FileInfo)
	for _, file := range files {
		if !file.IsDir {
			dir := file.ParentPath
			filesByDir[dir] = append(filesByDir[dir], file)
		}
	}

	// Check each directory against requirements
	for dir := range dirs {
		for pattern, requirement := range r.Requirements {
			if err := r.checkRequirement(dir, pattern, requirement, filesByDir[dir]); err != nil {
				displayPath := dir
				if displayPath == "" {
					displayPath = "."
				}
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    displayPath,
					Message: err.Error(),
				})
			}
		}
	}

	return violations
}

// checkRequirement checks a single file existence requirement for a directory
func (r *FileExistenceRule) checkRequirement(dir, pattern, requirement string, dirFiles []walker.FileInfo) error {
	minCount, maxCount, err := r.parseRequirement(requirement)
	if err != nil {
		return err
	}
	matchCount := r.countMatches(dir, pattern, dirFiles)
	if matchCount < minCount {
		return fmt.Errorf("requires at least %d file(s) matching '%s', found %d", minCount, pattern, matchCount)
	}
	if maxCount >= 0 && matchCount > maxCount {
		return fmt.Errorf("requires at most %d file(s) matching '%s', found %d", maxCount, pattern, matchCount)
	}
	return nil
}

func (r *FileExistenceRule) parseRequirement(requirement string) (int, int, error) {
	parts := strings.Split(requirement, ":")
	if len(parts) != 2 || parts[0] != "exists" {
		return 0, 0, fmt.Errorf("invalid requirement format: %s", requirement)
	}
	minCount, maxCount, err := r.parseCountSpec(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid count spec %q: %w", parts[1], err)
	}
	return minCount, maxCount, nil
}

func (r *FileExistenceRule) countMatches(dir, pattern string, dirFiles []walker.FileInfo) int {
	if pattern == ".dir" {
		return r.countSubdirs(dir, dirFiles)
	}
	return r.countFilePatterns(dir, pattern, dirFiles)
}

func (r *FileExistenceRule) countSubdirs(dir string, dirFiles []walker.FileInfo) int {
	count := 0
	for _, file := range dirFiles {
		if file.IsDir && file.ParentPath == dir {
			count++
		}
	}
	return count
}

func (r *FileExistenceRule) countFilePatterns(dir, pattern string, dirFiles []walker.FileInfo) int {
	patterns := strings.Split(pattern, "|")
	matched := make(map[string]bool)
	count := 0
	for _, file := range dirFiles {
		if file.ParentPath != dir {
			continue
		}
		for _, p := range patterns {
			if r.fileMatchesPattern(file, p) && !matched[file.Path] {
				count++
				matched[file.Path] = true
				break
			}
		}
	}
	return count
}

// parseCountSpec parses count specifications like "1", "0", "1-10"
func (r *FileExistenceRule) parseCountSpec(spec string) (min, max int, err error) {
	if strings.Contains(spec, "-") {
		parts := strings.Split(spec, "-")
		min, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid minimum count %q: %w", parts[0], err)
		}
		max, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid maximum count %q: %w", parts[1], err)
		}
		return min, max, nil
	}

	count, err := strconv.Atoi(spec)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid count %q: %w", spec, err)
	}
	return count, count, nil
}

// fileMatchesPattern checks if a file matches a pattern
func (r *FileExistenceRule) fileMatchesPattern(file walker.FileInfo, pattern string) bool {
	filename := filepath.Base(file.Path)

	// Exact match
	if filename == pattern {
		return true
	}

	// Glob match
	matched, err := filepath.Match(pattern, filename)
	if err == nil && matched {
		return true
	}

	return false
}

// NewFileExistenceRule creates a new FileExistenceRule
func NewFileExistenceRule(requirements map[string]string) *FileExistenceRule {
	return &FileExistenceRule{
		Requirements: requirements,
	}
}

// ValidateFileExistenceConfig surfaces malformed `exists:N` / `exists:N-M`
// specifications at config-load time rather than as per-directory violations
// at lint time. Returns the list of bad keys with reasons.
func ValidateFileExistenceConfig(requirements map[string]string) []string {
	var errs []string
	for pattern, requirement := range requirements {
		if err := validateRequirement(pattern, requirement); err != "" {
			errs = append(errs, err)
		}
	}
	return errs
}

func validateRequirement(pattern, requirement string) string {
	parts := strings.Split(requirement, ":")
	if len(parts) != 2 || parts[0] != "exists" {
		return fmt.Sprintf("%q: requirement must be 'exists:N' or 'exists:N-M' (got %q)", pattern, requirement)
	}
	spec := parts[1]
	if strings.Contains(spec, "-") {
		return validateRangeSpec(pattern, spec)
	}
	if _, err := strconv.Atoi(spec); err != nil {
		return fmt.Sprintf("%q: count %q is not an integer", pattern, spec)
	}
	return ""
}

func validateRangeSpec(pattern, spec string) string {
	rangeParts := strings.Split(spec, "-")
	if len(rangeParts) != 2 {
		return fmt.Sprintf("%q: range must be 'min-max' (got %q)", pattern, spec)
	}
	if _, err := strconv.Atoi(rangeParts[0]); err != nil {
		return fmt.Sprintf("%q: range minimum %q is not an integer", pattern, rangeParts[0])
	}
	if _, err := strconv.Atoi(rangeParts[1]); err != nil {
		return fmt.Sprintf("%q: range maximum %q is not an integer", pattern, rangeParts[1])
	}
	return ""
}
