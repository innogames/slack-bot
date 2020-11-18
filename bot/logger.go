package bot

import (
	"os"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// GetLogger provides logger instance for the given config
func GetLogger(cfg config.Config) *log.Logger {
	level, _ := log.ParseLevel(cfg.Logger.Level)

	log.SetOutput(os.Stdout)
	log.SetLevel(level)

	logger := log.New()
	if cfg.Logger.File != "" {
		logger.AddHook(lfshook.NewHook(
			cfg.Logger.File,
			&log.TextFormatter{},
		))
	}

	return logger
}

// get a log.Entry with some user related fields
func (b Bot) getUserBasedLogger(event slack.MessageEvent) *log.Entry {
	_, username := client.GetUser(event.User)

	channel := ""
	if event.Channel != "" && event.Channel[0] == 'D' {
		channel = "@" + username
	} else {
		_, channel = client.GetChannel(event.Channel)
	}
	if event.SubType == TypeInternal {
		channel += " (internal)"
	}

	return b.logger.
		WithField("channel", channel).
		WithField("user", username)
}
