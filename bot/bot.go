package bot

import (
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/interaction"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// TypeInternal is only used internally to identify internal slack messages.
// @deprecated do not use it anymore
const TypeInternal = "internal"

var linkRegexp = regexp.MustCompile("<[^\\s]+?\\|(.*?)>")

// Handler is the main bot interface
type Handler interface {
	HandleMessages(kill chan os.Signal)
}

// NewBot created main bot struct which holds the slack connection and dispatch messages to commands
func NewBot(cfg config.Config, slackClient *client.Slack, logger *log.Logger, commands *Commands) bot {
	return bot{
		config:       cfg,
		slackClient:  slackClient,
		logger:       logger,
		commands:     commands,
		allowedUsers: map[string]string{},
		userLocks:    map[string]*sync.Mutex{},
	}
}

type bot struct {
	config       config.Config
	slackClient  *client.Slack
	logger       *log.Logger
	auth         *slack.AuthTestResponse
	commands     *Commands
	allowedUsers map[string]string
	server       *interaction.Server
	userLocks    map[string]*sync.Mutex
}

// Init establishes the slack connection and load allowed users
func (b *bot) Init() (err error) {
	if b.config.Slack.Token == "" {
		return errors.Errorf("No slack.token provided in config!")
	}

	b.logger.Infof("Connecting to slack...")
	b.auth, err = b.slackClient.AuthTest()
	if err != nil {
		return errors.Wrap(err, "auth error")
	}

	go b.slackClient.ManageConnection()

	channels, err := b.slackClient.GetChannels(true)
	if err != nil {
		return errors.Wrap(err, "error while fetching public channels")
	}
	client.Channels = make(map[string]string, len(channels))
	for _, channel := range channels {
		client.Channels[channel.ID] = channel.Name
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
		b.server = interaction.NewServer(b.config.Server, b.logger, b.slackClient, b.allowedUsers)
		go b.server.StartServer()
	}

	b.logger.Infof("Loaded %d allowed users and %d channels", len(b.allowedUsers), len(client.Channels))
	b.logger.Infof("bot user: %s with ID: %s", b.auth.User, b.auth.UserID)
	b.logger.Infof("Initialized %d commands", b.commands.Count())

	return nil
}

// Disconnect will do a clean shutdown and kills all connections
func (b *bot) Disconnect() error {
	if b.server != nil {
		b.server.Stop()
	}

	return b.slackClient.Disconnect()
}

// load the public channels and list of all users from current space
func (b *bot) loadSlackData() error {
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
func (b bot) HandleMessages(kill chan os.Signal) {
	for {
		select {
		case msg := <-b.slackClient.IncomingEvents:
			// message received from user
			switch message := msg.Data.(type) {
			case *slack.MessageEvent:
				if b.shouldHandleMessage(message) {
					go b.handleMessage(*message)
				}
			case *slack.RTMError, *slack.UnmarshallingErrorEvent, *slack.RateLimitEvent:
				b.logger.Error(msg)
			case *slack.LatencyReport:
				b.logger.Debugf("Current latency: %v\n", message.Value)
			}
		case msg := <-client.InternalMessages:
			// e.g. triggered by "delay" or "macro" command. They are still executed in original event context
			// -> will post in same channel as the user posted the original command
			msg.SubType = TypeInternal
			go b.handleMessage(msg)
		case <-kill:
			b.Disconnect()
			b.logger.Warnf("Shutdown!")
			return
		}
	}
}

func (b bot) shouldHandleMessage(event *slack.MessageEvent) bool {
	// exclude all bot traffic
	if event.BotID != "" || event.User == "" || event.User == b.auth.UserID || event.SubType == "bot_message" {
		return false
	}

	// <@bot> was mentioned in a public channel
	if strings.Contains(event.Text, "<@"+b.auth.UserID+">") {
		return true
	}

	// Direct message channels always starts with 'D'
	if event.Channel[0] == 'D' {
		return true
	}

	return false
}

// remove @bot prefix of message and cleanup
func (b bot) trimMessage(msg string) string {
	msg = strings.Replace(msg, "<@"+b.auth.UserID+">", "", 1)
	msg = strings.Replace(msg, "‘", "'", -1)
	msg = strings.Replace(msg, "’", "'", -1)
	msg = linkRegexp.ReplaceAllString(msg, "$1")

	return strings.TrimSpace(msg)
}

// handleMessage process the incoming message and respond appropriately
func (b bot) handleMessage(event slack.MessageEvent) {
	event.Text = b.trimMessage(event.Text)
	if event.Text == "" {
		return
	}

	start := time.Now()
	logger := b.getLogger(event)

	// send "bot is typing" command
	b.slackClient.RTM.SendMessage(b.slackClient.NewTypingMessage(event.Channel))

	lock := b.getUserLock(event.User)
	defer lock.Unlock()

	_, existing := b.allowedUsers[event.User]
	if !existing && event.SubType != TypeInternal && b.config.Slack.TestEndpointUrl == "" {
		logger.Errorf("user %s is not allowed to execute message: %s", event.User, event.Text)
		b.slackClient.Reply(event, "Sorry, you are not whitelisted yet. Please ask the slack-bot admin to get access.")
		return
	}

	if !b.commands.Run(event) {
		logger.Infof("Unknown command: %s", event.Text)
		b.sendFallbackMessage(event)
	}

	logger.
		WithField("duration", util.FormatDuration(time.Now().Sub(start))).
		Infof("handled message: %s", event.Text)
}
