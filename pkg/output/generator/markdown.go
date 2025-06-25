package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

// MarkdownGenerator generates markdown output files
type MarkdownGenerator struct{}

// NewMarkdownGenerator creates a new markdown generator
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{}
}

// FormatType returns the format type this generator handles
func (mg *MarkdownGenerator) FormatType() OutputFormat {
	return FormatMarkdown
}

// Description returns a human-readable description of the format
func (mg *MarkdownGenerator) Description() string {
	return "Standard markdown files with organized directory structure"
}

// Generate creates markdown output files
func (mg *MarkdownGenerator) Generate(structure *generator.WikiStructure, pages map[string]*generator.WikiPage, options OutputOptions) (*OutputResult, error) {
	startTime := time.Now()

	// Create output directory structure
	if err := mg.organizeFiles(options.Directory); err != nil {
		return nil, fmt.Errorf("failed to create directory structure: %w", err)
	}

	var filesGenerated []string
	var totalSize int64
	var errors []error

	// Generate index file
	indexPath := filepath.Join(options.Directory, "index.md")
	if err := mg.generateIndex(structure, pages, indexPath, options); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate index: %w", err))
	} else {
		if stat, err := os.Stat(indexPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, indexPath)
	}

	// Generate individual page files
	pagesDir := filepath.Join(options.Directory, "pages")
	for pageID, page := range pages {
		fileName := mg.sanitizeFileName(page.Title) + ".md"
		pagePath := filepath.Join(pagesDir, fileName)

		if err := mg.generatePage(page, pagePath, structure, options); err != nil {
			errors = append(errors, fmt.Errorf("failed to generate page %s: %w", pageID, err))
			continue
		}

		if stat, err := os.Stat(pagePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, pagePath)
	}

	// Generate wiki structure JSON for reference
	structurePath := filepath.Join(options.Directory, "wiki-structure.json")
	if err := mg.generateWikiStructureJSON(structure, pages, structurePath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate wiki structure: %w", err))
	} else {
		if stat, err := os.Stat(structurePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, structurePath)
	}

	return &OutputResult{
		OutputDir:      options.Directory,
		FilesGenerated: filesGenerated,
		TotalFiles:     len(filesGenerated),
		TotalSize:      totalSize,
		GeneratedAt:    time.Now(),
		ProcessingTime: time.Since(startTime),
		Errors:         errors,
	}, nil
}

