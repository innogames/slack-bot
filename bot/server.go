package bot

// listen for incoming (button) interaction from slack server
// todo: WIP
/*
func (b *Bot) startServer(addr string) {
	http.HandleFunc("/slack/actions", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		body := buf.String()
		eventsAPIEvent, e := slackevents.ParseEvent(
			json.RawMessage(body),
			slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: b.config.Slack.VerificationToken}),
		)
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				_ = ev
			}
		}
	})

	go http.ListenAndServe(addr, nil)
}
*/
