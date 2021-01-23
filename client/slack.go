package client

//go:generate $GOPATH/bin/mockery --output ../mocks --name SlackClient

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// InternalMessages is internal queue of internal messages
var InternalMessages = make(chan msg.Message, 50)

// AuthResponse is holding some basic Slack metadata for the current connection, like Bot-Id, Workspace etc
var AuthResponse slack.AuthTestResponse

// Users is a lookup from user-id to user-name
var Users config.UserMap

// Channels is a map of each channelsId and the name
var Channels map[string]string

// GetSlackClient establishes a connection to the slack server.
// Either via "Socket Mode" or the legacy "RTM" connection
func GetSlackClient(cfg config.Slack) (*Slack, error) {
	if !strings.HasPrefix(cfg.Token, "xoxb-") {
		return nil, fmt.Errorf("slack.token needs to start with 'xoxb-'")
	}

	options := []slack.Option{
		slack.OptionHTTPClient(HTTPClient),
	}

	if cfg.TestEndpointURL != "" {
		options = append(options, slack.OptionAPIURL(cfg.TestEndpointURL))
	}

	if cfg.Debug {
		options = append(options, slack.OptionDebug(true))
	}

	if cfg.SocketToken != "" {
		if !strings.HasPrefix(cfg.SocketToken, "xapp-") {
			return nil, fmt.Errorf("cfg.socket_token needs to start to 'xapp-'")
		}
		options = append(options, slack.OptionAppLevelToken(cfg.SocketToken))
	}

	rawClient := slack.New(cfg.Token, options...)

	var rtm *slack.RTM
	var socket *socketmode.Client
	if cfg.SocketToken != "" {
		socket = socketmode.New(
			rawClient,
		)
	} else {
		rtm = rawClient.NewRTM()
	}

	return &Slack{Client: rawClient, RTM: rtm, Socket: socket, config: cfg}, nil
}

type SlackClient interface {
	// ReplyError Replies a error to the current channel/user/thread + log it!
	ReplyError(ref msg.Ref, err error)
	// SendMessage is the extended version of Reply and accepts any slack.MsgOption
	SendMessage(ref msg.Ref, text string, options ...slack.MsgOption) string
	// SendBlockMessage will send Slack Blocks/Sections to the target
	SendBlockMessage(ref msg.Ref, blocks []slack.Block, options ...slack.MsgOption) string

	// send a message to a user, using "@username or @U12334"
	SendToUser(user string, text string)

	RemoveReaction(reaction string, ref msg.Ref)
	AddReaction(reaction string, ref msg.Ref)
	GetReactions(item slack.ItemRef, params slack.GetReactionsParameters) ([]slack.ItemReaction, error)

	GetConversationHistory(*slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)

	// CanHandleInteractions checks if we have a slack connections which can inform us about events/interactions, like pressed buttons?
	CanHandleInteractions() bool
}

// wrapper to the Slack client which also holds the RTM connection (optional) and all needed config
type Slack struct {
	*slack.Client
	RTM    *slack.RTM
	Socket *socketmode.Client
	config config.Slack
}

func (s *Slack) AddReaction(reaction string, ref msg.Ref) {
	err := s.Client.AddReaction(reaction, slack.NewRefToMessage(ref.GetChannel(), ref.GetTimestamp()))
	if err != nil {
		log.Warn(errors.Wrapf(err, "Error while adding reaction: %s - %+v", reaction, ref))
	}
}

func (s *Slack) RemoveReaction(reaction string, ref msg.Ref) {
	err := s.Client.RemoveReaction(reaction, slack.NewRefToMessage(ref.GetChannel(), ref.GetTimestamp()))
	if err != nil {
		log.Warn(errors.Wrapf(err, "Error while removing reaction %s", reaction))
	}
}

// SendMessage is the "slow" reply via POST request, needed for Attachment or MsgRef
func (s *Slack) SendMessage(ref msg.Ref, text string, options ...slack.MsgOption) string {
	if ref.GetChannel() == "" {
		return ""
	}

	if len(options) == 0 && text == "" {
		return ""
	}

	defaultOptions := []slack.MsgOption{
		slack.MsgOptionTS(ref.GetThread()), // send in current thread by default
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(text, false),
		slack.MsgOptionDisableLinkUnfurl(),
	}

	options = append(defaultOptions, options...)
	_, msgTimestamp, err := s.PostMessage(
		ref.GetChannel(),
		options...,
	)

	if err != nil {
		log.
			WithField("user", ref.GetUser()).
			Errorf(err.Error())
	}

	return msgTimestamp
}

func (s *Slack) ReplyError(ref msg.Ref, err error) {
	log.WithError(err).Warnf("Error while sending reply")
	s.SendMessage(ref, err.Error())

	if s.config.ErrorChannel != "" {
		text := fmt.Sprintf(
			"<@%s> Error in command: %s",
			ref.GetUser(),
			err.Error(),
		)
		message := msg.Message{}
		message.Channel, _ = GetChannel(s.config.ErrorChannel)
		s.SendMessage(message, text)
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
		log.WithError(err).Errorf("Cannot open channel")
	}

	message := msg.Message{}
	message.Channel = channel.ID

	s.SendMessage(message, text)
}

// CanHandleInteractions checks if we have a slack connections which can inform us about events/interactions, like pressed buttons?
func (s *Slack) CanHandleInteractions() bool {
	return s.config.CanHandleInteractions()
}

// SendBlockMessage will send Slack Blocks/Sections to the target
func (s *Slack) SendBlockMessage(ref msg.Ref, blocks []slack.Block, options ...slack.MsgOption) string {
	allOptions := []slack.MsgOption{
		slack.MsgOptionBlocks(blocks...),
	}

	return s.SendMessage(ref, "", append(allOptions, options...)...)
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

// GetChannel returns channel-id and channel-name by an identifier which can be an id or a name
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

// GetTextBlock wraps a simple text in a Slack Block Section
func GetTextBlock(text string) *slack.SectionBlock {
	return slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: text,
		},
		nil,
		nil,
	)
}

// GetInteractionButton generates a block "Button" which is able to execute the given command once
func GetInteractionButton(ref msg.Ref, text string, command string, args ...string) *slack.ButtonBlockElement {
	var style slack.Style
	if len(args) > 0 {
		style = slack.Style(args[0])
	}

	id := util.RandString(32)

	message := ref.WithText(command)
	if err := storage.Write("interactions", id, message); err != nil {
		log.Warn(err)
	}

	buttonText := slack.NewTextBlockObject("plain_text", text, true, false)
	button := slack.NewButtonBlockElement("id", id, buttonText)
	button.Style = style

	return button
}

// GetSlackArchiveLink returns a permalink to the ref which can be shared
func GetSlackArchiveLink(message msg.Ref) string {
	return fmt.Sprintf(
		"https://%s.slack.com/archives/%s/p%s",
		strings.ToLower(AuthResponse.Team),
		message.GetChannel(),
		strings.ReplaceAll(message.GetTimestamp(), ".", ""),
	)
}
