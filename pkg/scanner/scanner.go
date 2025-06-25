package scanner

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/deepwiki-cli/deepwiki-cli/internal/logging"
)

// Scanner handles directory scanning and file analysis
type Scanner struct {
	options *ScanOptions
	stats   ScanStats
	mutex   sync.RWMutex
	logger  *logging.Logger
}

// ScanStats tracks scanning statistics
type ScanStats struct {
	FilesProcessed int
	DirsProcessed  int
	FilesFiltered  int
	Errors         []string
	StartTime      time.Time
	EndTime        time.Time
}

// NewScanner creates a new scanner with the given options
func NewScanner(options *ScanOptions) *Scanner {
	if options == nil {
		options = DefaultScanOptions()
	}

	return &Scanner{
		options: options,
		stats:   ScanStats{},
		logger:  logging.GetGlobalLogger().WithComponent("scanner"),
	}
}

// ScanDirectory scans a directory and returns information about all relevant files
func (s *Scanner) ScanDirectory(rootPath string) (*ScanResult, error) {
	s.mutex.Lock()
	s.stats = ScanStats{
		StartTime: time.Now(),
	}
	s.mutex.Unlock()

	// Convert to absolute path
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify the path exists and is a directory
	info, err := os.Stat(absRoot)
	if err != nil {
		s.logger.LogError(context.Background(), "failed to access directory", err, slog.String("path", absRoot))
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		s.logger.ErrorContext(context.Background(), "path is not a directory", slog.String("path", absRoot))
		return nil, fmt.Errorf("path is not a directory: %s", absRoot)
	}

	s.logger.InfoContext(context.Background(), "starting directory scan",
		slog.String("path", absRoot),
		slog.Bool("concurrent", s.options.Concurrent),
		slog.Int("max_files", s.options.MaxFiles),
		slog.Int("max_depth", s.options.MaxDepth),
	)

	var files []FileInfo
	var allErrors []string

	if s.options.Concurrent {
		files, allErrors = s.scanConcurrent(absRoot)
	} else {
		files, allErrors = s.scanSequential(absRoot)
	}

	s.mutex.Lock()
	s.stats.EndTime = time.Now()
	scanTime := s.stats.EndTime.Sub(s.stats.StartTime)
	s.mutex.Unlock()

	result := &ScanResult{
		RootPath:      absRoot,
		TotalFiles:    s.stats.FilesProcessed,
		TotalDirs:     s.stats.DirsProcessed,
		FilteredFiles: len(files),
		Files:         files,
		Errors:        allErrors,
		ScanTime:      scanTime,
	}

	return result, nil
}

// scanSequential performs sequential directory scanning
func (s *Scanner) scanSequential(rootPath string) ([]FileInfo, []string) {
	var files []FileInfo
	var errors []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, fmt.Sprintf("error accessing %s: %v", path, err))
			return nil // Continue walking
		}

		// Handle maximum depth
		if s.options.MaxDepth > 0 {
			relPath, _ := filepath.Rel(rootPath, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if depth > s.options.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Handle symlinks
		if !s.options.FollowSymlinks && isSymlink(info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			s.mutex.Lock()
			s.stats.DirsProcessed++
			s.mutex.Unlock()

			// Check if directory should be excluded
			if s.shouldExcludeDir(path, rootPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Process file
		fileInfo, shouldInclude, err := s.processFile(path, rootPath, info)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error processing %s: %v", path, err))
			return nil
		}

		s.mutex.Lock()
		s.stats.FilesProcessed++
		if shouldInclude {
			files = append(files, *fileInfo)
		} else {
			s.stats.FilesFiltered++
		}

		// Check max files limit
		if s.options.MaxFiles > 0 && len(files) >= s.options.MaxFiles {
			s.mutex.Unlock()
			return fmt.Errorf("max files limit reached")
		}
		s.mutex.Unlock()

		return nil
	})

	if err != nil && !strings.Contains(err.Error(), "max files limit") {
		errors = append(errors, fmt.Sprintf("walk error: %v", err))
	}

	return files, errors
}

// scanConcurrent performs concurrent directory scanning
func (s *Scanner) scanConcurrent(rootPath string) ([]FileInfo, []string) {
	// For now, implement a simple concurrent version
	// In a production implementation, you might want a more sophisticated worker pool
	return s.scanSequential(rootPath) // Fallback to sequential for now
}

