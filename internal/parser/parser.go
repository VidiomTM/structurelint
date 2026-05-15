package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Import represents an import statement found in a source file
type Import struct {
	SourceFile    string   // The file containing the import
	ImportPath    string   // The imported path/module
	IsRelative    bool     // Whether this is a relative import
	ImportedNames []string // Specific symbols imported (for Phase 2)
}

// Export represents an export statement found in a source file (Phase 2)
type Export struct {
	SourceFile string   // The file containing the export
	Names      []string // Exported symbol names
	IsDefault  bool     // Whether this is a default export
}

// Parser extracts imports from source files
type Parser struct {
	rootPath string
}

// New creates a new Parser
func New(rootPath string) *Parser {
	return &Parser{
		rootPath: rootPath,
	}
}

// ParseFile extracts imports from a single file
func (p *Parser) ParseFile(filePath string) ([]Import, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs":
		return p.parseTypeScriptJavaScript(filePath)
	case ".go":
		return p.parseGo(filePath)
	case ".py":
		return p.parsePython(filePath)
	case ".java":
		return p.parseJava(filePath)
	case ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp":
		return p.parseCpp(filePath)
	case ".cs":
		return p.parseCSharp(filePath)
	default:
		// Unsupported file type, return empty
		return []Import{}, nil
	}
}

// ParseExports extracts exports from a single file (Phase 2)
func (p *Parser) ParseExports(filePath string) ([]Export, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".mjs":
		return p.parseTypeScriptJavaScriptExports(filePath)
	case ".go":
		return p.parseGoExports(filePath)
	case ".py":
		return p.parsePythonExports(filePath)
	case ".java":
		return p.parseJavaExports(filePath)
	case ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp":
		return p.parseCppExports(filePath)
	case ".cs":
		return p.parseCSharpExports(filePath)
	default:
		// Unsupported file type, return empty
		return []Export{}, nil
	}
}

// parseTypeScriptJavaScript extracts imports from TypeScript/JavaScript files
func (p *Parser) parseTypeScriptJavaScript(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Regex patterns for different import styles
	// import foo from 'bar'
	// import { foo } from 'bar'
	// import * as foo from 'bar'
	// import 'bar'
	// const foo = require('bar')
	importRegex := regexp.MustCompile(`(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|import\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\))`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := importRegex.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			// Extract the import path from the matched groups
			importPath := ""
			for i := 1; i < len(match); i++ {
				if match[i] != "" {
					importPath = match[i]
					break
				}
			}

			if importPath != "" {
				isRelative := strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/")
				imports = append(imports, Import{
					SourceFile: filePath,
					ImportPath: importPath,
					IsRelative: isRelative,
				})
			}
		}
	}

	return imports, scanner.Err()
}

// parseGo extracts imports from Go files
func (p *Parser) parseGo(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Define regex patterns
	singleImportRegex := regexp.MustCompile(`^\s*import\s+"([^"]+)"`)
	importBlockStartRegex := regexp.MustCompile(`^\s*import\s+\(`)
	importInBlockRegex := regexp.MustCompile(`^\s*(?:\w+\s+)?"([^"]+)"`)

	inImportBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		
		inImportBlock = p.processGoImportLine(line, filePath, singleImportRegex, importBlockStartRegex, importInBlockRegex, inImportBlock, &imports)
	}

	return imports, scanner.Err()
}

func (p *Parser) processGoImportLine(line, filePath string, singleImportRegex, importBlockStartRegex, importInBlockRegex *regexp.Regexp, inImportBlock bool, imports *[]Import) bool {
	// Check for single-line import
	if match := singleImportRegex.FindStringSubmatch(line); match != nil {
		*imports = append(*imports, Import{
			SourceFile: filePath,
			ImportPath: match[1],
			IsRelative: strings.HasPrefix(match[1], "."),
		})
		return inImportBlock
	}

	// Check for import block start
	if importBlockStartRegex.MatchString(line) {
		return true
	}

	// Check for end of import block
	if inImportBlock && strings.TrimSpace(line) == ")" {
		return false
	}

	// Parse imports within block
	if inImportBlock {
		if match := importInBlockRegex.FindStringSubmatch(line); match != nil {
			*imports = append(*imports, Import{
				SourceFile: filePath,
				ImportPath: match[1],
				IsRelative: strings.HasPrefix(match[1], "."),
			})
		}
	}

	return inImportBlock
}

