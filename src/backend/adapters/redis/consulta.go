package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

const (
	consultaTokenKeyPrefix = "consulta:tok:"
	consultaFailKeyPrefix  = "consulta:fail:"
	consultaTTL            = 15 * time.Minute
)

type ConsultaStore struct {
	client redislib.Cmdable
}

func NewConsultaStore(client redislib.Cmdable) *ConsultaStore {
	return &ConsultaStore{client: client}
}

func (s *ConsultaStore) PutToken(ctx context.Context, token, empresaID, comprobanteID string) error {
	value := empresaID + "|" + comprobanteID
	if err := s.client.Set(ctx, consultaTokenKey(token), value, consultaTTL).Err(); err != nil {
		return fmt.Errorf("put consulta token: %w", err)
	}
	return nil
}

func (s *ConsultaStore) GetToken(ctx context.Context, token string) (empresaID, comprobanteID string, ok bool, err error) {
	value, err := s.client.Get(ctx, consultaTokenKey(token)).Result()
	if err == redislib.Nil {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, fmt.Errorf("get consulta token: %w", err)
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false, nil
	}
	return parts[0], parts[1], true, nil
}

func (s *ConsultaStore) IncrFail(ctx context.Context, ip string) (int64, error) {
	key := consultaFailKey(ip)
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("incr consulta fail: %w", err)
	}
	if count == 1 {
		if err := s.client.Expire(ctx, key, consultaTTL).Err(); err != nil {
			return count, fmt.Errorf("expire consulta fail: %w", err)
		}
	}
	return count, nil
}

func consultaTokenKey(token string) string {
	return consultaTokenKeyPrefix + token
}

func consultaFailKey(ip string) string {
	return consultaFailKeyPrefix + ip
}
