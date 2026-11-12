package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

// Store is the interface for agent memory persistence.
//
// Implementations must be safe for concurrent use.
type Store interface {
	// Get retrieves a value by key. Returns ErrNotFound if the key does not exist.
	Get(key string) (interface{}, error)
	// Set stores a value under the given key.
	Set(key string, value interface{}) error
	// Delete removes a key from the store. Returns ErrNotFound if the key does not exist.
	Delete(key string) error
	// List returns all keys currently stored.
	List() ([]string, error)
}

// ErrNotFound is returned when a key is not found in the store.
var ErrNotFound = fmt.Errorf("key not found")

// MemStore is an in-memory implementation of Store.
//
// It is safe for concurrent use.
type MemStore struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewMemStore creates a new empty in-memory store.
func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value by key.
func (s *MemStore) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

// Set stores a value under the given key.
func (s *MemStore) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Delete removes a key from the store.
func (s *MemStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}
	delete(s.data, key)
	return nil
}

// List returns all keys currently stored.
func (s *MemStore) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// Memory is a typed wrapper around a Store for agent memory.
//
// It provides typed get/set operations for agent state, conversation history,
// and learned facts.
type Memory struct {
	store Store
}

// NewMemory creates a new Memory backed by the given Store.
func NewMemory(store Store) *Memory {
	return &Memory{store: store}
}

// Get retrieves a typed value by key.
func (m *Memory) Get(key string, out interface{}) error {
	v, err := m.store.Get(key)
	if err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal stored value: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("unmarshal into target: %w", err)
	}
	return nil
}

// Set stores a typed value under the given key.
func (m *Memory) Set(key string, value interface{}) error {
	return m.store.Set(key, value)
}

// Delete removes a key from the store.
func (m *Memory) Delete(key string) error {
	return m.store.Delete(key)
}

// List returns all keys currently stored.
func (m *Memory) List() ([]string, error) {
	return m.store.List()
}

// FileStore is a file-backed implementation of Store.
//
// It persists memory to a JSON file on disk, allowing memory to survive
// process restarts.
type FileStore struct {
	*MemStore
	path string
}

// NewFileStore creates a new FileStore backed by the given file path.
//
// If the file already exists, its contents are loaded into the store.
func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{
		MemStore: NewMemStore(),
		path:   path,
	}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

// load reads the store from disk.
func (fs *FileStore) load() error {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var pkgs []struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(data, &pkgs); err != nil {
		return err
	}
	for _, p := range pkgs {
		fs.data[p.Key] = p.Value
	}
	return nil
}

// save writes the store to disk.
func (fs *FileStore) save() error {
	fs.mu.RLock()
	pkgs := make([]struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}, 0, len(fs.data))
	for k, v := range fs.data {
		pkgs = append(pkgs, struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		}{Key: k, Value: v})
	}
	fs.mu.RUnlock()

	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Key < pkgs[j].Key
	})

	data, err := json.MarshalIndent(pkgs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.path, data, 0644)
}

// Set stores a value under the given key and persists to disk.
func (fs *FileStore) Set(key string, value interface{}) error {
	if err := fs.MemStore.Set(key, value); err != nil {
		return err
	}
	return fs.save()
}

// Delete removes a key from the store and persists to disk.
func (fs *FileStore) Delete(key string) error {
	if err := fs.MemStore.Delete(key); err != nil {
		return err
	}
	return fs.save()
}