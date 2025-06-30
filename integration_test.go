package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/internal/config"
	"github.com/deepwiki-cli/deepwiki-cli/internal/logging"
	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
	"github.com/deepwiki-cli/deepwiki-cli/pkg/output"
	outputgen "github.com/deepwiki-cli/deepwiki-cli/pkg/output/generator"
	"github.com/deepwiki-cli/deepwiki-cli/pkg/scanner"
)

// TestIntegration_BasicFlow tests the basic integration flow
func TestIntegration_BasicFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary project directory
	projectDir := createTestProject(t, "go")
	defer os.RemoveAll(projectDir)

	outputDir := t.TempDir()

	// Test the basic flow: scan -> process -> generate
	ctx := context.Background()

	// 1. Load configuration
	cfg := config.DefaultConfig()
	cfg.Output.Directory = outputDir
	cfg.Output.Format = "markdown"

	// 2. Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// 3. Scan directory
	scanOptions := &scanner.ScanOptions{
		IncludeExtensions: cfg.Filters.IncludeExtensions,
		ExcludeDirs:       cfg.Filters.ExcludeDirs,
		ExcludeFiles:      cfg.Filters.ExcludeFiles,
		FollowSymlinks:    false,
		MaxDepth:          0,
		MaxFiles:          cfg.Processing.MaxFiles,
		AnalyzeContent:    true,
		MaxFileSize:       1024 * 1024,
		SkipBinaryFiles:   true,
		Concurrent:        true,
		MaxWorkers:        2,
	}

	fileScanner := scanner.NewScanner(scanOptions)
	scanResult, err := fileScanner.ScanDirectory(projectDir)
	if err != nil {
		t.Fatalf("Directory scan failed: %v", err)
	}

	if scanResult.FilteredFiles == 0 {
		t.Error("Expected to find at least some files")
	}

	// 4. Test output generation (without actual AI since we don't have API key)
	outputManager := output.NewOutputManager()

	// Create mock wiki structure and pages for testing
	mockStructure, mockPages := createMockWikiData()

	outputOptions := outputgen.OutputOptions{
		Format:      outputgen.FormatMarkdown,
		Directory:   outputDir,
		Language:    "en",
		ProjectName: "test-project",
		ProjectPath: projectDir,
	}

	result, err := outputManager.GenerateOutput(mockStructure, mockPages, outputOptions)
	if err != nil {
		t.Fatalf("Output generation failed: %v", err)
	}

	// 5. Verify output
	if result.TotalFiles == 0 {
		t.Error("Expected to generate at least some files")
	}

	// Check that key files exist
	expectedFiles := []string{
		filepath.Join(outputDir, "index.md"),
		filepath.Join(outputDir, "pages"),
		filepath.Join(outputDir, "wiki-structure.json"),
	}

	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file/directory not found: %s", expectedFile)
		}
	}

	logger.InfoContext(ctx, "Integration test completed successfully",
		"files_scanned", scanResult.TotalFiles,
		"files_filtered", scanResult.FilteredFiles,
		"output_files", result.TotalFiles,
		"output_size", result.TotalSize)
}

// TestIntegration_JSONOutput tests JSON output generation
func TestIntegration_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	outputDir := t.TempDir()
	outputManager := output.NewOutputManager()

	mockStructure, mockPages := createMockWikiData()

	outputOptions := outputgen.OutputOptions{
		Format:      outputgen.FormatJSON,
		Directory:   outputDir,
		Language:    "en",
		ProjectName: "test-project",
		ProjectPath: "/test/path",
	}

	result, err := outputManager.GenerateOutput(mockStructure, mockPages, outputOptions)
	if err != nil {
		t.Fatalf("JSON output generation failed: %v", err)
	}

	if result.TotalFiles == 0 {
		t.Error("Expected to generate at least some files")
	}

	// Check that JSON files exist
	expectedFiles := []string{
		filepath.Join(outputDir, "wiki.json"),
		filepath.Join(outputDir, "index.json"),
	}

	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected JSON file not found: %s", expectedFile)
		}
	}
}

