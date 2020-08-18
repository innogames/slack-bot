package storage

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
)

func TestRedisStorage(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	storage := NewRedisStorage(client)

	t.Run("test redis", func(t *testing.T) {
		testStorage(t, storage)
	})
}
