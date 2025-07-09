package generator

import (
	"fmt"
	"regexp"
	"strings"
)

// ContentProcessor handles cleaning and formatting of generated content
type ContentProcessor struct{}

// NewContentProcessor creates a new content processor
func NewContentProcessor() *ContentProcessor {
	return &ContentProcessor{}
}

// CleanMarkdown cleans and formats markdown content
func (cp *ContentProcessor) CleanMarkdown(content string) string {
	// Remove think blocks first
	content = cp.cleanThinkBlocks(content)

	// Remove markdown code block wrappers
	content = cp.cleanMarkdownCodeBlocks(content)

	// Remove empty code blocks
	content = cp.cleanEmptyCodeBlocks(content)

	// Remove excessive whitespace
	content = cp.normalizeWhitespace(content)

	// Fix heading formatting
	content = cp.fixHeadings(content)

	// Clean up code blocks
	content = cp.cleanCodeBlocks(content)

	// Fix list formatting
	content = cp.fixLists(content)

	// Clean up tables
	content = cp.cleanTables(content)

	// Remove duplicate empty lines
	content = cp.removeDuplicateEmptyLines(content)

	return strings.TrimSpace(content)
}

// ValidateMermaidDiagrams validates Mermaid diagram syntax
func (cp *ContentProcessor) ValidateMermaidDiagrams(content string) []string {
	var errors []string

	// Find all mermaid code blocks
	mermaidRegex := regexp.MustCompile("(?s)```mermaid\\s*\\n(.*?)\\n```")
	matches := mermaidRegex.FindAllStringSubmatch(content, -1)

	for i, match := range matches {
		if len(match) > 1 {
			diagramContent := strings.TrimSpace(match[1])
			if err := cp.validateMermaidSyntax(diagramContent); err != nil {
				errors = append(errors, fmt.Sprintf("Mermaid diagram %d: %s", i+1, err))
			}
		}
	}

	return errors
}

// ExtractSourceCitations extracts and validates source citations
func (cp *ContentProcessor) ExtractSourceCitations(content string) []SourceCitation {
	var citations []SourceCitation

	// Regex to match [filename.ext:line_numbers](#) format
	citationRegex := regexp.MustCompile(`\[([^:\]]+):([^\]]*)\]\(\)`)
	matches := citationRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			citation := SourceCitation{
				FileName:    match[1],
				LineNumbers: match[2],
				IsValid:     cp.validateCitation(match[1], match[2]),
			}
			citations = append(citations, citation)
		}
	}

	return citations
}

// FormatContent formats content according to style guidelines
func (cp *ContentProcessor) FormatContent(content string, options ContentFormatOptions) string {
	if options.RemoveEmptyLines {
		content = cp.removeDuplicateEmptyLines(content)
	}

	if options.FixHeadings {
		content = cp.fixHeadings(content)
	}

	if options.StandardizeCodeBlocks {
		content = cp.standardizeCodeBlocks(content)
	}

	if options.EnforceLineLength > 0 {
		content = cp.enforceLineLength(content, options.EnforceLineLength)
	}

	return content
}

// normalizeWhitespace removes excessive whitespace
func (cp *ContentProcessor) normalizeWhitespace(content string) string {
	// Replace multiple spaces with single space (except in code blocks)
	lines := strings.Split(content, "\n")
	var result []string

	inCodeBlock := false

	for _, line := range lines {
		// Check if we're entering or leaving a code block
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		// Don't normalize whitespace inside code blocks
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Normalize whitespace for regular text
		normalized := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(line), " ")
		result = append(result, normalized)
	}

	return strings.Join(result, "\n")
}

