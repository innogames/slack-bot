package client

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/slack-go/slack"
	"testing"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/stretchr/testify/assert"
)

func TestGetSlackClient(t *testing.T) {
	cfg := config.Slack{
		TestEndpointURL: "http://slack.example.com/",
		Debug:           true,
	}

	client := GetSlackClient(cfg)

	_, _, err := client.RTM.ConnectRTM()
	assert.Contains(t, err.Error(), "slack.example.com")
}

func TestGetSlackUser(t *testing.T) {
	Users = map[string]string{
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
	Channels = map[string]string{
		"C123": "dev",
		"C234": "general",
	}

	id, name := GetChannel("#C123")
	assert.Equal(t, "C123", id)
	assert.Equal(t, "dev", name)

	id, name = GetChannel("general")
	assert.Equal(t, "C234", id)
	assert.Equal(t, "general", name)

	id, name = GetChannel("foobar")
	assert.Equal(t, "", id)
	assert.Equal(t, "", name)
}

func TestGetMessageArchiveLink(t *testing.T) {
	AuthResponse = slack.AuthTestResponse{
		Team: "Test-Project",
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

func assertIDNameLookup(t *testing.T, identifier string, expectedID string, expectedName string) {
	t.Helper()

	id, name := GetUser(identifier)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedID, id)
}
