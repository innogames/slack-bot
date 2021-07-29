package bot

import (
	"os"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
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

	var channel string
	if strings.HasPrefix(ref.GetChannel(), "D") {
		channel = "@" + username
	} else {
		_, channel = client.GetChannelIDAndName(ref.GetChannel())
	}

	return log.
		WithField("channel", channel).
		WithField("user", username)
}
