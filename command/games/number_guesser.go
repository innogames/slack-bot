package games

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"math/rand"
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
	slackClient client.SlackClient
	games       runningGames
}

// NewNumberGuesserCommand is a very small game to guess a random number
func NewNumberGuesserCommand(slackClient client.SlackClient) bot.Command {
	return &numberGuesserCommand{slackClient, runningGames{}}
}

func (c *numberGuesserCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("start number guesser", c.Start),
		matcher.NewRegexpMatcher(`guess number (?P<number>\d+)`, c.Guess),
	)
}

func (c *numberGuesserCommand) Start(match matcher.Result, event slack.MessageEvent) {
	if _, ok := c.games[event.User]; ok {
		c.slackClient.Reply(event, "There is already a game :smile: use `guess number XX` instead")
		return
	}

	randomNumber := rand.Intn(maxNumber)
	game := &game{
		randomNumber: randomNumber,
		tries:        0,
	}
	c.games[event.User] = game

	c.slackClient.Reply(event, fmt.Sprintf("I chose a number between 0 an %d. Good luck! Use `guess number XX`", maxNumber))
}

func (c *numberGuesserCommand) Guess(match matcher.Result, event slack.MessageEvent) {
	currentGame, ok := c.games[event.User]
	if !ok {
		c.slackClient.Reply(event, "There is no game running. Use `start number guesser`")
		return
	}

	guess := match.GetInt("number")
	currentGame.tries++

	if guess == currentGame.randomNumber {
		c.slackClient.Reply(event, fmt.Sprintf("Wow! you got it in %d tries :beers:", currentGame.tries))
		delete(c.games, event.User)
		return
	}
	if currentGame.tries >= maxTries {
		c.slackClient.Reply(event, "Too many tries already...game over!")
		delete(c.games, event.User)
		return
	}

	if guess < currentGame.randomNumber {
		c.slackClient.Reply(event, "Higher :arrow_up_small:")
	} else {
		c.slackClient.Reply(event, "Lower :arrow_down_small:")
	}
}

func (c *numberGuesserCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "number guesser",
			Description: "small game to guess a random number",
			Examples: []string{
				"start number guesser",
				"guess number 100",
			},
		},
	}
}
