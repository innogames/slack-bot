package games

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
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

		mocks.AssertSlackMessage(slackClient, message, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber))

		actual := commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, len(gameCommand.games))

		game := gameCommand.games[message.User]
		assert.True(t, game.randomNumber >= 0)
		assert.True(t, game.randomNumber <= maxNumber)
		assert.Equal(t, 0, game.tries)

		// too low number
		message.Text = fmt.Sprintf("guess number %d", game.randomNumber-1)

		mocks.AssertSlackMessage(slackClient, message, "Higher :arrow_up_small:")

		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, game.tries)

		// too high number
		message.Text = fmt.Sprintf("guess number %d", game.randomNumber+1)

		mocks.AssertSlackMessage(slackClient, message, "Lower :arrow_down_small:")

		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 2, game.tries)

		// bingo! -> win message + remove game
		message.Text = fmt.Sprintf("guess number %d", game.randomNumber)

		mocks.AssertSlackMessage(slackClient, message, "Wow! you got it in 3 tries :beers:")

		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 0, len(gameCommand.games))
	})

	t.Run("Invalid states", func(t *testing.T) {
		assert.Equal(t, 0, len(gameCommand.games))

		// guess without running game -> error message
		message := msg.Message{}
		message.Text = "guess number 100"

		mocks.AssertSlackMessage(slackClient, message, "There is no game running. Use `start number guesser`")

		actual := commands.Run(message)
		assert.True(t, actual)

		// start the game
		message = msg.Message{}
		message.Text = "start number guesser"

		mocks.AssertSlackMessage(slackClient, message, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber))

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

		mocks.AssertSlackMessage(slackClient, message, "There is already a game :smile: use `guess number XX` instead")

		actual = commands.Run(message)
		assert.True(t, actual)
		assert.Equal(t, 1, len(gameCommand.games))
	})
}
