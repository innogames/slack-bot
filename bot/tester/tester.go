// Package tester provides functionality to test the bot against a fake slack server
package tester

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/innogames/slack-bot/v2/bot/msg"

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
	TestChannel = "dev"
	botID       = "W12345"
)

var FakeServerUrl string

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
				var blocks []map[string]interface{}
				_ = json.Unmarshal([]byte(blockJSON), &blocks)

				for _, block := range blocks {
					text += formatBlock(block) + "\n"
				}
			}

			_, _ = fmt.Fprint(output, formatSlackMessage(text)+"\n")

			response := slack.Message{}
			response.Text = text
			bytes, _ := json.Marshal(response)
			_, _ = w.Write(bytes)
		})
		c.Handle("/reactions.add", func(w http.ResponseWriter, r *http.Request) {
			// post the given reaction as unicode character in the terminal
			payload, _ := ioutil.ReadAll(r.Body)
			query, _ := url.ParseQuery(string(payload))
			emoji := query.Get("name")
			_, _ = fmt.Fprintln(output, util.Reaction(emoji).GetChar())

			response := slack.SlackResponse{}
			response.Ok = true
			bytes, _ := json.Marshal(response)
			_, _ = w.Write(bytes)
		})
		c.Handle("/command", func(writer http.ResponseWriter, request *http.Request) {
			// fake the buttons: pass the command in a hyper link
			commandText := request.URL.Query().Get("command")

			fmt.Fprintln(output, formatSlackMessage(fmt.Sprintf("Clicked link with message: *%s*", commandText)))
			writer.Write([]byte(fmt.Sprintf("Executed command '%s'. You can close the browser and go back to the terminal.", commandText)))
			HandleMessage(commandText)
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
	FakeServerUrl = fakeSlack.GetAPIURL()

	return fakeSlack
}

func HandleMessage(text string) {
	message := msg.Message{}
	message.Text = text
	message.Channel = TestChannel
	message.User = "cli"

	client.HandleMessageWithDoneHandler(message).Wait()
}

func formatBlock(block map[string]interface{}) string {
	text := ""

	switch block["type"] {
	case "section":
		return extractText(block)
	case "actions":
		for _, element := range block["elements"].([]interface{}) {
			buttonText := extractText(element.(map[string]interface{}))
			buttonValue := element.(map[string]interface{})["value"].(string)

			return fmt.Sprintf("<%scommand?command=%s|%s>", FakeServerUrl, buttonValue, buttonText)
		}
	default:
		return fmt.Sprintf("invalid block: %v", block)
	}

	return text
}

// bit hacky way to extract the text from some kind of block element
func extractText(block map[string]interface{}) string {
	if fields, ok := block["fields"]; ok {
		result := ""
		for _, field := range fields.([]interface{}) {
			result += extractText(field.(map[string]interface{}))
		}
		return result
	}

	// contains "text" element
	if txt, ok := block["text"].(string); ok {
		return txt
	} else if txt, ok := block["text"].(map[string]interface{}); ok {
		if value, ok := txt["value"]; ok {
			return value.(map[string]interface{})["text"].(string)
		}

		return txt["text"].(string)
	}

	return fmt.Sprintf("unknown: %v", block)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
