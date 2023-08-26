package storage

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// NewRedisStorage defined a redis bases storage to persist bot related information
func NewRedisStorage(client *redis.Client) Storage {
	return &redisStorage{
		client: client,
	}
}

type redisStorage struct {
	client *redis.Client
}

func (s redisStorage) Write(collection, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	ctx := context.Background()
	s.client.HSet(ctx, collection, key, string(data))

	return nil
}

func (s redisStorage) Read(collection, key string, v any) error {
	ctx := context.Background()
	res, err := s.client.HGet(ctx, collection, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(res), v)
}

func (s redisStorage) GetKeys(collection string) ([]string, error) {
	ctx := context.Background()
	res, err := s.client.HKeys(ctx, collection).Result()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s redisStorage) Delete(collection, key string) error {
	ctx := context.Background()
	_, err := s.client.HDel(ctx, collection, key).Result()

	return err
}
