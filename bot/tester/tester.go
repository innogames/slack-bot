// Package tester provides functionality to test the bot against a fake slack server
package tester

import (
	"fmt"
	"net/http"
	"os"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack/slacktest"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

// TestChannel is just a test channel name which is used for testing
const TestChannel = "#dev"
const botId = "W12345"

// StartBot will start this bot against the fake slack instance
func StartBot(cfg config.Config, logger *logrus.Logger) bot.Handler {
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
`, "T123", "bot", "teamId", botId)

	auth := func(c slacktest.Customize) {
		c.Handle("/auth.test", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(authResponse))
		})
	}

	fakeSlack := slacktest.NewTestServer(auth)
	fakeSlack.SetBotName("MyBotName")
	fakeSlack.BotID = botId
	fakeSlack.Start()

	cfg.Slack.Token = "not needed"
	cfg.Slack.TestEndpointUrl = fakeSlack.GetAPIURL()
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
