package bot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var (
	linkRegexp   = regexp.MustCompile(`<\S+?\|(.*?)>`)
	cleanMessage = strings.NewReplacer(
		"‘", "'",
		"’", "'",
		"“", "\"",
		"”", "\"",
		"\u00a0", " ", // NO-BREAK SPACE
	)
)

// NewBot created main Bot struct which holds the slack connection and dispatch messages to commands
func NewBot(cfg config.Config, slackClient *client.Slack, commands *Commands) *Bot {
	return &Bot{
		config:       cfg,
		slackClient:  slackClient,
		commands:     commands,
		allowedUsers: config.UserMap{},
	}
}

// Bot is the main object which is holding the connection to Slack and all possible commands
// it also registers the listener and handles topics like authentication and logging
type Bot struct {
	config       config.Config
	slackClient  *client.Slack
	auth         *slack.AuthTestResponse
	commands     *Commands
	allowedUsers config.UserMap
}

// Init establishes the slack connection and load allowed users
func (b *Bot) Init() (err error) {
	// set global default timezone
	if b.config.Timezone != "" {
		time.Local, err = time.LoadLocation(b.config.Timezone)
		if err != nil {
			return err
		}
	}

	log.Info("Connecting to slack...")
	b.auth, err = b.slackClient.AuthTest()
	if err != nil {
		return errors.Wrap(err, "auth error")
	}
	client.AuthResponse = *b.auth
	client.Channels, err = b.loadChannels()
	if err != nil {
		return errors.Wrap(err, "error while fetching public channels")
	}

	err = b.loadSlackData()
	if err != nil {
		return err
	}

	if b.slackClient.RTM != nil {
		log.Warn("You're using the deprecated Slack RTM API...we prefer using the Socket Mode API")
		go b.slackClient.RTM.ManageConnection()
	}

	log.Infof("Loaded %d allowed users and %d channels", len(b.allowedUsers), len(client.Channels))
	log.Infof("Bot user: %s with ID %s on workspace %s", b.auth.User, b.auth.UserID, b.auth.URL)
	log.Infof("Initialized %d commands", b.commands.Count())

	return nil
}

// loads a list of all public channels
func (b *Bot) loadChannels() (map[string]string, error) {
	var err error
	var cursor string
	var chunkedChannels []slack.Channel

	channels := make(map[string]string)

	// in CLI context we don't have to channels
	if b.config.Slack.IsFakeServer() {
		return channels, nil
	}

	for {
		options := &slack.GetConversationsParameters{
			Limit:           1000,
			Cursor:          cursor,
			ExcludeArchived: true,
		}

		chunkedChannels, cursor, err = b.slackClient.GetConversations(options)
		if err != nil {
			return channels, err
		}
		for _, channel := range chunkedChannels {
			channels[channel.ID] = channel.Name
		}
		if cursor == "" {
			break
		}
	}

	return channels, nil
}

// load the public channels and list of all users from current space
func (b *Bot) loadSlackData() error {
	// whitelist users by group
	for _, groupName := range b.config.Slack.AllowedGroups {
		group, err := b.slackClient.GetUserGroupMembers(groupName)
		if err != nil {
			return errors.Wrap(err, "error fetching user of group. You need a user token with 'usergroups:read' scope permission")
		}
		b.config.AllowedUsers = append(b.config.AllowedUsers, group...)
	}

	// load user list
	allUsers, err := b.slackClient.GetUsers()
	if err != nil {
		return errors.Wrap(err, "error fetching users")
	}

	for _, user := range allUsers {
		for _, allowedUserName := range b.config.AllowedUsers {
			if allowedUserName == user.Name || allowedUserName == user.ID {
				b.allowedUsers[user.ID] = user.Name
				break
			}
		}
	}

	client.Users = b.allowedUsers

	return nil
}

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
	if strings.HasPrefix(event.Channel, "D") {
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
		logger.Errorf("user %s is not allowed to execute message (missing in 'allowed_users' section): %s", message.User, message.Text)
		b.slackClient.SendMessage(message, fmt.Sprintf(
			"Sorry <@%s>, you are not whitelisted yet. Please ask a slack-bot admin to get access: %s",
			message.User,
			strings.Join(b.config.AdminUsers, ", "),
		))
		b.slackClient.AddReaction("❌", message)

		stats.IncreaseOne(stats.UnauthorizedCommands)
		return
	}

	// actual command execution!
	if !b.commands.Run(message) {
		logger.Infof("Unknown command: %s", message.Text)
		stats.IncreaseOne(stats.UnknownCommands)
		b.sendFallbackMessage(message)
	}

	// mark the message as handled...if someone needs this information
	if message.Done != nil {
		message.Done.Done()
	}

	logFields := log.Fields{
		// needed time of the actual command...until here
		"duration": util.FormatDuration(time.Since(start)),
	}

	// log the whole time from: client -> slack server -> bot server -> handle message
	if !message.IsInternalMessage() {
		logFields["durationWithLatency"] = util.FormatDuration(time.Since(message.GetTime()))
	}

	logger.
		WithFields(logFields).
		Infof("handled message: %s", message.Text)
}
