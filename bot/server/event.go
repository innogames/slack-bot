package server

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io/ioutil"
	"net/http"
)

// todo: this whole file is WIP and not ready for production yet - just use the RTM api for a stable API

// handle events from the event api: https://api.slack.com/events-api
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

			s.messageHandler(message)
			fmt.Printf("%+v", message)
		default:
			fmt.Printf("%T -> %s \n ", ev, ev)
		}
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r slackevents.ChallengeResponse
		err := json.Unmarshal(body, &r)
		if err != nil {
			s.error(w, e, http.StatusInternalServerError)
			return
		}

		log.Info("Server: auth successful")

		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		return
	}
}
