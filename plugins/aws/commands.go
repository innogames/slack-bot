package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
		Name: "aws",
		Init: getCommands,
	})
}

// getCommands will return a list of available AWS commands...if the config is set!
func getCommands(slackClient client.SlackClient, cfg config.Config) bot.Commands {
	var commands bot.Commands

	var pluginCfg Config
	if err := cfg.LoadPlugin("aws", &pluginCfg); err != nil {
		log.Error(errors.Wrap(err, "error while loading aws plugin config"))
		return commands
	}

	if !pluginCfg.IsEnabled() {
		return commands
	}

	ctx := context.Background()
	awsConfig, err := getAWSConfig(ctx)
	if nil != err {
		log.Error(errors.Wrap(err, "Error while getting aws sdk config"))
		return commands
	}

	awsBase := awsCommand{
		bot.BaseCommand{SlackClient: slackClient},
		awsConfig,
	}

	commands.AddCommand(
		newCloudFrontCommands(pluginCfg.CloudFront, awsBase),
		newEcsCommands(awsBase),
	)

	return commands
}
