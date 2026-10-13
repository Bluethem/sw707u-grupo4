package storage

import (
	"context"
	"fmt"
	"sync"
)

type Fake struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewFake() *Fake {
	return &Fake{data: make(map[string][]byte)}
}

func (f *Fake) Guardar(_ context.Context, key string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	f.data[key] = cp
	return nil
}

func (f *Fake) Get(key string) ([]byte, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b, ok := f.data[key]
	return b, ok
}

func (f *Fake) Leer(_ context.Context, key string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b, ok := f.data[key]
	if !ok {
		return nil, fmt.Errorf("storage fake: %q no encontrado", key)
	}
	cp := make([]byte, len(b))
	copy(cp, b)
	return cp, nil
}
