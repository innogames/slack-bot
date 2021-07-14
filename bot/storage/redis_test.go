package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
)

func TestRedisStorage(t *testing.T) {
	t.Run("test miniredis", func(t *testing.T) {
		server, err := miniredis.Run()
		if err != nil {
			panic(err)
		}
		defer server.Close()

		client := redis.NewClient(&redis.Options{
			Addr: server.Addr(),
		})

		storage := NewRedisStorage(client)

		testStorage(t, storage)
	})

	t.Run("test error handling", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr: "invalid.host",
		})

		storage := NewRedisStorage(client)

		var i int
		err := storage.Read("test", "foo", &i)
		assert.Equal(t, "dial tcp: address invalid.host: missing port in address", err.Error())

		keys, err := storage.GetKeys("test")
		assert.Len(t, keys, 0)
		assert.Equal(t, "dial tcp: address invalid.host: missing port in address", err.Error())
	})
}
