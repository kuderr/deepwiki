package generator

import (
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatMarkdown          OutputFormat = "markdown"
	FormatJSON              OutputFormat = "json"
	FormatDocusaurus2       OutputFormat = "docusaurus2"
	FormatDocusaurus3       OutputFormat = "docusaurus3"
	FormatSimpleDocusaurus2 OutputFormat = "simple-docusaurus2"
	FormatSimpleDocusaurus3 OutputFormat = "simple-docusaurus3"
)

// OutputOptions contains configuration for output generation
type OutputOptions struct {
	Format      OutputFormat `json:"format"`
	Directory   string       `json:"directory"`
	Language    string       `json:"language"`
	ProjectName string       `json:"projectName"`
	ProjectPath string       `json:"projectPath"`
}

// OutputResult represents the result of output generation
type OutputResult struct {
	OutputDir      string        `json:"outputDir"`
	FilesGenerated []string      `json:"filesGenerated"`
	TotalFiles     int           `json:"totalFiles"`
	TotalSize      int64         `json:"totalSize"`
	GeneratedAt    time.Time     `json:"generatedAt"`
	ProcessingTime time.Duration `json:"processingTime"`
	Errors         []error       `json:"errors,omitempty"`
}

// WikiIndex represents the structure of the wiki index
type WikiIndex struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Pages       []IndexPage            `json:"pages"`
	Sections    map[string][]IndexPage `json:"sections,omitempty"`
	GeneratedAt time.Time              `json:"generatedAt"`
	Version     string                 `json:"version"`
	Language    string                 `json:"language"`
	ProjectPath string                 `json:"projectPath"`
	Stats       IndexStats             `json:"stats"`
}

// IndexPage represents a page in the wiki index
type IndexPage struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	FilePath    string   `json:"filePath"`
	Importance  string   `json:"importance"`
	ParentID    string   `json:"parentId,omitempty"`
	Children    []string `json:"children,omitempty"`
	WordCount   int      `json:"wordCount"`
	SourceFiles int      `json:"sourceFiles"`
}

// IndexStats contains statistics about the generated wiki
type IndexStats struct {
	TotalPages       int `json:"totalPages"`
	TotalWords       int `json:"totalWords"`
	TotalFiles       int `json:"totalFiles"`
	HighImportance   int `json:"highImportance"`
	MediumImportance int `json:"mediumImportance"`
	LowImportance    int `json:"lowImportance"`
}

// NavigationItem represents an item in the navigation structure
type NavigationItem struct {
	ID         string
	Title      string
	FileName   string
	Importance string
	Position   int
}

// FormatGenerator defines the interface for output format generators
type FormatGenerator interface {
	// Generate creates output files in the specific format
	Generate(structure *generator.WikiStructure, pages map[string]*generator.WikiPage, options OutputOptions) (*OutputResult, error)

	// FormatType returns the format type this generator handles
	FormatType() OutputFormat

	// Description returns a human-readable description of the format
	Description() string
}
