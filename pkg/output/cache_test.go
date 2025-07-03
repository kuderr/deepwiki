package output

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/kuderr/deepwiki/pkg/generator"
)

func TestNewCacheManager(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)
	if cm == nil {
		t.Fatal("NewCacheManager returned nil")
	}

	if !cm.enabled {
		t.Error("Cache should be enabled")
	}

	if cm.cacheDir != tempDir {
		t.Errorf("Expected cache dir %s, got %s", tempDir, cm.cacheDir)
	}
}

func TestNewCacheManager_Disabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cm := NewCacheManager(logger, "", false)
	if cm == nil {
		t.Fatal("NewCacheManager returned nil")
	}

	if cm.enabled {
		t.Error("Cache should be disabled")
	}
}

func TestCacheManager_SetAndGet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	// Test basic string caching
	key := "test-key"
	value := "test-value"

	err := cm.Set(key, value, time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	var result string
	found := cm.Get(key, &result)
	if !found {
		t.Error("Get should have found the cached value")
	}

	if result != value {
		t.Errorf("Expected %s, got %s", value, result)
	}
}

func TestCacheManager_GetNonExistent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	var result string
	found := cm.Get("non-existent-key", &result)
	if found {
		t.Error("Get should not have found non-existent key")
	}
}

func TestCacheManager_Delete(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	key := "test-key"
	value := "test-value"

	// Set and verify
	cm.Set(key, value, time.Hour)
	var result string
	if !cm.Get(key, &result) {
		t.Fatal("Should have found the cached value")
	}

	// Delete and verify
	cm.Delete(key)
	if cm.Get(key, &result) {
		t.Error("Should not have found deleted key")
	}
}

func TestCacheManager_Clear(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	// Add multiple items
	cm.Set("key1", "value1", time.Hour)
	cm.Set("key2", "value2", time.Hour)

	// Clear cache
	err := cm.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify items are gone
	var result string
	if cm.Get("key1", &result) {
		t.Error("key1 should have been cleared")
	}
	if cm.Get("key2", &result) {
		t.Error("key2 should have been cleared")
	}
}

func TestCacheManager_SetCacheSettings(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	newMaxAge := 2 * time.Hour
	newMaxSize := int64(200 * 1024 * 1024)

	cm.SetCacheSettings(newMaxAge, newMaxSize)

	if cm.maxAge != newMaxAge {
		t.Errorf("Expected maxAge %v, got %v", newMaxAge, cm.maxAge)
	}

	if cm.maxSize != newMaxSize {
		t.Errorf("Expected maxSize %d, got %d", newMaxSize, cm.maxSize)
	}
}

func TestCacheManager_GetStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	stats := cm.GetStats()
	if !stats.Enabled {
		t.Error("Stats should show cache as enabled")
	}

	if stats.MemoryEntries != 0 {
		t.Errorf("Expected 0 memory entries, got %d", stats.MemoryEntries)
	}

	// Add an item and check stats
	cm.Set("test-key", "test-value", time.Hour)

	stats = cm.GetStats()
	if stats.MemoryEntries != 1 {
		t.Errorf("Expected 1 memory entry, got %d", stats.MemoryEntries)
	}
}

func TestCacheManager_CacheWikiPage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	page := &generator.WikiPage{
		ID:      "test-page",
		Title:   "Test Page",
		Content: "Test content",
	}

	err := cm.CacheWikiPage(page.ID, page)
	if err != nil {
		t.Fatalf("CacheWikiPage failed: %v", err)
	}

	// Retrieve the cached page
	cachedPage, found := cm.GetCachedWikiPage(page.ID)
	if !found {
		t.Error("Should have found cached wiki page")
	}

	if cachedPage.ID != page.ID {
		t.Errorf("Expected page ID %s, got %s", page.ID, cachedPage.ID)
	}

	if cachedPage.Title != page.Title {
		t.Errorf("Expected page title %s, got %s", page.Title, cachedPage.Title)
	}
}

