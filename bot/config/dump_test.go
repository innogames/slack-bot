package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDump(t *testing.T) {
	cfg := DefaultConfig
	actual := Dump(cfg)

	assert.Contains(t, actual, `"Timezone": ""`)
	assert.Contains(t, actual, `"File": "./bot.log"`)
}
