package client

import (
	"testing"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestGetSlackClient(t *testing.T) {
	logger, _ := test.NewNullLogger()

	cfg := config.Slack{
		TestEndpointUrl: "http://slack.example.com/",
		Debug:           true,
	}

	client := GetSlackClient(cfg, logger)

	_, _, err := client.RTM.ConnectRTM()
	assert.Contains(t, err.Error(), "slack.example.com")
}

func TestGetSlackUser(t *testing.T) {
	Users = map[string]string{
		"U121": "Jon Doe",
		"U122": "Doe Jon",
	}
	assertIdNameLookup(t, "Jon Doe", "U121", "Jon Doe")
	assertIdNameLookup(t, "@Jon Doe", "U121", "Jon Doe")
	assertIdNameLookup(t, "jOn Doe", "U121", "Jon Doe")
	assertIdNameLookup(t, "jOn", "", "")
	assertIdNameLookup(t, "", "", "")
	assertIdNameLookup(t, "Doe Jon", "U122", "Doe Jon")

	assertIdNameLookup(t, "U122", "U122", "Doe Jon")
	assertIdNameLookup(t, "U121", "U121", "Jon Doe")
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

func TestGetSlackLink(t *testing.T) {
	link := GetSlackLink("name", "url", "color")
	assert.Equal(t, "url", link.URL)
	assert.Equal(t, "name", link.Text)
}

func assertIdNameLookup(t *testing.T, identifier string, expectedId string, expectedName string) {
	id, name := GetUser(identifier)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedId, id)
}
