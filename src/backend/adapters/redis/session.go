package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "session:"

type SessionData struct {
	UserID    string `json:"user_id"`
	EmpresaID string `json:"empresa_id"`
	Rol       string `json:"rol"`
}

type SessionStore struct {
	client redislib.Cmdable
	ttl    time.Duration
}

func NewSessionStore(client redislib.Cmdable, ttl time.Duration) *SessionStore {
	return &SessionStore{client: client, ttl: ttl}
}

// GenerateToken creates an opaque session token with 256 bits of entropy.
func GenerateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func (s *SessionStore) Create(ctx context.Context, token string, data SessionData) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	if err := s.client.Set(ctx, sessionKey(token), encoded, s.ttl).Err(); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *SessionStore) Get(ctx context.Context, token string) (*SessionData, error) {
	encoded, err := s.client.Get(ctx, sessionKey(token)).Bytes()
	if err == redislib.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(encoded, &data); err != nil {
		return nil, fmt.Errorf("decode session: %w", err)
	}
	return &data, nil
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	if err := s.client.Del(ctx, sessionKey(token)).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func sessionKey(token string) string {
	return sessionKeyPrefix + token
}
