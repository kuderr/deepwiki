package rag

import (
	"sync"
	"time"
)

// cacheEntry represents a cached result with expiration
type cacheEntry struct {
	results []RetrievalResult
	expires time.Time
}

// InMemoryCache implements RetrievalCache using in-memory storage
type InMemoryCache struct {
	cache   map[string]*cacheEntry
	maxSize int
	mu      sync.RWMutex
	stats   map[string]interface{}
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(maxSize int) *InMemoryCache {
	return &InMemoryCache{
		cache:   make(map[string]*cacheEntry),
		maxSize: maxSize,
		stats: map[string]interface{}{
			"hits":      0,
			"misses":    0,
			"evictions": 0,
		},
	}
}

// Get retrieves results from cache
func (c *InMemoryCache) Get(key string) ([]RetrievalResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		c.stats["misses"] = c.stats["misses"].(int) + 1
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expires) {
		delete(c.cache, key)
		c.stats["misses"] = c.stats["misses"].(int) + 1
		return nil, false
	}

	c.stats["hits"] = c.stats["hits"].(int) + 1
	return entry.results, true
}

// Set stores results in cache
func (c *InMemoryCache) Set(key string, results []RetrievalResult, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &cacheEntry{
		results: results,
		expires: time.Now().Add(ttl),
	}
}

// Delete removes an entry from cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
}

// Clear removes all entries from cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
}

// Size returns the number of entries in cache
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// Stats returns cache statistics
func (c *InMemoryCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	stats := make(map[string]interface{})
	for k, v := range c.stats {
		stats[k] = v
	}
	stats["size"] = len(c.cache)

	return stats
}

// evictOldest removes the oldest entry from cache (simple LRU-like behavior)
func (c *InMemoryCache) evictOldest() {
	if len(c.cache) == 0 {
		return
	}

	// Find the entry with the earliest expiration time
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.cache {
		if first || entry.expires.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expires
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		c.stats["evictions"] = c.stats["evictions"].(int) + 1
	}
}
