package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kuderr/deepwiki/pkg/generator"
)

// SimpleDocusaurus3Generator generates Docusaurus v3-compatible markdown files with proper frontmatter
// This generator only creates markdown files with correct file structure and navigation metadata
// No configuration files (sidebars.ts, package.json, etc.) are generated
type SimpleDocusaurus3Generator struct{}

// NewSimpleDocusaurus3Generator creates a new Simple Docusaurus v3 generator
func NewSimpleDocusaurus3Generator() *SimpleDocusaurus3Generator {
	return &SimpleDocusaurus3Generator{}
}

// FormatType returns the format type this generator handles
func (sdg *SimpleDocusaurus3Generator) FormatType() OutputFormat {
	return FormatSimpleDocusaurus3
}

// Description returns a human-readable description of the format
func (sdg *SimpleDocusaurus3Generator) Description() string {
	return "Docusaurus v3-compatible markdown files only (no config files)"
}

// Generate creates Docusaurus v3-compatible markdown files with proper frontmatter and file structure
func (sdg *SimpleDocusaurus3Generator) Generate(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	options OutputOptions,
) (*OutputResult, error) {
	startTime := time.Now()

	// Create basic docs directory structure
	docsDir := filepath.Join(options.Directory, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create docs directory: %w", err)
	}

	var filesGenerated []string
	var totalSize int64
	var errors []error

	// Build navigation structure from pages
	navStructure := sdg.buildNavigationStructure(pages)

	// Generate intro.md (home page)
	introPath := filepath.Join(docsDir, "intro.md")
	if err := sdg.generateIntro(structure, pages, introPath, options, navStructure); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate intro: %w", err))
	} else {
		if stat, err := os.Stat(introPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, introPath)
	}

	// Generate individual page files with proper navigation
	for pageID, page := range pages {
		fileName := sdg.sanitizeFileName(page.Title) + ".md"
		pagePath := filepath.Join(docsDir, fileName)

		if err := sdg.generatePage(page, pagePath, structure, options, navStructure); err != nil {
			errors = append(errors, fmt.Errorf("failed to generate page %s: %w", pageID, err))
			continue
		}

		if stat, err := os.Stat(pagePath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, pagePath)
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

// buildNavigationStructure creates a navigation structure from pages
func (sdg *SimpleDocusaurus3Generator) buildNavigationStructure(
	pages map[string]*generator.WikiPage,
) map[string][]NavigationItem {
	structure := make(map[string][]NavigationItem)

	// Group pages by importance
	for _, page := range pages {
		importance := page.Importance
		if importance == "" {
			importance = "medium"
		}

		item := NavigationItem{
			ID:         page.ID,
			Title:      page.Title,
			FileName:   sdg.sanitizeFileName(page.Title),
			Importance: importance,
		}

		structure[importance] = append(structure[importance], item)
	}

	// Sort each group by title for consistent ordering
	for importance := range structure {
		sort.Slice(structure[importance], func(i, j int) bool {
			return structure[importance][i].Title < structure[importance][j].Title
		})

		// Assign positions
		for i := range structure[importance] {
			structure[importance][i].Position = i + 1
		}
	}

	return structure
}

// generateIntro generates the intro.md file with navigation links
func (sdg *SimpleDocusaurus3Generator) generateIntro(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	filePath string,
	options OutputOptions,
	navStructure map[string][]NavigationItem,
) error {
	var content strings.Builder

	// Write Docusaurus v3 frontmatter with enhanced features
	content.WriteString("---\n")
	content.WriteString("sidebar_position: 1\n")
	content.WriteString("slug: /\n")
	content.WriteString(fmt.Sprintf("title: %s\n", structure.Title))
	content.WriteString(fmt.Sprintf("description: %s\n", structure.Description))
	content.WriteString("tags:\n")
	content.WriteString("  - introduction\n")
	content.WriteString("  - getting-started\n")
	content.WriteString("---\n\n")

	// Write header
	content.WriteString(fmt.Sprintf("# %s\n\n", structure.Title))
	content.WriteString(fmt.Sprintf("%s\n\n", structure.Description))

	// Write navigation sections
	content.WriteString("## 📚 Documentation Sections\n\n")

	// High importance pages
	if items, exists := navStructure["high"]; exists && len(items) > 0 {
		content.WriteString("### 🔥 Essential Documentation\n\n")
		content.WriteString("Start here for the most important information about this project.\n\n")
		for _, item := range items {
			content.WriteString(fmt.Sprintf("- [%s](./%s) - Essential information\n", item.Title, item.FileName))
		}
		content.WriteString("\n")
	}

	// Medium importance pages
	if items, exists := navStructure["medium"]; exists && len(items) > 0 {
		content.WriteString("### 📋 Core Documentation\n\n")
		content.WriteString("Detailed documentation covering the main features and functionality.\n\n")
		for _, item := range items {
			content.WriteString(fmt.Sprintf("- [%s](./%s) - Core functionality\n", item.Title, item.FileName))
		}
		content.WriteString("\n")
	}

	// Low importance pages
	if items, exists := navStructure["low"]; exists && len(items) > 0 {
		content.WriteString("### 📝 Additional Information\n\n")
		content.WriteString("Supplementary documentation and reference materials.\n\n")
		for _, item := range items {
			content.WriteString(fmt.Sprintf("- [%s](./%s) - Additional details\n", item.Title, item.FileName))
		}
		content.WriteString("\n")
	}

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

// generatePage generates a Docusaurus v3-compatible page with proper frontmatter and navigation
func (sdg *SimpleDocusaurus3Generator) generatePage(
	page *generator.WikiPage,
	filePath string,
	structure *generator.WikiStructure,
	options OutputOptions,
	navStructure map[string][]NavigationItem,
) error {
	var content strings.Builder

	// Find this page in the navigation structure to determine position
	var sidebarPosition int
	found := false

	for importance, items := range navStructure {
		for _, item := range items {
			if item.ID == page.ID {
				found = true

				// Calculate sidebar position based on importance and position within group
				switch importance {
				case "high":
					sidebarPosition = 2 + item.Position // Start after intro (position 1)
				case "medium":
					highCount := len(navStructure["high"])
					sidebarPosition = 2 + highCount + item.Position
				case "low":
					highCount := len(navStructure["high"])
					mediumCount := len(navStructure["medium"])
					sidebarPosition = 2 + highCount + mediumCount + item.Position
				}
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		sidebarPosition = 100 // Default position for pages not in structure
	}

	// Generate slug from filename
	fileName := strings.TrimSuffix(filepath.Base(filePath), ".md")

	// Write enhanced Docusaurus v3 frontmatter with navigation
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("id: %s\n", page.ID))
	content.WriteString(fmt.Sprintf("title: %s\n", page.Title))
	content.WriteString(fmt.Sprintf("sidebar_position: %d\n", sidebarPosition))
	content.WriteString(fmt.Sprintf("slug: /%s\n", fileName))
	if page.Description != "" {
		content.WriteString(fmt.Sprintf("description: %s\n", page.Description))
	}

	// Add tags based on file extensions and importance
	content.WriteString("tags:\n")
	content.WriteString(fmt.Sprintf("  - %s\n", page.Importance))

	if len(page.FilePaths) > 0 {
		// Add language tags based on file extensions
		languages := make(map[string]bool)
		for _, path := range page.FilePaths {
			ext := filepath.Ext(path)
			switch ext {
			case ".go":
				languages["golang"] = true
			case ".js", ".jsx":
				languages["javascript"] = true
			case ".ts", ".tsx":
				languages["typescript"] = true
			case ".py":
				languages["python"] = true
			case ".java":
				languages["java"] = true
			case ".cpp", ".cc", ".cxx":
				languages["cpp"] = true
			case ".rs":
				languages["rust"] = true
			case ".md":
				languages["documentation"] = true
			}
		}
		for lang := range languages {
			content.WriteString(fmt.Sprintf("  - %s\n", lang))
		}
	}
	content.WriteString("---\n\n")

	// Write description if available
	if page.Description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", page.Description))
	}

	// Write main content
	content.WriteString(page.Content)
	content.WriteString("\n\n")

	return os.WriteFile(filePath, []byte(content.String()), 0o644)
}

func (sdg *SimpleDocusaurus3Generator) sanitizeFileName(name string) string {
	// Basic sanitization for Docusaurus-compatible filenames
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
