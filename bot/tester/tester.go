// Package tester provides functionality to test the bot against a fake slack server
package tester

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// TestChannel is just a test channel name which is used for testing
const (
	TestChannel = "dev"
	User        = "cli"
	botID       = "W12345"
)

// FakeServerURL is the host for the fake Slack server endpoint. It can be also used to send fake responses.
var FakeServerURL string

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

	color.Info.Printf(
		"Hey! I'm your Slack Emulator. Call or click '%s' to get a list of all supported commands\n",
		formatSlackMessage(commandButton("help", "help")),
	)

	return realBot
}

// HandleMessage is used in the CLI context to process the given message test for the "cli" user
func HandleMessage(text string) {
	message := msg.Message{}
	message.Text = text
	message.Channel = TestChannel
	message.User = User

	client.HandleMessageWithDoneHandler(message).Wait()
}

type usersResponse struct {
	Members []slack.User
}

// kinda dirty grown function to format a /chat.postMessage /chat.postEphemeral message on the command like...somehow
func messageHandler(w http.ResponseWriter, r *http.Request, output io.Writer) {
	payload, _ := ioutil.ReadAll(r.Body)
	query, _ := url.ParseQuery(string(payload))
	text := query.Get("text")

	// extract text from TextBlock
	if text == "" && query.Get("blocks") != "" {
		blockJSON := query.Get("blocks")
		var blocks []map[string]interface{}
		_ = json.Unmarshal([]byte(blockJSON), &blocks)

		for _, block := range blocks {
			text += formatBlock(block) + "\n"
		}
	} else if text == "" && query.Get("attachments") != "" {
		attachmentJSON := query.Get("attachments")
		var attachments []map[string]interface{}
		_ = json.Unmarshal([]byte(attachmentJSON), &attachments)

		for _, attachment := range attachments {
			if txt, ok := attachment["title"].(string); ok {
				text += txt + "\n"
			}
			for _, action := range attachment["actions"].([]interface{}) {
				actionMap := action.(map[string]interface{})

				if actionMap["type"] == "button" {
					text += fmt.Sprintf(
						"<%s|%s>\n",
						actionMap["url"],
						actionMap["text"],
					)
				} else {
					text += fmt.Sprintf("Attachment-actions are not supported yet:\n%v\n", action)
				}
			}
		}
	}

	_, _ = fmt.Fprint(output, formatSlackMessage(text)+"\n")

	response := slack.Message{}
	response.Text = text
	bytes, _ := json.Marshal(response)
	_, _ = w.Write(bytes)
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

			return commandButton(buttonText, buttonValue)
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
