package walker

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Jonathangadeaharder/structurelint/internal/parser"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Path       string             // Relative path from root
	AbsPath    string             // Absolute path
	IsDir      bool               // Whether this is a directory
	IsSymlink  bool               // Whether this entry is a symlink
	Depth      int                // Nesting depth from root
	ParentPath string             // Path of parent directory
	Directives []parser.Directive // Parsed structurelint directives from the file
}

// DirInfo represents aggregated information about a directory
type DirInfo struct {
	Path        string
	FileCount   int
	SubdirCount int
	Depth       int
}

// Walker walks a filesystem and collects information
type Walker struct {
	rootPath        string
	files           []FileInfo
	dirs            map[string]*DirInfo
	excludePatterns []string
}

// New creates a new Walker
func New(rootPath string) *Walker {
	return &Walker{
		rootPath: rootPath,
		files:    []FileInfo{},
		dirs:     make(map[string]*DirInfo),
	}
}

// WithExclude sets exclude patterns for the walker
func (w *Walker) WithExclude(patterns []string) *Walker {
	w.excludePatterns = patterns
	return w
}

// isExcluded checks if a path matches any exclude pattern
func (w *Walker) isExcluded(relPath string) bool {
	for _, pattern := range w.excludePatterns {
		if MatchesPattern(relPath, pattern) {
			return true
		}
	}
	return false
}

// Walk traverses the filesystem starting from the root path
func (w *Walker) Walk() error {
	absRoot, err := filepath.Abs(w.rootPath)
	if err != nil {
		return err
	}

	return filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		return w.processPath(relPath, path, d)
	})
}

// processPath handles a single path entry during the walk
func (w *Walker) processPath(relPath, absPath string, d fs.DirEntry) error {
	// Skip the root itself
	if relPath == "." {
		return nil
	}

	// Check for exclusions and skippable directories
	if skipAction := w.shouldSkip(relPath, d); skipAction != nil {
		return skipAction
	}

	depth := w.calculateDepth(relPath, d.IsDir())
	parentPath := w.normalizeParentPath(relPath)

	// Parse directives for regular files (not directories)
	var directives []parser.Directive
	if !d.IsDir() {
		directives = parser.ParseDirectives(absPath)
	}

	info := FileInfo{
		Path:       relPath,
		AbsPath:    absPath,
		IsDir:      d.IsDir(),
		IsSymlink:  d.Type()&fs.ModeSymlink != 0,
		Depth:      depth,
		ParentPath: parentPath,
		Directives: directives,
	}

	w.files = append(w.files, info)
	w.updateDirectoryStats(relPath, parentPath, depth, d.IsDir())

	return nil
}

