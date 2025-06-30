package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestDirectory creates a temporary test directory with sample files
func createTestDirectory(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "deepwiki-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create directory structure
	dirs := []string{
		"src",
		"src/components",
		"tests",
		"docs",
		"node_modules",
		".git",
		"dist",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0o755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := map[string]string{
		"README.md":              "# Test Project\n\nThis is a test project.",
		"package.json":           `{"name": "test", "version": "1.0.0"}`,
		"src/main.go":            "package main\n\nfunc main() {\n\tprintln(\"Hello World\")\n}",
		"src/utils.js":           "function add(a, b) {\n\treturn a + b;\n}",
		"src/components/App.tsx": "import React from 'react';\n\nexport default function App() {\n\treturn <div>Hello</div>;\n}",
		"tests/main_test.go":     "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}",
		"docs/api.md":            "# API Documentation\n\n## Endpoints",
		"node_modules/lib.js":    "// This should be excluded",
		".git/config":            "[core]\n\trepositoryformatversion = 0",
		"dist/bundle.js":         "// Built file",
		"image.png":              "fake binary content",
		"config.yaml":            "debug: true\nport: 8080",
		"Dockerfile":             "FROM node:16\nCOPY . .",
		"src/style.css":          "body { margin: 0; }",
		"script.sh":              "#!/bin/bash\necho hello",
	}

	for filename, content := range files {
		fullPath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	return tempDir
}

func TestNewScanner(t *testing.T) {
	scanner := NewScanner(nil)
	if scanner == nil {
		t.Fatal("NewScanner returned nil")
	}

	if scanner.options == nil {
		t.Fatal("Scanner options is nil")
	}

	if scanner.logger == nil {
		t.Fatal("Scanner logger is nil")
	}
}

func TestScanDirectory_Basic(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	scanner := NewScanner(DefaultScanOptions())
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if result == nil {
		t.Fatal("ScanResult is nil")
	}

	// Should find files but exclude node_modules, .git, dist
	if result.FilteredFiles == 0 {
		t.Error("Expected to find some files")
	}

	if result.TotalFiles == 0 {
		t.Error("Expected total files > 0")
	}

	if result.ScanTime == 0 {
		t.Error("Expected scan time > 0")
	}

	// Check that excluded directories are actually excluded
	for _, file := range result.Files {
		if strings.Contains(file.Path, "node_modules") {
			t.Errorf("node_modules file should be excluded: %s", file.Path)
		}
		if strings.Contains(file.Path, ".git") {
			t.Errorf(".git file should be excluded: %s", file.Path)
		}
		if strings.Contains(file.Path, "dist") {
			t.Errorf("dist file should be excluded: %s", file.Path)
		}
	}
}

func TestScanDirectory_FileTypes(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	scanner := NewScanner(DefaultScanOptions())
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Track found file types
	foundTypes := make(map[string]bool)
	for _, file := range result.Files {
		foundTypes[file.Extension] = true
	}

	// Should find various file types
	expectedTypes := []string{".md", ".go", ".js", ".tsx", ".json", ".yaml", ".css", ".sh"}
	for _, ext := range expectedTypes {
		if !foundTypes[ext] {
			t.Errorf("Expected to find files with extension %s", ext)
		}
	}
}

func TestScanDirectory_LanguageDetection(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	scanner := NewScanner(DefaultScanOptions())
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Check language detection
	languageMap := make(map[string]int)
	for _, file := range result.Files {
		if file.Language != "" {
			languageMap[file.Language]++
		}
	}

	// Should detect various languages
	expectedLanguages := []string{"Go", "JavaScript", "TypeScript React", "Markdown", "JSON", "YAML", "CSS", "Shell"}
	for _, lang := range expectedLanguages {
		if languageMap[lang] == 0 {
			t.Errorf("Expected to detect language %s", lang)
		}
	}
}

func TestScanDirectory_Categories(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	scanner := NewScanner(DefaultScanOptions())
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Check categories
	categoryMap := make(map[string]int)
	for _, file := range result.Files {
		if file.Category != "" {
			categoryMap[file.Category]++
		}
	}

	// Should categorize files correctly
	expectedCategories := []string{"code", "docs", "config", "test"}
	for _, category := range expectedCategories {
		if categoryMap[category] == 0 {
			t.Errorf("Expected to find files in category %s", category)
		}
	}
}

func TestScanDirectory_NonExistentPath(t *testing.T) {
	scanner := NewScanner(DefaultScanOptions())
	_, err := scanner.ScanDirectory("/non/existent/path")

	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestScanDirectory_FileAsPath(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	// Try to scan a file instead of directory
	filePath := filepath.Join(tempDir, "README.md")
	scanner := NewScanner(DefaultScanOptions())
	_, err := scanner.ScanDirectory(filePath)

	if err == nil {
		t.Error("Expected error when scanning a file instead of directory")
	}
}

func TestScanOptions_MaxFiles(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	options := DefaultScanOptions()
	options.MaxFiles = 3 // Limit to 3 files

	scanner := NewScanner(options)
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(result.Files) > 3 {
		t.Errorf("Expected at most 3 files, got %d", len(result.Files))
	}
}

func TestScanOptions_ExcludeDirectories(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	options := DefaultScanOptions()
	options.ExcludeDirs = []string{"src", "tests"} // Exclude src and tests

	scanner := NewScanner(options)
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Should not find files from src or tests directories
	for _, file := range result.Files {
		if strings.HasPrefix(file.Path, "src/") || strings.HasPrefix(file.Path, "tests/") {
			t.Errorf("File from excluded directory found: %s", file.Path)
		}
	}
}

func TestScanOptions_IncludeExtensions(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	options := DefaultScanOptions()
	options.IncludeExtensions = []string{".go", ".md"} // Only Go and Markdown files

	scanner := NewScanner(options)
	result, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Should only find .go and .md files
	for _, file := range result.Files {
		if file.Extension != ".go" && file.Extension != ".md" {
			t.Errorf("Unexpected file extension found: %s (file: %s)", file.Extension, file.Path)
		}
	}
}

func TestGetLanguageByExtension(t *testing.T) {
	tests := []struct {
		extension string
		expected  string
		category  FileCategory
	}{
		{".go", "Go", CategoryCode},
		{".py", "Python", CategoryCode},
		{".js", "JavaScript", CategoryCode},
		{".md", "Markdown", CategoryDocs},
		{".json", "JSON", CategoryConfig},
		{".unknown", "Unknown", CategoryUnknown},
	}

	for _, test := range tests {
		lang := GetLanguageByExtension(test.extension)
		if lang.Name != test.expected {
			t.Errorf("Expected language %s for extension %s, got %s", test.expected, test.extension, lang.Name)
		}
		if lang.Category != test.category {
			t.Errorf("Expected category %s for extension %s, got %s", test.category, test.extension, lang.Category)
		}
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"main_test.go", true},
		{"app.test.js", true},
		{"component.spec.tsx", true},
		{"test_utils.py", true},
		{"main.go", false},
		{"app.js", false},
		{"README.md", false},
	}

	for _, test := range tests {
		result := isTestFile(test.filename)
		if result != test.expected {
			t.Errorf("isTestFile(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestScannerStats(t *testing.T) {
	tempDir := createTestDirectory(t)
	defer os.RemoveAll(tempDir)

	scanner := NewScanner(DefaultScanOptions())
	_, err := scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	stats := scanner.GetStats()

	if stats.FilesProcessed == 0 {
		t.Error("Expected FilesProcessed > 0")
	}

	if stats.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	if stats.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}

	if stats.EndTime.Before(stats.StartTime) {
		t.Error("EndTime should be after StartTime")
	}
}

func TestDefaultScanOptions(t *testing.T) {
	options := DefaultScanOptions()

	if options == nil {
		t.Fatal("DefaultScanOptions returned nil")
	}

	if len(options.IncludeExtensions) == 0 {
		t.Error("Expected some include extensions")
	}

	if len(options.ExcludeDirs) == 0 {
		t.Error("Expected some exclude directories")
	}

	if options.MaxFiles <= 0 {
		t.Error("Expected MaxFiles > 0")
	}

	// Check that common extensions are included
	hasGo := false
	hasPython := false
	hasJS := false
	for _, ext := range options.IncludeExtensions {
		switch ext {
		case ".go":
			hasGo = true
		case ".py":
			hasPython = true
		case ".js":
			hasJS = true
		}
	}

	if !hasGo || !hasPython || !hasJS {
		t.Error("Expected common programming language extensions to be included")
	}

	// Check that common exclude directories are present
	hasNodeModules := false
	hasGit := false
	for _, dir := range options.ExcludeDirs {
		switch dir {
		case "node_modules":
			hasNodeModules = true
		case ".git":
			hasGit = true
		}
	}

	if !hasNodeModules || !hasGit {
		t.Error("Expected common directories to be excluded")
	}
}
