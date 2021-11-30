package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client/cloud/aws"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// help category to group all Jenkins command
var category = bot.Category{
	Name:        "Cloud-AWS",
	Description: "Interact with AWS resources: Clear CF caches",
	HelpURL:     "https://github.com/innogames/slack-bot",
}

// base command to access Slack+Jenkins directly
type awsCommand struct {
	bot.BaseCommand
	session *session.Session
}

// GetCommands will return a list of available Jenkins commands...if the config is set!
func GetCommands(cfg config.Aws, base bot.BaseCommand) bot.Commands {
	var commands bot.Commands

	if !cfg.IsEnabled() {
		return commands
	}

	session, err := aws.GetSession()
	if nil != err {
		log.Error(errors.Wrap(err, "Error while getting aws sdk session"))
		return commands
	}

	awsBase := awsCommand{
		base,
		session,
	}

	distributions := setCloudFrontDistributions(cfg)

	commands.AddCommand(
		newCloudFrontCommands(distributions, awsBase),
	)

	return commands
}

func setCloudFrontDistributions(cfg config.Aws) []config.AwsCfDistribution {
	c := []config.AwsCfDistribution{}

	for _, v := range cfg.CloudFront {
		c = append(c, config.AwsCfDistribution{
			ID:   v.ID,
			Name: v.Name,
		})
	}
	return c
}
