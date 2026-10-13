package redis

import (
	"context"
	"encoding/hex"
	"os"
	"testing"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	decoded, err := hex.DecodeString(token)
	if err != nil {
		t.Fatalf("GenerateToken() returned invalid hex: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("token has %d bytes, want 32", len(decoded))
	}
}

func TestSessionStoreCreateGetDelete(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL no está configurado; se omite el test de integración con Redis")
	}

	options, err := redislib.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse REDIS_URL: %v", err)
	}
	client := redislib.NewClient(options)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Redis no disponible: %v", err)
	}

	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Del(ctx, sessionKey(token)).Err() })

	ttl := 5 * time.Minute
	store := NewSessionStore(client, ttl)
	want := SessionData{
		UserID:    "usuario-1",
		EmpresaID: "empresa-1",
		Rol:       "administrador",
	}

	if err := store.Create(ctx, token, want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	remaining, err := client.TTL(ctx, sessionKey(token)).Result()
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if remaining <= 0 || remaining > ttl {
		t.Fatalf("TTL = %v, want > 0 and <= %v", remaining, ttl)
	}

	got, err := store.Get(ctx, token)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == nil || *got != want {
		t.Fatalf("Get() = %#v, want %#v", got, want)
	}

	if err := store.Delete(ctx, token); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	got, err = store.Get(ctx, token)
	if err != nil {
		t.Fatalf("Get() after Delete() error = %v", err)
	}
	if got != nil {
		t.Fatalf("Get() after Delete() = %#v, want nil", got)
	}
}
