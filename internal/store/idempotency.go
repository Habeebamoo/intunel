package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const idempotencyTTL = 24 * time.Hour

type IdempotencyStore struct {
	client *redis.Client
}

func NewIdempotencyStore(client *redis.Client) *IdempotencyStore {
	return &IdempotencyStore{client: client}
}

func (s *IdempotencyStore) key(idempotencyKey string) string {
	return fmt.Sprintf("idempotency:%s", idempotencyKey)
}

func (s *IdempotencyStore) Get(ctx context.Context, idempotencyKey string) (string, error) {
	return s.client.Get(ctx, s.key(idempotencyKey)).Result()
}

func (s *IdempotencyStore) Set(ctx context.Context, idempotencyKey string, response string) error {
	return s.client.Set(ctx, s.key(idempotencyKey), response, idempotencyTTL).Err()
}

func (s *IdempotencyStore) Exists(ctx context.Context, idempotencyKey string) (bool, error) {
	val, err := s.client.Exists(ctx, s.key(idempotencyKey)).Result()
	return val > 0, err
}