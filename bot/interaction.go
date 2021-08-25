package bot

import (
	"sync"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var interactionLock sync.Mutex

// this method is called, when a user pressed a button:
// - validates that the user is allowed to press the button
func (b *Bot) handleEvent(eventsAPIEvent slackevents.EventsAPIEvent) {
	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			if ev.SubType == "message_changed" {
				// don't listen to edited messages
				return
			}
			message := &slack.MessageEvent{
				Msg: slack.Msg{
					Text:            ev.Text,
					Channel:         ev.Channel,
					User:            ev.User,
					Timestamp:       ev.TimeStamp,
					ThreadTimestamp: ev.ThreadTimeStamp,
				},
			}
			b.HandleMessage(message)
		case *slackevents.AppMentionEvent:
			message := &slack.MessageEvent{
				Msg: slack.Msg{
					Text:            ev.Text,
					Channel:         ev.Channel,
					User:            ev.User,
					Timestamp:       ev.TimeStamp,
					ThreadTimestamp: ev.ThreadTimeStamp,
				},
			}
			b.HandleMessage(message)
		}
	default:
		log.Infof("unsupported Events API event received")
	}
}

func (b *Bot) handleInteraction(payload slack.InteractionCallback) bool {
	if !b.allowedUsers.Contains(payload.User.ID) {
		log.Warnf("User %s tried to execute a command", payload.User.ID)
		return false
	}

	action := payload.ActionCallback.BlockActions[0]
	command := action.Value

	if action.Value == "" {
		log.Infof("Action '%s' got already executed (user: %s)", action.Value, payload.User.Name)
		return false
	}

	interactionLock.Lock()
	defer interactionLock.Unlock()

	log.Infof(
		"Received interaction from user %s/%s (action-id: %s, command: %s)",
		payload.User.ID,
		payload.User.Name,
		action.Value,
		command,
	)

	ref := msg.MessageRef{
		Channel:        payload.Container.ChannelID,
		Thread:         payload.Container.ThreadTs,
		User:           payload.User.ID,
		Timestamp:      payload.Message.Timestamp,
		UpdatedMessage: true,
	}

	// update the original slack message (with the button) and disable the button
	newMessage := replaceClickedButton(&payload.Message, action.Value, " (clicked)")

	if b.slackClient.Socket != nil {
		b.slackClient.SendMessage(
			ref,
			newMessage.Text,
			slack.MsgOptionUpdate(newMessage.Timestamp),
			slack.MsgOptionAttachments(newMessage.Attachments...),
			slack.MsgOptionBlocks(newMessage.Blocks.BlockSet...),
		)
	}

	// execute the command which is stored for this interaction
	go b.ProcessMessage(ref.WithText(command), true)

	stats.IncreaseOne(stats.Interactions)

	return true
}

// replaces the clicked button: appends the "message" (like "already clicked") and changed the color to red
func replaceClickedButton(newMessage *slack.Message, actionID string, message string) slack.Message {
	for _, blocks := range newMessage.Blocks.BlockSet {
		if actionBlock, ok := blocks.(*slack.ActionBlock); ok {
			for _, block := range actionBlock.Elements.ElementSet {
				if buttonBlock, ok := block.(*slack.ButtonBlockElement); ok {
					if buttonBlock.Value == actionID {
						buttonBlock.Style = slack.StyleDanger
						buttonBlock.Value = "" // purge command from button
						buttonBlock.Text.Text += message
					}
				}
			}
		}
	}

	return *newMessage
}
