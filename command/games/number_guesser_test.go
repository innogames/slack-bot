package games

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNumberGuesser(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	command := NewNumberGuesserCommand(base)
	commands := bot.Commands{}
	commands.AddCommand(command)

	gameCommand := command.(*numberGuesserCommand)
	t.Run("Full Game", func(t *testing.T) {
		assert.Equal(t, 0, len(gameCommand.games))

		// start the game
		message := msg.Message{}
		message.Text = "start number guesser"
		slackClient.On("SendMessage", message, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber)).Return("")
		actual := commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, len(gameCommand.games))

		game := gameCommand.games[message.User]
		assert.True(t, game.randomNumber >= 0)
		assert.True(t, game.randomNumber <= maxNumber)
		assert.Equal(t, 0, game.tries)

		// too low number
		message.Text = fmt.Sprintf("guess number %d", game.randomNumber-1)
		slackClient.On("SendMessage", message, "Higher :arrow_up_small:").Return("")
		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, game.tries)

		// too high number
		message.Text = fmt.Sprintf("guess number %d", game.randomNumber+1)
		slackClient.On("SendMessage", message, "Lower :arrow_down_small:").Return("")
		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 2, game.tries)

		// bingo! -> win message + remove game
		message.Text = fmt.Sprintf("guess number %d", game.randomNumber)
		slackClient.On("SendMessage", message, "Wow! you got it in 3 tries :beers:").Return("")
		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 0, len(gameCommand.games))
	})

	t.Run("Invalid states", func(t *testing.T) {
		assert.Equal(t, 0, len(gameCommand.games))

		// guess without running game -> error message
		message := msg.Message{}
		message.Text = "guess number 100"
		slackClient.On("SendMessage", message, "There is no game running. Use `start number guesser`").Return("")
		actual := commands.Run(message)
		assert.True(t, actual)

		// start the game
		message = msg.Message{}
		message.Text = "start number guesser"
		slackClient.On("SendMessage", message, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber)).Return("")
		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, len(gameCommand.games))

		game := gameCommand.games[message.User]
		assert.True(t, game.randomNumber >= 0)
		assert.True(t, game.randomNumber <= maxNumber)
		assert.Equal(t, 0, game.tries)

		// start the game again -> error
		message = msg.Message{}
		message.Text = "start number guesser"
		slackClient.On("SendMessage", message, "There is already a game :smile: use `guess number XX` instead").Return("")
		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, len(gameCommand.games))
	})
}
