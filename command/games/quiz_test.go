package games

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"math/rand"
	"testing"
)

func TestQuiz(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	// mock test data
	rand.Seed(2)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./quiz_example.json")
		w.Write(file)
	}))
	defer ts.Close()

	command := NewQuizCommand(slackClient)
	command.apiUrl = ts.URL
	commands := bot.Commands{}
	commands.AddCommand(command)

	t.Run("Full Game", func(t *testing.T) {
		// start the game
		event := slack.MessageEvent{}
		event.Text = "quiz"
		slackClient.On("Reply", event,
			"Next question (#1) is of *\"hard\" difficulty* from the category: \"*Entertainment: Video Games*\"\n"+
				"According to Toby Fox, what was the method to creating the initial tune for Megalovania?\n"+
				"1.) Using a Composer Software\n"+
				"2.) Listened to birds at the park\n"+
				"3.) Singing into a Microphone\n4.) Playing a Piano\n"+
				":interrobang: Hint type `answer {number}` to send your answer :interrobang:",
		)
		actual := commands.Run(event)
		assert.Equal(t, true, actual)

		// wrong answer
		event = slack.MessageEvent{}
		event.Text = "answer 4"
		slackClient.On("Reply", event, "incorrect. try again")
		actual = commands.Run(event)
		assert.Equal(t, true, actual)

		// correct answer
		event = slack.MessageEvent{}
		event.Text = "answer 3"
		slackClient.On("Reply", event, ""+
			"Next question (#2) is of *\"easy\" difficulty* from the category: \"*Math*\"\n"+
			"Whats 1+4?\n"+
			"1.) 5\n"+
			"2.) 6\n"+
			":interrobang: Hint type `answer {number}` to send your answer :interrobang:",
		)
		slackClient.On("Reply", event, "correct")
		actual = commands.Run(event)
		assert.Equal(t, true, actual)

	})
}
