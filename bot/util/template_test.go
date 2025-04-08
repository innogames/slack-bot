package util

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctions(t *testing.T) {
	t.Run("makeSlice", func(t *testing.T) {
		actual := functions["makeSlice"].(func(args ...any) []any)("1", "2", "3")
		assert.Equal(t, []any{"1", "2", "3"}, actual)
	})

	t.Run("stringSlice", func(t *testing.T) {
		actual := functions["slice"].(func(string, int, int) string)("test me", 1, 3)
		assert.Equal(t, "es", actual)
	})
}

func TestTemplate(t *testing.T) {
	RegisterFunctions(template.FuncMap{
		"test": func() string {
			return "foo"
		},
	})

	text := "{{ test }} {{ $users := makeSlice \"2222\"}}{{ $users }} {{ .foo }}"

	temp, err := CompileTemplate(text)
	require.NoError(t, err)
	finalText, _ := EvalTemplate(temp, map[string]string{"foo": "bar"})

	assert.Equal(t, "foo [2222] bar", finalText)
}