// shouldSkip determines if a path should be skipped and returns the appropriate action
func (w *Walker) shouldSkip(relPath string, d fs.DirEntry) error {
	if w.isExcluded(relPath) {
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	if d.IsDir() && w.isIgnoredDir(relPath) {
		return filepath.SkipDir
	}

	return nil
}

// isIgnoredDir checks if a directory should be ignored
func (w *Walker) isIgnoredDir(relPath string) bool {
	baseName := filepath.Base(relPath)
	return baseName == ".git" || baseName == "node_modules" || baseName == "vendor"
}

// calculateDepth calculates the depth of a path
func (w *Walker) calculateDepth(relPath string, isDir bool) int {
	depth := strings.Count(relPath, string(filepath.Separator))
	if isDir {
		depth++ // Directories count themselves
	}
	return depth
}

// normalizeParentPath returns the normalized parent path
func (w *Walker) normalizeParentPath(relPath string) string {
	parentPath := filepath.Dir(relPath)
	if parentPath == "." {
		return ""
	}
	return parentPath
}

// updateDirectoryStats updates statistics for both the current directory and its parent
func (w *Walker) updateDirectoryStats(relPath, parentPath string, depth int, isDir bool) {
	if isDir {
		w.ensureDirExists(relPath, depth)
	}

	if parentPath != "" {
		w.ensureDirExists(parentPath, depth-1)
		w.updateParentCounts(parentPath, isDir)
	}
}

// ensureDirExists ensures a directory entry exists in the stats map
func (w *Walker) ensureDirExists(dirPath string, depth int) {
	if _, exists := w.dirs[dirPath]; !exists {
		w.dirs[dirPath] = &DirInfo{
			Path:  dirPath,
			Depth: depth,
		}
	}
}

// updateParentCounts updates file or subdir count for a parent directory
func (w *Walker) updateParentCounts(parentPath string, isDir bool) {
	if isDir {
		w.dirs[parentPath].SubdirCount++
	} else {
		w.dirs[parentPath].FileCount++
	}
}

// GetFiles returns all files found during the walk
func (w *Walker) GetFiles() []FileInfo {
	return w.files
}

// GetDirs returns directory statistics
func (w *Walker) GetDirs() map[string]*DirInfo {
	return w.dirs
}

// GetMaxDepth returns the maximum depth found in the filesystem
func (w *Walker) GetMaxDepth() int {
	maxDepth := 0
	for _, info := range w.files {
		if info.Depth > maxDepth {
			maxDepth = info.Depth
		}
	}
	return maxDepth
}

// MatchesPattern checks if a path matches a glob pattern.
// Supports `*` (no slash crossing), `?`, `[...]`, and `**` (any path).
// Trailing `/` on the pattern is treated as a directory prefix match.
func MatchesPattern(path, pattern string) bool {
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	if strings.HasSuffix(pattern, "/") {
		trimmed := strings.TrimSuffix(pattern, "/")
		if path == trimmed || strings.HasPrefix(path, trimmed+"/") {
			return true
		}
		if re := patternRegexp(trimmed); re != nil && re.MatchString(path) {
			return true
		}
	}

	if path == pattern {
		return true
	}

	re := patternRegexp(pattern)
	if re != nil && re.MatchString(path) {
		return true
	}

	if !strings.Contains(pattern, "**") && !strings.Contains(pattern, "/") {
		base := filepath.Base(path)
		matched, err := filepath.Match(pattern, base)
		if err == nil && matched {
			return true
		}
	}
	return false
}

var walkerGlobCache sync.Map

func patternRegexp(pattern string) *regexp.Regexp {
	if v, ok := walkerGlobCache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}
	var b strings.Builder
	b.WriteString("^")
	walkerPatternToRegex(pattern, &b)
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		return nil
	}
	walkerGlobCache.Store(pattern, re)
	return re
}

func walkerPatternToRegex(pattern string, b *strings.Builder) {
	i := 0
	for i < len(pattern) {
		switch c := pattern[i]; c {
		case '*':
			i = walkerAppendStar(pattern, i, b)
		case '?':
			b.WriteString("[^/]")
			i++
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
			i++
		case '[':
			i = walkerAppendCharClass(pattern, i, b)
		default:
			b.WriteByte(c)
			i++
		}
	}
}

func walkerAppendStar(pattern string, i int, b *strings.Builder) int {
	if i+1 < len(pattern) && pattern[i+1] == '*' {
		if i+2 < len(pattern) && pattern[i+2] == '/' {
			b.WriteString("(?:.*/)?")
			return i + 3
		}
		if i > 0 && pattern[i-1] == '/' {
			s := b.String()
			b.Reset()
			b.WriteString(strings.TrimSuffix(s, "/"))
			b.WriteString("(?:/.*)?")
			return i + 2
		}
		b.WriteString(".*")
		return i + 2
	}
	b.WriteString("[^/]*")
	return i + 1
}

func walkerAppendCharClass(pattern string, i int, b *strings.Builder) int {
	j := i + 1
	for j < len(pattern) && pattern[j] != ']' {
		j++
	}
	if j < len(pattern) {
		b.WriteString(pattern[i : j+1])
		return j + 1
	}
	b.WriteString("\\[")
	return i + 1
}
