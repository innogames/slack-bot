package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/slack-go/slack"
)

// help category to group all Jenkins command
var category = bot.Category{
	Name:        "Cloud-AWS",
	Description: "Interaction with defined aws lambdas",
	HelpURL:     "https://github.com/innogames/slack-bot",
}

type lambdaCommand struct {
	awsCommand
	service *lambda.Lambda
	cfg     []config.Lambda
}

// NewAwsCommand is a command to interact with aws resources
func newLambdaCommands(cfg []config.Lambda, base awsCommand) bot.Command {
	svc := lambda.New(base.session)
	return &lambdaCommand{base, svc, cfg}
}

func (c *lambdaCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("aws show", c.showLambdas),
		matcher.NewRegexpMatcher(`aws run (?P<LAMBDA>[\w|\d]+) (?P<PARAMS>[\w|\d|\W]+)`, c.invoke),
	)
}

func (c *lambdaCommand) showLambdas(match matcher.Result, message msg.Message) {
	// show list
	blocks := []slack.Block{}
	for _, v := range c.cfg {
		var name string = v.Name
		if v.Alias != "" {
			name = v.Alias
		}
		var description = "no defined description"
		if v.Description != "" {
			description = v.Description
		}
		txtObj := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("\"%s\": %s\n", name, description), false, false)
		if err := txtObj.Validate(); err != nil {
			fmt.Println(err.Error())
			return
		}
		blocks = append(blocks, slack.NewSectionBlock(txtObj, nil, nil))
	}
	c.SendBlockMessage(message, blocks)
}

func (c *lambdaCommand) invoke(match matcher.Result, message msg.Message) {
	return
}

func (c *lambdaCommand) GetHelp() []bot.Help {
	examples := []string{
		"aws test-lambda a,b,c",
	}

	help := make([]bot.Help, 0)
	help = append(help, bot.Help{
		Command:     "aws run <lambda-name> <params>",
		Description: "invoke selected AWS lambda with given parameters",
		Examples:    examples,
		Category:    category,
	})

	return help
}
