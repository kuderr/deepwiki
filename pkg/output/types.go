package output

import (
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
	outputgen "github.com/deepwiki-cli/deepwiki-cli/pkg/output/generator"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatMarkdown    OutputFormat = "markdown"
	FormatJSON        OutputFormat = "json"
	FormatDocusaurus2 OutputFormat = "docusaurus2"
	FormatDocusaurus3 OutputFormat = "docusaurus3"
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

// PageFile represents a generated page file
type PageFile struct {
	ID       string `json:"id"`
	FilePath string `json:"filePath"`
	Title    string `json:"title"`
	Size     int64  `json:"size"`
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

// Manager interface defines the contract for output generation
type Manager interface {
	GenerateOutput(
		structure *generator.WikiStructure,
		pages map[string]*generator.WikiPage,
		options outputgen.OutputOptions,
	) (*outputgen.OutputResult, error)
}
