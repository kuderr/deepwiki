package scanner

import (
	"time"
)

// FileInfo represents information about a scanned file
type FileInfo struct {
	// Basic file information
	Path         string    `json:"path"`         // Relative path from scan root
	AbsolutePath string    `json:"absolutePath"` // Absolute path
	Name         string    `json:"name"`         // File name with extension
	Extension    string    `json:"extension"`    // File extension (including dot)
	Size         int64     `json:"size"`         // File size in bytes
	ModTime      time.Time `json:"modTime"`      // Last modification time
	IsDir        bool      `json:"isDir"`        // Whether this is a directory

	// Content information
	IsText    bool   `json:"isText"`    // Whether file appears to be text
	IsBinary  bool   `json:"isBinary"`  // Whether file appears to be binary
	Language  string `json:"language"`  // Detected programming language
	LineCount int    `json:"lineCount"` // Number of lines (for text files)

	// Metadata
	Category   string `json:"category"`   // File category (code, docs, config, etc.)
	Importance int    `json:"importance"` // Importance score (1-5)
}

// ScanResult represents the result of a directory scan
type ScanResult struct {
	RootPath      string        `json:"rootPath"`      // Root path that was scanned
	TotalFiles    int           `json:"totalFiles"`    // Total number of files found
	TotalDirs     int           `json:"totalDirs"`     // Total number of directories found
	FilteredFiles int           `json:"filteredFiles"` // Number of files after filtering
	Files         []FileInfo    `json:"files"`         // List of scanned files
	Errors        []string      `json:"errors"`        // Any errors encountered during scan
	ScanTime      time.Duration `json:"scanTime"`      // Time taken to complete scan
}

// ScanOptions represents options for directory scanning
type ScanOptions struct {
	// Filtering options
	IncludeExtensions []string `json:"includeExtensions"` // Extensions to include
	ExcludeDirs       []string `json:"excludeDirs"`       // Directories to exclude
	ExcludeFiles      []string `json:"excludeFiles"`      // File patterns to exclude

	// Scanning options
	FollowSymlinks bool `json:"followSymlinks"` // Whether to follow symbolic links
	MaxDepth       int  `json:"maxDepth"`       // Maximum directory depth (0 = unlimited)
	MaxFiles       int  `json:"maxFiles"`       // Maximum number of files to process

	// Content analysis options
	AnalyzeContent  bool  `json:"analyzeContent"`  // Whether to analyze file content
	MaxFileSize     int64 `json:"maxFileSize"`     // Maximum file size to analyze (bytes)
	SkipBinaryFiles bool  `json:"skipBinaryFiles"` // Whether to skip binary files

	// Performance options
	Concurrent bool `json:"concurrent"` // Whether to use concurrent processing
	MaxWorkers int  `json:"maxWorkers"` // Maximum number of worker goroutines
}

// DefaultScanOptions returns default scanning options
func DefaultScanOptions() *ScanOptions {
	return &ScanOptions{
		IncludeExtensions: []string{
			// Code files
			".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".hpp",
			".rs", ".jsx", ".tsx", ".php", ".swift", ".cs", ".rb", ".kt",
			".scala", ".clj", ".hs", ".ml", ".fs", ".elm", ".dart", ".jl",

			// Web files
			".html", ".css", ".scss", ".sass", ".less", ".vue", ".svelte",

			// Config/Data files
			".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf",
			".xml", ".env", ".properties",

			// Documentation files
			".md", ".txt", ".rst", ".adoc", ".org",

			// Shell scripts
			".sh", ".bash", ".zsh", ".fish", ".ps1", ".bat", ".cmd",

			// Build files
			".makefile", ".dockerfile", ".gradle", ".maven",
		},
		ExcludeDirs: []string{
			// Version control
			".git", ".hg", ".svn", ".bzr",

			// Dependencies
			"node_modules", "vendor", ".cargo", "target",

			// Build outputs
			"dist", "build", "out", "bin", "obj", ".next", ".nuxt",

			// Caches
			".cache", "__pycache__", ".pytest_cache", ".mypy_cache",

			// Virtual environments
			".venv", "venv", ".virtualenv", "env",

			// IDE files
			".vscode", ".idea", ".vs", "*.xcworkspace", "*.xcodeproj",

			// Temporary files
			"tmp", "temp", ".tmp", ".temp",

			// Logs
			"logs", "log",
		},
		ExcludeFiles: []string{
			// Compiled files
			"*.exe", "*.dll", "*.so", "*.dylib", "*.a", "*.o", "*.obj",
			"*.class", "*.pyc", "*.pyo", "*.pyd",

			// Archives
			"*.zip", "*.tar", "*.gz", "*.bz2", "*.xz", "*.7z", "*.rar",

			// Lock files
			"package-lock.json", "yarn.lock", "composer.lock", "Pipfile.lock",
			"poetry.lock", "Cargo.lock", "go.sum",

			// Minified files
			"*.min.js", "*.min.css",

			// Large data files
			"*.sql", "*.db", "*.sqlite", "*.sqlite3",

			// Media files
			"*.png", "*.jpg", "*.jpeg", "*.gif", "*.svg", "*.ico",
			"*.mp3", "*.mp4", "*.avi", "*.mov", "*.wmv",
			"*.pdf", "*.doc", "*.docx", "*.xls", "*.xlsx", "*.ppt", "*.pptx",

			// OS files
			".DS_Store", "Thumbs.db", "desktop.ini",
		},
		FollowSymlinks:  false,
		MaxDepth:        0, // unlimited
		MaxFiles:        10000,
		AnalyzeContent:  true,
		MaxFileSize:     1024 * 1024, // 1MB
		SkipBinaryFiles: true,
		Concurrent:      true,
		MaxWorkers:      4,
	}
}