// TestIntegration_ConcurrentProcessing tests concurrent processing
func TestIntegration_ConcurrentProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger, _ := logging.NewLogger(nil)
	defer logger.Close()

	tracker := &mockProgressTracker{}
	processor := output.NewConcurrentProcessor(logger.Logger, 2, tracker)

	// Create mock pages for concurrent processing
	pages := createMockPages(10)

	processedCount := 0
	processFunc := func(page *generator.WikiPage) error {
		processedCount++
		time.Sleep(10 * time.Millisecond) // Simulate work
		return nil
	}

	ctx := context.Background()
	err := processor.ProcessPagesParallel(ctx, pages, processFunc)
	if err != nil {
		t.Fatalf("Concurrent processing failed: %v", err)
	}

	if processedCount != len(pages) {
		t.Errorf("Expected to process %d pages, got %d", len(pages), processedCount)
	}
}

// TestIntegration_MemoryManagement tests memory-efficient processing
func TestIntegration_MemoryManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger, _ := logging.NewLogger(nil)
	defer logger.Close()

	// Set a low memory limit for testing
	processor := output.NewMemoryEfficientProcessor(logger.Logger, 50) // 50MB limit

	processed := false
	processingFunc := func() error {
		processed = true
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	ctx := context.Background()
	err := processor.ProcessWithMemoryManagement(ctx, processingFunc)
	if err != nil {
		t.Fatalf("Memory management processing failed: %v", err)
	}

	if !processed {
		t.Error("Processing function was not called")
	}

	stats := processor.GetMemoryStats()
	if stats.GCCount == 0 {
		t.Error("Expected at least one GC run")
	}
}

// TestIntegration_CacheSystem tests the caching system
func TestIntegration_CacheSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger, _ := logging.NewLogger(nil)
	defer logger.Close()

	tempDir := t.TempDir()
	cacheManager := output.NewCacheManager(logger.Logger, tempDir, true)

	// Test wiki page caching
	testPage := &generator.WikiPage{
		ID:      "test-page",
		Title:   "Test Page",
		Content: "This is test content",
	}

	// Cache the page
	err := cacheManager.CacheWikiPage(testPage.ID, testPage)
	if err != nil {
		t.Fatalf("Failed to cache wiki page: %v", err)
	}

	// Retrieve from cache
	cachedPage, found := cacheManager.GetCachedWikiPage(testPage.ID)
	if !found {
		t.Error("Should have found cached page")
	}

	if cachedPage.ID != testPage.ID {
		t.Errorf("Expected page ID %s, got %s", testPage.ID, cachedPage.ID)
	}

	// Test cache stats
	stats := cacheManager.GetStats()
	if stats.MemoryEntries == 0 {
		t.Error("Expected at least one memory entry")
	}
}

// Helper functions

func createTestProject(t *testing.T, projectType string) string {
	tempDir, err := os.MkdirTemp("", "deepwiki-test-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	switch projectType {
	case "go":
		createGoProject(t, tempDir)
	case "python":
		createPythonProject(t, tempDir)
	case "javascript":
		createJavaScriptProject(t, tempDir)
	default:
		createGoProject(t, tempDir) // Default to Go
	}

	return tempDir
}

func createGoProject(t *testing.T, dir string) {
	// Create go.mod
	goMod := `module test-project

go 1.24

require (
	github.com/example/dep v1.0.0
)
`
	writeFile(t, filepath.Join(dir, "go.mod"), goMod)

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	writeFile(t, filepath.Join(dir, "main.go"), mainGo)

	// Create a package
	pkgDir := filepath.Join(dir, "pkg", "utils")
	os.MkdirAll(pkgDir, 0o755)

	utilsGo := `package utils

// Calculator provides basic math operations
type Calculator struct{}

// Add adds two numbers
func (c *Calculator) Add(a, b int) int {
	return a + b
}

// Subtract subtracts two numbers
func (c *Calculator) Subtract(a, b int) int {
	return a - b
}
`
	writeFile(t, filepath.Join(pkgDir, "calculator.go"), utilsGo)

	// Create README
	codeBlock := "```bash\ngo run main.go\n```"
	readme := `# Test Project

This is a test project for deepwiki-cli integration testing.

## Features

- Basic calculator functionality
- Example Go project structure
- Integration test support

## Usage

` + codeBlock

	writeFile(t, filepath.Join(dir, "README.md"), readme)
}