// organizeFiles creates the necessary directory structure
func (mg *MarkdownGenerator) organizeFiles(outputDir string) error {
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "pages"),
		filepath.Join(outputDir, "assets"),
		filepath.Join(outputDir, "assets", "diagrams"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (mg *MarkdownGenerator) generateIndex(structure *generator.WikiStructure, pages map[string]*generator.WikiPage, filePath string, options OutputOptions) error {
	var content strings.Builder

	// Write header
	content.WriteString(fmt.Sprintf("# %s\n\n", structure.Title))
	content.WriteString(fmt.Sprintf("%s\n\n", structure.Description))

	// Write metadata
	content.WriteString("## 📊 Wiki Information\n\n")
	content.WriteString(fmt.Sprintf("- **Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("- **Project**: %s\n", options.ProjectName))
	content.WriteString(fmt.Sprintf("- **Language**: %s\n", options.Language))
	content.WriteString(fmt.Sprintf("- **Total Pages**: %d\n", len(pages)))
	content.WriteString(fmt.Sprintf("- **Version**: %s\n\n", structure.Version))

	// Write page index
	content.WriteString("## 📚 Pages\n\n")

	// Group pages by importance
	importanceGroups := map[string][]*generator.WikiPage{
		"high":   {},
		"medium": {},
		"low":    {},
	}

	for _, page := range pages {
		importance := page.Importance
		if importance == "" {
			importance = "medium"
		}
		importanceGroups[importance] = append(importanceGroups[importance], page)
	}

	// Write high importance pages
	if len(importanceGroups["high"]) > 0 {
		content.WriteString("### 🔥 High Importance\n\n")
		for _, page := range importanceGroups["high"] {
			fileName := mg.sanitizeFileName(page.Title) + ".md"
			content.WriteString(fmt.Sprintf("- [%s](pages/%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write medium importance pages
	if len(importanceGroups["medium"]) > 0 {
		content.WriteString("### 📋 Medium Importance\n\n")
		for _, page := range importanceGroups["medium"] {
			fileName := mg.sanitizeFileName(page.Title) + ".md"
			content.WriteString(fmt.Sprintf("- [%s](pages/%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write low importance pages
	if len(importanceGroups["low"]) > 0 {
		content.WriteString("### 📝 Additional Information\n\n")
		for _, page := range importanceGroups["low"] {
			fileName := mg.sanitizeFileName(page.Title) + ".md"
			content.WriteString(fmt.Sprintf("- [%s](pages/%s) - %s\n", page.Title, fileName, page.Description))
		}
		content.WriteString("\n")
	}

	// Write footer
	content.WriteString("---\n\n")
	content.WriteString("*Generated by DeepWiki CLI*\n")

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (mg *MarkdownGenerator) generatePage(page *generator.WikiPage, filePath string, structure *generator.WikiStructure, options OutputOptions) error {
	var content strings.Builder

	// Write header
	content.WriteString(fmt.Sprintf("# %s\n\n", page.Title))

	// Write metadata
	content.WriteString("<details>\n")
	content.WriteString("<summary>📋 Page Information</summary>\n\n")
	content.WriteString(fmt.Sprintf("- **Page ID**: %s\n", page.ID))
	content.WriteString(fmt.Sprintf("- **Importance**: %s\n", page.Importance))
	content.WriteString(fmt.Sprintf("- **Word Count**: %d\n", page.WordCount))
	content.WriteString(fmt.Sprintf("- **Source Files**: %d\n", page.SourceFiles))
	content.WriteString(fmt.Sprintf("- **Generated**: %s\n", page.CreatedAt.Format("2006-01-02 15:04:05")))

	if len(page.FilePaths) > 0 {
		content.WriteString("- **Source Files**:\n")
		for _, path := range page.FilePaths {
			content.WriteString(fmt.Sprintf("  - `%s`\n", path))
		}
	}

	if len(page.RelatedPages) > 0 {
		content.WriteString("- **Related Pages**:\n")
		for _, relatedID := range page.RelatedPages {
			content.WriteString(fmt.Sprintf("  - %s\n", relatedID))
		}
	}

	content.WriteString("\n</details>\n\n")

	// Write description if available
	if page.Description != "" {
		content.WriteString(fmt.Sprintf("## Overview\n\n%s\n\n", page.Description))
	}

	// Write main content
	content.WriteString(page.Content)
	content.WriteString("\n\n")

	// Write navigation
	content.WriteString("---\n\n")
	content.WriteString("## Navigation\n\n")
	content.WriteString("- [🏠 Home](../index.md)\n")
	content.WriteString("- [📚 All Pages](../index.md#-pages)\n")

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (mg *MarkdownGenerator) generateWikiStructureJSON(structure *generator.WikiStructure, pages map[string]*generator.WikiPage, filePath string) error {
	data := map[string]interface{}{
		"structure": structure,
		"pages":     pages,
		"metadata": map[string]interface{}{
			"generatedAt": time.Now(),
			"totalPages":  len(pages),
		},
	}

	return mg.writeJSONFile(data, filePath)
}

func (mg *MarkdownGenerator) writeJSONFile(data interface{}, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (mg *MarkdownGenerator) sanitizeFileName(name string) string {
	// Basic sanitization
	result := strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "/", "-")
	result = strings.ReplaceAll(result, "\\", "-")
	result = strings.ReplaceAll(result, ":", "-")
	result = strings.ReplaceAll(result, "*", "-")
	result = strings.ReplaceAll(result, "?", "-")
	result = strings.ReplaceAll(result, "\"", "-")
	result = strings.ReplaceAll(result, "<", "-")
	result = strings.ReplaceAll(result, ">", "-")
	result = strings.ReplaceAll(result, "|", "-")

	// Remove multiple consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim dashes from start and end
	result = strings.Trim(result, "-")

	// Ensure not empty
	if result == "" {
		result = "untitled"
	}

	return result
}
