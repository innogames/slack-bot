package tester

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slacktest"
)

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
			messageHandler(w, r, output)
		})
		c.Handle("/chat.postEphemeral", func(w http.ResponseWriter, r *http.Request) {
			messageHandler(w, r, output)
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

			_, _ = fmt.Fprintln(output, formatSlackMessage(fmt.Sprintf("Clicked link with message: *%s*", commandText)))
			_, _ = writer.Write([]byte(fmt.Sprintf(
				"Executed command '%s'. You can close the browser and go back to the terminal.",
				html.EscapeString(commandText),
			)))
			HandleMessage(commandText)
		})

		c.Handle("/apps.connections.open", func(w http.ResponseWriter, r *http.Request) {
			// just do noting
		})

		c.Handle("/", func(w http.ResponseWriter, r *http.Request) {
			// check for unhandled methods
			fmt.Println(r.RequestURI)
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
	FakeServerURL = fakeSlack.GetAPIURL()

	return fakeSlack
}
