package bot

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/server"
	"github.com/innogames/slack-bot/bot/stats"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"regexp"
	"strings"
	"sync"
	"time"
)

var linkRegexp = regexp.MustCompile(`<\S+?\|(.*?)>`)
var cleanMessage = strings.NewReplacer(
	"‘", "'",
	"’", "'",
)

// NewBot created main Bot struct which holds the slack connection and dispatch messages to commands
func NewBot(cfg config.Config, slackClient *client.Slack, commands *Commands) *Bot {
	return &Bot{
		config:       cfg,
		slackClient:  slackClient,
		commands:     commands,
		allowedUsers: map[string]string{},
		userLocks:    map[string]*sync.Mutex{},
	}
}

type Bot struct {
	config       config.Config
	slackClient  *client.Slack
	auth         *slack.AuthTestResponse
	commands     *Commands
	server       *server.Server
	allowedUsers map[string]string
	userLocks    map[string]*sync.Mutex
}

// Init establishes the slack connection and load allowed users
func (b *Bot) Init() (err error) {
	if b.config.Slack.Token == "" {
		return errors.New("no slack.token provided in config")
	}

	log.Info("Connecting to slack...")
	b.auth, err = b.slackClient.AuthTest()
	if err != nil {
		return errors.Wrap(err, "auth error")
	}
	client.BotUserID = b.auth.UserID
	client.Channels, err = b.loadChannels()
	if err != nil {
		return errors.Wrap(err, "error while fetching public channels")
	}

	err = b.loadSlackData()
	if err != nil {
		return err
	}

	if len(b.config.Slack.AutoJoinChannels) > 0 {
		for _, channel := range b.config.Slack.AutoJoinChannels {
			_, err := b.slackClient.JoinChannel(channel)
			if err != nil {
				return err
			}
		}

		log.Infof("Auto joined channels: %s", strings.Join(b.config.Slack.AutoJoinChannels, ", "))
	}

	if b.config.Server.IsEnabled() {
		b.server = server.NewServer(b.config.Server, b.slackClient)
		go b.server.StartServer()
	}

	go b.slackClient.RTM.ManageConnection()

	log.Infof("Loaded %d allowed users and %d channels", len(b.allowedUsers), len(client.Channels))
	log.Infof("Bot user: %s with ID %s on workspace %s", b.auth.User, b.auth.UserID, b.auth.URL)
	log.Infof("Initialized %d commands", b.commands.Count())

	return nil
}

func (b *Bot) loadChannels() (map[string]string, error) {
	var err error
	var cursor string
	var chunkedChannels []slack.Channel

	channels := make(map[string]string)

	// in CLI context we don't have to channels
	if b.config.Slack.TestEndpointURL != "" {
		return channels, nil
	}

	for {
		options := &slack.GetConversationsParameters{
			Limit:           1000,
			Cursor:          cursor,
			ExcludeArchived: "true",
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

// DisconnectRTM will do a clean shutdown and kills all connections
func (b *Bot) DisconnectRTM() error {
	if b.server != nil {
		return b.server.Stop()
	}

	if b.slackClient.RTM != nil {
		return b.slackClient.RTM.Disconnect()
	}

	return nil
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

// HandleMessages is blocking method to handle new incoming events
func (b *Bot) HandleMessages(ctx *util.ServerContext) {
	ctx.RegisterChild()
	defer ctx.ChildDone()

	for {
		select {
		case event := <-b.slackClient.RTM.IncomingEvents:
			// message received from user
			switch message := event.Data.(type) {
			case *slack.HelloEvent:
				log.Info("Hello, the RTM connection is ready!")
			case *slack.MessageEvent:
				if b.canHandleMessage(message) {
					go b.handleMessage(msg.FromSlackEvent(*message), true)
				}
			case *slack.RTMError, *slack.UnmarshallingErrorEvent, *slack.RateLimitEvent, *slack.ConnectionErrorEvent:
				log.Error(event)
			case *slack.LatencyReport:
				log.Debugf("Current latency: %v", message.Value)
			}
		case message := <-client.InternalMessages:
			// e.g. triggered by "delay" or "macro" command. They are still executed in original event context
			// -> will post in same channel as the user posted the original command
			message.InternalMessage = true
			go b.handleMessage(message, false)
		case <-ctx.Done():
			if err := b.DisconnectRTM(); err != nil {
				log.Error(err)
			}
			return
		}
	}
}

// check if a user message was targeted to @bot or is a direct message to the bot
func (b *Bot) canHandleMessage(event *slack.MessageEvent) bool {
	// exclude all Bot traffic
	if event.User == "" || event.User == b.auth.UserID || event.SubType == "bot_message" {
		return false
	}

	// <@Bot> was mentioned in a public channel
	if strings.Contains(event.Text, "<@"+b.auth.UserID+">") {
		return true
	}

	// Direct message channels always starts with 'D'
	if event.Channel[0] == 'D' {
		return true
	}

	return false
}

// remove @Bot prefix of message and cleans the message
func (b *Bot) cleanMessage(msg string, fromUserContext bool) string {
	msg = strings.ReplaceAll(msg, "<@"+b.auth.UserID+">", "")
	msg = cleanMessage.Replace(msg)

	msg = strings.TrimSpace(msg)

	// remove links from incoming messages. for internal ones they might be wanted, as they contain valid links with texts
	if fromUserContext {
		msg = linkRegexp.ReplaceAllString(msg, "$1")
	}

	return msg
}

// handleMessage process the incoming message and respond appropriately
func (b *Bot) handleMessage(message msg.Message, fromUserContext bool) {
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
	lock := b.getUserLock(message.User)
	defer lock.Unlock()

	stats.IncreaseOne(stats.TotalCommands)

	_, existing := b.allowedUsers[message.User]
	if !existing && fromUserContext && b.config.Slack.TestEndpointURL == "" {
		logger.Errorf("user %s is not allowed to execute message (missing in 'allowed_users' section): %s", message.User, message.Text)
		b.slackClient.SendMessage(message, fmt.Sprintf(
			"Sorry, you are not whitelisted yet. Please ask a slack-bot admin to get access: %s",
			strings.Join(b.config.AdminUsers, ", "),
		))
		stats.IncreaseOne(stats.UnauthorizedCommands)
		return
	}

	if !b.commands.Run(message) {
		logger.Infof("Unknown command: %s", message.Text)
		stats.IncreaseOne(stats.UnknownCommands)
		b.sendFallbackMessage(message)
	}

	logger.
		WithField("duration", util.FormatDuration(time.Since(start))).
		Infof("handled message: %s", message.Text)
}
