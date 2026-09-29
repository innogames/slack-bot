// Package aws is a slack-bot plugin to interact with AWS resources: CloudFront and ECS
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/pkg/errors"
)

// help category to group all AWS commands
var category = bot.Category{
	Name:        "Cloud-AWS",
	Description: "Interact with AWS resources: CF && ECS",
}

// base command to access Slack+AWS directly
type awsCommand struct {
	bot.BaseCommand
	cfg aws.Config
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name:  "aws",
		Setup: setup,
	})
}

// setup returns the AWS commands, the AWS credentials are loaded by the default AWS SDK credential chain
func setup(ctx *bot.PluginContext) (bot.Commands, error) {
	var commands bot.Commands

	cfg := Config{}
	if err := ctx.LoadConfig(&cfg); err != nil {
		return commands, err
	}

	awsConfig, err := getAWSConfig(context.Background())
	if err != nil {
		return commands, errors.Wrap(err, "error while getting aws sdk config")
	}

	base := awsCommand{
		ctx.BaseCommand(),
		awsConfig,
	}

	commands.AddCommand(
		newCloudFrontCommands(cfg.CloudFront, base),
		newEcsCommands(base),
	)

	return commands, nil
}
