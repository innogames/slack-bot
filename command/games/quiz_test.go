package games

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
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

	command := NewQuizCommand(slackClient).(*quizCommand)
	command.apiURL = ts.URL
	commands := bot.Commands{}
	commands.AddCommand(command)

	t.Run("Full Game", func(t *testing.T) {
		// start the game
		message := msg.Message{}
		message.Text = "quiz"
		slackClient.On("SendMessage", message,
			"Next question (#1) is of *\"hard\" difficulty* from the category: \"*Entertainment: Video Games*\"\n"+
				"According to Toby Fox, what was the method to creating the initial tune for Megalovania?\n"+
				"1.) Using a Composer Software\n"+
				"2.) Listened to birds at the park\n"+
				"3.) Singing into a Microphone\n4.) Playing a Piano\n"+
				":interrobang: Hint type `answer {number}` to send your answer :interrobang:",
		).Return("")
		actual := commands.Run(message)
		assert.Equal(t, true, actual)

		// wrong answer
		message = msg.Message{}
		message.Text = "answer 4"
		slackClient.On("SendMessage", message, "incorrect. try again").Return("")
		actual = commands.Run(message)
		assert.Equal(t, true, actual)

		// correct answer
		message = msg.Message{}
		message.Text = "answer 3"
		slackClient.On("SendMessage", message, ""+
			"Next question (#2) is of *\"easy\" difficulty* from the category: \"*Math*\"\n"+
			"Whats 1+4?\n"+
			"1.) 5\n"+
			"2.) 6\n"+
			":interrobang: Hint type `answer {number}` to send your answer :interrobang:",
		).Return("")
		slackClient.On("SendMessage", message, "correct").Return("")
		actual = commands.Run(message)
		assert.Equal(t, true, actual)
	})
}
