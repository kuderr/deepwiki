package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/generator"
	outputgen "github.com/kuderr/deepwiki/pkg/output/generator"
)

func TestNewOutputManager(t *testing.T) {
	manager := NewOutputManager()
	if manager == nil {
		t.Fatal("NewOutputManager returned nil")
	}
}

func TestOutputManager_GenerateOutput_Markdown(t *testing.T) {
	manager := NewOutputManager()
	tempDir := t.TempDir()

	// Create test data
	structure := &generator.WikiStructure{
		ID:          "test-wiki",
		Title:       "Test Wiki",
		Description: "A test wiki for unit testing",
		Version:     "1.0.0",
		Language:    "en",
		ProjectPath: "/test/project",
		CreatedAt:   time.Now(),
	}

	pages := map[string]*generator.WikiPage{
		"page1": {
			ID:          "page1",
			Title:       "Test Page 1",
			Description: "First test page",
			Content:     "# Test Content\n\nThis is test content.",
			Importance:  "high",
			WordCount:   10,
			SourceFiles: 2,
			FilePaths:   []string{"test1.go", "test2.go"},
			CreatedAt:   time.Now(),
		},
		"page2": {
			ID:          "page2",
			Title:       "Test Page 2",
			Description: "Second test page",
			Content:     "# Another Test\n\nMore test content.",
			Importance:  "medium",
			WordCount:   8,
			SourceFiles: 1,
			FilePaths:   []string{"test3.go"},
			CreatedAt:   time.Now(),
		},
	}

	options := outputgen.OutputOptions{
		Format:      outputgen.FormatMarkdown,
		Directory:   tempDir,
		Language:    "en",
		ProjectName: "test-project",
		ProjectPath: "/test/project",
	}

	result, err := manager.GenerateOutput(structure, pages, options)
	if err != nil {
		t.Fatalf("GenerateOutput failed: %v", err)
	}

	// Verify result
	if result.OutputDir != tempDir {
		t.Errorf("Expected output dir %s, got %s", tempDir, result.OutputDir)
	}

	if result.TotalFiles < 3 { // index.md + 2 pages + wiki-structure.json
		t.Errorf("Expected at least 3 files, got %d", result.TotalFiles)
	}

	// Check that files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "index.md"),
		filepath.Join(tempDir, "pages", "test-page-1.md"),
		filepath.Join(tempDir, "pages", "test-page-2.md"),
		filepath.Join(tempDir, "wiki-structure.json"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file was not created: %s", file)
		}
	}

	// Verify index.md content
	indexContent, err := os.ReadFile(filepath.Join(tempDir, "index.md"))
	if err != nil {
		t.Fatalf("Failed to read index.md: %v", err)
	}

	indexStr := string(indexContent)
	if !strings.Contains(indexStr, "Test Wiki") {
		t.Error("index.md should contain wiki title")
	}
	if !strings.Contains(indexStr, "High Importance") {
		t.Error("index.md should contain high importance section")
	}
	if !strings.Contains(indexStr, "Test Page 1") {
		t.Error("index.md should contain page 1 title")
	}
}

func TestOutputManager_GenerateOutput_JSON(t *testing.T) {
	manager := NewOutputManager()
	tempDir := t.TempDir()

	// Create test data
	structure := &generator.WikiStructure{
		ID:          "test-wiki",
		Title:       "Test Wiki",
		Description: "A test wiki for unit testing",
		Version:     "1.0.0",
		Language:    "en",
		ProjectPath: "/test/project",
		CreatedAt:   time.Now(),
	}

	pages := map[string]*generator.WikiPage{
		"page1": {
			ID:          "page1",
			Title:       "Test Page 1",
			Description: "First test page",
			Content:     "Test content",
			Importance:  "high",
			WordCount:   5,
			SourceFiles: 1,
			CreatedAt:   time.Now(),
		},
	}

	options := outputgen.OutputOptions{
		Format:      outputgen.FormatJSON,
		Directory:   tempDir,
		Language:    "en",
		ProjectName: "test-project",
		ProjectPath: "/test/project",
	}

	result, err := manager.GenerateOutput(structure, pages, options)
	if err != nil {
		t.Fatalf("GenerateOutput failed: %v", err)
	}

	// Verify result
	if result.OutputDir != tempDir {
		t.Errorf("Expected output dir %s, got %s", tempDir, result.OutputDir)
	}

	// Check that files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "wiki.json"),
		filepath.Join(tempDir, "pages", "page1.json"),
		filepath.Join(tempDir, "index.json"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file was not created: %s", file)
		}
	}

	// Verify wiki.json content
	wikiContent, err := os.ReadFile(filepath.Join(tempDir, "wiki.json"))
	if err != nil {
		t.Fatalf("Failed to read wiki.json: %v", err)
	}

	var wikiData map[string]interface{}
	if err := json.Unmarshal(wikiContent, &wikiData); err != nil {
		t.Fatalf("Failed to parse wiki.json: %v", err)
	}

	if wikiData["structure"] == nil {
		t.Error("wiki.json should contain structure")
	}
	if wikiData["pages"] == nil {
		t.Error("wiki.json should contain pages")
	}
}

