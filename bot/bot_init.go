package bot

// this file contains everything important to bootstrap the Bot like connecting to Slack and loading the commands

import (
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// NewBot created main Bot struct which holds the slack connection and dispatch messages to commands
func NewBot(cfg config.Config, slackClient *client.Slack, commands *Commands) *Bot {
	return &Bot{
		config:       cfg,
		slackClient:  slackClient,
		commands:     commands,
		allowedUsers: config.UserMap{},
		locks:        util.NewGroupedLogger(),
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
	locks        util.GroupedLock[string]
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
	client.AllChannels, err = b.loadChannels()
	if err != nil {
		return errors.Wrap(err, "error while fetching public channels")
	}

	err = b.loadSlackData()
	if err != nil {
		return err
	}

	pluginCommands := LoadPlugins(b)
	b.commands.Merge(pluginCommands)

	log.Infof("Loaded %d allowed users and %d channels", len(b.allowedUsers), len(client.AllChannels))
	log.Infof("Bot user: @%s with ID %s on workspace %s", b.auth.User, b.auth.UserID, b.auth.URL)

	pluginCommands := loadPlugins(b, b.slackClient)
	log.Infof("Loaded %d plugin commands", len(pluginCommands.commands))
	b.commands.Merge(pluginCommands)

	commands := b.commands.GetCommandNames()
	log.Infof("Initialized %d commands:", len(commands))
	log.Info(strings.Join(commands, ", "))

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
	// whitelist users by group, can be the plain group-id, @dev-group1
	if len(b.config.Slack.AllowedGroups) > 0 {
		allUserGroups, err := b.slackClient.GetUserGroups()
		if err != nil {
			log.Fatalf("error fetching user groups: %s", err)
		}

		for _, groupID := range b.config.Slack.AllowedGroups {
			for _, group := range allUserGroups {
				if "@"+group.Handle == groupID {
					groupID = group.ID
					break
				}
			}

			group, err := b.slackClient.GetUserGroupMembers(groupID)
			if err != nil {
				return errors.Wrap(err, "error fetching user of group. You need a user token with 'usergroups:read' scope permission")
			}
			log.Infof("Found %d users in group %s", len(group), groupID)
			b.config.AllowedUsers = append(b.config.AllowedUsers, group...)
		}
	}

	// load user list
	allUsers, err := b.slackClient.GetUsers()
	if err != nil {
		return errors.Wrap(err, "error fetching users")
	}

	client.AllUsers = make(config.UserMap, len(allUsers))
	for _, user := range allUsers {
		client.AllUsers[user.ID] = user.Name
		for _, allowedUserName := range b.config.AllowedUsers {
			if allowedUserName == user.Name || allowedUserName == user.ID {
				b.allowedUsers[user.ID] = user.Name
				break
			}
		}
	}

	return nil
}
