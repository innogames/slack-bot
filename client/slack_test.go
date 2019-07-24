package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func assertIdNameLookup(t *testing.T, identifier string, expectedId string, expectedName string) {
	id, name := GetUser(identifier)
	assert.Equal(t, expectedName, name)
	assert.Equal(t, expectedId, id)
}
