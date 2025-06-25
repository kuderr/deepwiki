package output

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/deepwiki-cli/deepwiki-cli/pkg/generator"
)

// CacheManager manages caching for repeated operations
type CacheManager struct {
	logger      *slog.Logger
	cacheDir    string
	enabled     bool
	maxAge      time.Duration
	maxSize     int64 // Max cache size in bytes
	mu          sync.RWMutex
	memoryCache map[string]*CacheEntry
	diskCache   *DiskCache
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key        string      `json:"key"`
	Data       interface{} `json:"data"`
	CreatedAt  time.Time   `json:"created_at"`
	AccessedAt time.Time   `json:"accessed_at"`
	Size       int64       `json:"size"`
	Hash       string      `json:"hash"`
}

// DiskCache manages persistent caching to disk
type DiskCache struct {
	baseDir   string
	indexFile string
	index     map[string]*CacheEntry
	mu        sync.RWMutex
}

// CacheStats provides statistics about cache usage
type CacheStats struct {
	Enabled       bool      `json:"enabled"`
	MemoryEntries int       `json:"memory_entries"`
	DiskEntries   int       `json:"disk_entries"`
	MemorySize    int64     `json:"memory_size"`
	DiskSize      int64     `json:"disk_size"`
	HitRate       float64   `json:"hit_rate"`
	MissRate      float64   `json:"miss_rate"`
	TotalRequests int64     `json:"total_requests"`
	TotalHits     int64     `json:"total_hits"`
	TotalMisses   int64     `json:"total_misses"`
	OldestEntry   time.Time `json:"oldest_entry"`
	NewestEntry   time.Time `json:"newest_entry"`
	LastCleanup   time.Time `json:"last_cleanup"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(logger *slog.Logger, cacheDir string, enabled bool) *CacheManager {
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "deepwiki-cache")
	}

	cm := &CacheManager{
		logger:      logger.With("component", "cache"),
		cacheDir:    cacheDir,
		enabled:     enabled,
		maxAge:      24 * time.Hour,    // Default 24 hours
		maxSize:     100 * 1024 * 1024, // Default 100MB
		memoryCache: make(map[string]*CacheEntry),
	}

	if enabled {
		if err := cm.initializeCache(); err != nil {
			logger.Warn("failed to initialize cache, disabling", "error", err)
			cm.enabled = false
		}
	}

	return cm
}

// SetCacheSettings configures cache parameters
func (cm *CacheManager) SetCacheSettings(maxAge time.Duration, maxSize int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.maxAge = maxAge
	cm.maxSize = maxSize
}

// Get retrieves an item from cache
func (cm *CacheManager) Get(key string, result interface{}) bool {
	if !cm.enabled {
		return false
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check memory cache first
	if entry, exists := cm.memoryCache[key]; exists {
		if cm.isEntryValid(entry) {
			entry.AccessedAt = time.Now()
			if err := cm.copyData(entry.Data, result); err == nil {
				return true
			}
		} else {
			delete(cm.memoryCache, key)
		}
	}

	// Check disk cache
	if cm.diskCache != nil {
		if data, err := cm.diskCache.Get(key); err == nil {
			if err := cm.copyData(data, result); err == nil {
				// Promote to memory cache
				cm.memoryCache[key] = &CacheEntry{
					Key:        key,
					Data:       data,
					CreatedAt:  time.Now(),
					AccessedAt: time.Now(),
					Size:       int64(len(fmt.Sprintf("%v", data))),
					Hash:       cm.generateHash(data),
				}
				return true
			}
		}
	}

	return false
}

// Set stores an item in cache
func (cm *CacheManager) Set(key string, data interface{}, ttl time.Duration) error {
	if !cm.enabled {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	size := int64(len(fmt.Sprintf("%v", data)))
	entry := &CacheEntry{
		Key:        key,
		Data:       data,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Size:       size,
		Hash:       cm.generateHash(data),
	}

	// Store in memory cache
	cm.memoryCache[key] = entry

	// Store in disk cache if available
	if cm.diskCache != nil {
		if err := cm.diskCache.Set(key, data, ttl); err != nil {
			cm.logger.Warn("failed to store in disk cache", "key", key, "error", err)
		}
	}

	// Clean up if necessary
	cm.cleanupIfNeeded()

	return nil
}

// Delete removes an item from cache
func (cm *CacheManager) Delete(key string) {
	if !cm.enabled {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.memoryCache, key)

	if cm.diskCache != nil {
		cm.diskCache.Delete(key)
	}
}

// Clear removes all items from cache
func (cm *CacheManager) Clear() error {
	if !cm.enabled {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.memoryCache = make(map[string]*CacheEntry)

	if cm.diskCache != nil {
		return cm.diskCache.Clear()
	}

	return nil
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := CacheStats{
		Enabled:       cm.enabled,
		MemoryEntries: len(cm.memoryCache),
	}

	if !cm.enabled {
		return stats
	}

	// Calculate memory stats
	var memorySize int64
	var oldest, newest time.Time

	for _, entry := range cm.memoryCache {
		memorySize += entry.Size

		if oldest.IsZero() || entry.CreatedAt.Before(oldest) {
			oldest = entry.CreatedAt
		}
		if newest.IsZero() || entry.CreatedAt.After(newest) {
			newest = entry.CreatedAt
		}
	}

	stats.MemorySize = memorySize
	stats.OldestEntry = oldest
	stats.NewestEntry = newest

	// Get disk cache stats
	if cm.diskCache != nil {
		diskStats := cm.diskCache.GetStats()
		stats.DiskEntries = diskStats.Entries
		stats.DiskSize = diskStats.Size
	}

	// TODO: Implement hit/miss rate tracking

	return stats
}

// Cleanup removes expired and least recently used items
func (cm *CacheManager) Cleanup() error {
	if !cm.enabled {
		return nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.performCleanup()
}

// CacheWikiPage caches a generated wiki page
func (cm *CacheManager) CacheWikiPage(pageID string, page *generator.WikiPage) error {
	key := fmt.Sprintf("wiki_page_%s", pageID)
	return cm.Set(key, page, cm.maxAge)
}

// GetCachedWikiPage retrieves a cached wiki page
func (cm *CacheManager) GetCachedWikiPage(pageID string) (*generator.WikiPage, bool) {
	key := fmt.Sprintf("wiki_page_%s", pageID)
	var page generator.WikiPage
	if cm.Get(key, &page) {
		return &page, true
	}
	return nil, false
}

// CacheWikiStructure caches a generated wiki structure
func (cm *CacheManager) CacheWikiStructure(projectPath string, structure *generator.WikiStructure) error {
	key := fmt.Sprintf("wiki_structure_%s", cm.generatePathHash(projectPath))
	return cm.Set(key, structure, cm.maxAge)
}

// GetCachedWikiStructure retrieves a cached wiki structure
func (cm *CacheManager) GetCachedWikiStructure(projectPath string) (*generator.WikiStructure, bool) {
	key := fmt.Sprintf("wiki_structure_%s", cm.generatePathHash(projectPath))
	var structure generator.WikiStructure
	if cm.Get(key, &structure) {
		return &structure, true
	}
	return nil, false
}

// Private methods

func (cm *CacheManager) initializeCache() error {
	// Create cache directory
	if err := os.MkdirAll(cm.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Initialize disk cache
	diskCache, err := NewDiskCache(cm.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to initialize disk cache: %w", err)
	}

	cm.diskCache = diskCache
	return nil
}

func (cm *CacheManager) isEntryValid(entry *CacheEntry) bool {
	return time.Since(entry.CreatedAt) < cm.maxAge
}

func (cm *CacheManager) copyData(src, dst interface{}) error {
	// TODO: Implement proper deep copy
	// For now, use JSON serialization as a simple approach
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func (cm *CacheManager) generateHash(data interface{}) string {
	h := md5.New()
	json.NewEncoder(h).Encode(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (cm *CacheManager) generatePathHash(path string) string {
	h := md5.New()
	h.Write([]byte(path))
	return hex.EncodeToString(h.Sum(nil))
}

func (cm *CacheManager) cleanupIfNeeded() {
	// Check if cleanup is needed based on size or age
	var totalSize int64
	for _, entry := range cm.memoryCache {
		totalSize += entry.Size
	}

	if totalSize > cm.maxSize || len(cm.memoryCache) > 1000 {
		cm.performCleanup()
	}
}

func (cm *CacheManager) performCleanup() error {
	now := time.Now()

	// Remove expired entries
	for key, entry := range cm.memoryCache {
		if now.Sub(entry.CreatedAt) > cm.maxAge {
			delete(cm.memoryCache, key)
		}
	}

	// If still over limit, remove least recently used
	if len(cm.memoryCache) > 800 { // Keep some buffer
		type entryWithKey struct {
			key   string
			entry *CacheEntry
		}

		var entries []entryWithKey
		for key, entry := range cm.memoryCache {
			entries = append(entries, entryWithKey{key, entry})
		}

		// Sort by access time (oldest first)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].entry.AccessedAt.After(entries[j].entry.AccessedAt) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Remove oldest 200 entries
		for i := 0; i < 200 && i < len(entries); i++ {
			delete(cm.memoryCache, entries[i].key)
		}
	}

	// Cleanup disk cache
	if cm.diskCache != nil {
		return cm.diskCache.Cleanup(cm.maxAge)
	}

	return nil
}

// DiskCache implementation

// NewDiskCache creates a new disk cache
func NewDiskCache(baseDir string) (*DiskCache, error) {
	dc := &DiskCache{
		baseDir:   baseDir,
		indexFile: filepath.Join(baseDir, "cache_index.json"),
		index:     make(map[string]*CacheEntry),
	}

	// Load existing index
	if err := dc.loadIndex(); err != nil {
		// Create new index if loading fails
		dc.index = make(map[string]*CacheEntry)
	}

	return dc, nil
}

// Set stores data to disk
func (dc *DiskCache) Set(key string, data interface{}, ttl time.Duration) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Serialize data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Create file path
	filePath := filepath.Join(dc.baseDir, key+".json")

	// Write to disk
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return err
	}

	// Update index
	dc.index[key] = &CacheEntry{
		Key:        key,
		CreatedAt:  time.Now(),
		AccessedAt: time.Now(),
		Size:       int64(len(jsonData)),
	}

	return dc.saveIndex()
}

// Get retrieves data from disk
func (dc *DiskCache) Get(key string) (interface{}, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	entry, exists := dc.index[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	filePath := filepath.Join(dc.baseDir, key+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Update access time
	entry.AccessedAt = time.Now()

	return result, nil
}

// Delete removes data from disk
func (dc *DiskCache) Delete(key string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	filePath := filepath.Join(dc.baseDir, key+".json")
	os.Remove(filePath) // Ignore error if file doesn't exist

	delete(dc.index, key)
	return dc.saveIndex()
}

// Clear removes all cached data
func (dc *DiskCache) Clear() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Remove all cache files
	entries, err := os.ReadDir(dc.baseDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" && entry.Name() != "cache_index.json" {
			os.Remove(filepath.Join(dc.baseDir, entry.Name()))
		}
	}

	dc.index = make(map[string]*CacheEntry)
	return dc.saveIndex()
}

// GetStats returns disk cache statistics
func (dc *DiskCache) GetStats() struct {
	Entries int
	Size    int64
} {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	var totalSize int64
	for _, entry := range dc.index {
		totalSize += entry.Size
	}

	return struct {
		Entries int
		Size    int64
	}{
		Entries: len(dc.index),
		Size:    totalSize,
	}
}

// Cleanup removes expired entries from disk
func (dc *DiskCache) Cleanup(maxAge time.Duration) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()

	for key, entry := range dc.index {
		if now.Sub(entry.CreatedAt) > maxAge {
			filePath := filepath.Join(dc.baseDir, key+".json")
			os.Remove(filePath)
			delete(dc.index, key)
		}
	}

	return dc.saveIndex()
}

func (dc *DiskCache) loadIndex() error {
	data, err := os.ReadFile(dc.indexFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &dc.index)
}

func (dc *DiskCache) saveIndex() error {
	data, err := json.Marshal(dc.index)
	if err != nil {
		return err
	}

	return os.WriteFile(dc.indexFile, data, 0644)
}
