package config

import (
	"path/filepath"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.conf")

	// Create new config
	cfg := New(configPath)

	// Set some values
	if err := cfg.Set("TEST_KEY", "test_value"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	if err := cfg.Set("ANOTHER_KEY", "another_value"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Load config in new instance
	cfg2 := New(configPath)
	if err := cfg2.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify values
	if val := cfg2.GetOrDefault("TEST_KEY", ""); val != "test_value" {
		t.Errorf("GetOrDefault() = %v, want %v", val, "test_value")
	}

	if val := cfg2.GetOrDefault("ANOTHER_KEY", ""); val != "another_value" {
		t.Errorf("GetOrDefault() = %v, want %v", val, "another_value")
	}
}

func TestConfigGet(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := New(filepath.Join(tmpDir, "test.conf"))

	// Set a value
	if err := cfg.Set("KEY1", "value1"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Test Get for existing key
	val, err := cfg.Get("KEY1")
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if val != "value1" {
		t.Errorf("Get() = %v, want %v", val, "value1")
	}

	// Test Get for non-existing key
	_, err = cfg.Get("NONEXISTENT")
	if err == nil {
		t.Error("Get() error = nil, want error for non-existent key")
	}
}

func TestConfigGetOrDefault(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := New(filepath.Join(tmpDir, "test.conf"))

	// Test with default
	val := cfg.GetOrDefault("NONEXISTENT", "default_value")
	if val != "default_value" {
		t.Errorf("GetOrDefault() = %v, want %v", val, "default_value")
	}

	// Set a value and test
	if err := cfg.Set("KEY1", "value1"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	val = cfg.GetOrDefault("KEY1", "default")
	if val != "value1" {
		t.Errorf("GetOrDefault() = %v, want %v", val, "value1")
	}
}

func TestConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := New(filepath.Join(tmpDir, "test.conf"))

	// Test non-existent key
	if cfg.Exists("NONEXISTENT") {
		t.Error("Exists() = true, want false for non-existent key")
	}

	// Set a value and test
	if err := cfg.Set("KEY1", "value1"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	if !cfg.Exists("KEY1") {
		t.Error("Exists() = false, want true for existing key")
	}
}

func TestConfigDelete(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := New(filepath.Join(tmpDir, "test.conf"))

	// Set and delete
	if err := cfg.Set("KEY1", "value1"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	if !cfg.Exists("KEY1") {
		t.Error("Key should exist after Set()")
	}

	if err := cfg.Delete("KEY1"); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
	if cfg.Exists("KEY1") {
		t.Error("Key should not exist after Delete()")
	}
}

func TestConfigLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := New(filepath.Join(tmpDir, "nonexistent.conf"))

	// Should not error when loading non-existent file
	err := cfg.Load()
	if err != nil {
		t.Errorf("Load() on non-existent file error = %v, want nil", err)
	}
}

func TestConfigFilePath(t *testing.T) {
	expectedPath := "/tmp/test.conf"
	cfg := New(expectedPath)

	if cfg.FilePath() != expectedPath {
		t.Errorf("FilePath() = %v, want %v", cfg.FilePath(), expectedPath)
	}
}

func TestConfigLazyLoadOnRead(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lazy.conf")

	writer := New(configPath)
	if err := writer.Set("LAZY_KEY", "lazy_value"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	reader := New(configPath)
	val, err := reader.Get("LAZY_KEY")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if val != "lazy_value" {
		t.Errorf("Get() = %v, want lazy_value", val)
	}

	if got := reader.GetOrDefault("LAZY_KEY", "default"); got != "lazy_value" {
		t.Errorf("GetOrDefault() = %v, want lazy_value", got)
	}

	if !reader.Exists("LAZY_KEY") {
		t.Errorf("Exists() = false, want true")
	}
}
