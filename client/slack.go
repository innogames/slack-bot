package client

//go:generate $GOPATH/bin/mockery --output ../mocks --name SlackClient

import (
	"fmt"
	"strings"
	"sync"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// InternalMessages is internal queue of internal messages
// @deprecated -> use HandleMessageWithDoneHandler instead
var InternalMessages = make(chan msg.Message, 50)

// HandleMessage will register the given message in the queue...and returns a sync.WaitGroup which can be used to see when the message is handled
func HandleMessage(message msg.Message) {
	message.Text = strings.TrimSpace(message.Text)
	if message.Text == "" {
		return
	}
	InternalMessages <- message
}

// HandleMessageWithDoneHandler will register the given message in the queue...and returns a sync.WaitGroup which can be used to see when the message is handled
func HandleMessageWithDoneHandler(message msg.Message) *sync.WaitGroup {
	done := message.AddDoneHandler()

	if message.Text != "" {
		HandleMessage(message)
	} else {
		// if we have no text, mark the message as processed to avoid open lock
		done.Done()
	}

	return done
}

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
		return nil, fmt.Errorf("config slack.token needs to start with 'xoxb-'")
	}

	options := []slack.Option{
		slack.OptionHTTPClient(GetHTTPClient()),
	}

	if cfg.TestEndpointURL != "" {
		options = append(options, slack.OptionAPIURL(cfg.TestEndpointURL))
	}

	if cfg.Debug {
		options = append(options, slack.OptionDebug(true))
	}

	if cfg.SocketToken != "" {
		if !strings.HasPrefix(cfg.SocketToken, "xapp-") {
			return nil, fmt.Errorf("config slack.socket_token needs to start to 'xapp-'")
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

// SlackClient is the main interface which is used for all commands to interact with slack
// for tests we have a mock for it, for production use, we use a slack.Client with some custom logic
type SlackClient interface {
	// ReplyError Replies a error to the current channel/user/thread + log it!
	ReplyError(ref msg.Ref, err error)
	// SendMessage is the extended version of Reply and accepts any slack.MsgOption
	SendMessage(ref msg.Ref, text string, options ...slack.MsgOption) string

	// SendEphemeralMessage sends a message just visible to the current user
	SendEphemeralMessage(ref msg.Ref, text string, options ...slack.MsgOption)

	// SendBlockMessage will send Slack Blocks/Sections to the target
	SendBlockMessage(ref msg.Ref, blocks []slack.Block, options ...slack.MsgOption) string

	// SendToUser sends a simple text message to a user, using "@username or @U12334"
	SendToUser(user string, text string)
	RemoveReaction(reaction util.Reaction, ref msg.Ref)
	AddReaction(reaction util.Reaction, ref msg.Ref)
	GetReactions(item slack.ItemRef, params slack.GetReactionsParameters) ([]slack.ItemReaction, error)

	// GetConversationHistory loads the message history from slack
	GetConversationHistory(*slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)

	// CanHandleInteractions checks if we have a slack connections which can inform us about events/interactions, like pressed buttons?
	CanHandleInteractions() bool
}

// Slack is wrapper to the slack.Client which also holds the RTM connection OR the socketmode.Client and all needed config
type Slack struct {
	*slack.Client
	RTM    *slack.RTM
	Socket *socketmode.Client
	config config.Slack
}

// AddReaction will add a reaction from the given message
func (s *Slack) AddReaction(reaction util.Reaction, ref msg.Ref) {
	err := s.Client.AddReaction(reaction.ToSlackReaction(), slack.NewRefToMessage(ref.GetChannel(), ref.GetTimestamp()))
	if err != nil {
		log.WithError(err).Warnf("Error while adding reaction %s", reaction)
	}
}

// RemoveReaction will remove a reaction from the given message
func (s *Slack) RemoveReaction(reaction util.Reaction, ref msg.Ref) {
	err := s.Client.RemoveReaction(reaction.ToSlackReaction(), slack.NewRefToMessage(ref.GetChannel(), ref.GetTimestamp()))
	if err != nil {
		log.WithError(err).Warnf("Error while removing reaction %s", reaction)
	}
}

// SendEphemeralMessage sends a message just visible to the current user
// see https://api.slack.com/methods/chat.postEphemeral
func (s *Slack) SendEphemeralMessage(ref msg.Ref, text string, options ...slack.MsgOption) {
	_, err := s.Client.PostEphemeral(
		ref.GetChannel(),
		ref.GetUser(),
		slack.MsgOptionAsUser(true),
		slack.MsgOptionTS(ref.GetThread()), // send in current thread by default
		slack.MsgOptionText(text, false),
	)
	if err != nil {
		log.Warn(errors.Wrapf(err, "Error while sending Ephemeral message %s", err))
	}
}

// SendMessage is the "slow" reply via POST request, needed for Attachment or MsgRef
func (s *Slack) SendMessage(ref msg.Ref, text string, options ...slack.MsgOption) string {
	if ref.GetChannel() == "" {
		log.Warnf("no channel given: %s", ref)
		return ""
	}

	if len(options) == 0 && text == "" {
		// ignore empty messages
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
			Error(err.Error())
	}

	return msgTimestamp
}

// ReplyError send a error message as a reply to the user and log it in the log + send it to ErrorChannel (if defined)
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
		message.Channel, _ = GetChannelIDAndName(s.config.ErrorChannel)
		s.SendMessage(message, text)
	}
}

// SendToUser sends a message to any user via IM channel
func (s *Slack) SendToUser(user string, text string) {
	// check if a real username was passed -> we need the user-id here
	userID, _ := GetUserIDAndName(user)
	if userID == "" {
		log.Errorf("Invalid user: %s", user)
		return
	}

	options := &slack.OpenConversationParameters{
		Users: []string{userID},
	}

	channel, _, _, err := s.Client.OpenConversation(options)
	if err != nil {
		log.WithError(err).Errorf("Cannot open channel")
		return
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

// GetUserIDAndName returns the user-id and user-name based on a identifier. If can get a user-id or name
func GetUserIDAndName(identifier string) (id string, name string) {
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

// GetChannelIDAndName returns channel-id and channel-name by an identifier which can be an id or a name
func GetChannelIDAndName(identifier string) (id string, name string) {
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

// GetSlackLink generates a "link button" as a slack.AttachmentAction which will
// open the given URL in the Slack client (when pressed)
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

// GetContextBlock generates a "Context block"
// https://api.slack.com/reference/block-kit/blocks#context
func GetContextBlock(text string) *slack.ContextBlock {
	return slack.NewContextBlock(
		"",
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: text,
		},
	)
}

// GetInteractionButton generates a block "Button" which is able to execute the given command once
// https://api.slack.com/reference/block-kit/blocks#actions
func GetInteractionButton(text string, command string, args ...slack.Style) *slack.ButtonBlockElement {
	var style slack.Style
	if len(args) > 0 {
		style = args[0]
	}

	buttonText := slack.NewTextBlockObject("plain_text", text, true, false)
	button := slack.NewButtonBlockElement("id", command, buttonText)
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
