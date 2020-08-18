package interaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) interactionHandler(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, s.cfg.VerificationSecret)
	if err != nil {
		s.error(w, errors.Wrap(err, "Could not initialize SecretVerifier"), http.StatusInternalServerError)
		return
	}

	re, _ := ioutil.ReadAll(r.Body)
	verifier.Write(re)
	r.Body = ioutil.NopCloser(bytes.NewReader(re))

	if err = verifier.Ensure(); err != nil {
		s.error(w, errors.Wrap(err, "Used invalid signature"), http.StatusUnauthorized)
		return
	}

	var payload slack.InteractionCallback
	err = json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		s.error(w, errors.Wrap(err, "Could not parse action response JSON"), http.StatusInternalServerError)
		return
	}

	if _, ok := s.allowedUsers[payload.User.ID]; !ok {
		s.error(
			w,
			errors.Wrap(err, fmt.Sprintf("%s tried to perform an interaction which is not whitelisted", payload.User.Name)),
			http.StatusInternalServerError,
		)
		return
	}

	if len(payload.ActionCallback.BlockActions) == 0 {
		// no valid action defined
		w.WriteHeader(http.StatusOK)
		return
	}

	action := payload.ActionCallback.BlockActions[0]
	var event slack.MessageEvent
	err = storage.Read("interactions", action.Value, &event)
	if err != nil {
		// already performed action -> do nothing
		w.WriteHeader(200)
		return
	}
	storage.Delete("interactions", action.Value)
	event.User = payload.User.ID

	client.InternalMessages <- event

	newMessage := getChangedMessage(payload.Message, action.Value)
	w.WriteHeader(http.StatusOK)

	response := slackevents.MessageActionResponse{}
	response.ReplaceOriginal = true
	response.Text = fmt.Sprintf("<@%s> performed action at %s", payload.User.Name, time.Now())

	s.slackClient.SendMessage(
		event,
		newMessage.Text,
		slack.MsgOptionUpdate(newMessage.Timestamp),
		slack.MsgOptionAttachments(newMessage.Attachments...),
		slack.MsgOptionBlocks(newMessage.Blocks.BlockSet...),
	)
}

func (s *Server) error(w http.ResponseWriter, err error, status int) {
	s.logger.Errorf(err.Error())
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
