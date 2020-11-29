package bot

import (
	"github.com/innogames/slack-bot/bot/msg"
	"os"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// GetLogger provides logger instance for the given config
func GetLogger(cfg config.Logger) *log.Logger {
	level, _ := log.ParseLevel(cfg.Level)

	log.SetOutput(os.Stdout)
	log.SetLevel(level)

	logger := log.New()

	if cfg.File != "" {
		logger.AddHook(lfshook.NewHook(
			cfg.File,
			&log.TextFormatter{},
		))
	}

	return logger
}

// get a log.Entry with some user related fields
func (b *Bot) getUserBasedLogger(message msg.Message) *log.Entry {
	_, username := client.GetUser(message.User)

	channel := ""
	if message.Channel != "" && message.Channel[0] == 'D' {
		channel = "@" + username
	} else {
		_, channel = client.GetChannel(message.Channel)
	}

	return b.logger.
		WithField("channel", channel).
		WithField("user", username)
}
