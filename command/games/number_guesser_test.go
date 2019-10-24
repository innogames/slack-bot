package games

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNumberGuesser(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	command := NewNumberGuesserCommand(slackClient)
	commands := bot.Commands{}
	commands.AddCommand(command)

	gameCommand := command.(*numberGuesserCommand)
	t.Run("Full Game", func(t *testing.T) {
		assert.Equal(t, 0, len(gameCommand.games))

		// start the game
		event := slack.MessageEvent{}
		event.Text = "start number guesser"
		slackClient.On("Reply", event, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber))
		actual := commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Equal(t, 1, len(gameCommand.games))

		game := gameCommand.games[event.User]
		assert.True(t, game.randomNumber >= 0)
		assert.True(t, game.randomNumber <= maxNumber)
		assert.Equal(t, 0, game.tries)

		// too low number
		event.Text = fmt.Sprintf("guess number %d", game.randomNumber-1)
		slackClient.On("Reply", event, "Higher :arrow_up_small:")
		actual = commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Equal(t, 1, game.tries)

		// too high number
		event.Text = fmt.Sprintf("guess number %d", game.randomNumber+1)
		slackClient.On("Reply", event, "Lower :arrow_down_small:")
		actual = commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Equal(t, 2, game.tries)

		// bingo! -> win message + remove game
		event.Text = fmt.Sprintf("guess number %d", game.randomNumber)
		slackClient.On("Reply", event, "Wow! you got it in 3 tries :beers:")
		actual = commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Equal(t, 0, len(gameCommand.games))
	})

	t.Run("Invalid states", func(t *testing.T) {
		assert.Equal(t, 0, len(gameCommand.games))

		// guess without running game -> error message
		event := slack.MessageEvent{}
		event.Text = fmt.Sprintf("guess number 100")
		slackClient.On("Reply", event, "There is no game running. Use `start number guesser`")
		actual := commands.Run(event)
		assert.Equal(t, true, actual)

		// start the game
		event = slack.MessageEvent{}
		event.Text = "start number guesser"
		slackClient.On("Reply", event, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber))
		actual = commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Equal(t, 1, len(gameCommand.games))

		game := gameCommand.games[event.User]
		assert.True(t, game.randomNumber >= 0)
		assert.True(t, game.randomNumber <= maxNumber)
		assert.Equal(t, 0, game.tries)

		// start the game again -> error
		event = slack.MessageEvent{}
		event.Text = "start number guesser"
		slackClient.On("Reply", event, "There is already a game :smile: use `guess number XX` instead")
		actual = commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Equal(t, 1, len(gameCommand.games))
	})
}
