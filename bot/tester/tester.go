// Package tester provides functionality to test the bot against a fake slack server
package tester

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

// TestChannel is just a test channel name which is used for testing
const (
	TestChannel = "#dev"
	botID       = "W12345"
)

// StartBot will start this bot against the fake slack instance
func StartBot(cfg config.Config) *bot.Bot {
	slackClient, err := client.GetSlackClient(cfg.Slack)
	checkError(err)

	commands := command.GetCommands(
		slackClient,
		cfg,
	)
	realBot := bot.NewBot(
		cfg,
		slackClient,
		commands,
	)

	err = realBot.Init()
	checkError(err)

	return realBot
}

type usersResponse struct {
	Members []slack.User
}

// StartFakeSlack will start a http server which implements the basic Slack API
func StartFakeSlack(cfg *config.Config, output io.Writer) *slacktest.Server {
	// handle requests sto the mocked slack server and react on them for the "cli" tool
	handler := func(c slacktest.Customize) {
		c.Handle("/users.list", func(w http.ResponseWriter, _ *http.Request) {
			users := usersResponse{
				Members: []slack.User{},
			}
			bytes, _ := json.Marshal(users)
			_, _ = w.Write(bytes)
		})
		c.Handle("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
			payload, _ := ioutil.ReadAll(r.Body)
			query, _ := url.ParseQuery(string(payload))
			text := query.Get("text")

			// extract text from TextBlock
			if text == "" {
				blockJSON := query.Get("blocks")
				var blocks []slack.SectionBlock
				_ = json.Unmarshal([]byte(blockJSON), &blocks)

				text = blocks[0].Text.Text
			}

			_, _ = fmt.Fprintf(output, formatSlackMessage(text)+"\n")

			response := slack.Message{}
			response.Text = text
			bytes, _ := json.Marshal(response)
			_, _ = w.Write(bytes)
		})
		c.Handle("/reactions.add", func(w http.ResponseWriter, r *http.Request) {
			payload, _ := ioutil.ReadAll(r.Body)
			query, _ := url.ParseQuery(string(payload))
			emoji := query.Get("name")
			_, _ = fmt.Fprintln(output, util.Reaction(emoji).GetChar())

			response := slack.SlackResponse{}
			response.Ok = true
			bytes, _ := json.Marshal(response)
			_, _ = w.Write(bytes)
		})
	}

	fakeSlack := slacktest.NewTestServer(handler)
	fakeSlack.SetBotName("MyBotName")
	fakeSlack.BotID = botID
	fakeSlack.Start()

	cfg.Slack.Token = "xoxb-fake"
	cfg.Slack.TestEndpointURL = fakeSlack.GetAPIURL()
	cfg.AllowedUsers = []string{
		"W012A3CDE",
	}

	return fakeSlack
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
