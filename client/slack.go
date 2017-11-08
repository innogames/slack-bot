package client

import (
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"strings"
)

// InternalMessages is internal queue of internal messages
var InternalMessages = make(chan slack.MessageEvent, 100)

// Users is a lookup from user-id to user-name
var Users map[string]string
var Channels map[string]string

func GetSlackClient(cfg config.Slack, logger *logrus.Logger) *Slack {
	options := make([]slack.Option, 0)
	if cfg.TestEndpointUrl != "" {
		options = append(options, slack.OptionAPIURL(cfg.TestEndpointUrl))
	}

	if cfg.Debug {
		options = append(options, slack.OptionDebug(true))
	}

	rtm := slack.New(cfg.Token, options...).NewRTM()
	slackClient := &Slack{RTM: *rtm, logger: logger, config: cfg}

	return slackClient
}

type SlackClient interface {
	Reply(event slack.MessageEvent, text string)
	ReplyError(event slack.MessageEvent, err error)
	SendMessage(event slack.MessageEvent, text string, options ...slack.MsgOption) string
	SendToUser(user string, text string) string
	RemoveReaction(name string, item slack.ItemRef)
	AddReaction(name string, item slack.ItemRef)
}

type Slack struct {
	slack.RTM
	config config.Slack
	logger *logrus.Logger
}

// Reply fast reply via RTM websocket
func (s Slack) Reply(event slack.MessageEvent, text string) {
	s.RTM.SendMessage(s.RTM.NewOutgoingMessage(
		text,
		event.Channel,
		slack.RTMsgOptionTS(event.ThreadTimestamp),
	))
}

func (s Slack) AddReaction(name string, item slack.ItemRef) {
	go s.Client.AddReaction(name, item)
}

func (s Slack) RemoveReaction(name string, item slack.ItemRef) {
	go s.Client.RemoveReaction(name, item)
}

// SendMessage is the "slow" reply via POST request, needed for Attachment or MsgRef
func (s Slack) SendMessage(event slack.MessageEvent, text string, options ...slack.MsgOption) string {
	if event.Channel == "" {
		// todo log error
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

func (s Slack) ReplyError(event slack.MessageEvent, err error) {
	s.logger.WithError(err).Warnf("Error while sending reply")
	s.Reply(event, err.Error())
}

// SendToUser sends a message to any user via IM channel
func (s Slack) SendToUser(user string, text string) string {
	// check if a real username was passed -> we need the user-id here
	user, _ = GetUser(user)

	_, _, channel, err := s.Client.OpenIMChannel(user)
	if err != nil {
		s.logger.WithError(err).Errorf("Cannot open channel")
	}

	event := slack.MessageEvent{}
	event.Channel = channel

	return s.SendMessage(event, text)
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
