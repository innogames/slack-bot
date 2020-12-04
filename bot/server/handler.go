package server

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io/ioutil"
	"net/http"
	"time"
)

// todo: this whole file is WIP and not ready for production yet - just use the RTM api for a stable API

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (s *Server) eventHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if !s.verifyRequest(w, r, body) {
		return
	}

	eventsAPIEvent, e := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if e != nil {
		s.error(w, e, http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent

		switch ev := innerEvent.Data.(type) {
		case slackevents.MessageEvent:
			client.InternalMessages <- msg.Message{
				// todo fill
			}
		default:
			fmt.Printf("%T -> %s \n ", ev, ev)
		}
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal(body, &r)
		if err != nil {
			s.error(w, e, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		return
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'm your friendly slack-bot"))
}

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

	var message msg.Message

	err = storage.Read("interactions", action.Value, &message)
	if err != nil {
		// already performed action -> do nothing
		w.WriteHeader(200)
		return
	}
	storage.Delete("interactions", action.Value)

	message.User = payload.User.ID

	client.InternalMessages <- message

	newMessage := getChangedMessage(payload.Message, action.Value)
	w.WriteHeader(http.StatusOK)

	response := slackevents.MessageActionResponse{}
	response.ReplaceOriginal = true
	response.Text = fmt.Sprintf("<@%s> performed action at %s", payload.User.Name, time.Now())

	s.slackClient.SendMessage(
		message,
		newMessage.Text,
		slack.MsgOptionUpdate(newMessage.Timestamp),
		slack.MsgOptionAttachments(newMessage.Attachments...),
		slack.MsgOptionBlocks(newMessage.Blocks.BlockSet...),
	)
}

func (s *Server) verifyRequest(w http.ResponseWriter, r *http.Request, body []byte) bool {
	// verifyRequest signature
	sv, err := slack.NewSecretsVerifier(r.Header, s.cfg.SigningSecret)
	if err != nil {
		s.error(w, err, http.StatusUnauthorized)
		return false
	}

	sv.Write(body)
	if err := sv.Ensure(); err != nil {
		s.error(w, err, http.StatusUnauthorized)
		return false
	}

	return true
}

func (s *Server) error(w http.ResponseWriter, err error, status int) {
	log.Error(err)
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
