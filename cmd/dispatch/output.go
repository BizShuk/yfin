package dispatch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeJSONAtomic(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	return writeBytesAtomic(path, append(data, '\n'))
}

func writeErrorAtomic(path string, err error) error {
	if err == nil {
		return fmt.Errorf("write %s: nil error", path)
	}
	return writeBytesAtomic(path, []byte(err.Error()+"\n"))
}

func writeBytesAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory %s: %w", dir, err)
	}
	f, err := os.CreateTemp(dir, ".yfin-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output in %s: %w", dir, err)
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0o644); err != nil {
		_ = f.Close()
		return fmt.Errorf("set temporary output permissions: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("write temporary output: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync temporary output: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace output %s: %w", path, err)
	}
	return nil
}
