package calendar

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
)

// NewCalendarCommand listen for ical events and executed defined command when a event started
func NewCalendarCommand(cfg []config.Calendar, logger *logrus.Logger) bot.Command {
	if len(cfg) == 0 {
		return nil
	}

	go func() {
		events := WaitForEvents(cfg)
		for event := range events {
			for _, command := range event.Event.Commands {
				macro, _ := util.CompileTemplate(command)
				text, err := util.EvalTemplate(macro, event.Params)
				if err != nil {
					logger.Error(err)
					continue
				}

				message := slack.MessageEvent{}
				message.Channel = event.Event.Channel
				message.Text = text
				message.User = "calendar"
				client.InternalMessages <- message
				logger.Info(message.Text)
			}
		}
	}()

	return nil
}
