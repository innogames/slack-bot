package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelp(t *testing.T) {
	help := Help{
		Command: "test",
		Examples: []string{
			"i can do this",
			"i can do foo",
		},
	}

	t.Run("GetKeywords", func(t *testing.T) {
		expected := []string{
			"test",
			"i can do this",
			"i can do foo",
		}
		actual := help.GetKeywords()
		assert.Equal(t, expected, actual)
	})
}
