package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAppName(t *testing.T) {
	if AppName != "airwallex-cli" {
		t.Errorf("AppName = %q, want 'airwallex-cli'", AppName)
	}
}

func TestConfigDir(t *testing.T) {
	t.Run("with XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		got, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		want := "/custom/config/airwallex-cli"
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})

	t.Run("without XDG_CONFIG_HOME on darwin", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("skipping darwin-specific test")
		}
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		got, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		want := filepath.Join(home, "Library", "Application Support", "airwallex-cli")
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})

	t.Run("without XDG_CONFIG_HOME on linux", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("skipping linux-specific test")
		}
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		got, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		want := filepath.Join(home, ".config", "airwallex-cli")
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})
}

func TestDataDir(t *testing.T) {
	t.Run("with XDG_DATA_HOME set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/custom/data")
		got, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		want := "/custom/data/airwallex-cli"
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	})

	t.Run("without XDG_DATA_HOME on darwin", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("skipping darwin-specific test")
		}
		_ = os.Unsetenv("XDG_DATA_HOME")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		got, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		want := filepath.Join(home, "Library", "Application Support", "airwallex-cli")
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	})

	t.Run("without XDG_DATA_HOME on linux", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("skipping linux-specific test")
		}
		_ = os.Unsetenv("XDG_DATA_HOME")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		got, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		want := filepath.Join(home, ".local", "share", "airwallex-cli")
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	})
}

func TestCacheDir(t *testing.T) {
	t.Run("with XDG_CACHE_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/custom/cache")
		got, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}
		want := "/custom/cache/airwallex-cli"
		if got != want {
			t.Errorf("CacheDir() = %q, want %q", got, want)
		}
	})

	t.Run("without XDG_CACHE_HOME on darwin", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("skipping darwin-specific test")
		}
		_ = os.Unsetenv("XDG_CACHE_HOME")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		got, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}
		want := filepath.Join(home, "Library", "Caches", "airwallex-cli")
		if got != want {
			t.Errorf("CacheDir() = %q, want %q", got, want)
		}
	})

	t.Run("without XDG_CACHE_HOME on linux", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("skipping linux-specific test")
		}
		_ = os.Unsetenv("XDG_CACHE_HOME")
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		got, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}
		want := filepath.Join(home, ".cache", "airwallex-cli")
		if got != want {
			t.Errorf("CacheDir() = %q, want %q", got, want)
		}
	})
}

func TestPathConsistency(t *testing.T) {
	t.Run("all paths use same app name", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/test/config")
		t.Setenv("XDG_DATA_HOME", "/test/data")
		t.Setenv("XDG_CACHE_HOME", "/test/cache")

		configPath, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		dataPath, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		cachePath, err := CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() error = %v", err)
		}

		if filepath.Base(configPath) != AppName {
			t.Errorf("ConfigDir() base = %q, want %q", filepath.Base(configPath), AppName)
		}
		if filepath.Base(dataPath) != AppName {
			t.Errorf("DataDir() base = %q, want %q", filepath.Base(dataPath), AppName)
		}
		if filepath.Base(cachePath) != AppName {
			t.Errorf("CacheDir() base = %q, want %q", filepath.Base(cachePath), AppName)
		}
	})
}
