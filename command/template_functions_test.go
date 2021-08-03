package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/util"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/mocks"
)

func TestTemplateFunctions(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := []config.TemplateFunction{
		{
			Name:     "staticValue",
			Template: "foobar",
		},
		{
			Name:      "simpleVariable",
			Template:  "foo {{.Name}}",
			Arguments: []string{"Name"},
		},
		{
			Name:     "callOtherFunction",
			Template: "{{staticValue}}",
		},
		{
			Name:     "invalidFunction",
			Template: "{{invalidCall}}",
		},
	}

	command := bot.Commands{}
	command.AddCommand(NewCommands(base, []config.Command{}, cfg))

	testCases := []struct {
		Template string
		Expected string
	}{
		{
			Template: "{{staticValue}}",
			Expected: "foobar",
		},
		{
			Template: "{{simpleVariable \"param1\"}}",
			Expected: "foo param1",
		},
		{
			Template: "{{simpleVariable}}",
			Expected: "simpleVariable: not enough parameters",
		},
		{
			Template: "{{callOtherFunction}}",
			Expected: "foobar",
		},
		{
			Template: "{{invalidFunction}}",
			Expected: "template: {{invalidCall}}:1: function \"invalidCall\" not defined",
		},
	}

	t.Run("Render template functions", func(t *testing.T) {
		for _, testCase := range testCases {
			mocks.AssertRenderedTemplate(
				t,
				testCase.Template,
				testCase.Expected,
				util.Parameters{},
			)
		}
	})
}
