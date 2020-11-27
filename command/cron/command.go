package cron

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	cronLib "github.com/robfig/cron/v3"
)

// NewCronCommand registers cron which are configurable in the yaml config
func NewCronCommand(slackClient client.SlackClient, logger *logrus.Logger, crons []config.Cron) bot.Command {
	if len(crons) == 0 {
		return nil
	}

	cron := cronLib.New()
	cmd := &command{slackClient, crons, cron, logger}

	for _, cronCommand := range crons {
		_, err := cron.AddFunc(cronCommand.Schedule, cmd.getCallback(cronCommand))
		if err != nil {
			logger.Error(err)
		}
	}

	cron.Start()
	logger.Infof("Initialized %d crons", len(crons))

	return cmd
}

type command struct {
	slackClient client.SlackClient
	cfg         []config.Cron
	cron        *cronLib.Cron
	logger      *logrus.Logger
}

func (c *command) getCallback(cron config.Cron) func() {
	// todo validate template before execution. but this is tricky as some functions gets registered later on...
	return func() {
		for _, commandTemplate := range cron.Commands {
			command, err := util.CompileTemplate(commandTemplate)
			if err != nil {
				c.logger.Error(err)
				continue
			}
			text, err := util.EvalTemplate(command, util.Parameters{})
			if err != nil {
				c.logger.Error(err)
				continue
			}

			newMessage := slack.MessageEvent{}
			newMessage.User = "cron"
			newMessage.Channel, _ = client.GetChannel(cron.Channel)
			newMessage.Text = text
			client.InternalMessages <- msg.FromSlackEvent(newMessage)
		}
	}
}