func TestCacheManager_CacheWikiStructure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	structure := &generator.WikiStructure{
		ID:          "test-wiki",
		Title:       "Test Wiki",
		Description: "Test description",
		Language:    "English", // types.LanguageEnglish
	}

	projectPath := "/test/project"
	err := cm.CacheWikiStructure(projectPath, structure)
	if err != nil {
		t.Fatalf("CacheWikiStructure failed: %v", err)
	}

	// Retrieve the cached structure
	cachedStructure, found := cm.GetCachedWikiStructure(projectPath)
	if !found {
		t.Error("Should have found cached wiki structure")
		return
	}

	if cachedStructure == nil {
		t.Error("Cached structure should not be nil")
		return
	}

	if cachedStructure.ID != structure.ID {
		t.Errorf("Expected structure ID %s, got %s", structure.ID, cachedStructure.ID)
	}

	if cachedStructure.Title != structure.Title {
		t.Errorf("Expected structure title %s, got %s", structure.Title, cachedStructure.Title)
	}
}

func TestCacheManager_Cleanup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tempDir := t.TempDir()

	cm := NewCacheManager(logger, tempDir, true)

	// Set a short max age for testing
	cm.SetCacheSettings(50*time.Millisecond, 1024*1024)

	// Add an item
	cm.Set("test-key", "test-value", 50*time.Millisecond)

	// Verify item was added
	var result string
	if !cm.Get("test-key", &result) {
		t.Fatal("Item should have been cached")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	err := cm.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Item should be gone
	if cm.Get("test-key", &result) {
		t.Error("Expired item should have been cleaned up")
	}
}

func TestCacheManager_DisabledOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cm := NewCacheManager(logger, "", false)

	// All operations should be no-ops when disabled
	err := cm.Set("key", "value", time.Hour)
	if err != nil {
		t.Errorf("Set should not error when disabled: %v", err)
	}

	var result string
	found := cm.Get("key", &result)
	if found {
		t.Error("Get should return false when disabled")
	}

	cm.Delete("key") // Should not panic

	err = cm.Clear()
	if err != nil {
		t.Errorf("Clear should not error when disabled: %v", err)
	}

	err = cm.Cleanup()
	if err != nil {
		t.Errorf("Cleanup should not error when disabled: %v", err)
	}
}

func TestNewDiskCache(t *testing.T) {
	tempDir := t.TempDir()

	dc, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("NewDiskCache failed: %v", err)
	}

	if dc == nil {
		t.Fatal("NewDiskCache returned nil")
	}

	if dc.baseDir != tempDir {
		t.Errorf("Expected baseDir %s, got %s", tempDir, dc.baseDir)
	}
}

func TestDiskCache_SetAndGet(t *testing.T) {
	tempDir := t.TempDir()

	dc, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("NewDiskCache failed: %v", err)
	}

	key := "test-key"
	value := "test-value"

	err = dc.Set(key, value, time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	result, err := dc.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Note: JSON unmarshaling may change type
	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}
}

func TestDiskCache_Delete(t *testing.T) {
	tempDir := t.TempDir()

	dc, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("NewDiskCache failed: %v", err)
	}

	key := "test-key"
	value := "test-value"

	// Set and verify
	dc.Set(key, value, time.Hour)
	_, err = dc.Get(key)
	if err != nil {
		t.Fatal("Should have found the cached value")
	}

	// Delete and verify
	err = dc.Delete(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = dc.Get(key)
	if err == nil {
		t.Error("Should not have found deleted key")
	}
}

func TestDiskCache_Clear(t *testing.T) {
	tempDir := t.TempDir()

	dc, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("NewDiskCache failed: %v", err)
	}

	// Add multiple items
	dc.Set("key1", "value1", time.Hour)
	dc.Set("key2", "value2", time.Hour)

	// Clear cache
	err = dc.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify items are gone
	_, err1 := dc.Get("key1")
	_, err2 := dc.Get("key2")

	if err1 == nil {
		t.Error("key1 should have been cleared")
	}
	if err2 == nil {
		t.Error("key2 should have been cleared")
	}
}

func TestDiskCache_GetStats(t *testing.T) {
	tempDir := t.TempDir()

	dc, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("NewDiskCache failed: %v", err)
	}

	stats := dc.GetStats()
	if stats.Entries != 0 {
		t.Errorf("Expected 0 entries, got %d", stats.Entries)
	}

	// Add an item and check stats
	dc.Set("test-key", "test-value", time.Hour)

	stats = dc.GetStats()
	if stats.Entries != 1 {
		t.Errorf("Expected 1 entry, got %d", stats.Entries)
	}

	if stats.Size <= 0 {
		t.Errorf("Expected positive size, got %d", stats.Size)
	}
}
