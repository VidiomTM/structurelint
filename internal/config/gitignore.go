package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadGitignorePatterns loads patterns from a .gitignore file
// Returns patterns that can be used as exclusion patterns
func LoadGitignorePatterns(rootDir string) ([]string, error) {
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	// If .gitignore doesn't exist, return empty list
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return []string{}, nil
	}

	file, err := os.Open(gitignorePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle negation patterns (we'll skip these for now as they're complex)
		if strings.HasPrefix(line, "!") {
			continue
		}

		// Convert gitignore pattern to glob pattern
		pattern := normalizeGitignorePattern(line)
		if pattern != "" {
			patterns = append(patterns, pattern)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// normalizeGitignorePattern converts a .gitignore pattern to a glob pattern
func normalizeGitignorePattern(pattern string) string {
	// Remove trailing whitespace
	pattern = strings.TrimSpace(pattern)

	if pattern == "" {
		return ""
	}

	// If pattern starts with /, it's relative to root
	if strings.HasPrefix(pattern, "/") {
		pattern = strings.TrimPrefix(pattern, "/")
		// Ensure it matches only at root
		if !strings.Contains(pattern, "*") && !strings.HasSuffix(pattern, "/") {
			// It's a specific file/directory at root
			return pattern
		}
	}

	// If pattern ends with /, it's a directory
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		// Match directory and all contents
		return pattern + "/**"
	}

	// If pattern doesn't contain /, it matches anywhere
	if !strings.Contains(pattern, "/") {
		return "**/" + pattern
	}

	// Otherwise, use as-is
	return pattern
}

// MergeWithGitignore merges .gitignore patterns with existing exclusions
// It avoids duplicates and returns the combined list
func MergeWithGitignore(existing []string, gitignorePatterns []string) []string {
	// Create a set to track existing patterns
	seen := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(gitignorePatterns))

	// Add existing patterns
	for _, pattern := range existing {
		if !seen[pattern] {
			seen[pattern] = true
			result = append(result, pattern)
		}
	}

	// Add gitignore patterns (avoiding duplicates)
	for _, pattern := range gitignorePatterns {
		if !seen[pattern] {
			seen[pattern] = true
			result = append(result, pattern)
		}
	}

	return result
}
