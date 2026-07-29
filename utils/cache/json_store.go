// Package cache owns reusable yfin cache mechanics.
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// JSONStore persists JSON documents below one fixed root.
type JSONStore struct {
	root string
}

// NewJSONStore creates a store rooted at root.
func NewJSONStore(root string) (*JSONStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("cache: root is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("cache: resolve root: %w", err)
	}
	if err := os.MkdirAll(absolute, 0o755); err != nil {
		return nil, fmt.Errorf("cache: create root %s: %w", absolute, err)
	}
	return &JSONStore{root: absolute}, nil
}

// Root returns the absolute cache root.
func (s *JSONStore) Root() string {
	return s.root
}

// Read decodes one JSON document.
func (s *JSONStore) Read(relativePath string, destination any) error {
	path, err := s.resolve(relativePath)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cache: read %s: %w", relativePath, err)
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return fmt.Errorf("cache: decode %s: %w", relativePath, err)
	}
	return nil
}

// Write atomically replaces one JSON document.
func (s *JSONStore) Write(relativePath string, value any) error {
	path, err := s.resolve(relativePath)
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("cache: encode %s: %w", relativePath, err)
	}
	body = append(body, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cache: create directory %s: %w", dir, err)
	}
	file, err := os.CreateTemp(dir, ".yfin-*.tmp")
	if err != nil {
		return fmt.Errorf("cache: create temporary file in %s: %w", dir, err)
	}
	temporaryPath := file.Name()
	defer os.Remove(temporaryPath)

	if err := file.Chmod(0o644); err != nil {
		_ = file.Close()
		return fmt.Errorf("cache: set temporary permissions: %w", err)
	}
	if _, err := file.Write(body); err != nil {
		_ = file.Close()
		return fmt.Errorf("cache: write temporary file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("cache: sync temporary file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("cache: close temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("cache: replace %s: %w", relativePath, err)
	}
	return nil
}

// List returns direct JSON children of relativeDir as root-relative paths.
func (s *JSONStore) List(relativeDir string) ([]string, error) {
	dir, err := s.resolve(relativeDir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cache: list %s: %w", relativeDir, err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(filepath.Clean(relativeDir), entry.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}

func (s *JSONStore) resolve(relativePath string) (string, error) {
	if strings.TrimSpace(relativePath) == "" || filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("cache: path must be a non-empty relative path: %q", relativePath)
	}
	clean := filepath.Clean(relativePath)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("cache: path escapes root: %q", relativePath)
	}
	path := filepath.Join(s.root, clean)
	relativeToRoot, err := filepath.Rel(s.root, path)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("cache: path escapes root: %q", relativePath)
	}
	return path, nil
}
