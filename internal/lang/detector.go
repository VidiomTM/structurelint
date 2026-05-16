package lang

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Language represents a programming language detected in the project
type Language int

const (
	Unknown Language = iota
	Go
	Python
	TypeScript
	JavaScript
	React
	Rust
	Java
	CSharp
	Ruby
)

// String returns the string representation of a language
func (l Language) String() string {
	switch l {
	case Go:
		return "Go"
	case Python:
		return "Python"
	case TypeScript:
		return "TypeScript"
	case JavaScript:
		return "JavaScript"
	case React:
		return "React"
	case Rust:
		return "Rust"
	case Java:
		return "Java"
	case CSharp:
		return "C#"
	case Ruby:
		return "Ruby"
	default:
		return "Unknown"
	}
}

// LanguageInfo contains information about a detected language
type LanguageInfo struct {
	Language     Language
	RootDir      string   // Directory where manifest was found
	ManifestFile string   // The manifest file that identified this language
	SubLanguages []Language // e.g., React is a sub-language of JavaScript/TypeScript
}

// DefaultNamingConvention returns the default naming convention for a language
func (l Language) DefaultNamingConvention() string {
	switch l {
	case Go:
		return "snake_case"
	case Python:
		return "snake_case"
	case TypeScript, JavaScript:
		return "camelCase"
	case React:
		return "PascalCase" // For components
	case Rust:
		return "snake_case"
	case Java:
		return "PascalCase" // For classes
	case CSharp:
		return "PascalCase"
	case Ruby:
		return "snake_case"
	default:
		return "snake_case"
	}
}

// Detector detects programming languages in a project
type Detector struct {
	rootDir string
}

// NewDetector creates a new language detector
func NewDetector(rootDir string) *Detector {
	return &Detector{rootDir: rootDir}
}

// Detect scans the directory tree and detects all languages
func (d *Detector) Detect() ([]*LanguageInfo, error) {
	var languages []*LanguageInfo
	visited := make(map[string]bool)

	err := filepath.Walk(d.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip common directories to ignore
		if info.IsDir() {
			base := filepath.Base(path)
			if shouldSkipDir(base) {
				return filepath.SkipDir
			}
		}

		// Check for manifest files
		if !info.IsDir() {
			langInfo := d.detectFromFile(path, info)
			if langInfo != nil {
				// Avoid duplicates by checking the root directory
				if !visited[langInfo.RootDir] {
					visited[langInfo.RootDir] = true
					languages = append(languages, langInfo)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return languages, nil
}

// detectFromFile detects language from a specific file
func (d *Detector) detectFromFile(path string, info os.FileInfo) *LanguageInfo {
	filename := info.Name()
	dir := filepath.Dir(path)

	switch filename {
	case "go.mod":
		return &LanguageInfo{
			Language:     Go,
			RootDir:      dir,
			ManifestFile: path,
		}

	case "pyproject.toml", "setup.py", "requirements.txt":
		return &LanguageInfo{
			Language:     Python,
			RootDir:      dir,
			ManifestFile: path,
		}

	case "package.json":
		// Check if it's TypeScript or JavaScript, and if React is used
		return d.detectJavaScriptProject(path, dir)

	case "tsconfig.json":
		return &LanguageInfo{
			Language:     TypeScript,
			RootDir:      dir,
			ManifestFile: path,
		}

	case "Cargo.toml":
		return &LanguageInfo{
			Language:     Rust,
			RootDir:      dir,
			ManifestFile: path,
		}

	case "pom.xml", "build.gradle", "build.gradle.kts":
		return &LanguageInfo{
			Language:     Java,
			RootDir:      dir,
			ManifestFile: path,
		}

	case "Gemfile":
		return &LanguageInfo{
			Language:     Ruby,
			RootDir:      dir,
			ManifestFile: path,
		}

	default:
		// Check for C# project files
		if strings.HasSuffix(filename, ".csproj") || strings.HasSuffix(filename, ".sln") {
			return &LanguageInfo{
				Language:     CSharp,
				RootDir:      dir,
				ManifestFile: path,
			}
		}
	}

	return nil
}

// detectJavaScriptProject detects JavaScript/TypeScript and React from package.json
func (d *Detector) detectJavaScriptProject(packageJSONPath, dir string) *LanguageInfo {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return defaultJSInfo(JavaScript, dir, packageJSONPath)
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return defaultJSInfo(JavaScript, dir, packageJSONPath)
	}

	isTypeScript := d.hasTypeScript(pkg)
	subLanguages := d.detectSubLanguages(pkg)
	primaryLang := d.resolvePrimaryLang(isTypeScript, subLanguages, dir)

	return &LanguageInfo{
		Language:     primaryLang,
		RootDir:      dir,
		ManifestFile: packageJSONPath,
		SubLanguages: subLanguages,
	}
}

func defaultJSInfo(lang Language, dir, path string) *LanguageInfo {
	return &LanguageInfo{
		Language:     lang,
		RootDir:      dir,
		ManifestFile: path,
	}
}

func (d *Detector) hasTypeScript(pkg struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}) bool {
	if pkg.Dependencies != nil {
		if _, ok := pkg.Dependencies["typescript"]; ok {
			return true
		}
	}
	if pkg.DevDependencies != nil {
		if _, ok := pkg.DevDependencies["typescript"]; ok {
			return true
		}
	}
	return false
}

func (d *Detector) detectSubLanguages(pkg struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}) []Language {
	if pkg.Dependencies != nil {
		if _, ok := pkg.Dependencies["react"]; ok {
			return []Language{React}
		}
	}
	return nil
}

func (d *Detector) resolvePrimaryLang(isTypeScript bool, subLanguages []Language, dir string) Language {
	if isTypeScript {
		return TypeScript
	}
	for _, l := range subLanguages {
		if l == React {
			if _, err := os.Stat(filepath.Join(dir, "tsconfig.json")); err == nil {
				return TypeScript
			}
		}
	}
	return JavaScript
}

// shouldSkipDir returns true if the directory should be skipped during detection
func shouldSkipDir(dirname string) bool {
	skipDirs := []string{
		"node_modules",
		".git",
		"vendor",
		".venv",
		"venv",
		"env",
		"__pycache__",
		"dist",
		"build",
		"target",
		"bin",
		"obj",
		".next",
		".nuxt",
		"coverage",
	}

	for _, skip := range skipDirs {
		if dirname == skip {
			return true
		}
	}

	return false
}

// DetectInDirectory is a convenience function to detect languages in a directory
func DetectInDirectory(dir string) ([]*LanguageInfo, error) {
	detector := NewDetector(dir)
	return detector.Detect()
}