func createPythonProject(t *testing.T, dir string) {
	// Create requirements.txt
	requirements := `requests==2.28.0
numpy==1.24.0
`
	writeFile(t, filepath.Join(dir, "requirements.txt"), requirements)

	// Create main.py
	mainPy := `#!/usr/bin/env python3
"""
Main module for the test project.
"""

from src.calculator import Calculator

def main():
    calc = Calculator()
    result = calc.add(2, 3)
    print(f"2 + 3 = {result}")

if __name__ == "__main__":
    main()
`
	writeFile(t, filepath.Join(dir, "main.py"), mainPy)

	// Create package
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0o755)

	calculatorPy := `"""
Calculator module providing basic math operations.
"""

class Calculator:
    """A simple calculator class."""
    
    def add(self, a: int, b: int) -> int:
        """Add two numbers."""
        return a + b
    
    def subtract(self, a: int, b: int) -> int:
        """Subtract two numbers."""
        return a - b
`
	writeFile(t, filepath.Join(srcDir, "calculator.py"), calculatorPy)
	writeFile(t, filepath.Join(srcDir, "__init__.py"), "")
}

func createJavaScriptProject(t *testing.T, dir string) {
	// Create package.json
	packageJson := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Test project for deepwiki-cli",
  "main": "index.js",
  "scripts": {
    "start": "node index.js",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.0"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}
`
	writeFile(t, filepath.Join(dir, "package.json"), packageJson)

	// Create index.js
	indexJs := `const Calculator = require('./src/calculator');

function main() {
    const calc = new Calculator();
    const result = calc.add(2, 3);
    console.log("2 + 3 = ${result}");
}

main();
`
	writeFile(t, filepath.Join(dir, "index.js"), indexJs)

	// Create src directory
	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0o755)

	calculatorJs := `/**
 * Calculator class providing basic math operations.
 */
class Calculator {
    /**
     * Add two numbers.
     * @param {number} a - First number
     * @param {number} b - Second number
     * @returns {number} Sum of a and b
     */
    add(a, b) {
        return a + b;
    }
    
    /**
     * Subtract two numbers.
     * @param {number} a - First number
     * @param {number} b - Second number
     * @returns {number} Difference of a and b
     */
    subtract(a, b) {
        return a - b;
    }
}

module.exports = Calculator;
`
	writeFile(t, filepath.Join(srcDir, "calculator.js"), calculatorJs)
}

func writeFile(t *testing.T, path, content string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// Mock progress tracker for testing
type mockProgressTracker struct{}

func (m *mockProgressTracker) StartTask(taskName string, total int)       {}
func (m *mockProgressTracker) UpdateProgress(current int, message string) {}
func (m *mockProgressTracker) CompleteTask(message string)                {}
func (m *mockProgressTracker) SetError(err error)                         {}

func createMockWikiData() (*generator.WikiStructure, map[string]*generator.WikiPage) {
	structure := &generator.WikiStructure{
		ID:          "test-wiki",
		Title:       "Test Project Documentation",
		Description: "Comprehensive documentation for the test project",
		Version:     "1.0.0",
	}

	pages := map[string]*generator.WikiPage{
		"overview": {
			ID:      "overview",
			Title:   "Project Overview",
			Content: "# Project Overview\n\nThis is a test project...",
		},
		"architecture": {
			ID:      "architecture",
			Title:   "System Architecture",
			Content: "# System Architecture\n\nThe system consists of...",
		},
		"api": {
			ID:      "api",
			Title:   "API Documentation",
			Content: "# API Documentation\n\nThe API provides...",
		},
	}

	return structure, pages
}

func createMockPages(count int) []*generator.WikiPage {
	pages := make([]*generator.WikiPage, count)
	for i := 0; i < count; i++ {
		pages[i] = &generator.WikiPage{
			ID:      fmt.Sprintf("page-%d", i),
			Title:   fmt.Sprintf("Test Page %d", i),
			Content: fmt.Sprintf("Content for page %d", i),
		}
	}
	return pages
}
