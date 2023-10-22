package cron

import (
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	cronLib "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// NewCronCommand registers cron which are configurable in the yaml config
func NewCronCommand(base bot.BaseCommand, crons []config.Cron) bot.Command {
	if len(crons) == 0 {
		return nil
	}

	cron := cronLib.New(
		cronLib.WithLogger(newCronLogger()),
	)
	cmd := &command{base, crons, cron}

	for _, cronCommand := range crons {
		_, err := cron.AddFunc(cronCommand.Schedule, cmd.getCallback(cronCommand))
		if err != nil {
			log.Error(err)
		}
	}

	return cmd
}

type command struct {
	bot.BaseCommand
	cfg  []config.Cron
	cron *cronLib.Cron
}

// RunAsync provide proper Cron start/stop in a async context
func (c *command) RunAsync(ctx *util.ServerContext) {
	ctx.RegisterChild()
	defer ctx.ChildDone()

	c.cron.Start()
	log.Infof("Initialized %d crons", len(c.cfg))

	<-ctx.Done()
}

func (c *command) getCallback(cron config.Cron) func() {
	return func() {
		for _, commandTemplate := range cron.Commands {
			command, err := util.CompileTemplate(commandTemplate)
			if err != nil {
				log.Error(err)
				continue
			}
			text, err := util.EvalTemplate(command, util.Parameters{})
			if err != nil {
				log.Error(err)
				continue
			}

			for _, line := range strings.Split(text, "\n") {
				newMessage := msg.Message{}
				newMessage.User = "cron"
				newMessage.Channel, _ = client.GetChannelIDAndName(cron.Channel)
				newMessage.Text = line
				client.HandleMessage(newMessage)
			}
		}
	}
}
