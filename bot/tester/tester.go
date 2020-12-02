// Package tester provides functionality to test the bot against a fake slack server
package tester

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

// TestChannel is just a test channel name which is used for testing
const TestChannel = "#dev"
const botID = "W12345"

// StartBot will start this bot against the fake slack instance
func StartBot(cfg config.Config) *bot.Bot {
	slackClient := client.GetSlackClient(cfg.Slack)

	commands := command.GetCommands(
		slackClient,
		cfg,
	)
	realBot := bot.NewBot(
		cfg,
		slackClient,
		commands,
	)

	err := realBot.Init()
	checkError(err)

	return realBot
}

type usersResponse struct {
	Members []slack.User
}

// StartFakeSlack will start a http server which implements the basic Slack API
func StartFakeSlack(cfg *config.Config, output io.Writer) *slacktest.Server {
	authResponse := fmt.Sprintf(`
	{
		"ok": true,
		"url": "https://localhost.localdomain/",
		"team": "%s",
		"user": "%s",
		"team_id": "%s",
		"user_id": "%s"
	}
`, "T123", "bot", "teamId", botID)

	// handle requests sto the mocked slack server and react on them for the "cli" tool
	handler := func(c slacktest.Customize) {
		c.Handle("/handler.test", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(authResponse))
		})
		c.Handle("/users.list", func(w http.ResponseWriter, _ *http.Request) {
			users := usersResponse{
				Members: []slack.User{},
			}
			bytes, _ := json.Marshal(users)
			w.Write(bytes)
		})
		c.Handle("/chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
			payload, _ := ioutil.ReadAll(r.Body)
			query, _ := url.ParseQuery(string(payload))
			text := query.Get("text")
			fmt.Fprintf(output, formatSlackMessage(text)+"\n")

			response := slack.Message{}
			response.Text = text
			bytes, _ := json.Marshal(response)
			w.Write(bytes)
		})
		c.Handle("/reactions.add", func(w http.ResponseWriter, r *http.Request) {
			payload, _ := ioutil.ReadAll(r.Body)
			fmt.Fprintln(output, getEmoji(string(payload)))
		})
	}

	fakeSlack := slacktest.NewTestServer(handler)
	fakeSlack.SetBotName("MyBotName")
	fakeSlack.BotID = botID
	fakeSlack.Start()

	cfg.Slack.Token = "not needed"
	cfg.Slack.TestEndpointURL = fakeSlack.GetAPIURL()
	cfg.AllowedUsers = []string{
		"W012A3CDE",
	}

	return fakeSlack
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
