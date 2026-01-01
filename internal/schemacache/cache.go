package schemacache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

// Cache stores beneficiary schemas locally with TTL
type Cache struct {
	mu  sync.RWMutex
	dir string
	ttl time.Duration
}

// cacheEntry wraps schema with timestamp
type cacheEntry struct {
	Schema   *api.Schema `json:"schema"`
	CachedAt time.Time   `json:"cached_at"`
}

// New creates a new schema cache
func New(dir string, ttl time.Duration) *Cache {
	return &Cache{dir: dir, ttl: ttl}
}

// CacheKey generates a cache key from parameters
func CacheKey(bankCountry, entityType, transferMethod string) string {
	if transferMethod == "" {
		transferMethod = "LOCAL"
	}
	return fmt.Sprintf("%s_%s_%s",
		strings.ToUpper(bankCountry),
		strings.ToUpper(entityType),
		strings.ToUpper(transferMethod))
}

// Get retrieves a cached schema if valid
func (c *Cache) Get(key string) (*api.Schema, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := c.path(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Check if expired - don't remove under read lock, just return miss
	// Let Set() overwrite or use Prune() for cleanup
	if time.Since(entry.CachedAt) > c.ttl {
		return nil, false
	}

	return entry.Schema, true
}

// Set stores a schema in the cache
func (c *Cache) Set(key string, schema *api.Schema) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(c.dir, 0700); err != nil {
		return err
	}

	entry := cacheEntry{
		Schema:   schema,
		CachedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path(key), data, 0600)
}

// Clear removes all cached schemas
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			_ = os.Remove(filepath.Join(c.dir, e.Name()))
		}
	}
	return nil
}

// Prune removes all expired entries from the cache
func (c *Cache) Prune() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		path := filepath.Join(c.dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var entry cacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			// Invalid entry, remove it
			_ = os.Remove(path)
			continue
		}

		if time.Since(entry.CachedAt) > c.ttl {
			_ = os.Remove(path)
		}
	}
	return nil
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.dir, key+".json")
}
