package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client/cloud/aws"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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

	lambda := setAWSLambda(cfg)

	commands.AddCommand(
		newLambdaCommands(lambda, awsBase),
	)

	return commands
}

func setAWSLambda(cfg config.Aws) []config.Lambda {
	c := []config.Lambda{}

	for _, v := range cfg.Lambda {
		c = append(c, config.Lambda{
			Name:        v.Name,
			Alias:       v.Alias,
			Description: v.Description,
		})
	}
	return c
}
