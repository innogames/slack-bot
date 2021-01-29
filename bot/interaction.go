package bot

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"sync"
	"time"
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

	interactionLock.Lock()
	defer interactionLock.Unlock()

	action := payload.ActionCallback.BlockActions[0]

	// check in storage if there is still a interaction stored on our side
	var message slack.MessageEvent
	err := storage.Read("interactions", action.Value, &message)
	if err != nil {
		log.Warnf("Action '%s' got already executed (user: %s)", action.Value, payload.User.Name)
		return
	}
	storage.Delete("interactions", action.Value)

	log.Infof(
		"Received interaction from user %s/%s (action-id: %s, command: %s)",
		payload.User.ID,
		payload.User.Name,
		action.Value,
		message.Text,
	)

	// execute the command which is stored for this interaction
	message.User = payload.User.ID
	b.HandleMessage(&message)

	// update the original slack message (with the button) and disable the button
	newMessage := replaceClickedButton(&payload.Message, action.Value, " (clicked)")
	response := slackevents.MessageActionResponse{}
	response.ReplaceOriginal = true
	response.Text = fmt.Sprintf("<@%s> performed action at %s", payload.User.Name, time.Now())

	if b.slackClient.Socket == nil {
		return
	}
	b.slackClient.SendMessage(
		msg.FromSlackEvent(&message),
		newMessage.Text,
		slack.MsgOptionUpdate(newMessage.Timestamp),
		slack.MsgOptionAttachments(newMessage.Attachments...),
		slack.MsgOptionBlocks(newMessage.Blocks.BlockSet...),
	)
}

func (b *Bot) cleanOldInteractions() (deleted int) {
	timeCheck := time.Now().Add(-time.Hour * 24)
	var message slack.MessageEvent
	keys, _ := storage.GetKeys("interactions")

	for _, key := range keys {
		storage.Read("interactions", key, &message)
		if msg.FromSlackEvent(&message).GetTime().Before(timeCheck) {
			storage.Delete("interactions", key)
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
