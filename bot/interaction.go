package bot

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"sync"
	"time"
)

var interactionLock sync.Mutex

// internal storage collection which contains all stored interactions aka buttons which can get pressed
const storageCollection = "interactions"

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

func (b *Bot) handleInteraction(payload slack.InteractionCallback) {
	if !b.allowedUsers.Contains(payload.User.ID) {
		log.Warnf("User %s tried to execute a command", payload.User.ID)
		return
	}

	action := payload.ActionCallback.BlockActions[0]

	if action.Value == "" {
		log.Infof("Action '%s' got already executed (user: %s)", action.Value, payload.User.Name)
		return
	}

	interactionLock.Lock()
	defer interactionLock.Unlock()

	// check in storage if there is still a interaction stored on our side
	var messageEvent slack.MessageEvent
	err := storage.Read(storageCollection, action.Value, &messageEvent)
	if err != nil {
		log.Warnf("Action '%s' is invalid (user: %s)", action.Value, payload.User.Name)
		return
	}
	storage.Delete(storageCollection, action.Value)
	messageEvent.Timestamp = payload.Message.Timestamp
	log.Infof(
		"Received interaction from user %s/%s (action-id: %s, command: %s)",
		payload.User.ID,
		payload.User.Name,
		action.Value,
		messageEvent.Text,
	)

	message := msg.FromSlackEvent(&messageEvent)
	message.UpdatedMessage = true

	// update the original slack message (with the button) and disable the button
	newMessage := replaceClickedButton(&payload.Message, action.Value, " (clicked)")

	if b.slackClient.Socket != nil {
		b.slackClient.SendMessage(
			message,
			newMessage.Text,
			slack.MsgOptionUpdate(newMessage.Timestamp),
			slack.MsgOptionAttachments(newMessage.Attachments...),
			slack.MsgOptionBlocks(newMessage.Blocks.BlockSet...),
		)
	}

	// execute the command which is stored for this interaction
	messageEvent.User = payload.User.ID
	go b.handleMessage(message, true)
}

func (b *Bot) cleanOldInteractions() (deleted int) {
	timeCheck := time.Now().Add(-time.Hour * 24)
	var message slack.MessageEvent
	keys, _ := storage.GetKeys(storageCollection)

	for _, key := range keys {
		storage.Read(storageCollection, key, &message)
		if msg.FromSlackEvent(&message).GetTime().Before(timeCheck) {
			storage.Delete(storageCollection, key)
			deleted++
		}
	}
	return
}

// replaces the clicked button: appends the "message" (like "already clicked") and changed the color to red
func replaceClickedButton(newMessage *slack.Message, actionID string, message string) slack.Message {
	for _, blocks := range newMessage.Blocks.BlockSet {
		if actionBlock, ok := blocks.(*slack.ActionBlock); ok {
			for _, block := range actionBlock.Elements.ElementSet {
				if buttonBlock, ok := block.(*slack.ButtonBlockElement); ok {
					if buttonBlock.Value == actionID {
						buttonBlock.Style = slack.StyleDanger
						buttonBlock.Value = ""
						buttonBlock.Text.Text += message
					}
				}
			}
		}
	}

	return *newMessage
}
