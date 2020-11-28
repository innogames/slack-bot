package client

//go:generate $GOPATH/bin/mockery --output ../mocks --name SlackClient

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"strings"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// InternalMessages is internal queue of internal messages
var InternalMessages = make(chan msg.Message, 50)

// BotUserID is filled with the slack user id of the bot
var BotUserID = ""

// Users is a lookup from user-id to user-name
var Users map[string]string

// Channels is a map of each channelsId and the name
var Channels map[string]string

// GetSlackClient establishes a RTM connection to the slack server
func GetSlackClient(cfg config.Slack, logger *logrus.Logger) *Slack {
	options := make([]slack.Option, 0)
	if cfg.TestEndpointURL != "" {
		options = append(options, slack.OptionAPIURL(cfg.TestEndpointURL))
	}

	// todo add slack.OptionLog() and proxy to logger
	if cfg.Debug {
		options = append(options, slack.OptionDebug(true))
	}

	rawClient := slack.New(cfg.Token, options...)
	rtm := rawClient.NewRTM()

	return &Slack{Client: rawClient, RTM: rtm, logger: logger, config: cfg}
}

type SlackClient interface {
	// Reply a message to the current channel/user/thread
	Reply(event slack.MessageEvent, text string, options ...slack.MsgOption)

	// ReplyError Replies a error to the current channel/user/thread + log it!
	ReplyError(event slack.MessageEvent, err error)

	// SendMessage is the extended version of Reply and accepts any slack.MsgOption
	SendMessage(event slack.MessageEvent, text string, options ...slack.MsgOption) string

	SendToUser(user string, text string)

	RemoveReaction(name string, item slack.ItemRef)
	AddReaction(name string, item slack.ItemRef)
	GetReactions(item slack.ItemRef, params slack.GetReactionsParameters) ([]slack.ItemReaction, error)
}

type Slack struct {
	*slack.Client
	RTM    *slack.RTM
	config config.Slack
	logger *logrus.Logger
}

// Reply fast reply via RTM websocket
// todo merge with SendMessage which is doing the same stuff now
func (s *Slack) Reply(event slack.MessageEvent, text string, options ...slack.MsgOption) {
	s.SendMessage(event, text, options...)
}

func (s *Slack) AddReaction(name string, item slack.ItemRef) {
	s.Client.AddReaction(name, item)
}

func (s *Slack) RemoveReaction(name string, item slack.ItemRef) {
	s.Client.RemoveReaction(name, item)
}

// SendMessage is the "slow" reply via POST request, needed for Attachment or MsgRef
func (s *Slack) SendMessage(event slack.MessageEvent, text string, options ...slack.MsgOption) string {
	if event.Channel == "" {
		return ""
	}

	if len(options) == 0 && text == "" {
		return ""
	}

	defaultOptions := []slack.MsgOption{
		slack.MsgOptionTS(event.ThreadTimestamp), // send in current thread by default
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(text, false),
		slack.MsgOptionDisableLinkUnfurl(),
	}

	options = append(defaultOptions, options...)
	_, msgTimestamp, err := s.PostMessage(
		event.Channel,
		options...,
	)

	if err != nil {
		s.logger.
			WithField("user", event.User).
			Errorf(err.Error())
	}

	return msgTimestamp
}

func (s *Slack) ReplyError(event slack.MessageEvent, err error) {
	s.logger.WithError(err).Warnf("Error while sending reply")
	s.Reply(event, err.Error())

	if s.config.ErrorChannel != "" {
		text := fmt.Sprintf(
			"<@%s> Error in command `%s`: %s",
			event.User,
			event.Msg.Text,
			err.Error(),
		)
		event.Channel, _ = GetChannel(s.config.ErrorChannel)
		s.SendMessage(event, text)
	}
}

// SendToUser sends a message to any user via IM channel
func (s *Slack) SendToUser(user string, text string) {
	// check if a real username was passed -> we need the user-id here
	user, _ = GetUser(user)

	options := &slack.OpenConversationParameters{
		Users: []string{user},
	}

	channel, _, _, err := s.Client.OpenConversation(options)
	if err != nil {
		s.logger.WithError(err).Errorf("Cannot open channel")
	}

	event := slack.MessageEvent{}
	event.Channel = channel.ID

	s.SendMessage(event, text)
}

func GetUser(identifier string) (id string, name string) {
	identifier = strings.TrimPrefix(identifier, "@")
	if name, ok := Users[identifier]; ok {
		return identifier, name
	}

	identifier = strings.ToLower(identifier)
	for id, name := range Users {
		if strings.EqualFold(name, identifier) {
			return id, name
		}
	}

	return "", ""
}

func GetChannel(identifier string) (id string, name string) {
	identifier = strings.TrimPrefix(identifier, "#")
	if name, ok := Channels[identifier]; ok {
		return identifier, name
	}

	identifier = strings.ToLower(identifier)
	for id, name := range Channels {
		if strings.EqualFold(name, identifier) {
			return id, name
		}
	}

	return "", ""
}

func GetSlackLink(name string, url string, args ...string) slack.AttachmentAction {
	style := "default"

	if len(args) > 0 {
		style = args[0]
	}

	action := slack.AttachmentAction{}
	action.Style = style
	action.Type = "button"
	action.Text = name
	action.URL = url

	return action
}

func GetInteraction(event slack.MessageEvent, text string, command string, args ...string) *slack.ActionBlock {
	var style slack.Style
	if len(args) > 0 {
		style = slack.Style(args[0])
	}

	id := util.RandString(32)

	event.Text = command
	storage.Write("interactions", id, event)

	buttonText := slack.NewTextBlockObject("plain_text", text, true, false)
	button := slack.NewButtonBlockElement("id", id, buttonText)
	button.Style = style

	return slack.NewActionBlock("", button)
}
