package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// todo: this whole file is WIP and not ready for production yet - just use the RTM api for a stable API

// handle request from interactive events: https://api.slack.com/messaging/interactivity
func (s *Server) interactionHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if !s.verifyRequest(w, r, body) {
		return
	}

	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		s.error(w, errors.Wrap(err, "Could not parse action response JSON"), http.StatusInternalServerError)
		return
	}

	if len(payload.ActionCallback.BlockActions) == 0 {
		// no valid action defined
		w.WriteHeader(http.StatusOK)
		return
	}

	action := payload.ActionCallback.BlockActions[0]

	// check in storage if there is still a interaction stored on our side
	var message slack.MessageEvent
	err = storage.Read("interactions", action.Value, &message)
	if err != nil {
		// already performed action -> do nothing
		w.WriteHeader(http.StatusOK)
		return
	}
	storage.Delete("interactions", action.Value)

	// execute the command which is stored for this interaction
	message.User = payload.User.ID
	s.messageHandler(&message)

	// update the original slack message (with the button) and disable the button
	newMessage := getChangedMessage(&payload.Message, action.Value)
	response := slackevents.MessageActionResponse{}
	response.ReplaceOriginal = true
	response.Text = fmt.Sprintf("<@%s> performed action at %s", payload.User.Name, time.Now())

	s.slackClient.SendMessage(
		msg.FromSlackEvent(&message),
		newMessage.Text,
		slack.MsgOptionUpdate(newMessage.Timestamp),
		slack.MsgOptionAttachments(newMessage.Attachments...),
		slack.MsgOptionBlocks(newMessage.Blocks.BlockSet...),
	)
}

// copy the given message and disable the button which got pressed and mark it as clicked
func getChangedMessage(newMessage *slack.Message, actionID string) slack.Message {
	for _, blocks := range newMessage.Blocks.BlockSet {
		if actionBlock, ok := blocks.(*slack.ActionBlock); ok {
			for _, block := range actionBlock.Elements.ElementSet {
				if buttonBlock, ok := block.(*slack.ButtonBlockElement); ok {
					if buttonBlock.Value == actionID {
						buttonBlock.Value = ""
						buttonBlock.Text.Text += " (already clicked)"
					}
				}
			}
		}
	}

	return *newMessage
}
