package command

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const storageKey = "user_history"

// NewRetryCommand store the history of the commands of the user sent to the bot in a local storage
// With "retry" the most recent command of the channel will be repeated
func NewRetryCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	return &retryCommand{
		base,
		cfg,
	}
}

type retryCommand struct {
	bot.BaseCommand
	cfg *config.Config
}

func (c *retryCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher(`retry`, c.retry),
		matcher.NewTextMatcher(`repeat`, c.retry),
		matcher.NewRegexpMatcher(`<?https://(.*?)\.slack\.com/archives/(?P<channel>\w+)/p(?P<timestamp>\d{16})>?`, c.slackMessage),
		matcher.WildcardMatcher(c.storeLastCommand),
	)
}

// retry the last stored message
func (c *retryCommand) retry(match matcher.Result, message msg.Message) {
	key := message.GetUniqueKey()

	var lastCommand string
	err := storage.Read(storageKey, key, &lastCommand)
	if err != nil {
		log.Warn(errors.Wrap(err, "error while loading user history"))
	}

	if lastCommand != "" {
		c.SendMessage(message, fmt.Sprintf("Executing command: %s", lastCommand))

		client.HandleMessage(message.WithText(lastCommand))
	} else {
		c.SendMessage(message, "Sorry, no history found.")
	}
}

// store any message in the storage to get repeatable
func (c *retryCommand) storeLastCommand(ref msg.Ref, text string) bool {
	if ref.IsInternalMessage() {
		return false
	}

	key := ref.GetUniqueKey()
	err := storage.Write(storageKey, key, text)
	if err != nil {
		log.Error(err)
	}

	return false
}

// re-execute a slack message
func (c *retryCommand) slackMessage(match matcher.Result, message msg.Message) {
	channel := match.GetString("channel")
	timestamp := match.GetString("timestamp")

	m, err := c.SlackClient.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Latest:    timestamp[0:10] + "." + timestamp[10:],
		Inclusive: true,
		Limit:     1,
	})
	if err != nil {
		c.ReplyError(message, fmt.Errorf("can't load original message: %w", err))
		return
	}
	historyMessage := msg.FromSlackEvent(&slack.MessageEvent{
		Msg: m.Messages[0].Msg,
	})
	historyMessage.Channel = channel
	// check if the user is allowed to re-post the message: either the author of the message or an admin
	if historyMessage.User != message.User && !c.cfg.AdminUsers.Contains(message.User) {
		c.SendMessage(message, "this is not your message")
		return
	}
	c.AddReaction("✅", message)
	client.HandleMessage(historyMessage)
}

func (c *retryCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "repeat (or retry)",
			Description: "retires the last executed command",
			Category:    helperCategory,
			Examples: []string{
				"retry",
				"repeat",
			},
		},
	}
}
