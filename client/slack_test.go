package client

import (
	"errors"
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestGetSlackClient(t *testing.T) {
	t.Run("Connect with invalid token", func(t *testing.T) {
		cfg := config.Slack{
			TestEndpointURL: "http://slack.example.com/",
			Token:           "??",
			Debug:           true,
		}

		client, err := GetSlackClient(cfg)
		assert.Equal(t, "config slack.token needs to start with 'xoxb-'", err.Error())
		assert.Nil(t, client)
	})

	t.Run("Connect with invalid socket token", func(t *testing.T) {
		cfg := config.Slack{
			TestEndpointURL: "http://slack.example.com/",
			Token:           "xoxb-123",
			SocketToken:     "???",
			Debug:           true,
		}

		client, err := GetSlackClient(cfg)
		assert.Equal(t, "config slack.socket_token needs to start with 'xapp-'", err.Error())
		assert.Nil(t, client)
	})

	t.Run("Connect with invalid socket-token", func(t *testing.T) {
		cfg := config.Slack{
			TestEndpointURL: "http://slack.example.com/",
			Token:           "xoxb-yep",
			SocketToken:     "sometoken",
			Debug:           true,
		}

		client, err := GetSlackClient(cfg)
		assert.Equal(t, "config slack.socket_token needs to start with 'xapp-'", err.Error())
		assert.Nil(t, client)
	})

	t.Run("Send to user", func(t *testing.T) {
		slackClient := &Slack{}
		slackClient.SendToUser("user", "foo")
	})
}

func TestGetSlackUser(t *testing.T) {
	AllUsers = map[string]string{
		"U121": "Jon Doe",
		"U122": "Doe Jon",
	}
	assertIDNameLookup(t, "Jon Doe", "U121", "Jon Doe")
	assertIDNameLookup(t, "@Jon Doe", "U121", "Jon Doe")
	assertIDNameLookup(t, "jOn Doe", "U121", "Jon Doe")
	assertIDNameLookup(t, "jOn", "", "")
	assertIDNameLookup(t, "", "", "")
	assertIDNameLookup(t, "Doe Jon", "U122", "Doe Jon")

	assertIDNameLookup(t, "U122", "U122", "Doe Jon")
	assertIDNameLookup(t, "U121", "U121", "Jon Doe")
}

func TestGetSlackChannel(t *testing.T) {
	AllChannels = map[string]string{
		"C123": "dev",
		"C234": "general",
	}

	id, name := GetChannelIDAndName("#C123")
	assert.Equal(t, "C123", id)
	assert.Equal(t, "dev", name)

	id, name = GetChannelIDAndName("general")
	assert.Equal(t, "C234", id)
	assert.Equal(t, "general", name)

	id, name = GetChannelIDAndName("foobar")
	assert.Empty(t, id)
	assert.Empty(t, name)
}

func TestGetMessageArchiveLink(t *testing.T) {
	AuthResponse = slack.AuthTestResponse{
		Team: "Test-Project",
		URL:  "https://test-project.slack.com/",
	}

	message := msg.MessageRef{}
	message.Timestamp = "1610699454.002000"
	message.Channel = "DKJAPDWV8"
	actual := GetSlackArchiveLink(message)

	expected := "https://test-project.slack.com/archives/DKJAPDWV8/p1610699454002000"
	assert.Equal(t, expected, actual)
}

func TestGetSlackLink(t *testing.T) {
	link := GetSlackLink("name", "url", "color")
	assert.Equal(t, "url", link.URL)
	assert.Equal(t, "name", link.Text)
}

func TestSendMessage(t *testing.T) {
	client := &Slack{}

	t.Run("No text", func(t *testing.T) {
		ref := msg.MessageRef{}
		ref.Channel = "C1233"
		actual := client.SendMessage(ref, "")
		assert.Empty(t, actual)
	})

	t.Run("No target", func(t *testing.T) {
		ref := msg.MessageRef{}
		ref.Channel = ""
		actual := client.SendMessage(ref, "test")
		assert.Empty(t, actual)
	})

	t.Run("ReplyError", func(t *testing.T) {
		ref := msg.MessageRef{}
		err := errors.New("test error")
		client.ReplyError(ref, err)
	})
}

func assertIDNameLookup(t *testing.T, identifier string, expectedID string, expectedName string) {
	t.Helper()

	id, name := GetUserIDAndName(identifier)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedID, id)
}
