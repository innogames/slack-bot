package bot

import (
	"os"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// InitLogger provides logger instance for the given config
func InitLogger(cfg config.Logger) {
	level, err := log.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(os.Stdout)
	log.SetLevel(level)

	if cfg.File != "" {
		log.AddHook(lfshook.NewHook(
			cfg.File,
			&log.TextFormatter{},
		))
	}
}

// get a log.Entry with some user related fields
func (b *Bot) getUserBasedLogger(ref msg.Ref) *log.Entry {
	_, username := client.GetUserIDAndName(ref.GetUser())

	channel := ""
	if ref.GetChannel() != "" && ref.GetChannel()[0] == 'D' {
		channel = "@" + username
	} else {
		_, channel = client.GetChannelIDAndName(ref.GetChannel())
	}

	return log.
		WithField("channel", channel).
		WithField("user", username)
}
