package schemacache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/salmonumbrella/airwallex-cli/internal/api"
)

func TestCache_GetSet(t *testing.T) {
	tmpDir := t.TempDir()
	cache := New(tmpDir, 24*time.Hour)

	schema := &api.Schema{
		Fields: []api.SchemaField{
			{Key: "account_name", Required: true},
		},
	}

	key := CacheKey("US", "COMPANY", "LOCAL")

	// Should not exist initially
	if _, ok := cache.Get(key); ok {
		t.Fatal("expected cache miss")
	}

	// Set and get
	if err := cache.Set(key, schema); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got.Fields) != 1 || got.Fields[0].Key != "account_name" {
		t.Errorf("got wrong schema: %+v", got)
	}
}

func TestCache_Expiry(t *testing.T) {
	tmpDir := t.TempDir()
	cache := New(tmpDir, 1*time.Millisecond) // Very short TTL

	schema := &api.Schema{Fields: []api.SchemaField{{Key: "test"}}}
	key := CacheKey("US", "COMPANY", "LOCAL")

	cache.Set(key, schema)
	time.Sleep(10 * time.Millisecond)

	if _, ok := cache.Get(key); ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestCacheKey(t *testing.T) {
	key := CacheKey("US", "COMPANY", "SWIFT")
	want := "US_COMPANY_SWIFT"
	if key != want {
		t.Errorf("key = %s, want %s", key, want)
	}
}

func TestCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cache := New(tmpDir, 24*time.Hour)

	schema := &api.Schema{Fields: []api.SchemaField{{Key: "test"}}}
	cache.Set(CacheKey("US", "COMPANY", "LOCAL"), schema)
	cache.Set(CacheKey("GB", "PERSONAL", "SWIFT"), schema)

	if err := cache.Clear(); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	if _, ok := cache.Get(CacheKey("US", "COMPANY", "LOCAL")); ok {
		t.Fatal("expected cache miss after clear")
	}
}

func TestCacheKey_DefaultTransferMethod(t *testing.T) {
	key := CacheKey("US", "COMPANY", "")
	want := "US_COMPANY_LOCAL"
	if key != want {
		t.Errorf("key = %s, want %s", key, want)
	}
}

func TestCache_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	cache := New(tmpDir, 50*time.Millisecond)

	schema := &api.Schema{Fields: []api.SchemaField{{Key: "test"}}}

	// 1. Entry that will expire
	expiredKey := CacheKey("US", "COMPANY", "LOCAL")
	if err := cache.Set(expiredKey, schema); err != nil {
		t.Fatalf("set expired entry: %v", err)
	}

	// Wait for the first entry to expire
	time.Sleep(60 * time.Millisecond)

	// 2. Valid entry (not expired)
	validKey := CacheKey("GB", "PERSONAL", "SWIFT")
	if err := cache.Set(validKey, schema); err != nil {
		t.Fatalf("set valid entry: %v", err)
	}

	// 3. Corrupt JSON file written directly to disk
	corruptPath := filepath.Join(tmpDir, "CORRUPT_ENTRY.json")
	if err := os.WriteFile(corruptPath, []byte("not-valid-json{{{"), 0600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	// Run prune
	if err := cache.Prune(); err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	// Valid entry should still exist
	if _, ok := cache.Get(validKey); !ok {
		t.Error("expected valid entry to survive prune")
	}

	// Expired entry should be removed
	if _, ok := cache.Get(expiredKey); ok {
		t.Error("expected expired entry to be pruned")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, expiredKey+".json")); !os.IsNotExist(err) {
		t.Error("expected expired file to be deleted from disk")
	}

	// Corrupt file should be removed
	if _, err := os.Stat(corruptPath); !os.IsNotExist(err) {
		t.Error("expected corrupt file to be deleted from disk")
	}
}

func TestCache_SetCreatesDir(t *testing.T) {
	// Use a nested path that doesn't exist yet
	nestedDir := filepath.Join(t.TempDir(), "deeply", "nested", "cache")
	cache := New(nestedDir, 24*time.Hour)

	schema := &api.Schema{Fields: []api.SchemaField{{Key: "account_name"}}}
	key := CacheKey("US", "COMPANY", "LOCAL")

	if err := cache.Set(key, schema); err != nil {
		t.Fatalf("set should create directory and succeed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}

	// Verify value is readable
	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit after set")
	}
	if len(got.Fields) != 1 || got.Fields[0].Key != "account_name" {
		t.Errorf("got wrong schema: %+v", got)
	}
}

func TestCache_GetCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	cache := New(tmpDir, 24*time.Hour)

	key := CacheKey("US", "COMPANY", "LOCAL")
	corruptPath := filepath.Join(tmpDir, key+".json")

	// Write garbage data to the cache file path
	if err := os.WriteFile(corruptPath, []byte("{{{not json!!!"), 0600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	// Get should return false, not panic or error
	schema, ok := cache.Get(key)
	if ok {
		t.Fatal("expected cache miss for corrupt file")
	}
	if schema != nil {
		t.Fatal("expected nil schema for corrupt file")
	}
}

func TestCache_ClearNonExistentDir(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "does-not-exist")
	cache := New(nonExistent, 24*time.Hour)

	// Clear on non-existent directory should be a no-op, not an error
	if err := cache.Clear(); err != nil {
		t.Fatalf("clear on non-existent dir should return nil, got: %v", err)
	}
}

func TestCache_PruneNonExistentDir(t *testing.T) {
	nonExistent := filepath.Join(t.TempDir(), "does-not-exist")
	cache := New(nonExistent, 24*time.Hour)

	// Prune on non-existent directory should be a no-op, not an error
	if err := cache.Prune(); err != nil {
		t.Fatalf("prune on non-existent dir should return nil, got: %v", err)
	}
}

func TestCacheKey_CaseNormalization(t *testing.T) {
	key := CacheKey("us", "company", "swift")
	want := "US_COMPANY_SWIFT"
	if key != want {
		t.Errorf("key = %s, want %s", key, want)
	}

	// Also test mixed case
	key2 := CacheKey("Gb", "Personal", "Local")
	want2 := "GB_PERSONAL_LOCAL"
	if key2 != want2 {
		t.Errorf("key = %s, want %s", key2, want2)
	}
}
