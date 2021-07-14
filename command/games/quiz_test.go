package games

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestQuiz(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	// mock test data
	rand.Seed(2)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./quiz_example.json")
		w.Write(file)
	}))

	defer ts.Close()

	command := NewQuizCommand(base).(*quizCommand)
	command.apiURL = ts.URL
	commands := bot.Commands{}
	commands.AddCommand(command)

	t.Run("Full Game", func(t *testing.T) {
		// start the game
		message := msg.Message{}
		message.Text = "start quiz"

		expected := `[` +
			`{"type":"section","text":{"type":"mrkdwn","text":"Next question (#1) is of *\"hard\" difficulty* from the category: \"*Entertainment: Video Games*\"\nAccording to Toby Fox, what was the method to creating the initial tune for Megalovania?\n"}},` +
			`{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Using a Composer Software","emoji":true},"action_id":"id","value":"answer Using a Composer Software"}]},` +
			`{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Listened to birds at the park","emoji":true},"action_id":"id","value":"answer Listened to birds at the park"}]},` +
			`{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Singing into a Microphone","emoji":true},"action_id":"id","value":"answer Singing into a Microphone"}]},` +
			`{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"Playing a Piano","emoji":true},"action_id":"id","value":"answer Playing a Piano"}]}` +
			`]`
		mocks.AssertSlackBlocks(t, slackClient, message, expected)

		actual := commands.Run(message)
		assert.True(t, actual)

		// wrong answer
		message = msg.Message{}
		message.Text = "answer 4"
		mocks.AssertSlackMessage(slackClient, message, "incorrect. try again")
		actual = commands.Run(message)
		assert.True(t, actual)

		// correct answer
		message = msg.Message{}
		message.Text = "answer 3"

		expected = `[` +
			`{"type":"section","text":{"type":"mrkdwn","text":"Next question (#2) is of *\"easy\" difficulty* from the category: \"*Math*\"\nWhats 1+4?\n"}},` +
			`{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"5","emoji":true},"action_id":"id","value":"answer 5"}]},` +
			`{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"6","emoji":true},"action_id":"id","value":"answer 6"}]}` +
			`]`
		mocks.AssertSlackBlocks(t, slackClient, message, expected)

		mocks.AssertSlackMessage(slackClient, message, "correct")
		actual = commands.Run(message)
		assert.True(t, actual)
	})
}
