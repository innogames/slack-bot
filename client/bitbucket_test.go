package client

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/stretchr/testify/assert"
)

func TestBitbucket(t *testing.T) {
	t.Run("no host", func(t *testing.T) {
		cfg := config.Bitbucket{}

		client, err := GetBitbucketClient(cfg)

		assert.Nil(t, client)
		assert.Equal(t, "bitbucket: No host given", err.Error())
	})

	t.Run("no credentials", func(t *testing.T) {
		cfg := config.Bitbucket{
			Host: "https://bitbucket.example.com",
		}

		client, err := GetBitbucketClient(cfg)
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with username/password", func(t *testing.T) {
		cfg := config.Bitbucket{
			Host:     "https://bitbucket.example.com",
			Username: "myUsername",
			Password: "myPassword",
		}

		client, err := GetBitbucketClient(cfg)

		assert.Nil(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with apiKey", func(t *testing.T) {
		cfg := config.Bitbucket{
			Host:   "https://bitbucket.example.com",
			APIKey: "myApiKey",
		}

		client, err := GetBitbucketClient(cfg)

		assert.Nil(t, err)
		assert.NotNil(t, client)
	})
}
