package generator

import (
	"strings"
	"testing"
)

func TestXMLParser_ExtractXMLBlock(t *testing.T) {
	parser := NewXMLParser()

	testCases := []struct {
		name     string
		content  string
		tagName  string
		expected string
	}{
		{
			name:     "simple XML block",
			content:  "Some text <wiki_structure><title>Test</title></wiki_structure> more text",
			tagName:  "wiki_structure",
			expected: "<wiki_structure><title>Test</title></wiki_structure>",
		},
		{
			name:     "XML block with attributes",
			content:  "Text <wiki_structure version=\"1.0\"><title>Test</title></wiki_structure> more",
			tagName:  "wiki_structure",
			expected: "<wiki_structure><title>Test</title></wiki_structure>",
		},
		{
			name:     "no XML block found",
			content:  "Just some regular text without XML",
			tagName:  "wiki_structure",
			expected: "",
		},
		{
			name:     "multiline XML block",
			content:  "Text\n<wiki_structure>\n  <title>Test</title>\n  <description>Desc</description>\n</wiki_structure>\nMore text",
			tagName:  "wiki_structure",
			expected: "<wiki_structure>\n  <title>Test</title>\n  <description>Desc</description>\n</wiki_structure>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractXMLBlock(tc.content, tc.tagName)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestXMLParser_ParseWikiStructure(t *testing.T) {
	parser := NewXMLParser()

	validXML := `
Some OpenAI response text before...

<wiki_structure>
<title>Test Project Wiki</title>
<description>A comprehensive wiki for the test project</description>
<pages>
<page>
<id>overview</id>
<title>Project Overview</title>
<description>Introduction to the project</description>
<importance>high</importance>
</page>
<page>
<id>architecture</id>
<title>System Architecture</title>
<description>Technical architecture overview</description>
<importance>high</importance>
<parent_id>overview</parent_id>
</page>
</pages>
</wiki_structure>

Some more text after...
`

	result, err := parser.ParseWikiStructure(validXML)
	if err != nil {
		t.Fatalf("Failed to parse valid XML: %v", err)
	}

	if result.Title != "Test Project Wiki" {
		t.Errorf("Expected title 'Test Project Wiki', got %q", result.Title)
	}

	if len(result.Pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(result.Pages))
	}

	// Check first page
	page1 := result.Pages[0]
	if page1.ID != "overview" {
		t.Errorf("Expected first page ID 'overview', got %q", page1.ID)
	}
	if page1.Title != "Project Overview" {
		t.Errorf("Expected first page title 'Project Overview', got %q", page1.Title)
	}
	if page1.Importance != "high" {
		t.Errorf("Expected first page importance 'high', got %q", page1.Importance)
	}

	// Check second page
	page2 := result.Pages[1]
	if page2.ParentID != "overview" {
		t.Errorf("Expected second page parent ID 'overview', got %q", page2.ParentID)
	}
}

func TestXMLParser_ParseWikiStructure_InvalidXML(t *testing.T) {
	parser := NewXMLParser()

	testCases := []struct {
		name string
		xml  string
	}{
		{
			name: "no XML block",
			xml:  "Just some regular text without any XML structure",
		},
		{
			name: "malformed XML",
			xml:  "<wiki_structure><title>Test</title><unclosed>",
		},
		{
			name: "missing title",
			xml:  "<wiki_structure><description>Desc</description><pages></pages></wiki_structure>",
		},
		{
			name: "no pages",
			xml:  "<wiki_structure><title>Test</title><description>Desc</description></wiki_structure>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.ParseWikiStructure(tc.xml)
			if err == nil {
				t.Error("Expected error for invalid XML, got nil")
			}
		})
	}
}

func TestXMLParser_ValidateWikiStructure(t *testing.T) {
	parser := NewXMLParser()

	validStructure := &WikiStructureResponse{
		Title:       "Test Wiki",
		Description: "Test description",
		Pages: []WikiPageRequest{
			{
				ID:          "page1",
				Title:       "Page 1",
				Description: "First page",
				Importance:  "high",
			},
			{
				ID:          "page2",
				Title:       "Page 2",
				Description: "Second page",
				Importance:  "medium",
				ParentID:    "page1",
			},
		},
	}

	err := parser.validateWikiStructure(validStructure)
	if err != nil {
		t.Errorf("Expected valid structure to pass validation, got error: %v", err)
	}
}

