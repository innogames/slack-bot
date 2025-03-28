package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("parse", func(t *testing.T) {
		tests := []struct {
			input    string
			expected Result
		}{
			{
				input:    `option1=value1 option2='value 2' option3="value 3"`,
				expected: Result{"option1": "value1", "option2": "value 2", "option3": "value 3"},
			},
			{
				input:    `option1=value1 option2=value2`,
				expected: Result{"option1": "value1", "option2": "value2"},
			},
			{
				input:    `option1='some value' option2="another value" option3=val`,
				expected: Result{"option1": "some value", "option2": "another value", "option3": "val"},
			},
			{
				input:    `option1=value1`,
				expected: Result{"option1": "value1"},
			},
			{
				input: `option1="value with spaces" option2='single quoted' option3=plain`,
				expected: Result{
					"option1": "value with spaces",
					"option2": "single quoted",
					"option3": "plain",
				},
			},
			{
				input:    `foo=bar test`,
				expected: Result{"foo": "bar", "test": "true"},
			},
			{
				input:    `foo=bar test=true`,
				expected: Result{"foo": "bar", "test": "true"},
			},
			{
				input:    `--foo=bar --test=true`,
				expected: Result{"foo": "bar", "test": "true"},
			},
		}

		for _, test := range tests {
			result := parseOptions(test.input)
			assert.Equal(t, test.expected, result, "For input '%s', expected %v but got %v", test.input, test.expected, result)
		}
	})

	t.Run("Test command", func(t *testing.T) {
		matchTest := []struct {
			command  string
			input    string
			expected bool
			match    Result
		}{
			{"test", "", false, nil},
			{"test", "no test", false, nil},
			{"test", "test foo=bar", true, Result{"foo": "bar"}},
			{"test", "test foo=bar", true, Result{"foo": "bar"}},
			{"test", "test foo=bar jolo", true, Result{"foo": "bar", "jolo": "true"}},
			{"test", "test foo=bar jo='test foo'", true, Result{"foo": "bar", "jo": "test foo"}},
		}

		for _, testCase := range matchTest {
			subject := NewOptionMatcher(testCase.command, []string{"foo", "jolo", "jo"}, testRunner, nil)

			message := msg.Message{}
			message.Text = testCase.input
			run, match := subject.Match(message)
			if testCase.expected {
				assert.NotNil(t, run, testCase.input)
			} else {
				assert.Nil(t, run)
			}
			assert.Equal(t, testCase.match, match, testCase.input)
		}
	})
}
