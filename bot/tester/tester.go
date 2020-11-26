// Package tester provides functionality to test the bot against a fake slack server
package tester

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

// TestChannel is just a test channel name which is used for testing
const TestChannel = "#dev"
const botID = "W12345"

// StartBot will start this bot against the fake slack instance
func StartBot(cfg config.Config, logger *logrus.Logger) *bot.Bot {
	slackClient := client.GetSlackClient(cfg.Slack, logger)

	commands := command.GetCommands(
		slackClient,
		cfg,
		logger,
	)
	realBot := bot.NewBot(
		cfg,
		slackClient,
		logger,
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
func StartFakeSlack(cfg *config.Config) *slacktest.Server {
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

	auth := func(c slacktest.Customize) {
		c.Handle("/auth.test", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(authResponse))
		})
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

			fmt.Print("\n")
			fmt.Println(query.Get("text"))
		})
	}

	fakeSlack := slacktest.NewTestServer(auth)
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

// GetNullLogger will just ignore all logs
func GetNullLogger() *logrus.Logger {
	logger, _ := test.NewNullLogger()

	return logger
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
