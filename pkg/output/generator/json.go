package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

// JSONGenerator generates JSON output files
type JSONGenerator struct{}

// NewJSONGenerator creates a new JSON generator
func NewJSONGenerator() *JSONGenerator {
	return &JSONGenerator{}
}

// FormatType returns the format type this generator handles
func (jg *JSONGenerator) FormatType() OutputFormat {
	return FormatJSON
}

// Description returns a human-readable description of the format
func (jg *JSONGenerator) Description() string {
	return "Structured JSON format for programmatic consumption"
}

// Generate creates JSON output files
func (jg *JSONGenerator) Generate(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	options OutputOptions,
) (*OutputResult, error) {
	startTime := time.Now()

	// Create output directory
	if err := os.MkdirAll(options.Directory, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var filesGenerated []string
	var totalSize int64
	var errors []error

	// Generate main wiki JSON file
	wikiData := map[string]interface{}{
		"structure": structure,
		"pages":     pages,
		"metadata": map[string]interface{}{
			"generatedAt": time.Now(),
			"projectName": options.ProjectName,
			"projectPath": options.ProjectPath,
			"language":    options.Language,
			"totalPages":  len(pages),
		},
	}

	wikiPath := filepath.Join(options.Directory, "wiki.json")
	if err := jg.writeJSONFile(wikiData, wikiPath); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate wiki JSON: %w", err))
	} else {
		if stat, err := os.Stat(wikiPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, wikiPath)
	}

	// Generate individual page JSON files
	pagesDir := filepath.Join(options.Directory, "pages")
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		errors = append(errors, fmt.Errorf("failed to create pages directory: %w", err))
	} else {
		for pageID, page := range pages {
			fileName := pageID + ".json"
			pagePath := filepath.Join(pagesDir, fileName)

			if err := jg.writeJSONFile(page, pagePath); err != nil {
				errors = append(errors, fmt.Errorf("failed to generate page JSON %s: %w", pageID, err))
				continue
			}

			if stat, err := os.Stat(pagePath); err == nil {
				totalSize += stat.Size()
			}
			filesGenerated = append(filesGenerated, pagePath)
		}
	}

	// Generate index JSON
	indexPath := filepath.Join(options.Directory, "index.json")
	if err := jg.generateIndex(structure, pages, indexPath, options); err != nil {
		errors = append(errors, fmt.Errorf("failed to generate JSON index: %w", err))
	} else {
		if stat, err := os.Stat(indexPath); err == nil {
			totalSize += stat.Size()
		}
		filesGenerated = append(filesGenerated, indexPath)
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

func (jg *JSONGenerator) generateIndex(
	structure *generator.WikiStructure,
	pages map[string]*generator.WikiPage,
	filePath string,
	options OutputOptions,
) error {
	// Build index structure
	indexPages := make([]IndexPage, 0, len(pages))
	stats := IndexStats{}

	for _, page := range pages {
		indexPage := IndexPage{
			ID:          page.ID,
			Title:       page.Title,
			Description: page.Description,
			FilePath:    fmt.Sprintf("pages/%s.json", page.ID),
			Importance:  page.Importance,
			ParentID:    page.ParentID,
			WordCount:   page.WordCount,
			SourceFiles: page.SourceFiles,
		}
		indexPages = append(indexPages, indexPage)

		// Update stats
		stats.TotalPages++
		stats.TotalWords += page.WordCount
		stats.TotalFiles += page.SourceFiles

		switch page.Importance {
		case "high":
			stats.HighImportance++
		case "medium":
			stats.MediumImportance++
		case "low":
			stats.LowImportance++
		}
	}

	index := WikiIndex{
		Title:       structure.Title,
		Description: structure.Description,
		Pages:       indexPages,
		GeneratedAt: time.Now(),
		Version:     structure.Version,
		Language:    options.Language,
		ProjectPath: options.ProjectPath,
		Stats:       stats,
	}

	return jg.writeJSONFile(index, filePath)
}

func (jg *JSONGenerator) writeJSONFile(data interface{}, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0o644)
}
