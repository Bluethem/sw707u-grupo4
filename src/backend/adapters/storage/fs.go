package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type FS struct {
	dir string
}

func NewFS(dir string) *FS {
	return &FS{dir: dir}
}

func (f *FS) Guardar(_ context.Context, key string, data []byte) error {
	if err := os.MkdirAll(f.dir, 0o755); err != nil {
		return fmt.Errorf("storage fs mkdir: %w", err)
	}
	path := filepath.Join(f.dir, filepath.Base(key))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("storage fs guardar %q: %w", key, err)
	}
	return nil
}

func (f *FS) Leer(_ context.Context, key string) ([]byte, error) {
	path := filepath.Join(f.dir, filepath.Base(key))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("storage fs leer %q: %w", key, err)
	}
	return data, nil
}
