package bot

import (
	"github.com/innogames/slack-bot/bot/stats"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/server"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var linkRegexp = regexp.MustCompile(`<\S+?\|(.*?)>`)
var cleanMessage = strings.NewReplacer(
	"‘", "'",
	"’", "'",
)

// NewBot created main Bot struct which holds the slack connection and dispatch messages to commands
func NewBot(cfg config.Config, slackClient *client.Slack, logger *log.Logger, commands *Commands) *Bot {
	return &Bot{
		config:       cfg,
		slackClient:  slackClient,
		logger:       logger,
		commands:     commands,
		allowedUsers: map[string]string{},
		userLocks:    map[string]*sync.Mutex{},
	}
}

type Bot struct {
	config       config.Config
	slackClient  *client.Slack
	logger       *log.Logger
	auth         *slack.AuthTestResponse
	commands     *Commands
	allowedUsers map[string]string
	server       *server.Server
	userLocks    map[string]*sync.Mutex
}

// Init establishes the slack connection and load allowed users
func (b *Bot) Init() (err error) {
	if b.config.Slack.Token == "" {
		return errors.Errorf("No slack.token provided in config!")
	}

	b.logger.Info("Connecting to slack...")
	b.auth, err = b.slackClient.AuthTest()
	if err != nil {
		return errors.Wrap(err, "auth error")
	}
	client.BotUserID = b.auth.UserID

	err = b.loadChannels()
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

		b.logger.Infof("Auto joined channels: %s", strings.Join(b.config.Slack.AutoJoinChannels, ", "))
	}

	if b.config.Server.IsEnabled() {
		b.server = server.NewServer(b.config.Server, b.logger, b.slackClient, b.allowedUsers)
		go b.server.StartServer()
	}

	go b.slackClient.RTM.ManageConnection()

	b.logger.Infof("Loaded %d allowed users and %d channels", len(b.allowedUsers), len(client.Channels))
	b.logger.Infof("Bot user: %s with ID %s on workspace %s", b.auth.User, b.auth.UserID, b.auth.URL)
	b.logger.Infof("Initialized %d commands", b.commands.Count())

	return nil
}

func (b *Bot) loadChannels() error {
	var err error
	var cursor string
	var chunkedChannels []slack.Channel

	// in CLI context we don't have to channels
	if b.config.Slack.TestEndpointURL != "" {
		return nil
	}

	client.Channels = make(map[string]string)

	// todo proper pagination
	for err == nil {
		options := &slack.GetConversationsParameters{
			Limit:           500,
			Cursor:          cursor,
			ExcludeArchived: "true",
		}

		chunkedChannels, cursor, err = b.slackClient.GetConversations(options)
		if err != nil {
			return err
		}
		for _, channel := range chunkedChannels {
			client.Channels[channel.ID] = channel.Name
		}
		if cursor == "" {
			break
		}
	}

	return nil
}

// DisconnectRTM will do a clean shutdown and kills all connections
func (b *Bot) DisconnectRTM() error {
	if b.server != nil {
		b.server.Stop()
	}

	return b.slackClient.RTM.Disconnect()
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
func (b *Bot) HandleMessages(kill chan os.Signal) {
	for {
		select {
		case msg := <-b.slackClient.RTM.IncomingEvents:
			// message received from user
			switch message := msg.Data.(type) {
			case *slack.HelloEvent:
				b.logger.Info("Hello, the RTM connection is ready!")
			case *slack.MessageEvent:
				if b.shouldHandleMessage(message) {
					go b.handleMessage(*message, true)
				}
			case *slack.RTMError, *slack.UnmarshallingErrorEvent, *slack.RateLimitEvent, *slack.ConnectionErrorEvent:
				b.logger.Error(msg)
			case *slack.LatencyReport:
				b.logger.Debugf("Current latency: %v\n", message.Value)
			}
		case msg := <-client.InternalMessages:
			// e.g. triggered by "delay" or "macro" command. They are still executed in original event context
			// -> will post in same channel as the user posted the original command
			msg.InternalMessage = true
			event := msg.ToSlackEvent()
			go b.handleMessage(event, false)
		case <-kill:
			b.DisconnectRTM()
			b.logger.Warn("Shutdown!")
			return
		}
	}
}

func (b *Bot) shouldHandleMessage(event *slack.MessageEvent) bool {
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

// remove @Bot prefix of message and cleanup
func (b *Bot) trimMessage(msg string) string {
	msg = strings.ReplaceAll(msg, "<@"+b.auth.UserID+">", "")
	msg = cleanMessage.Replace(msg)

	return strings.TrimSpace(msg)
}

// handleMessage process the incoming message and respond appropriately
func (b *Bot) handleMessage(event slack.MessageEvent, fromUserContext bool) {
	event.Text = b.trimMessage(event.Text)

	// remove links from incoming messages. for internal ones they might be wanted, as they contain valid links with texts
	if !fromUserContext {
		event.Text = linkRegexp.ReplaceAllString(event.Text, "$1")
	}

	if event.Text == "" {
		return
	}

	start := time.Now()
	logger := b.getUserBasedLogger(event)

	// send "Bot is typing" command
	if b.slackClient.RTM != nil {
		b.slackClient.RTM.SendMessage(b.slackClient.RTM.NewTypingMessage(event.Channel))
	}

	// prevent messages from one user processed in parallel (usual + internal ones)
	lock := b.getUserLock(event.User)
	defer lock.Unlock()

	stats.IncreaseOne(stats.TotalCommands)

	_, existing := b.allowedUsers[event.User]
	if !existing && !fromUserContext && b.config.Slack.TestEndpointURL == "" {
		logger.Errorf("user %s is not allowed to execute message (missing in 'allowed_users' section): %s", event.User, event.Text)
		// todo pass imploded cfg.AdminUsers here...if set
		b.slackClient.Reply(event, "Sorry, you are not whitelisted yet. Please ask the slack-bot admin to get access.")
		stats.IncreaseOne(stats.UnauthorizedCommands)
		return
	}

	if !b.commands.Run(event) {
		logger.Infof("Unknown command: %s", event.Text)
		stats.IncreaseOne(stats.UnknownCommands)
		b.sendFallbackMessage(event)
	}

	logger.
		WithField("duration", util.FormatDuration(time.Since(start))).
		Infof("handled message: %s", event.Text)
}
