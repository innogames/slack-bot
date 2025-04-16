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

	t.Run("makeMap", func(t *testing.T) {
		fn := functions["makeMap"].(func(args ...any) (map[string]any, error))

		t.Run("make map successfully", func(t *testing.T) {
			actual, err := fn("foo", 1, "bar", true, "baz", "xyz")
			require.NoError(t, err)
			assert.Equal(t, map[string]any{"foo": 1, "bar": true, "baz": "xyz"}, actual)
		})

		t.Run("fail with wrong amount of arguments", func(t *testing.T) {
			actual, err := fn("foo", 1, "bar", true, "baz")
			require.EqualError(t, err, "makeMap: expected alternating key-value pairs as arguments")
			assert.Nil(t, actual)
		})

		t.Run("fail with key not a string", func(t *testing.T) {
			actual, err := fn("foo", 1, "bar", true, 42, "xyz")
			require.EqualError(t, err, "makeMap: arg at index 4: key must be string")
			assert.Nil(t, actual)
		})
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
