package storage_test

import (
	"context"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/storage"
)

func TestFS_GuardarLeer(t *testing.T) {
	dir := t.TempDir()
	s := storage.NewFS(dir)
	if err := s.Guardar(context.Background(), "a.xml", []byte("<x/>")); err != nil {
		t.Fatal(err)
	}
	b, err := s.Leer(context.Background(), "a.xml")
	if err != nil || string(b) != "<x/>" {
		t.Fatalf("got %q err=%v", b, err)
	}
}