// parsePython extracts imports from Python files
func (p *Parser) parsePython(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// from foo import bar
	// import foo
	// import foo.bar
	importRegex := regexp.MustCompile(`^\s*(?:from\s+([\w.]+)\s+import|import\s+([\w.]+))`)

	for scanner.Scan() {
		line := scanner.Text()
		if match := importRegex.FindStringSubmatch(line); match != nil {
			importPath := ""
			if match[1] != "" {
				importPath = match[1]
			} else if match[2] != "" {
				importPath = match[2]
			}

			if importPath != "" {
				isRelative := strings.HasPrefix(importPath, ".")
				imports = append(imports, Import{
					SourceFile: filePath,
					ImportPath: importPath,
					IsRelative: isRelative,
				})
			}
		}
	}

	return imports, scanner.Err()
}

// parseJava extracts imports from Java files
func (p *Parser) parseJava(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Extract package name from current file to determine relative imports
	var currentPackage string

	// Regex patterns for Java imports
	// import com.example.MyClass;
	// import com.example.*;
	// import static com.example.MyClass.staticMethod;
	packageRegex := regexp.MustCompile(`^\s*package\s+([\w.]+);`)
	importRegex := regexp.MustCompile(`^\s*import\s+(?:static\s+)?([\w.]+)(?:\.\*)?\s*;`)

	for scanner.Scan() {
		line := scanner.Text()

		// Extract package declaration
		if match := packageRegex.FindStringSubmatch(line); match != nil {
			currentPackage = match[1]
			continue
		}

		// Extract import statements
		if match := importRegex.FindStringSubmatch(line); match != nil {
			importPath := match[1]

			// Determine if relative (same package prefix)
			isRelative := false
			if currentPackage != "" && strings.HasPrefix(importPath, currentPackage+".") {
				isRelative = true
			}

			imports = append(imports, Import{
				SourceFile: filePath,
				ImportPath: importPath,
				IsRelative: isRelative,
			})
		}
	}

	return imports, scanner.Err()
}

// parseJavaExports extracts public classes/interfaces/methods from Java
func (p *Parser) parseJavaExports(filePath string) ([]Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var exports []Export
	scanner := bufio.NewScanner(file)

	// Regex patterns for public declarations
	// public class ClassName
	// public interface InterfaceName
	// public enum EnumName
	// public abstract class AbstractClass
	// public final class FinalClass
	classRegex := regexp.MustCompile(`^\s*public\s+(?:abstract\s+|final\s+)?(?:class|interface|enum)\s+(\w+)`)

	var exportNames []string

	for scanner.Scan() {
		line := scanner.Text()

		// Extract public class/interface/enum declarations
		if match := classRegex.FindStringSubmatch(line); match != nil {
			exportNames = append(exportNames, match[1])
		}
	}

	if len(exportNames) > 0 {
		exports = append(exports, Export{
			SourceFile: filePath,
			Names:      exportNames,
			IsDefault:  false,
		})
	}

	return exports, scanner.Err()
}

// parseCpp extracts includes from C++ files
func (p *Parser) parseCpp(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Regex pattern for C++ includes
	// #include "myheader.h"       (relative, local)
	// #include <iostream>         (system/library)
	// #include <boost/algorithm.hpp>  (external library)
	includeRegex := regexp.MustCompile(`^\s*#include\s+([<"])([^>"]+)[>"]`)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip lines that are commented out
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		if match := includeRegex.FindStringSubmatch(line); match != nil {
			delimiter := match[1]
			includePath := match[2]

			// Quote includes ("") = relative/local
			// Angle bracket includes (<>) = system/external
			isRelative := delimiter == "\""

			imports = append(imports, Import{
				SourceFile: filePath,
				ImportPath: includePath,
				IsRelative: isRelative,
			})
		}
	}

	return imports, scanner.Err()
}

