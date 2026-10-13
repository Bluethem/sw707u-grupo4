package storage

import (
	"context"
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