// processFile processes a single file and returns its information
func (s *Scanner) processFile(path, rootPath string, info os.FileInfo) (*FileInfo, bool, error) {
	// Get relative path
	relPath, err := filepath.Rel(rootPath, path)
	if err != nil {
		return nil, false, err
	}

	// Basic file info
	fileInfo := &FileInfo{
		Path:         relPath,
		AbsolutePath: path,
		Name:         info.Name(),
		Extension:    strings.ToLower(filepath.Ext(info.Name())),
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		IsDir:        info.IsDir(),
	}

	// Check if file should be excluded
	if s.shouldExcludeFile(fileInfo) {
		return fileInfo, false, nil
	}

	// Check file size limit
	if s.options.MaxFileSize > 0 && fileInfo.Size > s.options.MaxFileSize {
		return fileInfo, false, nil
	}

	// Analyze content if requested
	if s.options.AnalyzeContent {
		if err := s.analyzeFile(fileInfo); err != nil {
			// Don't fail completely, just log the error
			return fileInfo, true, nil
		}
	}

	// Set language and category information
	lang := GetLanguageByExtension(fileInfo.Extension)
	fileInfo.Language = lang.Name
	fileInfo.Category = string(lang.Category)
	fileInfo.Importance = lang.Importance

	// Special handling for test files
	if isTestFile(fileInfo.Name) {
		fileInfo.Category = string(CategoryTest)
		fileInfo.Importance = 2
	}

	return fileInfo, true, nil
}

// analyzeFile analyzes the content of a file
func (s *Scanner) analyzeFile(fileInfo *FileInfo) error {
	// Skip binary files if requested
	if s.options.SkipBinaryFiles && isBinaryFile(fileInfo.AbsolutePath) {
		fileInfo.IsBinary = true
		return nil
	}

	// Try to read the file
	file, err := os.Open(fileInfo.AbsolutePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read first few bytes to determine if it's text
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return err
	}

	// Check if it's valid UTF-8 text
	fileInfo.IsText = utf8.Valid(buffer[:n])
	fileInfo.IsBinary = !fileInfo.IsText

	if fileInfo.IsText {
		// Count lines
		file.Seek(0, 0) // Reset to beginning
		scanner := bufio.NewScanner(file)
		lineCount := 0
		for scanner.Scan() {
			lineCount++
		}
		fileInfo.LineCount = lineCount
	}

	return nil
}

// shouldExcludeDir determines if a directory should be excluded
func (s *Scanner) shouldExcludeDir(dirPath, rootPath string) bool {
	relPath, _ := filepath.Rel(rootPath, dirPath)
	dirName := filepath.Base(dirPath)

	for _, excludePattern := range s.options.ExcludeDirs {
		if matched, _ := filepath.Match(excludePattern, dirName); matched {
			return true
		}
		if matched, _ := filepath.Match(excludePattern, relPath); matched {
			return true
		}
		// Also check if the pattern is a substring
		if strings.Contains(relPath, excludePattern) {
			return true
		}
	}

	return false
}

// shouldExcludeFile determines if a file should be excluded
func (s *Scanner) shouldExcludeFile(fileInfo *FileInfo) bool {
	// Check extension inclusion list
	if len(s.options.IncludeExtensions) > 0 {
		included := false
		for _, ext := range s.options.IncludeExtensions {
			if fileInfo.Extension == ext {
				included = true
				break
			}
		}
		if !included {
			return true
		}
	}

	// Check file exclusion patterns
	for _, excludePattern := range s.options.ExcludeFiles {
		if matched, _ := filepath.Match(excludePattern, fileInfo.Name); matched {
			return true
		}
		if matched, _ := filepath.Match(excludePattern, fileInfo.Path); matched {
			return true
		}
	}

	return false
}

// Helper functions

// isSymlink checks if a file is a symbolic link
func isSymlink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}

// isBinaryFile attempts to determine if a file is binary
func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil || n == 0 {
		return false
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	// Check if it's valid UTF-8
	return !utf8.Valid(buffer[:n])
}

// isTestFile determines if a file is a test file based on naming conventions
func isTestFile(filename string) bool {
	lower := strings.ToLower(filename)

	// Common test file patterns
	testPatterns := []string{
		"_test.", ".test.", "_spec.", ".spec.",
		"test_", "spec_", "tests.", "specs.",
	}

	for _, pattern := range testPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	// Check if it's in a test directory
	testDirs := []string{"test", "tests", "spec", "specs", "__tests__"}
	for _, testDir := range testDirs {
		if strings.Contains(lower, testDir) {
			return true
		}
	}

	return false
}

// GetStats returns the current scanning statistics
func (s *Scanner) GetStats() ScanStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}
