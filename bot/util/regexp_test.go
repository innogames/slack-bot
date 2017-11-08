package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegexp(t *testing.T) {
	t.Run("RegexpResultToParams", func(t *testing.T) {
		re := CompileRegexp(`(?P<first>\w+) (?P<second>\w+)`)
		match := re.FindStringSubmatch("foo bar")

		actual := RegexpResultToParams(re, match)
		assert.Equal(t, map[string]string{
			FullMatch: "foo bar",
			"first":   "foo",
			"second":  "bar",
		}, actual)
	})

	t.Run("RegexpEmptyResultToParams", func(t *testing.T) {
		re := CompileRegexp(`(?P<first>\w+) (?P<second>\w+)`)
		match := re.FindStringSubmatch("nixda")

		actual := RegexpResultToParams(re, match)
		assert.Equal(t, map[string]string{}, actual)
	})
}