func TestOutputManager_GenerateOutput_Docusaurus2(t *testing.T) {
	manager := NewOutputManager()
	tempDir := t.TempDir()

	// Create minimal test data
	structure := &generator.WikiStructure{
		ID:          "test-wiki",
		Title:       "Test Wiki",
		Description: "A test wiki",
		Version:     "1.0.0",
		Language:    "en",
		ProjectPath: "/test/project",
		CreatedAt:   time.Now(),
	}

	pages := map[string]*generator.WikiPage{
		"page1": {
			ID:          "page1",
			Title:       "Test Page",
			Description: "Test page",
			Content:     "Test content",
			Importance:  "high",
			WordCount:   5,
			SourceFiles: 1,
			CreatedAt:   time.Now(),
		},
	}

	options := outputgen.OutputOptions{
		Format:      outputgen.FormatDocusaurus2,
		Directory:   tempDir,
		Language:    "en",
		ProjectName: "test-project",
		ProjectPath: "/test/project",
	}

	result, err := manager.GenerateOutput(structure, pages, options)
	if err != nil {
		t.Fatalf("GenerateOutput failed: %v", err)
	}

	// Check that Docusaurus files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "docs", "intro.md"),
		filepath.Join(tempDir, "sidebars.js"),
		filepath.Join(tempDir, "docusaurus.config.js"),
		filepath.Join(tempDir, "package.json"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file was not created: %s", file)
		}
	}

	if result.TotalFiles < 4 {
		t.Errorf("Expected at least 4 files, got %d", result.TotalFiles)
	}
}

func TestOutputManager_GenerateOutput_Docusaurus3(t *testing.T) {
	manager := NewOutputManager()
	tempDir := t.TempDir()

	// Create minimal test data
	structure := &generator.WikiStructure{
		ID:          "test-wiki",
		Title:       "Test Wiki",
		Description: "A test wiki",
		Version:     "1.0.0",
		Language:    "en",
		ProjectPath: "/test/project",
		CreatedAt:   time.Now(),
	}

	pages := map[string]*generator.WikiPage{
		"page1": {
			ID:          "page1",
			Title:       "Test Page",
			Description: "Test page",
			Content:     "Test content",
			Importance:  "high",
			WordCount:   5,
			SourceFiles: 1,
			CreatedAt:   time.Now(),
		},
	}

	options := outputgen.OutputOptions{
		Format:      outputgen.FormatDocusaurus3,
		Directory:   tempDir,
		Language:    "en",
		ProjectName: "test-project",
		ProjectPath: "/test/project",
	}

	result, err := manager.GenerateOutput(structure, pages, options)
	if err != nil {
		t.Fatalf("GenerateOutput failed: %v", err)
	}

	// Check that Docusaurus v3 files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "docs", "intro.md"),
		filepath.Join(tempDir, "sidebars.ts"),
		filepath.Join(tempDir, "docusaurus.config.ts"),
		filepath.Join(tempDir, "package.json"),
		filepath.Join(tempDir, "tsconfig.json"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file was not created: %s", file)
		}
	}

	if result.TotalFiles < 5 {
		t.Errorf("Expected at least 5 files, got %d", result.TotalFiles)
	}
}

func TestOutputManager_UnsupportedFormat(t *testing.T) {
	manager := NewOutputManager()
	tempDir := t.TempDir()

	structure := &generator.WikiStructure{
		Title: "Test",
	}
	pages := map[string]*generator.WikiPage{}

	options := outputgen.OutputOptions{
		Format:    outputgen.OutputFormat("unsupported"),
		Directory: tempDir,
	}

	_, err := manager.GenerateOutput(structure, pages, options)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("Expected unsupported format error, got: %v", err)
	}
}
