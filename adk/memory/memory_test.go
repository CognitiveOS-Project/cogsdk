package memory

import (
	"path/filepath"
	"testing"
)

func TestMemStore(t *testing.T) {
	s := NewMemStore()

	// Get on empty key returns ErrNotFound
	if _, err := s.Get("missing"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Set and Get
	if err := s.Set("key1", "value1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	v, err := s.Get("key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if v != "value1" {
		t.Fatalf("expected 'value1', got %v", v)
	}

	// Delete
	if err := s.Delete("key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := s.Get("key1"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	// Delete non-existent key
	if err := s.Delete("missing"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// List
	if err := s.Set("a", 1); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := s.Set("b", 2); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	keys, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestMemory(t *testing.T) {
	store := NewMemStore()
	m := NewMemory(store)

	// Set and Get typed value
	if err := m.Set("count", 42); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	var count int
	if err := m.Get("count", &count); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if count != 42 {
		t.Fatalf("expected 42, got %d", count)
	}

	// Delete
	if err := m.Delete("count"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	var count2 int
	if err := m.Get("count", &count2); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFileStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memory.json")
	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	_ = fs.Set("key", "value")
	fs2, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore reload failed: %v", err)
	}
	v, err := fs2.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if v != "value" {
		t.Fatalf("expected 'value', got %v", v)
	}
}