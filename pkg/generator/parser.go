package generator

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// XMLParser handles parsing of XML responses from OpenAI
type XMLParser struct{}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{}
}

// ParseWikiStructure parses the XML response containing wiki structure
func (p *XMLParser) ParseWikiStructure(xmlContent string) (*WikiStructureResponse, error) {
	// Extract XML content between <wiki_structure> tags
	xmlContent = p.extractXMLBlock(xmlContent, "wiki_structure")
	if xmlContent == "" {
		return nil, fmt.Errorf("no wiki_structure XML block found in response")
	}

	var response WikiStructureResponse
	if err := xml.Unmarshal([]byte(xmlContent), &response); err != nil {
		return nil, fmt.Errorf("failed to parse wiki structure XML: %w", err)
	}

	// Validate the response
	if err := p.validateWikiStructure(&response); err != nil {
		return nil, fmt.Errorf("invalid wiki structure: %w", err)
	}

	return &response, nil
}

// extractXMLBlock extracts XML content between specified tags
func (p *XMLParser) extractXMLBlock(content, tagName string) string {
	// Create regex pattern to match the XML block (including newlines)
	pattern := fmt.Sprintf(`(?s)<%s[^>]*>(.*?)</%s>`, tagName, tagName)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)

	if len(matches) < 2 {
		return ""
	}

	// Reconstruct the full XML block
	return fmt.Sprintf("<%s>%s</%s>", tagName, matches[1], tagName)
}

// validateWikiStructure validates the parsed wiki structure
func (p *XMLParser) validateWikiStructure(structure *WikiStructureResponse) error {
	if structure.Title == "" {
		return fmt.Errorf("wiki structure must have a title")
	}

	if len(structure.Pages) == 0 {
		return fmt.Errorf("wiki structure must have at least one page")
	}

	// Validate individual pages
	pageIDs := make(map[string]bool)
	for i, page := range structure.Pages {
		if page.ID == "" {
			return fmt.Errorf("page %d must have an ID", i)
		}
		if page.Title == "" {
			return fmt.Errorf("page %d (%s) must have a title", i, page.ID)
		}
		if pageIDs[page.ID] {
			return fmt.Errorf("duplicate page ID: %s", page.ID)
		}
		pageIDs[page.ID] = true

		// Validate importance level
		if page.Importance != "" && !isValidImportance(page.Importance) {
			return fmt.Errorf("page %s has invalid importance level: %s", page.ID, page.Importance)
		}
	}

	// Validate parent-child relationships
	for _, page := range structure.Pages {
		if page.ParentID != "" && !pageIDs[page.ParentID] {
			return fmt.Errorf("page %s references non-existent parent: %s", page.ID, page.ParentID)
		}
	}

	return nil
}

// isValidImportance checks if the importance level is valid
func isValidImportance(importance string) bool {
	validLevels := []string{"high", "medium", "low"}
	importance = strings.ToLower(strings.TrimSpace(importance))
	for _, level := range validLevels {
		if importance == level {
			return true
		}
	}
	return false
}

// ConvertToWikiStructure converts WikiStructureResponse to WikiStructure
func (p *XMLParser) ConvertToWikiStructure(response *WikiStructureResponse, options GenerationOptions) *WikiStructure {
	structure := &WikiStructure{
		ID:          generateID("wiki", options.ProjectName),
		Title:       response.Title,
		Description: response.Description,
		Language:    options.Language,
		ProjectPath: options.ProjectPath,
		Version:     "1.0",
		CreatedAt:   time.Now(),
		Pages:       make([]WikiPage, len(response.Pages)),
	}

	for i, pageReq := range response.Pages {
		structure.Pages[i] = WikiPage{
			ID:          pageReq.ID,
			Title:       pageReq.Title,
			Description: pageReq.Description,
			Importance:  normalizeImportance(pageReq.Importance),
			ParentID:    pageReq.ParentID,
			CreatedAt:   time.Now(),
		}
	}

	return structure
}

// normalizeImportance normalizes importance level to standard values
func normalizeImportance(importance string) string {
	importance = strings.ToLower(strings.TrimSpace(importance))
	switch importance {
	case "high", "critical", "essential":
		return "high"
	case "medium", "normal", "standard":
		return "medium"
	case "low", "optional", "nice-to-have":
		return "low"
	default:
		return "medium" // default to medium if unrecognized
	}
}

// generateID generates a unique ID for wiki components
func generateID(prefix, name string) string {
	// TODO: Enhance ID generation with UUIDs or collision detection for uniqueness
	sanitized := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(name, "-"))
	sanitized = regexp.MustCompile(`-+`).ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	if sanitized == "" {
		sanitized = "unnamed"
	}

	return fmt.Sprintf("%s-%s", prefix, sanitized)
}
