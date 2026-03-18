package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// FileStore is a file-backed key-value store implementing httpclient.Store.
// Data is persisted as a JSON object. Permissions: dir 0700, file 0600.
type FileStore struct {
	path string
	mu   sync.Mutex
	data map[string]string
}

// NewFileStore creates a FileStore backed by the file at path.
// The parent directory is created if it does not exist.
func NewFileStore(path string) (*FileStore, error) {
	s := &FileStore{
		path: path,
		data: make(map[string]string),
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	_ = s.loadFromDisk() // ignore error when file does not exist yet
	return s, nil
}

func (s *FileStore) loadFromDisk() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.data)
}

func (s *FileStore) saveToDisk() error {
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Load returns all stored key-value pairs.
func (s *FileStore) Load(_ context.Context) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]string, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result, nil
}

// Save stores value for key and persists to disk.
func (s *FileStore) Save(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.saveToDisk()
}

// IsSet returns true if key has a non-empty stored value.
func (s *FileStore) IsSet(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[key] != ""
}
