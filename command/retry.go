package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
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
		matcher.NewTextMatcher(`retry`, c.Retry),
		matcher.NewTextMatcher(`repeat`, c.Retry),
		matcher.NewRegexpMatcher(`<?https://(.*)slack.com/archives/(?P<channel>\w+)/p(?P<timestamp>\d{16})>?`, c.SlackMessage),
		matcher.WildcardMatcher(c.Store),
	)
}

// retry the last stored message
func (c *retryCommand) Retry(match matcher.Result, message msg.Message) {
	key := message.GetUniqueKey()

	var lastCommand string
	_ = storage.Read(storageKey, key, &lastCommand)
	if lastCommand != "" {
		c.SendMessage(message, fmt.Sprintf("Executing command: %s", lastCommand))

		client.HandleMessage(message.WithText(lastCommand))
	} else {
		c.SendMessage(message, "Sorry, no history found.")
	}
}

// store any message in the storage to get repeatable
func (c *retryCommand) Store(ref msg.Ref, text string) bool {
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
func (c *retryCommand) SlackMessage(match matcher.Result, message msg.Message) {
	channel := match.GetString("channel")
	timestamp := match.GetString("timestamp")

	m, err := c.SlackClient.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Latest:    timestamp[0:10] + "." + timestamp[10:],
		Inclusive: true,
		Limit:     1,
	})
	if err != nil {
		c.ReplyError(message, err)
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
	c.AddReaction("white_check_mark", message)
	client.HandleMessage(historyMessage)
}

func (c *retryCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "repeat",
			Description: "repeat the last executed command",
			Examples: []string{
				"retry",
				"repeat",
			},
		},
	}
}