// parseCppExports extracts public symbols from C++ headers
func (p *Parser) parseCppExports(filePath string) ([]Export, error) {
	ext := filepath.Ext(filePath)

	// Only parse header files (.h, .hpp)
	if ext != ".h" && ext != ".hpp" {
		return []Export{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var exports []Export
	scanner := bufio.NewScanner(file)

	classRegex := regexp.MustCompile(`^\s*(?:class|struct)\s+(\w+)`)
	namespaceRegex := regexp.MustCompile(`^\s*namespace\s+(\w+)`)

	var exportNames []string
	inComment := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Process comment state
		inComment = p.processCppComment(trimmed, inComment)
		if inComment || trimmed == "" {
			continue
		}

		// Extract declarations
		p.extractCppDeclarations(trimmed, classRegex, namespaceRegex, &exportNames)
	}

	if len(exportNames) > 0 {
		exports = append(exports, Export{
			SourceFile: filePath,
			Names:      exportNames,
			IsDefault:  false,
		})
	}

	return exports, scanner.Err()
}

func (p *Parser) processCppComment(trimmed string, inComment bool) bool {
	// Skip single-line comments
	if strings.HasPrefix(trimmed, "//") {
		return inComment
	}

	// Handle inline block comments
	if strings.Contains(trimmed, "/*") && strings.Contains(trimmed, "*/") {
		return inComment
	}

	// Start of multi-line comment
	if strings.Contains(trimmed, "/*") {
		return true
	}

	// End of multi-line comment
	if strings.Contains(trimmed, "*/") {
		return false
	}

	return inComment
}

func (p *Parser) extractCppDeclarations(trimmed string, classRegex, namespaceRegex *regexp.Regexp, exportNames *[]string) {
	// Extract class/struct declarations
	if match := classRegex.FindStringSubmatch(trimmed); match != nil {
		*exportNames = append(*exportNames, match[1])
	}

	// Extract namespace declarations
	if match := namespaceRegex.FindStringSubmatch(trimmed); match != nil {
		*exportNames = append(*exportNames, match[1])
	}
}

// parseCSharp extracts using statements from C# files
func (p *Parser) parseCSharp(filePath string) ([]Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []Import
	scanner := bufio.NewScanner(file)

	// Extract namespace from current file to determine relative imports
	var currentNamespace string

	// Regex patterns for C# using statements
	// using System;
	// using System.Collections.Generic;
	// using MyNamespace.SubNamespace;
	// using static System.Math;
	// using Alias = System.Text.StringBuilder;
	namespaceRegex := regexp.MustCompile(`^\s*namespace\s+([\w.]+)`)
	usingRegex := regexp.MustCompile(`^\s*using\s+(?:static\s+)?(?:\w+\s*=\s*)?([\w.]+)\s*;`)

	for scanner.Scan() {
		line := scanner.Text()

		// Extract namespace declaration
		if match := namespaceRegex.FindStringSubmatch(line); match != nil {
			currentNamespace = match[1]
			continue
		}

		// Extract using statements
		if match := usingRegex.FindStringSubmatch(line); match != nil {
			importPath := match[1]

			// Determine if relative (same namespace hierarchy)
			// Only consider exact namespace prefix matches as relative
			// to avoid false positives with third-party libraries
			isRelative := false
			if currentNamespace != "" && strings.HasPrefix(importPath, currentNamespace+".") {
				isRelative = true
			}

			// Also check if it's in the same root namespace as the current file
			// but exclude common third-party and system namespaces
			if currentNamespace != "" && !isRelative {
				commonExternalNamespaces := map[string]bool{
					"System": true, "Microsoft": true, "Newtonsoft": true,
					"AutoMapper": true, "Serilog": true, "NLog": true,
					"Xunit": true, "NUnit": true, "Moq": true,
					"FluentAssertions": true, "MediatR": true,
				}

				currentRoot := strings.Split(currentNamespace, ".")[0]
				importRoot := strings.Split(importPath, ".")[0]

				// Only mark as relative if same root AND not a known external namespace
				if currentRoot == importRoot && !commonExternalNamespaces[currentRoot] {
					isRelative = true
				}
			}

			imports = append(imports, Import{
				SourceFile: filePath,
				ImportPath: importPath,
				IsRelative: isRelative,
			})
		}
	}

	return imports, scanner.Err()
}

// parseCSharpExports extracts public types from C# files
func (p *Parser) parseCSharpExports(filePath string) ([]Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var exports []Export
	scanner := bufio.NewScanner(file)

	// Regex patterns for public declarations
	// public class ClassName
	// public interface IInterfaceName
	// public struct StructName
	// public enum EnumName
	// public delegate ...
	// public sealed class SealedClass
	// public abstract class AbstractClass
	typeRegex := regexp.MustCompile(`^\s*public\s+(?:sealed\s+|abstract\s+|static\s+|partial\s+)?(?:class|interface|struct|enum|delegate)\s+(\w+)`)

	var exportNames []string

	for scanner.Scan() {
		line := scanner.Text()

		// Extract public type declarations
		if match := typeRegex.FindStringSubmatch(line); match != nil {
			exportNames = append(exportNames, match[1])
		}
	}

	if len(exportNames) > 0 {
		exports = append(exports, Export{
			SourceFile: filePath,
			Names:      exportNames,
			IsDefault:  false,
		})
	}

	return exports, scanner.Err()
}

// ResolveImportPath resolves a relative import path to a path within the project
func (p *Parser) ResolveImportPath(sourceFile, importPath string) string {
	if !strings.HasPrefix(importPath, ".") {
		// Not a relative import, return as-is
		return importPath
	}

	// Get the directory of the source file
	sourceDir := filepath.Dir(sourceFile)

	// Resolve the import path relative to the source directory
	resolvedPath := filepath.Join(sourceDir, importPath)

	// Clean the path to resolve .. and .
	resolvedPath = filepath.Clean(resolvedPath)

	return resolvedPath
}
