package storage

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"go-analytics-service/internal/analytics"
)

type RedisStore struct {
	client *redis.Client
	key    string
}

func NewRedisStore(addr string) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[REDIS] ping failed: %v", err)
	} else {
		log.Printf("[REDIS] connected to %s", addr)
	}

	return &RedisStore{
		client: rdb,
		key:    "metrics",
	}
}

func (s *RedisStore) SaveMetric(m analytics.Metric) {
	if s == nil || s.client == nil {
		return
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("[REDIS] marshal metric: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := s.client.LPush(ctx, s.key, data).Err(); err != nil {
		log.Printf("[REDIS] LPUSH failed: %v", err)
	}
}
