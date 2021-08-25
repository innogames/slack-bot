package games

import (
	"fmt"
	"math/rand"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

const (
	maxNumber = 1000
	maxTries  = 10
)

type game struct {
	randomNumber int
	tries        int
}
type runningGames map[string]*game

type numberGuesserCommand struct {
	bot.BaseCommand
	games runningGames
}

// NewNumberGuesserCommand is a very small game to guess a random number
func NewNumberGuesserCommand(base bot.BaseCommand) bot.Command {
	return &numberGuesserCommand{base, runningGames{}}
}

func (c *numberGuesserCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("start number guesser", c.start),
		matcher.NewRegexpMatcher(`guess number (?P<number>\d+)`, c.guess),
	)
}

func (c *numberGuesserCommand) start(match matcher.Result, message msg.Message) {
	if _, ok := c.games[message.GetUser()]; ok {
		c.SendMessage(message, "There is already a game :smile: use `guess number XX` instead")
		return
	}

	randomNumber := rand.Intn(maxNumber) //nolint:gosec
	game := &game{
		randomNumber: randomNumber,
		tries:        0,
	}
	c.games[message.GetUser()] = game

	c.SendMessage(message, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber))
}

func (c *numberGuesserCommand) guess(match matcher.Result, message msg.Message) {
	currentGame, ok := c.games[message.GetUser()]
	if !ok {
		c.SendMessage(message, "There is no game running. Use `start number guesser`")
		return
	}

	guess := match.GetInt("number")
	currentGame.tries++

	if guess == currentGame.randomNumber {
		c.SendMessage(message, fmt.Sprintf("Wow! you got it in %d tries :beers:", currentGame.tries))
		delete(c.games, message.GetUser())
		return
	}
	if currentGame.tries >= maxTries {
		c.SendMessage(message, "Too many tries already...game over!")
		delete(c.games, message.GetUser())
		return
	}

	if guess < currentGame.randomNumber {
		c.SendMessage(message, "Higher :arrow_up_small:")
	} else {
		c.SendMessage(message, "Lower :arrow_down_small:")
	}
}

func (c *numberGuesserCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "start number guesser",
			Description: "small game to guess a random number",
			Category:    category,
			Examples: []string{
				"start number guesser",
				"guess number 100",
			},
		},
	}
}
