package generator

import (
	"time"
)

// WikiStructure represents the overall structure of a wiki
type WikiStructure struct {
	ID          string     `json:"id"          xml:"id"`
	Title       string     `json:"title"       xml:"title"`
	Description string     `json:"description" xml:"description"`
	Pages       []WikiPage `json:"pages"       xml:"pages>page"`
	CreatedAt   time.Time  `json:"createdAt"   xml:"createdAt"`
	Language    string     `json:"language"    xml:"language"`
	ProjectPath string     `json:"projectPath" xml:"projectPath"`
	Version     string     `json:"version"     xml:"version"`
}

// WikiPage represents a single wiki page
type WikiPage struct {
	ID           string    `json:"id"                     xml:"id"`
	Title        string    `json:"title"                  xml:"title"`
	Description  string    `json:"description"            xml:"description"`
	Content      string    `json:"content"                xml:"content"`
	FilePaths    []string  `json:"filePaths"              xml:"filePaths>path"`
	Importance   string    `json:"importance"             xml:"importance"` // high, medium, low
	ParentID     string    `json:"parentId,omitempty"     xml:"parentId,omitempty"`
	RelatedPages []string  `json:"relatedPages,omitempty" xml:"relatedPages>pageId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"              xml:"createdAt"`
	WordCount    int       `json:"wordCount"              xml:"wordCount"`
	SourceFiles  int       `json:"sourceFiles"            xml:"sourceFiles"`
}

// GenerationOptions contains options for wiki generation
type GenerationOptions struct {
	ProjectName     string
	ProjectPath     string
	Language        string
	OutputFormat    string
	MaxConcurrency  int
	ProgressTracker ProgressTracker
}

// GenerationResult represents the result of wiki generation
type GenerationResult struct {
	Structure      *WikiStructure
	Pages          map[string]*WikiPage
	GeneratedAt    time.Time
	TotalPages     int
	TotalWords     int
	ProcessingTime time.Duration
	Errors         []error
}

// ProgressTracker interface for tracking generation progress
type ProgressTracker interface {
	StartTask(taskName string, total int)
	UpdateProgress(current int, message string)
	CompleteTask(message string)
	SetError(err error)
}

// NoOpProgressTracker is a no-op implementation of ProgressTracker
type NoOpProgressTracker struct{}

func (n *NoOpProgressTracker) StartTask(taskName string, total int)       {}
func (n *NoOpProgressTracker) UpdateProgress(current int, message string) {}
func (n *NoOpProgressTracker) CompleteTask(message string)                {}
func (n *NoOpProgressTracker) SetError(err error)                         {}

// FileReference represents a reference to a source file with line numbers
type FileReference struct {
	FilePath    string `json:"filePath"`
	StartLine   int    `json:"startLine,omitempty"`
	EndLine     int    `json:"endLine,omitempty"`
	Description string `json:"description,omitempty"`
}

// WikiStructureResponse represents the XML response for wiki structure
type WikiStructureResponse struct {
	Title       string            `xml:"title"`
	Description string            `xml:"description"`
	Pages       []WikiPageRequest `xml:"pages>page"`
}

// WikiPageRequest represents a page in the structure generation request
type WikiPageRequest struct {
	ID          string `xml:"id"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Importance  string `xml:"importance"`
	ParentID    string `xml:"parent_id,omitempty"`
}

// GenerationStats tracks statistics during generation
type GenerationStats struct {
	FilesProcessed    int
	PagesGenerated    int
	TotalTokensUsed   int
	EmbeddingsCreated int
	APICallsMade      int
	StartTime         time.Time
	EndTime           time.Time
}

// Add returns a new GenerationStats with added values
func (gs *GenerationStats) Add(other GenerationStats) GenerationStats {
	return GenerationStats{
		FilesProcessed:    gs.FilesProcessed + other.FilesProcessed,
		PagesGenerated:    gs.PagesGenerated + other.PagesGenerated,
		TotalTokensUsed:   gs.TotalTokensUsed + other.TotalTokensUsed,
		EmbeddingsCreated: gs.EmbeddingsCreated + other.EmbeddingsCreated,
		APICallsMade:      gs.APICallsMade + other.APICallsMade,
		StartTime:         gs.StartTime,
		EndTime:           other.EndTime,
	}
}

// Duration returns the total processing duration
func (gs *GenerationStats) Duration() time.Duration {
	if gs.EndTime.IsZero() {
		return time.Since(gs.StartTime)
	}
	return gs.EndTime.Sub(gs.StartTime)
}
