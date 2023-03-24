package bot

// this file contains the main code to handle a message from users

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
)

var (
	linkRegexp = regexp.MustCompile(`<\S+?\|(.*?)>`)

	// clean copy&paste crap from Mac etc
	cleanMessage = strings.NewReplacer(
		"‘", "'",
		"’", "'",
		"“", "\"",
		"”", "\"",
		"\u00a0", " ", // NO-BREAK SPACE
	)
)

// HandleMessage is the entry point for incoming slack messages:
// - checks if the message is relevant (direct message to bot or mentioned via @bot)
// - is the user allowed to interact with the bot?
// - find the matching command and execute it
func (b *Bot) HandleMessage(message *slack.MessageEvent) {
	if b.canHandleMessage(message) {
		go b.ProcessMessage(msg.FromSlackEvent(message), true)
	}
}

// check if a user message was targeted to @bot or is a direct message to the bot. We also block traffic from other bots.
func (b *Bot) canHandleMessage(event *slack.MessageEvent) bool {
	// exclude all Bot traffic
	if event.User == "" || event.User == b.auth.UserID || event.SubType == slack.MsgSubTypeBotMessage {
		return false
	}

	// <@Bot> was mentioned in a public channel
	if strings.Contains(event.Text, "<@"+b.auth.UserID+">") {
		return true
	}

	// Direct message channels always starts with 'D'
	if slackutilsx.DetectChannelType(event.Channel) == slackutilsx.CTypeDM {
		return true
	}

	return false
}

// remove @Bot prefix of message and cleans unwanted characters from the message
func (b *Bot) cleanMessage(text string, fromUserContext bool) string {
	text = strings.ReplaceAll(text, "<@"+b.auth.UserID+">", "")
	text = cleanMessage.Replace(text)

	text = strings.Trim(text, "*")
	text = strings.TrimSpace(text)

	// remove links from incoming messages. for internal ones they might be wanted, as they contain valid links with texts
	if fromUserContext {
		text = linkRegexp.ReplaceAllString(text, "$1")
	}

	return text
}

// ProcessMessage process the incoming message and respond appropriately
func (b *Bot) ProcessMessage(message msg.Message, fromUserContext bool) {
	message.Text = b.cleanMessage(message.Text, fromUserContext)
	if message.Text == "" {
		return
	}

	start := time.Now()
	logger := b.getUserBasedLogger(message)

	// send "Bot is typing" command
	if b.slackClient.RTM != nil {
		b.slackClient.RTM.SendMessage(b.slackClient.RTM.NewTypingMessage(message.Channel))
	}

	// prevent messages from one user processed in parallel (usual + internal ones)
	if message.Done == nil {
		lock := getUserLock(message.User)
		defer lock.Unlock()
	}

	stats.IncreaseOne(stats.TotalCommands)

	// check if user is allowed to interact with the bot
	existing := b.allowedUsers.Contains(message.User)
	if !existing && fromUserContext && !b.config.Slack.IsFakeServer() {
		_, userName := client.GetUserIDAndName(message.User)

		logger.Errorf("user %s (%s) is not allowed to execute message (missing in 'allowed_users' section): %s", userName, message.User, message.Text)

		errorMessage := "Sorry, you are not whitelisted in the config yet."
		if len(b.config.AdminUsers) > 0 {
			errorMessage += " Please ask a slack-bot admin to get access: "
			for _, admin := range b.config.AdminUsers {
				adminID, _ := client.GetUserIDAndName(admin)
				errorMessage += fmt.Sprintf(
					"<@%s>",
					adminID,
				)
			}
		}
		b.slackClient.SendMessage(message, errorMessage)
		b.slackClient.AddReaction("❌", message)

		stats.IncreaseOne(stats.UnauthorizedCommands)
		return
	}

	// actual command execution!
	var commandName string
	var match bool
	if match, commandName = b.commands.RunWithName(message); !match {
		logger.Infof("Unknown command: %s", message.Text)
		stats.IncreaseOne(stats.UnknownCommands)
		b.sendFallbackMessage(message)
	}

	// mark the message as handled...if someone needs this information
	if message.Done != nil {
		message.Done.Done()
	}

	if commandName != "" {
		stats.IncreaseOne("handled_" + strings.ReplaceAll(commandName, ".", "_"))
	}

	logFields := log.Fields{
		// needed time of the actual command...until here
		"duration": util.FormatDuration(time.Since(start)),
		"command":  commandName,
	}

	// log the whole time from: client -> slack server -> bot server -> handle message
	if !message.IsInternalMessage() {
		logFields["durationWithLatency"] = util.FormatDuration(time.Since(message.GetTime()))
	}

	logger.
		WithFields(logFields).
		Infof("handled message: %s", message.Text)
}