// FileCategory represents different categories of files
type FileCategory string

const (
	CategoryCode    FileCategory = "code"
	CategoryTest    FileCategory = "test"
	CategoryConfig  FileCategory = "config"
	CategoryDocs    FileCategory = "docs"
	CategoryBuild   FileCategory = "build"
	CategoryData    FileCategory = "data"
	CategoryAssets  FileCategory = "assets"
	CategoryUnknown FileCategory = "unknown"
)

// LanguageInfo represents information about a programming language
type LanguageInfo struct {
	Name       string       `json:"name"`
	Extensions []string     `json:"extensions"`
	Category   FileCategory `json:"category"`
	Importance int          `json:"importance"` // 1-5 scale
}

// GetLanguageByExtension returns language information for a file extension
func GetLanguageByExtension(ext string) *LanguageInfo {
	languages := map[string]*LanguageInfo{
		// High importance languages
		".go":   {Name: "Go", Extensions: []string{".go"}, Category: CategoryCode, Importance: 5},
		".py":   {Name: "Python", Extensions: []string{".py"}, Category: CategoryCode, Importance: 5},
		".js":   {Name: "JavaScript", Extensions: []string{".js"}, Category: CategoryCode, Importance: 5},
		".ts":   {Name: "TypeScript", Extensions: []string{".ts"}, Category: CategoryCode, Importance: 5},
		".java": {Name: "Java", Extensions: []string{".java"}, Category: CategoryCode, Importance: 5},
		".cpp":  {Name: "C++", Extensions: []string{".cpp", ".cxx", ".cc"}, Category: CategoryCode, Importance: 5},
		".c":    {Name: "C", Extensions: []string{".c"}, Category: CategoryCode, Importance: 5},
		".rs":   {Name: "Rust", Extensions: []string{".rs"}, Category: CategoryCode, Importance: 5},

		// Medium-high importance
		".jsx":   {Name: "React", Extensions: []string{".jsx"}, Category: CategoryCode, Importance: 4},
		".tsx":   {Name: "TypeScript React", Extensions: []string{".tsx"}, Category: CategoryCode, Importance: 4},
		".php":   {Name: "PHP", Extensions: []string{".php"}, Category: CategoryCode, Importance: 4},
		".cs":    {Name: "C#", Extensions: []string{".cs"}, Category: CategoryCode, Importance: 4},
		".rb":    {Name: "Ruby", Extensions: []string{".rb"}, Category: CategoryCode, Importance: 4},
		".swift": {Name: "Swift", Extensions: []string{".swift"}, Category: CategoryCode, Importance: 4},
		".kt":    {Name: "Kotlin", Extensions: []string{".kt"}, Category: CategoryCode, Importance: 4},

		// Documentation (high importance)
		".md":  {Name: "Markdown", Extensions: []string{".md"}, Category: CategoryDocs, Importance: 5},
		".rst": {Name: "reStructuredText", Extensions: []string{".rst"}, Category: CategoryDocs, Importance: 4},
		".txt": {Name: "Text", Extensions: []string{".txt"}, Category: CategoryDocs, Importance: 3},

		// Configuration (medium importance)
		".json": {Name: "JSON", Extensions: []string{".json"}, Category: CategoryConfig, Importance: 3},
		".yaml": {Name: "YAML", Extensions: []string{".yaml", ".yml"}, Category: CategoryConfig, Importance: 3},
		".yml":  {Name: "YAML", Extensions: []string{".yaml", ".yml"}, Category: CategoryConfig, Importance: 3},
		".toml": {Name: "TOML", Extensions: []string{".toml"}, Category: CategoryConfig, Importance: 3},
		".xml":  {Name: "XML", Extensions: []string{".xml"}, Category: CategoryConfig, Importance: 2},

		// Web files
		".html": {Name: "HTML", Extensions: []string{".html", ".htm"}, Category: CategoryCode, Importance: 3},
		".css":  {Name: "CSS", Extensions: []string{".css"}, Category: CategoryCode, Importance: 3},
		".scss": {Name: "Sass", Extensions: []string{".scss"}, Category: CategoryCode, Importance: 3},

		// Build files
		".dockerfile": {
			Name:       "Dockerfile",
			Extensions: []string{".dockerfile"},
			Category:   CategoryBuild,
			Importance: 3,
		},
		".makefile": {Name: "Makefile", Extensions: []string{".makefile"}, Category: CategoryBuild, Importance: 3},

		// Shell scripts
		".sh":   {Name: "Shell", Extensions: []string{".sh"}, Category: CategoryCode, Importance: 3},
		".bash": {Name: "Bash", Extensions: []string{".bash"}, Category: CategoryCode, Importance: 3},
		".ps1":  {Name: "PowerShell", Extensions: []string{".ps1"}, Category: CategoryCode, Importance: 3},

		// Test files (detected by naming pattern)
		"_test.go": {Name: "Go Test", Extensions: []string{"_test.go"}, Category: CategoryTest, Importance: 2},
		"_test.py": {Name: "Python Test", Extensions: []string{"_test.py"}, Category: CategoryTest, Importance: 2},
		".test.js": {Name: "JavaScript Test", Extensions: []string{".test.js"}, Category: CategoryTest, Importance: 2},
		".spec.js": {Name: "JavaScript Spec", Extensions: []string{".spec.js"}, Category: CategoryTest, Importance: 2},
	}

	if lang, exists := languages[ext]; exists {
		return lang
	}

	return &LanguageInfo{
		Name:       "Unknown",
		Extensions: []string{ext},
		Category:   CategoryUnknown,
		Importance: 1,
	}
}
