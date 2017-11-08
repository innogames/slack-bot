package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFunctions(t *testing.T) {
	t.Run("makeSlice", func(t *testing.T) {
		actual := functions["makeSlice"].(func(args ...interface{}) []interface{})("1", "2", "3")
		assert.Equal(t, []interface{}{"1", "2", "3"}, actual)
	})

	t.Run("stringSlice", func(t *testing.T) {
		actual := functions["slice"].(func(string, int, int) string)("test me", 1, 3)
		assert.Equal(t, "es", actual)
	})
}

func TestTemplate(t *testing.T) {
	text := "{{ $users := makeSlice \"2222\"}}{{ $users }} {{ .foo }}"

	temp, _ := CompileTemplate(text)
	finalText, _ := EvalTemplate(temp, map[string]string{"foo": "bar"})

	assert.Equal(t, finalText, "[2222] bar")
}
