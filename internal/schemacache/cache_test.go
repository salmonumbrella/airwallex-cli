package schemacache

import (
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