func TestXMLParser_ValidateWikiStructure_Errors(t *testing.T) {
	parser := NewXMLParser()

	testCases := []struct {
		name      string
		structure *WikiStructureResponse
		expectErr string
	}{
		{
			name: "missing title",
			structure: &WikiStructureResponse{
				Description: "Test description",
				Pages: []WikiPageRequest{
					{ID: "page1", Title: "Page 1"},
				},
			},
			expectErr: "must have a title",
		},
		{
			name: "no pages",
			structure: &WikiStructureResponse{
				Title:       "Test Wiki",
				Description: "Test description",
				Pages:       []WikiPageRequest{},
			},
			expectErr: "must have at least one page",
		},
		{
			name: "page missing ID",
			structure: &WikiStructureResponse{
				Title:       "Test Wiki",
				Description: "Test description",
				Pages: []WikiPageRequest{
					{Title: "Page 1"},
				},
			},
			expectErr: "must have an ID",
		},
		{
			name: "page missing title",
			structure: &WikiStructureResponse{
				Title:       "Test Wiki",
				Description: "Test description",
				Pages: []WikiPageRequest{
					{ID: "page1"},
				},
			},
			expectErr: "must have a title",
		},
		{
			name: "duplicate page ID",
			structure: &WikiStructureResponse{
				Title:       "Test Wiki",
				Description: "Test description",
				Pages: []WikiPageRequest{
					{ID: "page1", Title: "Page 1"},
					{ID: "page1", Title: "Page 2"},
				},
			},
			expectErr: "duplicate page ID",
		},
		{
			name: "invalid importance",
			structure: &WikiStructureResponse{
				Title:       "Test Wiki",
				Description: "Test description",
				Pages: []WikiPageRequest{
					{ID: "page1", Title: "Page 1", Importance: "invalid"},
				},
			},
			expectErr: "invalid importance level",
		},
		{
			name: "non-existent parent",
			structure: &WikiStructureResponse{
				Title:       "Test Wiki",
				Description: "Test description",
				Pages: []WikiPageRequest{
					{ID: "page1", Title: "Page 1", ParentID: "nonexistent"},
				},
			},
			expectErr: "references non-existent parent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := parser.validateWikiStructure(tc.structure)
			if err == nil {
				t.Error("Expected validation error, got nil")
			} else if !strings.Contains(err.Error(), tc.expectErr) {
				t.Errorf("Expected error containing %q, got %q", tc.expectErr, err.Error())
			}
		})
	}
}

func TestIsValidImportance(t *testing.T) {
	testCases := []struct {
		importance string
		expected   bool
	}{
		{"high", true},
		{"medium", true},
		{"low", true},
		{"HIGH", true},       // case insensitive
		{"  medium  ", true}, // whitespace trimmed
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isValidImportance(tc.importance)
		if result != tc.expected {
			t.Errorf("isValidImportance(%q) = %v, expected %v", tc.importance, result, tc.expected)
		}
	}
}

func TestNormalizeImportance(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"high", "high"},
		{"HIGH", "high"},
		{"critical", "high"},
		{"essential", "high"},
		{"medium", "medium"},
		{"normal", "medium"},
		{"standard", "medium"},
		{"low", "low"},
		{"optional", "low"},
		{"nice-to-have", "low"},
		{"unknown", "medium"}, // default
		{"", "medium"},        // default
	}

	for _, tc := range testCases {
		result := normalizeImportance(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeImportance(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestGenerateID(t *testing.T) {
	testCases := []struct {
		prefix   string
		name     string
		expected string
	}{
		{"wiki", "My Project", "wiki-my-project"},
		{"page", "System Architecture", "page-system-architecture"},
		{"wiki", "test@#$%project", "wiki-test-project"},
		{"page", "  multiple   spaces  ", "page-multiple-spaces"},
		{"wiki", "", "wiki-unnamed"},
		{"test", "---", "test-unnamed"},
	}

	for _, tc := range testCases {
		result := generateID(tc.prefix, tc.name)
		if result != tc.expected {
			t.Errorf("generateID(%q, %q) = %q, expected %q", tc.prefix, tc.name, result, tc.expected)
		}
	}
}

func TestConvertToWikiStructure(t *testing.T) {
	parser := NewXMLParser()

	response := &WikiStructureResponse{
		Title:       "Test Project Wiki",
		Description: "A test wiki",
		Pages: []WikiPageRequest{
			{
				ID:          "overview",
				Title:       "Overview",
				Description: "Project overview",
				Importance:  "high",
			},
			{
				ID:          "architecture",
				Title:       "Architecture",
				Description: "System architecture",
				Importance:  "medium",
				ParentID:    "overview",
			},
		},
	}

	options := GenerationOptions{
		ProjectName: "test-project",
		ProjectPath: "/test/path",
		Language:    "en",
	}

	structure := parser.ConvertToWikiStructure(response, options)

	if structure.Title != "Test Project Wiki" {
		t.Errorf("Expected title 'Test Project Wiki', got %q", structure.Title)
	}

	if structure.Language != "en" {
		t.Errorf("Expected language 'en', got %q", structure.Language)
	}

	if len(structure.Pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(structure.Pages))
	}

	// Check first page conversion
	page1 := structure.Pages[0]
	if page1.ID != "overview" {
		t.Errorf("Expected page ID 'overview', got %q", page1.ID)
	}
	if page1.Importance != "high" {
		t.Errorf("Expected page importance 'high', got %q", page1.Importance)
	}

	// Check second page conversion
	page2 := structure.Pages[1]
	if page2.ParentID != "overview" {
		t.Errorf("Expected page parent ID 'overview', got %q", page2.ParentID)
	}
	if page2.Importance != "medium" {
		t.Errorf("Expected page importance 'medium', got %q", page2.Importance)
	}
}