// fixHeadings ensures proper heading formatting
func (cp *ContentProcessor) fixHeadings(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Fix heading spacing (ensure space after #)
		if strings.HasPrefix(trimmed, "#") {
			// Count the number of # characters
			level := 0
			for _, char := range trimmed {
				if char == '#' {
					level++
				} else {
					break
				}
			}

			// Extract heading text
			headingText := strings.TrimSpace(trimmed[level:])
			if headingText != "" {
				result = append(result, strings.Repeat("#", level)+" "+headingText)
			} else {
				result = append(result, line) // Keep original if malformed
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// cleanCodeBlocks ensures proper code block formatting
func (cp *ContentProcessor) cleanCodeBlocks(content string) string {
	// Fix code block language tags
	codeBlockRegex := regexp.MustCompile("(?m)^```([a-zA-Z0-9]*)")
	content = codeBlockRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := strings.TrimPrefix(match, "```")
		if parts == "" {
			return "```"
		}
		return "```" + strings.ToLower(parts)
	})

	return content
}

// fixLists ensures proper list formatting
func (cp *ContentProcessor) fixLists(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Fix unordered lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			// Ensure consistent bullet style (use -)
			text := strings.TrimSpace(trimmed[2:])
			result = append(result, "- "+text)
		} else if regexp.MustCompile(`^\d+\.\s`).MatchString(trimmed) {
			// Ordered list - ensure space after period
			parts := regexp.MustCompile(`^(\d+\.)(.*)$`).FindStringSubmatch(trimmed)
			if len(parts) == 3 {
				text := strings.TrimSpace(parts[2])
				result = append(result, parts[1]+" "+text)
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// cleanTables ensures proper table formatting
func (cp *ContentProcessor) cleanTables(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this line looks like a table row
		if strings.Contains(trimmed, "|") && len(strings.Split(trimmed, "|")) >= 3 {
			if !inTable {
				inTable = true
			}

			// Clean up table row
			cells := strings.Split(trimmed, "|")
			var cleanCells []string

			for _, cell := range cells {
				cleanCells = append(cleanCells, strings.TrimSpace(cell))
			}

			result = append(result, strings.Join(cleanCells, " | "))
		} else {
			if inTable {
				inTable = false
			}
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// removeDuplicateEmptyLines removes multiple consecutive empty lines
func (cp *ContentProcessor) removeDuplicateEmptyLines(content string) string {
	// Replace multiple consecutive empty lines with single empty line
	multipleEmptyLines := regexp.MustCompile(`\n\s*\n\s*\n+`)
	return multipleEmptyLines.ReplaceAllString(content, "\n\n")
}

// cleanThinkBlocks removes <think></think> blocks with their content
func (cp *ContentProcessor) cleanThinkBlocks(content string) string {
	// Remove <think></think> blocks including their content
	thinkBlockRegex := regexp.MustCompile(`(?s)<think>.*?</think>`)
	return thinkBlockRegex.ReplaceAllString(content, "")
}

// cleanEmptyCodeBlocks removes empty ``` blocks
func (cp *ContentProcessor) cleanEmptyCodeBlocks(content string) string {
	// Remove empty code blocks (``` followed by optional whitespace and closing ```)
	emptyCodeBlockRegex := regexp.MustCompile("(?m)^\\s*```\\s*\n\\s*```\\s*$")
	return emptyCodeBlockRegex.ReplaceAllString(content, "")
}

// cleanMarkdownCodeBlocks removes ````markdown start and end lines but keeps content
func (cp *ContentProcessor) cleanMarkdownCodeBlocks(content string) string {
	// Remove ````markdown wrappers but keep the content inside
	markdownBlockRegex := regexp.MustCompile("(?s)````markdown\\s*\n(.*?)\n````")
	return markdownBlockRegex.ReplaceAllString(content, "$1")
}

// validateMermaidSyntax performs basic validation of Mermaid diagram syntax
func (cp *ContentProcessor) validateMermaidSyntax(diagram string) error {
	trimmed := strings.TrimSpace(diagram)

	if trimmed == "" {
		return fmt.Errorf("empty diagram")
	}

	// Check for valid diagram types
	validTypes := []string{
		"flowchart", "graph", "sequenceDiagram", "classDiagram",
		"stateDiagram", "erDiagram", "gantt", "pie", "gitgraph",
	}

	firstLine := strings.Split(trimmed, "\n")[0]
	firstWord := strings.Fields(firstLine)[0]

	for _, validType := range validTypes {
		if strings.HasPrefix(firstWord, validType) {
			return nil // Valid diagram type found
		}
	}

	return fmt.Errorf("unknown diagram type: %s", firstWord)
}

// validateCitation validates a source citation
func (cp *ContentProcessor) validateCitation(filename, lineNumbers string) bool {
	// Basic validation
	if filename == "" {
		return false
	}

	// Check if line numbers format is valid (if provided)
	if lineNumbers != "" {
		// Should be either "123" or "123-456"
		lineNumRegex := regexp.MustCompile(`^\d+(-\d+)?$`)
		return lineNumRegex.MatchString(lineNumbers)
	}

	return true
}

// standardizeCodeBlocks ensures consistent code block formatting
func (cp *ContentProcessor) standardizeCodeBlocks(content string) string {
	// Ensure code blocks have proper language tags
	codeBlockRegex := regexp.MustCompile("(?m)^```\\s*$")
	content = codeBlockRegex.ReplaceAllString(content, "```text")

	return content
}

// enforceLineLength wraps long lines (excluding code blocks)
func (cp *ContentProcessor) enforceLineLength(content string, maxLength int) string {
	lines := strings.Split(content, "\n")
	var result []string

	inCodeBlock := false

	for _, line := range lines {
		// Check if we're entering or leaving a code block
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		// Don't wrap lines inside code blocks
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Wrap long lines
		if len(line) > maxLength {
			wrapped := cp.wrapLine(line, maxLength)
			result = append(result, wrapped...)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// wrapLine wraps a single line at word boundaries
func (cp *ContentProcessor) wrapLine(line string, maxLength int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}

	var result []string
	var currentLine strings.Builder

	for _, word := range words {
		// Check if adding this word would exceed the limit
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > maxLength && currentLine.Len() > 0 {
			// Start a new line
			result = append(result, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	// Add the last line
	if currentLine.Len() > 0 {
		result = append(result, currentLine.String())
	}

	return result
}

// SourceCitation represents a source code citation
type SourceCitation struct {
	FileName    string `json:"fileName"`
	LineNumbers string `json:"lineNumbers"`
	IsValid     bool   `json:"isValid"`
}

// ContentFormatOptions contains options for content formatting
type ContentFormatOptions struct {
	RemoveEmptyLines      bool `json:"removeEmptyLines"`
	FixHeadings           bool `json:"fixHeadings"`
	StandardizeCodeBlocks bool `json:"standardizeCodeBlocks"`
	EnforceLineLength     int  `json:"enforceLineLength"` // 0 = no limit
}
