package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDump(t *testing.T) {
	cfg := DefaultConfig
	actual := Dump(cfg)

	assert.Contains(t, actual, `timezone: ""`)
	assert.Contains(t, actual, `debug: false`)
}
