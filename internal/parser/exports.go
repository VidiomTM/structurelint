// Package parser provides functionality for parsing source code exports.
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// parseTypeScriptJavaScriptExports extracts exports from TypeScript/JavaScript files
func (p *Parser) parseTypeScriptJavaScriptExports(filePath string) ([]Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var exports []Export
	scanner := bufio.NewScanner(file)

	exportDefaultRegex := regexp.MustCompile(`^\s*export\s+default\s+`)
	exportNamedRegex := regexp.MustCompile(`^\s*export\s+\{\s*([^}]+)\s*\}`)
	exportDeclarationRegex := regexp.MustCompile(`^\s*export\s+(?:const|let|var|function|class|interface|type|enum)\s+(\w+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for default export
		if exportDefaultRegex.MatchString(line) {
			exports = append(exports, Export{
				SourceFile: filePath,
				Names:      []string{"default"},
				IsDefault:  true,
			})
			continue
		}

		// Check for named exports { foo, bar }
		if match := exportNamedRegex.FindStringSubmatch(line); match != nil {
			names := strings.Split(match[1], ",")
			var cleanNames []string
			for _, name := range names {
				// Handle "foo as bar" syntax
				parts := strings.Split(strings.TrimSpace(name), " as ")
				cleanName := strings.TrimSpace(parts[0])
				if cleanName != "" {
					cleanNames = append(cleanNames, cleanName)
				}
			}
			if len(cleanNames) > 0 {
				exports = append(exports, Export{
					SourceFile: filePath,
					Names:      cleanNames,
					IsDefault:  false,
				})
			}
			continue
		}

		// Check for export declarations
		if match := exportDeclarationRegex.FindStringSubmatch(line); match != nil {
			exports = append(exports, Export{
				SourceFile: filePath,
				Names:      []string{match[1]},
				IsDefault:  false,
			})
		}
	}

	return exports, scanner.Err()
}

// parseGoExports extracts exports from Go files
func (p *Parser) parseGoExports(filePath string) ([]Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var exports []Export
	scanner := bufio.NewScanner(file)

	// In Go, exports are indicated by uppercase first letter
	// func Foo(), type Bar struct, const Baz
	funcRegex := regexp.MustCompile(`^\s*func\s+([A-Z]\w*)`)
	typeRegex := regexp.MustCompile(`^\s*type\s+([A-Z]\w*)`)
	constVarRegex := regexp.MustCompile(`^\s*(?:const|var)\s+([A-Z]\w*)`)

	for scanner.Scan() {
		line := scanner.Text()

		if match := funcRegex.FindStringSubmatch(line); match != nil {
			exports = append(exports, Export{
				SourceFile: filePath,
				Names:      []string{match[1]},
				IsDefault:  false,
			})
		} else if match := typeRegex.FindStringSubmatch(line); match != nil {
			exports = append(exports, Export{
				SourceFile: filePath,
				Names:      []string{match[1]},
				IsDefault:  false,
			})
		} else if match := constVarRegex.FindStringSubmatch(line); match != nil {
			exports = append(exports, Export{
				SourceFile: filePath,
				Names:      []string{match[1]},
				IsDefault:  false,
			})
		}
	}

	return exports, scanner.Err()
}

// parsePythonExports extracts exports from Python files
func (p *Parser) parsePythonExports(filePath string) ([]Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var exports []Export
	scanner := bufio.NewScanner(file)

	// In Python, check for __all__ definition
	allRegex := regexp.MustCompile(`^\s*__all__\s*=\s*\[([^\]]+)\]`)
	// Also check for top-level def/class
	defClassRegex := regexp.MustCompile(`^(?:def|class)\s+([a-zA-Z_]\w*)`)

	var allExports []string
	var topLevelDefs []string

	for scanner.Scan() {
		line := scanner.Text()

		// Check for __all__
		if match := allRegex.FindStringSubmatch(line); match != nil {
			items := strings.Split(match[1], ",")
			for _, item := range items {
				cleanItem := strings.Trim(strings.TrimSpace(item), `"'`)
				if cleanItem != "" {
					allExports = append(allExports, cleanItem)
				}
			}
		}

		// Check for top-level definitions
		if match := defClassRegex.FindStringSubmatch(line); match != nil {
			name := match[1]
			// Skip private (starting with _)
			if !strings.HasPrefix(name, "_") {
				topLevelDefs = append(topLevelDefs, name)
			}
		}
	}

	// Prefer __all__ if defined, otherwise use top-level defs
	if len(allExports) > 0 {
		exports = append(exports, Export{
			SourceFile: filePath,
			Names:      allExports,
			IsDefault:  false,
		})
	} else if len(topLevelDefs) > 0 {
		exports = append(exports, Export{
			SourceFile: filePath,
			Names:      topLevelDefs,
			IsDefault:  false,
		})
	}

	return exports, scanner.Err()
}
